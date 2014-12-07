
# get couchbase cluster ip from etcd
COUCHBASE_CLUSTER=$(etcdctl get /services/couchbase/bootstrap_ip)

# get usrname/pass from etcd
CB_USERNAME_PASSWORD=$(etcdctl get /services/couchbase/userpass)
IFS=':' read -a array <<< "$CB_USERNAME_PASSWORD"
CB_USERNAME=${array[0]}
CB_PASSWORD=${array[1]}

# create cbfs bucket
sudo docker run tleyden5iwx/couchbase-server-3.0.1 /opt/couchbase/bin/couchbase-cli bucket-create -c $COUCHBASE_CLUSTER -u $CB_USERNAME -p $CB_PASSWORD --bucket=cbfs --bucket-ramsize=512

# kick off 3 cbfs nodes (TODO: num nodes should be a parameter)
wget https://raw.githubusercontent.com/tleyden/elastic-thought/master/docker/fleet/cbfs_node.service.template
for i in `seq 1 3`; do cp cbfs_node.service.template cbfs_node.$i.service; done
fleetctl start cbfs_node.*.service


