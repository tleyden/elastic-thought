
Scalable REST API wrapper for the [Caffe](caffe.berkeleyvision.org) deep learning framework. 

## Components

![ElasticThought Components](http://tleyden-misc.s3.amazonaws.com/blog_images/elasticthought-components.png)

* [CoreOS](https://coreos.com/) / [Docker](https://www.docker.com/) - OS / Container
* [Caffe](http://caffe.berkeleyvision.org/) - core deep learning framework
* [Couchbase Server](http://www.couchbase.com/nosql-databases/couchbase-server) - NoSQL Database used as system of record
* [Sync Gateway](https://github.com/couchbase/sync_gateway) - REST adapter layer for Couchbase Server + Mobile Sync gateway
* [CBFS](https://github.com/couchbaselabs/cbfs) - Couchbase Distributed File System used as blob store
* [NSQ](http://nsq.io/) - Distributed message queue
* [ElasticThought REST Service](https://github.com/tleyden/elastic-thought/) - REST API server written in Go

## Deployment Architecture

Here is what a typical cluster might look like:

![ElasticThought Deployment](http://tleyden-misc.s3.amazonaws.com/blog_images/elasticthought-stack.png) 

Although not shown, all components are running inside of docker containers.

## Roadmap

*Current Status: on step 1, everything under heavy construction, not ready for public consumption yet*

1. --> Support a single caffe use case: IMAGE_DATA caffe layer using a single test set with a single training set
1. Support the majority of caffe use cases
1. Package everything up to make it easy to deploy  <-- initial release
1. Ability to auto-scale worker instances up and down based on how many jobs are in the message queue.
1. Attempt to add support for other deep learning frameworks: pylearn2, cuda-convnet, etc.
1. Build a Web App on top of the REST API that leverages [PouchDB](https://github.com/pouchdb/pouchdb)
1. Build Android and iOS mobile apps on top of the REST API that leverage [Couchbase Mobile](https://github.com/couchbase/mobile)


## Documentation 

* [REST API](http://docs.elasticthought.apiary.io/)
* [Godocs](http://godoc.org/github.com/tleyden/elastic-thought)

## Quick Start

# Install go1.3 or later

```

```

# Start Nsq

```
$ nsqlookupd & 
$ nsqd --lookupd-tcp-address=127.0.0.1:4160 &
$ nsqadmin --lookupd-http-address=127.0.0.1:4161 &
```

# Start Sync Gatewway

```
$ ./run.sh config.json
```

# Start httpd

# Start cbfs

# Start Couchbase Server


## Todo

- Validate solver.prototxt: 
  - it should assert "net" arg is empty
  - it should add a value for "net", which should be absolute path to net prototxt file
  - the worker should rewrite the solver_mode CPU/GPU based on worker capabilities 

## License

Apache 2