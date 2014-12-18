
function untilsuccessful() {
	"$@"
	while [ $? -ne 0 ]; do
		echo Retrying...
		sleep 1
		"$@"
	done
}

usage="./elasticthought-cluster-init.sh -v 3.0.1 -n 3 -u "user:passw0rd" -p gpu"

while getopts ":v:n:u:p:" opt; do
      case $opt in
        v  ) version=$OPTARG ;;
        n  ) numnodes=$OPTARG ;;
        u  ) userpass=$OPTARG ;;
        p  ) processor=$OPTARG ;;
        \? ) echo $usage
             exit 1 ;; 
      esac
done

# make sure required args were given
if [[ -z "$version" || -z "$numnodes" || -z "$userpass" || -z "$processor" ]] ; then
    echo "Required argument was empty"
    echo $usage
    exit 1 
fi

# validate processor arg: cpu or gpu
if [ "$processor" != "cpu" ] && [ "$processor" != "gpu" ]; then
    echo "You passed an invalid value for processor.  Must be cpu or gpu"
    exit 1 
fi

if [ "$processor" == "gpu" ]; then
    NUM_NVIDIA=$(lspci | grep -i nvidia | wc -l)
    if (( $NUM_NVIDIA <= 0 )); then
	echo "No nvidia graphics cards found.  Did you use correct AMI?"
	exit 1 
    fi
fi 

# parse user/pass into variables
IFS=':' read -a array <<< "$userpass"
CB_USERNAME=${array[0]}
CB_PASSWORD=${array[1]}

# Kick off couchbase cluster 
echo "Kick off couchbase cluster"
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
while (( $NUM_COUCHBASE_SERVERS != $numnodes )); do
    echo "Retrying... $NUM_COUCHBASE_SERVERS != $numnodes"
    NUM_COUCHBASE_SERVERS=$(sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli server-list -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD | wc -l)
    sleep 5
done
echo "Done waiting: $numnodes Couchbase Servers are running"

fleetctl list-units

# rebalance cluster
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli rebalance -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD

# create cbfs bucket
TOTAL_MEM_MB=$(free -m | awk '/^Mem:/{print $2}')
CBFS_BUCKET_SIZE_MB=$(($TOTAL_MEM_MB * 10 / 100))
echo "Create a cbfs bucket of size: $CBFS_BUCKET_SIZE_MB"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=$CBFS_BUCKET_SIZE_MB
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

fleetctl list-units

# create elastic-thought bucket
ET_BUCKET_SIZE_MB=$(($TOTAL_MEM_MB * 20 / 100))
echo "Create elastic-thought bucket with size: $ET_BUCKET_SIZE_MB"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=elastic-thought --bucket-ramsize=$ET_BUCKET_SIZE_MB

# run sed on sync gateway config template
COUCHBASE_IP_PORT=$COUCHBASE_CLUSTER:8091
sed -e "s/COUCHBASE_IP_PORT/${COUCHBASE_IP_PORT}/" elastic-thought/docker/templates/sync_gateway/sync_gw_config.json > /tmp/sync_gw_config.json
echo "Generated sync gateway config"
cat /tmp/sync_gw_config.json

# upload sync gateway config to cbfs
ip=$(hostname -i | tr -d ' ')
echo "Upload sync gateway config to cbfs: http://$ip:8484/"
untilsuccessful sudo docker run --net=host -v /tmp:/tmp tleyden5iwx/cbfs cbfsclient http://$ip:8484/ upload /tmp/sync_gw_config.json /sync_gw_config.json 

# kick off sync gateway 
echo "Kick off sync gateway"
mkdir sync-gateway && \
  cd sync-gateway && \
  wget https://raw.githubusercontent.com/tleyden/sync-gateway-coreos/master/scripts/sync-gw-cluster-init.sh && \
  chmod +x sync-gw-cluster-init.sh && \
  ./sync-gw-cluster-init.sh -n $numnodes -c "master" -g http://$ip:8484/sync_gw_config.json -v 0 && \
  cd ~ 

# wait for all sync gw nodes to come up 
echo "Wait for sync gateway nodes to come up"
for i in `seq 1 $numnodes`; do 
    untilsuccessful etcdctl get /services/sync_gw/sync_gw_node@$i
done
echo "Done: sync gateway nodes up"

fleetctl list-units

# kick off elastic-thought httpd daemons
echo "Kick off elastic thought httpd daemons"
cd elastic-thought/docker/fleet && fleetctl submit elastic_thought_$processor@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start elastic_thought_$processor@$i.service; done

fleetctl list-units

echo "Done!  In a few minutes, your ElasticThought REST API will be available to use on <public-ip-any-node>:8080 -- check status with fleetctl list-units"
