
Scalable REST API wrapper for the [Caffe](caffe.berkeleyvision.org) deep learning framework. 

## The problem

After using Caffe for a while, I started finding it inconvenient to run things on my own laptop.  Often I'd kick off a job and then need to throw my laptop in my backpack and run out the door, and the job would stop running.  

I came to the realization that I needed to run Caffe in the cloud, and came up with the following requirements:

* Run multiple training jobs in parallel
* Queue up a lot of training jobs at once 
* Tune the number of workers that process jobs on the queue 
* Interact with it via a REST API (and later build Web/Mobile apps on top of it)
* Support teams: multi-tenancy to allow multiple users to interact with it, where each user only sees their own data

## Components

![ElasticThought Components](http://tleyden-misc.s3.amazonaws.com/blog_images/elasticthought-components.png)


* [Caffe](http://caffe.berkeleyvision.org/) - core deep learning framework
* [Couchbase Server](http://www.couchbase.com/nosql-databases/couchbase-server) - Distributed document database used as an object store ([source code](https://github.com/couchbase/manifest))
* [Sync Gateway](https://github.com/couchbase/sync_gateway) - REST adapter layer for Couchbase Server + Mobile Sync gateway
* [CBFS](https://github.com/couchbaselabs/cbfs) - Couchbase Distributed File System used as blob store
* [NSQ](http://nsq.io/) - Distributed message queue
* [ElasticThought REST Service](https://github.com/tleyden/elastic-thought/) - REST API server written in Go

## Deployment Architecture

Here is what a typical cluster might look like:

![ElasticThought Deployment](http://tleyden-misc.s3.amazonaws.com/blog_images/elasticthought-stack.png) 

If running on AWS, each [CoreOS](https://coreos.com/) instance would be running on its own EC2 instance.

Although not shown, all components would be running inside of [Docker](https://www.docker.com/) containers.

[CoreOS Fleet](https://coreos.com/docs/launching-containers/launching/launching-containers-fleet/) would be leveraged to auto-restart any failed components, including Caffe workers.

## Roadmap

*Current Status: on step 1, everything under heavy construction, not ready for public consumption yet*

1. --> Support a single caffe use case: IMAGE_DATA caffe layer using a single test set with a single training set
1. Support the majority of caffe use cases
1. Package everything up to make it easy to deploy  <-- initial release
1. Ability to auto-scale worker instances up and down based on how many jobs are in the message queue.
1. Attempt to add support for other deep learning frameworks: pylearn2, cuda-convnet, etc.
1. Build a Web App on top of the REST API that leverages [PouchDB](https://github.com/pouchdb/pouchdb)
1. Build Android and iOS mobile apps on top of the REST API that leverages [Couchbase Mobile](https://github.com/couchbase/couchbase-lite-android)


## Design goals

* 100% Open Source (Apache 2 / BSD), including all components used.
* Architected to enable *warehouse scale* computing
* No IAAS lockin -- easily migrate between AWS, GCE, or your own private data center
* Ability to scale *down* as well as up

## Documentation 

* [REST API](http://docs.elasticthought.apiary.io/)
* [Godocs](http://godoc.org/github.com/tleyden/elastic-thought)

## Grid Computing

ElasticThought is not trying to be a grid computing (aka distributed computation) solution.  

For that, check out:

* [ParameterServer](http://parameterserver.org/)
* [Caffe Issue 876](https://github.com/BVLC/caffe/issues/876)

## Quick Start

*Note: this will be much easier after everything is packaged as fleetctl scripts, for the meantime these are just notes to myself*

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
- Add ability to cancel a training job in progress

## Related Work



## License

Apache 2