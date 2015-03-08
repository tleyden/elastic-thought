
sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet stop --all-units

sudo docker run --net=host tleyden5iwx/couchbase-cluster-go update-wrapper couchbase-fleet destroy --all-units

DELETE_DATA="sudo rm -rf /opt/couchbase/var/* /var/lib/cbfs/*"

fleetctl list-machines | grep -v MACHINE | awk '{print $2}' | xargs -I{} ssh {} 'echo Delete /opt/couchbase/var and /var/lib/cbfs on `hostname` && $DELETE_DATA'


