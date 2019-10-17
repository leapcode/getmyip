FROM debian:stretch AS build
RUN apt-get -q update && env DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential pkg-config golang-go git ca-certificates \
&& rm -rf /var/lib/apt/lists/*

# don't need to do bash tricks to keep the layers small, as this is a multi-stage build
ENV GOPATH=/go
WORKDIR $GOPATH
RUN go get 0xacab.org/leap/getmyip
RUN strip $GOPATH/bin/getmyip

FROM registry.git.autistici.org/ai3/docker/chaperone-base
RUN echo 'deb http://deb.debian.org/debian stretch contrib' > /etc/apt/sources.list.d/contrib.list
RUN apt-get -q update && env DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    geoipupdate ca-certificates \
&& rm -rf /var/lib/apt/lists/*
RUN cp /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
RUN /usr/bin/geoipupdate
COPY --from=build /go/bin/getmyip /usr/local/bin/getmyip
COPY chaperone.d/ /etc/chaperone.d

ENTRYPOINT ["/usr/local/bin/chaperone"]
