
fleetctl list-units | awk {'print $1'} | grep -v UNIT | xargs fleetctl destroy

fleetctl list-unit-files | awk {'print $1'} | grep -v UNIT | xargs fleetctl destroy

rm -rf couchbase-server-docker elastic-thought sync-gateway
