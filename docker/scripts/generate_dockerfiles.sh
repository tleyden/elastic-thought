#!/usr/bin/env bash

go run generate_dockerfiles/generate_dockerfiles.go cpu develop > ../cpu/develop/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go cpu master > ../cpu/master/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go gpu develop > ../gpu/develop/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go gpu master > ../gpu/master/Dockerfile


