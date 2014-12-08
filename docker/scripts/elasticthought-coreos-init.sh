
# TODO: this script needs to wait for the couchbase bootstrap node to be running
# before executing

# wait for couchbase server to come up
# TODO: come up with better way than this
echo "Sleeping to wait for Couchbase Server"
sleep 180


# get couchbase cluster ip from etcd
COUCHBASE_CLUSTER=$(etcdctl get /services/couchbase/bootstrap_ip)

if [ -z "$COUCHBASE_CLUSTER" ]; then
    echo "COUCHBASE_CLUSTER is empty"
    exit 1 
fi

# get usrname/pass from etcd
CB_USERNAME_PASSWORD=$(etcdctl get /services/couchbase/userpass)
IFS=':' read -a array <<< "$CB_USERNAME_PASSWORD"
CB_USERNAME=${array[0]}
CB_PASSWORD=${array[1]}

# create cbfs bucket
sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=512

# kick off 3 cbfs nodes (TODO: num nodes should be a parameter)
git clone https://github.com/tleyden/elastic-thought.git
cp elastic-thought/docker/fleet/cbfs_node.service.template .
for i in `seq 1 3`; do cp cbfs_node.service.template cbfs_node.$i.service; done
fleetctl start cbfs_node.*.service

# wait for cbfs nodes to come up
# TODO: come up with better way than this
echo "Sleeping to wait for CBFS"
sleep 180 

# create elastic-thought bucket
sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=elastic-thought --bucket-ramsize=1024

# run sed on sync gateway config template
COUCHBASE_IP_PORT=$COUCHBASE_CLUSTER:8091
sed -e "s/COUCHBASE_IP_PORT/${COUCHBASE_IP_PORT}/" elastic-thought/docker/templates/sync_gateway/sync_gw_config.json > /tmp/sync_gw_config.json

# upload to cbfs
ip=$(hostname -i | tr -d ' ')
sudo docker run --net=host -v /tmp:/tmp tleyden5iwx/cbfs cbfsclient http://$ip:8484/ upload /tmp/sync_gw_config.json /sync_gw_config.json 

# kick off sync gateway 
mkdir sync-gateway && \
  cd sync-gateway && \
  wget https://raw.githubusercontent.com/tleyden/sync-gateway-coreos/master/scripts/cluster-init.sh && \
  chmod +x cluster-init.sh && \
  ./cluster-init.sh -n 3 -c "master" -g http://$ip:8484/sync_gw_config.json

# wait for sync gw and nsq to come up 
# TODO: come up with better way than this
echo "Sleeping to wait for Sync Gateway"
sleep 180 

# kick off elasticthought httpd-worker (goroutine)
echo "Starting elastic thought httpd (blocking call)"
sudo docker run --net=host tleyden5iwx/elastic-thought httpd 

# TODO: 
# kick off nsqlookupd + sidekick
# kick off nsq (3 nodes)
# kick off elasticthought worker (caffe worker)
