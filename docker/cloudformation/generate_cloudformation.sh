#!/usr/bin/env bash

go run generate_cloudformation.go elastic_thought cpu > elastic_thought_cpu.template
go run generate_cloudformation.go elastic_thought gpu > elastic_thought_gpu.template
go run generate_cloudformation.go sync_gateway cpu > sync_gateway.template
go run generate_cloudformation.go cbfs cpu > cbfs.template
go run generate_cloudformation.go couchbase_server cpu > couchbase_server.template
