#!/usr/bin/env bash

go run generate_dockerfiles/generate_dockerfiles.go cpu develop > ../cpu/develop/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go cpu master > ../cpu/master/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go gpu develop > ../gpu/develop/Dockerfile
go run generate_dockerfiles/generate_dockerfiles.go gpu master > ../gpu/master/Dockerfile

# README
cp ../../README.md ../cpu/develop
cp ../../README.md ../cpu/master 
cp ../../README.md ../gpu/develop
cp ../../README.md ../gpu/master 

# Scripts
cp refresh-elastic-thought* ../cpu/develop/scripts && chmod +x ../cpu/develop/scripts/*
cp refresh-elastic-thought* ../cpu/master/scripts && chmod +x ../cpu/master/scripts/*
cp refresh-elastic-thought* ../gpu/develop/scripts && chmod +x ../gpu/develop/scripts/*
cp refresh-elastic-thought* ../gpu/master/scripts && chmod +x ../gpu/master/scripts/*
