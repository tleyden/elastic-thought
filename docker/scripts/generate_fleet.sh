#!/usr/bin/env bash

go run generate_fleet/generate_fleet.go cpu > ../fleet/elastic_thought_cpu@.service
go run generate_fleet/generate_fleet.go gpu > ../fleet/elastic_thought_gpu@.service

