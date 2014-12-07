#!/usr/bin/env bash

go run generate_cloudformation.go elastic_thought > elastic_thought.template
go run generate_cloudformation.go sync_gateway > sync_gateway.template
go run generate_cloudformation.go cbfs > cbfs.template
go run generate_cloudformation.go couchbase_server > couchbase_server.template
