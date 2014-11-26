FROM tleyden5iwx/caffe

MAINTAINER Traun Leyden tleyden@couchbase.com

ENV GOPATH /opt/go
ENV PATH $GOPATH/bin:$PATH

RUN apt-get update && \
    apt-get -q -y install mercurial && \
    apt-get -q -y install make && \
    apt-get -q -y install binutils && \
    apt-get -q -y install bison && \
    apt-get -q -y install build-essential

RUN mkdir -p $GOPATH

# Since Go 1.3 is required, and ubuntu 14.04 still uses Go 1.2

# Install Go 1.3 manually
RUN curl -O https://storage.googleapis.com/golang/go1.3.1.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.3.1.linux-amd64.tar.gz

# Install Go 1.3 via GVM -- not working will
RUN bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
RUN source /root/.gvm/scripts/gvm
RUN gvm install go1.3.3 && \
    gvm use go1.3.3

# Go get ElasticThought
RUN go get -u -v -t github.com/tleyden/elastic-thought



