
# TODO: this script needs to wait for the couchbase bootstrap node to be running
# before executing

function untilsuccessful() {
	"$@"
	while [ $? -ne 0 ]; do
		echo Retrying...
		sleep 1
		"$@"
	done
}

while getopts ":v:n:u:" opt; do
      case $opt in
        v  ) version=$OPTARG ;;
        n  ) numnodes=$OPTARG ;;
        u  ) userpass=$OPTARG ;;
        \? ) echo $usage
             exit 1 ;; 
      esac
done

# parse user/pass into variables
IFS=':' read -a array <<< "$userpass"
CB_USERNAME=${array[0]}
CB_PASSWORD=${array[1]}

# Kick off couchbase cluster 
wget https://raw.githubusercontent.com/couchbaselabs/couchbase-server-docker/master/scripts/cluster-init.sh
chmod +x cluster-init.sh
./cluster-init.sh -v $version -n $numnodes -u $userpass

if [ $? -ne 0 ]; then
    echo "Error executing cluster-init.sh"
    exit 1 
fi

# Wait until bootstrap node is up
echo "Wait until Couchbase bootstrap node is up"
while [ -z "$COUCHBASE_CLUSTER" ]; do
    echo Retrying...
    COUCHBASE_CLUSTER=$(etcdctl get /services/couchbase/bootstrap_ip)
    sleep 5
done

echo "Couchbase Server bootstrap ip: $COUCHBASE_CLUSTER"

# wait until all couchbase nodes come up
echo "Wait until $numnodes Couchbase Servers running"
NUM_COUCHBASE_SERVERS="0"
while [ "$NUM_COUCHBASE_SERVERS" -ne $numnodes ]; do
    echo Retrying...
    NUM_COUCHBASE_SERVERS=$(sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli server-list -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD | wc -l)
    sleep 5
done
echo "Done waiting: $numnodes Couchbase Servers are running"

# rebalance cluster
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli rebalance -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD

# create cbfs bucket
echo "Create a cbfs bucket"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=512
echo "Done: created a cbfs bucket"

# kick off 3 cbfs nodes (TODO: num nodes should be a parameter)
echo "Kick off cbfs nodes"
git clone https://github.com/tleyden/elastic-thought.git
cd elastic-thought/docker/fleet && fleetctl submit cbfs_node@.service && fleetctl submit cbfs_announce@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start cbfs_node@$i.service; fleetctl start cbfs_announce@$i.service; done

# wait for all cbfs nodes to come up 
echo "Wait for cbfs nodes to come up"
for i in `seq 1 $numnodes`; do 
    untilsuccessful etcdctl get /services/cbfs/cbfs_node@$i
done
echo "Done: cbfs nodes up"

# create elastic-thought bucket
echo "Create elastic-thought bucket"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=elastic-thought --bucket-ramsize=1024

# run sed on sync gateway config template
COUCHBASE_IP_PORT=$COUCHBASE_CLUSTER:8091
sed -e "s/COUCHBASE_IP_PORT/${COUCHBASE_IP_PORT}/" elastic-thought/docker/templates/sync_gateway/sync_gw_config.json > /tmp/sync_gw_config.json

# upload sync gateway config to cbfs
echo "Upload sync gateway config to cbfs"
ip=$(hostname -i | tr -d ' ')
sudo docker run --net=host -v /tmp:/tmp tleyden5iwx/cbfs cbfsclient http://$ip:8484/ upload /tmp/sync_gw_config.json /sync_gw_config.json 

# kick off sync gateway 
echo "Kick off sync gateway"
mkdir sync-gateway && \
  cd sync-gateway && \
  wget https://raw.githubusercontent.com/tleyden/sync-gateway-coreos/master/scripts/cluster-init.sh && \
  chmod +x cluster-init.sh && \
  ./cluster-init.sh -n $numnodes -c "master" -g http://$ip:8484/sync_gw_config.json && \
  cd ~ 

# wait for all sync gw nodes to come up 
echo "Wait for sync gateway nodes to come up"
for i in `seq 1 $numnodes`; do 
    untilsuccessful etcdctl get /services/sync_gw/sync_gw_node@$i
done
echo "Done: sync gateway nodes up"

# kick off elastic-thought httpd daemons
echo "Kick off elastic thought httpd daemons"
cd elastic-thought/docker/fleet && fleetctl submit elastic_thought_gpu@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start elastic_thought_gpu@$i.service; done

echo "Done!  Your ElasticThought REST API is available to use on <public-ip-any-node>:8080"
