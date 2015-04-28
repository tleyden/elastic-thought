[![Build Status](https://drone.io/github.com/tleyden/elastic-thought/status.png)](https://drone.io/github.com/tleyden/elastic-thought/latest) [![GoDoc](https://godoc.org/github.com/tleyden/elastic-thought?status.png)](https://godoc.org/github.com/tleyden/elastic-thought) [![Coverage Status](https://coveralls.io/repos/tleyden/elastic-thought/badge.svg?branch=master)](https://coveralls.io/r/tleyden/elastic-thought?branch=master) [![Join the chat at https://gitter.im/tleyden/elastic-thought](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tleyden/elastic-thought?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Scalable REST API wrapper for the [Caffe](http://caffe.berkeleyvision.org) deep learning framework. 

## The problem

Caffe is an awesome deep learning framework, but running it on a single laptop or desktop computer isn't nearly as productive as running it in the cloud at scale.

ElasticThought gives you the ability to:

* Run multiple Caffe training jobs in parallel
* Queue up training jobs
* Tune the number of workers that process jobs on the queue 
* Interact with it via a REST API (and later build Web/Mobile apps on top of it)
* Multi-tenancy to allow multiple users to interact with it, each having access to only their own data

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

It would be possible to start more nodes which only had Caffe GPU workers running.

## Roadmap

*Current Status: everything under heavy construction, not ready for public consumption yet*

1. **[done]** Working end-to-end with IMAGE_DATA caffe layer using a single test set with a single training set, and ability to query trained set.
1. **[done]** Support LEVELDB / LMDB data formats, to run mnist example.
1. **[in progress]** Support the majority of caffe use cases
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
* This README

## System Requirements

ElasticThought requires CoreOS to run.

If you want to access the GPU, you will need to do extra work to get [CoreOS working with Nvidia CUDA GPU Drivers](http://tleyden.github.io/blog/2014/11/04/coreos-with-nvidia-cuda-gpu-drivers/)


## Installing elastic-thought on AWS (Production mode)

It should be possible to install elastic-thought anywhere that CoreOS is supported.  Currently, there are instructions for AWS and Vagrant (below).

### Launch EC2 instances via CloudFormation script

*Note: the instance will launch in **us-east-1***.  If you want to launch in another region, please [file an issue](https://github.com/tleyden/elastic-thought/issues).

* [Launch CPU Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#cstack=sn%7ECouchbase-CoreOS%7Cturl%7Ehttp://tleyden-misc.s3.amazonaws.com/elastic-thought/cloudformation/elastic_thought_cpu.template) or [Launch GPU Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#cstack=sn%7ECouchbase-CoreOS%7Cturl%7Ehttp://tleyden-misc.s3.amazonaws.com/elastic-thought/cloudformation/elastic_thought_gpu.template) 
* Choose 3 node cluster with m3.medium or g2.2xlarge (GPU case) instance type
* All other values should be default

### Verify CoreOS cluster

Run:

```
$ fleetctl list-machines
```

Which should show all the CoreOS machines in your cluster.  (this uses etcd under the hood, so will also validate that etcd is setup correctly).

### Kick off ElasticThought

Ssh into one of the machines (doesn't matter which): `ssh -A core@ec2-54-160-96-153.compute-1.amazonaws.com`

```
$ wget https://raw.githubusercontent.com/tleyden/elastic-thought/master/docker/scripts/elasticthought-cluster-init.sh
$ chmod +x elasticthought-cluster-init.sh
$ ./elasticthought-cluster-init.sh -v 3.0.1 -n 3 -u "user:passw0rd" -p gpu 
```

Once it launches, verify your cluster by running `fleetctl list-units`.  

It should look like this:

```
UNIT						MACHINE				ACTIVE	SUB
cbfs_announce@1.service                         2340c553.../10.225.17.229       active	running
cbfs_announce@2.service                         fbd4562e.../10.182.197.145      active	running
cbfs_announce@3.service                         0f5e2e11.../10.168.212.210      active	running
cbfs_node@1.service                             2340c553.../10.225.17.229       active	running
cbfs_node@2.service                             fbd4562e.../10.182.197.145      active	running
cbfs_node@3.service                             0f5e2e11.../10.168.212.210      active	running
couchbase_bootstrap_node.service                0f5e2e11.../10.168.212.210      active	running
couchbase_bootstrap_node_announce.service       0f5e2e11.../10.168.212.210      active	running
couchbase_node.1.service                        2340c553.../10.225.17.229       active	running
couchbase_node.2.service                        fbd4562e.../10.182.197.145      active	running
elastic_thought_gpu@1.service                   2340c553.../10.225.17.229       active	running
elastic_thought_gpu@2.service                   fbd4562e.../10.182.197.145      active	running
elastic_thought_gpu@3.service                   0f5e2e11.../10.168.212.210      active	running
sync_gw_announce@1.service                      2340c553.../10.225.17.229       active	running
sync_gw_announce@2.service                      fbd4562e.../10.182.197.145      active	running
sync_gw_announce@3.service                      0f5e2e11.../10.168.212.210      active	running
sync_gw_node@1.service                          2340c553.../10.225.17.229       active	running
sync_gw_node@2.service                          fbd4562e.../10.182.197.145      active	running
sync_gw_node@3.service                          0f5e2e11.../10.168.212.210      active	running
```

At this point you should be able to access the [REST API](http://docs.elasticthought.apiary.io/) on the public ip any of the three Sync Gateway machines.

## Installing elastic-thought on a single CoreOS host (Development mode)

If you are on OSX, you'll first need to install Vagrant, VirtualBox, and CoreOS.  See [CoreOS on Vagrant](https://coreos.com/docs/running-coreos/platforms/vagrant/) for instructions.  

Here's what will be created:

                                                                          
                                                                          
               ┌─────────────────────────────────────────────────────────┐
               │                       CoreOS Host                       │
               │  ┌──────────────────────────┐  ┌─────────────────────┐  │
               │  │     Docker Container     │  │  Docker Container   │  │
               │  │   ┌───────────────────┐  │  │    ┌────────────┐   │  │
               │  │   │  Elastic Thought  │  │  │    │Sync Gateway│   │  │
               │  │   │      Server       │  │  │    │  Database  │   │  │
               │  │   │   ┌───────────┐   │  │  │    │            │   │  │
               │  │   │   │In-process │   │◀─┼──┼───▶│            │   │  │
               │  │   │   │   Caffe   │   │  │  │    │            │   │  │
               │  │   │   │  worker   │   │  │  │    │            │   │  │
               │  │   │   └───────────┘   │  │  │    └────────────┘   │  │
               │  │   └───────────────────┘  │  └─────────────────────┘  │
               │  └──────────────────────────┘                           │
               └─────────────────────────────────────────────────────────┘
	       

```
$ vagrant ssh core-01
$ docker run --name sync-gateway -P couchbase/sync-gateway sync-gw-start -c feature/forestdb_bucket -g http://git.io/vfF4b
$ docker run --name elastic-thought -P --link sync-gateway:sync-gateway tleyden5iwx/elastic-thought-cpu-develop bash -c 'refresh-elastic-thought; elastic-thought --sync-gw http://sync-gateway:4984'
```
	
## Installing elastic-thought on Vagrant

### Update Vagrant

Make sure you're running a current version of Vagrant, otherwise the plugin install below may [fail](https://github.com/mitchellh/vagrant/issues/3769).

```
$ vagrant -v
1.7.1
```

### Install CoreOS on Vagrant

Clone the coreos/vagrant fork that has been customized for running ElasticThought.

```
$ cd ~/Vagrant 
$ git clone git@github.com:tleyden/coreos-vagrant.git
$ cd coreos-vagrant
$ cp config.rb.sample config.rb
$ cp user-data.sample user-data
```

By default this will run a **two node** cluster, if you want to change this, update the `$num_instances` variable in the `config.rb` file.

### Run CoreOS

```
$ vagrant up
```

Ssh in:

```
$ vagrant ssh core-01 -- -A
```

If you see:

```
Failed Units: 1
  user-cloudinit@var-lib-coreos\x2dvagrant-vagrantfile\x2duser\x2ddata.service
```

Jump to **Workaround CoreOS + Vagrant issues** below.

Verify things started up correctly:

```
core@core-01 ~ $ fleectctl list-machines
```

If you get errors like:

```
2015/03/26 16:58:50 INFO client.go:291: Failed getting response from http://127.0.0.1:4001/: dial tcp 127.0.0.1:4001: connection refused
2015/03/26 16:58:50 ERROR client.go:213: Unable to get result for {Get /_coreos.com/fleet/machines}, retrying in 100ms
```

Jump to **Workaround CoreOS + Vagrant issues** below.

### Workaround CoreOS + Vagrant issues:

First exit out of CoreOS:

```
core@core-01 ~ $ exit
```

On your OSX workstation, try the following workaround:

```
$ sed -i '' 's/420/0644/' user-data
$ sed -i '' 's/484/0744/' user-data
$ vagrant reload --provision
```

Ssh back in:

```
$ vagrant ssh core-01 -- -A
```

Verify it worked:

```
core@core-01 ~ $ fleectctl list-machines
```

You should see:

```
MACHINE		IP		METADATA
ce0fec18...	172.17.8.102	-
d6402b24...	172.17.8.101	-
```

I filed [CoreOS cloudinit issue 328](https://github.com/coreos/coreos-cloudinit/issues/328) to figure out why this error is happening (possibly related issues: [CoreOS cloudinit issue 261](https://github.com/coreos/coreos-cloudinit/issues/261) or [CoreOS cloudinit issue 190](https://github.com/coreos/bugs/issues/190))


### Continue steps above 

Scroll up to the **Installing elastic-thought on AWS** section and start with **Verify CoreOS cluster**

## FAQ

* Is this useful for grid computing / distributed computation?  **Ans**:  No, this is not trying to be a grid computing (aka distributed computation) solution.  You may want to check out [Caffe Issue 876](https://github.com/BVLC/caffe/issues/876) or [ParameterServer](http://parameterserver.org/)
  
## License

Apache 2
