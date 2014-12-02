#!/usr/bin/env bash

# Pin to a particular version of caffe.proto, so it doesn't change unexpectedly
wget https://raw.githubusercontent.com/BVLC/caffe/7c3c089a59c9f0301ec57a2b1317c588e2be5a8e/src/caffe/proto/caffe.proto

# This requires some dependencies to work, see http://tleyden.github.io/blog/2014/12/02/getting-started-with-go-and-protocol-buffers/
protoc --go_out=. *.proto

# Copy source to appropriate place in source tree
mv caffe.pb.go ../../caffe/

# Clean up
rm *.proto