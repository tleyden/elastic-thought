
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
while [ -z "$COUCHBASE_CLUSTER" ]; do
    echo Retrying...
    COUCHBASE_CLUSTER=$(etcdctl get /services/couchbase/bootstrap_ip)
    sleep 5
done

echo "Couchbase Server bootstrap ip: $COUCHBASE_CLUSTER"

# rebalance cluster
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli rebalance -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD

# create cbfs bucket
sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=512

# kick off 3 cbfs nodes (TODO: num nodes should be a parameter)
git clone https://github.com/tleyden/elastic-thought.git
cd elastic-thought/docker/fleet && fleetctl submit cbfs_node@.service && fleetctl submit cbfs_announce@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start cbfs_node@$i.service; fleetctl start cbfs_announce@$i.service; done

# wait for all 3 cbfs nodes to come up 
while [ -z "$CBFS_UP" ]; do
    COUNTER=0
    for i in `seq 1 $numnodes`; do 
	NODE_UP=$(etcdctl get /services/cbfs/cbfs_node@$i)
	if [ -n $NODE_UP ]; then
	    COUNTER=$[$COUNTER +1]
	fi
    done
    if (( $COUNTER == 4 )); then
	CBFS_UP="true"
    else
	echo "Sleeping .. will retry"
	sleep 5
    fi 
done


# create elastic-thought bucket
sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=elastic-thought --bucket-ramsize=1024

# run sed on sync gateway config template
COUCHBASE_IP_PORT=$COUCHBASE_CLUSTER:8091
sed -e "s/COUCHBASE_IP_PORT/${COUCHBASE_IP_PORT}/" elastic-thought/docker/templates/sync_gateway/sync_gw_config.json > /tmp/sync_gw_config.json

# upload sync gateway config to cbfs
ip=$(hostname -i | tr -d ' ')
sudo docker run --net=host -v /tmp:/tmp tleyden5iwx/cbfs cbfsclient http://$ip:8484/ upload /tmp/sync_gw_config.json /sync_gw_config.json 

# kick off sync gateway 
mkdir sync-gateway && \
  cd sync-gateway && \
  wget https://raw.githubusercontent.com/tleyden/sync-gateway-coreos/master/scripts/cluster-init.sh && \
  chmod +x cluster-init.sh && \
  ./cluster-init.sh -n $numnodes -c "master" -g http://$ip:8484/sync_gw_config.json

# wait for all 3 sync gw nodes to come up 
while [ -z "$SYNC_GW_UP" ]; do
    COUNTER=0
    for i in `seq 1 $numnodes`; do 
	NODE_UP=$(etcdctl get /services/sync_gw/sync_gw_node@$i)
	if [ -n $NODE_UP ]; then
	    COUNTER=$[$COUNTER +1]
	fi
    done
    if (( $COUNTER == 4 )); then
	SYNC_GW_UP="true"
    else
	echo "Sleeping .. will retry"
	sleep 5
    fi 
done



# Todo: use coreos init for this
# kick off elasticthought httpd-worker (goroutine)
echo "Starting elastic thought httpd (blocking call)"
sudo docker run --net=host tleyden5iwx/elastic-thought httpd 

# TODO: 
# kick off nsqlookupd + sidekick
# kick off nsq (3 nodes)
# kick off elasticthought worker (caffe worker)
