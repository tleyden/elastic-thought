
sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet stop --all-units

sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet destroy --all-units

fleetctl list-machines | grep -v MACHINE | awk '{print $2}' | xargs -I{} ssh {} 'sudo rm -rf /opt/couchbase/var/ && sudo rm -rf /var/lib/cbfs/'

etcdctl rmdir /couchbase.com/couchbase-node-state
