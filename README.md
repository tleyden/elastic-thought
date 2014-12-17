
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

## Kick things off: Aws

### Launch EC2 instances via CloudFormation script

*Note: the instance will launch in **us-east-1***.  If you want to launch in another region, please [file an issue](https://github.com/tleyden/elastic-thought/issues).

* [Launch CPU Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#cstack=sn%7ECouchbase-CoreOS%7Cturl%7Ehttp://tleyden-misc.s3.amazonaws.com/elastic-thought/cloudformation/elastic_thought_cpu.template) or [Launch GPU Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#cstack=sn%7ECouchbase-CoreOS%7Cturl%7Ehttp://tleyden-misc.s3.amazonaws.com/elastic-thought/cloudformation/elastic_thought_gpu.template) 
* Choose 3 node cluster with m3.medium or g2.2xlarge (GPU case) instance type
* All other values should be default

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
cbfs_announce@1.service				2340c553.../10.225.17.229	active	running
cbfs_announce@2.service				fbd4562e.../10.182.197.145	active	running
cbfs_announce@3.service				0f5e2e11.../10.168.212.210	active	running
cbfs_node@1.service				2340c553.../10.225.17.229	active	running
cbfs_node@2.service				fbd4562e.../10.182.197.145	active	running
cbfs_node@3.service				0f5e2e11.../10.168.212.210	active	running
couchbase_bootstrap_node.service		0f5e2e11.../10.168.212.210	active	running
couchbase_bootstrap_node_announce.service	0f5e2e11.../10.168.212.210	active	running
couchbase_node.1.service			2340c553.../10.225.17.229	active	running
couchbase_node.2.service			fbd4562e.../10.182.197.145	active	running
elastic_thought_gpu@1.service			2340c553.../10.225.17.229	active	running
elastic_thought_gpu@2.service			fbd4562e.../10.182.197.145	active	running
elastic_thought_gpu@3.service			0f5e2e11.../10.168.212.210	active	running
sync_gw_announce@1.service			2340c553.../10.225.17.229	active	running
sync_gw_announce@2.service			fbd4562e.../10.182.197.145	active	running
sync_gw_announce@3.service			0f5e2e11.../10.168.212.210	active	running
sync_gw_node@1.service				2340c553.../10.225.17.229	active	running
sync_gw_node@2.service				fbd4562e.../10.182.197.145	active	running
sync_gw_node@3.service				0f5e2e11.../10.168.212.210	active	running
```

At this point you should be able to access the [REST API](http://docs.elasticthought.apiary.io/) on the public ip any of the three Sync Gateway machines.

## Kick things off: Vagrant

### Update Vagrant

Make sure you're running a current version of Vagrant, otherwise the plugin install below may [fail](https://github.com/mitchellh/vagrant/issues/3769).

```
$ vagrant -v
1.7.1
```

### Install CoreOS

See https://coreos.com/docs/running-coreos/platforms/vagrant/

### Hack Vagrant hostname to use an IP

This is a dirty hack to get around the following Couchbase Server problem:

* The fleet script uses %H to get the hostname and passes that into the couchbase-start script
* Couchbase expects a resolvable, fully qualified domain name for the hostname.
* If you give it a host like "core-01", it will return an error "Requested address "core-01" is not allowed: short names are not allowed. Couchbase Server requires at least one dot in a name"
* My first attempt was to create entries in /etc/hosts on CoreOS, but this didn't work because the Docker container that is running Couchbase Server won't see these mappings.

As a hack, I ended up setting the hostname to the ip, which so far has not caused any issues.

The alternative was to modify the fleet script to use $COREOS_PUBLIC_IPV4 as described [here](https://groups.google.com/d/msg/coreos-user/Dkl0Rj3Rg6w/urcUFThQKjYJ), or use some bash commands to discover the public ip.  However that had the drawback of changing what already works fine on AWS.

Make the following change in your Vagrant file:

**Before**

```
config.vm.hostname = ip
...
...
ip = "172.17.8.#{i+100}"
```

**After**

```
ip = "172.17.8.#{i+100}"
config.vm.hostname = ip
```

### Install vagrant-hostmanager

This is required on Vagrant to map hosts to ip addresses.

From the core-os directory created in the step above, run:

```
$ vagrant plugin install vagrant-hostmanager
$ vagrant hostmanager
```

**Note: this isn't working for me, as [reported here](https://github.com/smdahlen/vagrant-hostmanager/issues/127)**

### Manually update /etc/hosts

If you can get the vagrant-hostmanager working, you can skip this part.  

For **each coreos machine**:

```
$ vagrant ssh core-01 -- -A 
$ sudo vi /etc/hosts
```

when vi is up, paste in:

```
172.17.8.101 core-01.yourhost.com
172.17.8.102 core-02.yourhost.com
172.17.8.103 core-03.yourhost.com
```

But change the ip's to match your cluster.


## License

Apache 2