
FROM tleyden5iwx/caffe-gpu-develop

MAINTAINER Traun Leyden tleyden@couchbase.com

ENV GOPATH /opt/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV GOROOT /usr/local/go

RUN apt-get update && \
    apt-get -q -y install mercurial && \
    apt-get -q -y install make && \
    apt-get -q -y install binutils && \
    apt-get -q -y install bison && \
    apt-get -q -y install build-essential

RUN mkdir -p $GOPATH

# Install Go 1.3 manually (since Go 1.3 is required, and ubuntu 14.04 still uses Go 1.2)
RUN curl -O https://storage.googleapis.com/golang/go1.3.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.3.1.linux-amd64.tar.gz

# Go get ElasticThought
RUN go get -u -v -t github.com/tleyden/elastic-thought && \
    go get -u -v -t github.com/tleyden/elastic-thought/cli/httpd && \
    go get -u -v -t github.com/tleyden/elastic-thought/cli/worker

# Copy binaries 
RUN cp /opt/go/bin/worker /usr/local/bin && \
    cp /opt/go/bin/httpd /usr/local/bin
