
# this gives us access to $COREOS_PRIVATE_IPV4 etc
source /etc/environment

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

# make sure we're on the bleeding edge and have latest code
etcdctl set /couchbase.com/enable-code-refresh true

# Kick off couchbase cluster 
echo "Kick off couchbase cluster"
sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet launch-cbs --version 3.0.1 --num-nodes $numnodes --userpass "$CB_USERNAME:$CB_PASSWORD"

if [ $? != 0 ]; then 
    echo "Failed to kick off couchbase server";
    exit 1
fi

# get an ip address of a running node in the cluster
COUCHBASE_CLUSTER=$(sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-cluster get-live-node-ip)
echo "Couchbase cluster node: $COUCHBASE_CLUSTER"

if [[ -z "$COUCHBASE_CLUSTER" ]] ; then
    echo "Failed to find ip of couchbase node"
    exit 1 
fi


# list units
fleetctl list-units

# create cbfs bucket
TOTAL_MEM_MB=$(free -m | awk '/^Mem:/{print $2}')
CBFS_BUCKET_SIZE_MB=$(($TOTAL_MEM_MB * 10 / 100))
echo "Create a cbfs bucket of size: $CBFS_BUCKET_SIZE_MB"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=$CBFS_BUCKET_SIZE_MB
echo "Done: created a cbfs bucket"

# kick off cbfs nodes
echo "Kick off cbfs nodes"
git clone https://github.com/tleyden/elastic-thought.git
cd elastic-thought/docker/fleet && fleetctl submit cbfs_node@.service && fleetctl submit cbfs_announce@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start cbfs_node@$i.service; fleetctl start cbfs_announce@$i.service; done


# wait for all cbfs nodes to come up 
echo "Wait for cbfs nodes to come up"
for i in `seq 1 $numnodes`; do 
    untilsuccessful etcdctl get /couchbase.com/cbfs/cbfs_node@$i
done
echo "Done: cbfs nodes up"

# list units
fleetctl list-units

# create elastic-thought bucket
ET_BUCKET_SIZE_MB=$(($TOTAL_MEM_MB * 20 / 100))
echo "Create elastic-thought bucket with size: $ET_BUCKET_SIZE_MB"
untilsuccessful sudo docker run tleyden5iwx/couchbase-server-$version /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=elastic-thought --bucket-ramsize=$ET_BUCKET_SIZE_MB

# run sed on sync gateway config template
# TODO: replace the sed stuff with sg-config-rewrite docker container invocation
COUCHBASE_IP_PORT=$COUCHBASE_CLUSTER:8091
sed -e "s/COUCHBASE_IP_PORT/${COUCHBASE_IP_PORT}/" elastic-thought/docker/templates/sync_gateway/sync_gw_config.json > /tmp/sync_gw_config.json
echo "Generated sync gateway config"
cat /tmp/sync_gw_config.json

# upload sync gateway config to cbfs
echo "Upload sync gateway config to cbfs: http://$COREOS_PRIVATE_IPV4:8484/"
untilsuccessful sudo docker run --net=host -v /tmp:/tmp tleyden5iwx/cbfs cbfsclient http://$COREOS_PRIVATE_IPV4:8484/ upload /tmp/sync_gw_config.json /sync_gw_config.json 

# kick off sync gateway 
echo "Kick off sync gateway"
sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper sync-gw-cluster launch-sgw --num-nodes=$numnodes --config-url=http://$COREOS_PRIVATE_IPV4:8484/sync_gw_config.json 

if [ $? != 0 ]; then 
    echo "Failed to kick off sync gateway";
    exit 1
fi

# wait for all sync gw nodes to come up 
# TODO: need sync gw sidekicks which publish to /couchbase.com/sgw-node-state/x..
# TODO: but only AFTER its detected to be running
echo "TODO: Wait for sync gateway nodes to come up"
echo "Done: sync gateway nodes up"

# list units
fleetctl list-units

# run elastic-thought environment sanity check
echo "Kick off elastic thought environment check"
untilsuccessful sudo docker run --net=host tleyden5iwx/elastic-thought-$processor-develop bash -c "refresh-elastic-thought; envcheck ${numnodes}"

# kick off elastic-thought httpd daemons
echo "Kick off elastic thought httpd daemons"
cd elastic-thought/docker/fleet && fleetctl submit elastic_thought_$processor@.service && cd ~
for i in `seq 1 $numnodes`; do fleetctl start elastic_thought_$processor@$i.service; done

fleetctl list-units

echo "Done!  In a few minutes, your ElasticThought REST API will be available to use on <public-ip-any-node>:8080 -- check status with fleetctl list-units"
