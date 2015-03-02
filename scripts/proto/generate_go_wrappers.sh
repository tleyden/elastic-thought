#!/usr/bin/env bash

# get dependencies
apt-get install protobuf-compiler
go get -u -v github.com/golang/protobuf/proto
go get -u -v github.com/golang/protobuf/protoc-gen-go

# Latest master branch version
wget https://raw.githubusercontent.com/BVLC/caffe/master/src/caffe/proto/caffe.proto

# This requires some dependencies to work, see http://tleyden.github.io/blog/2014/12/02/getting-started-with-go-and-protocol-buffers/
protoc --go_out=. *.proto

# Copy source to appropriate place in source tree
mv caffe.pb.go ../../caffe/

# Clean up
rm *.proto
