FROM debian:stable AS build
RUN echo 'deb http://deb.debian.org/debian buster contrib' > /etc/apt/sources.list.d/contrib.list
RUN apt-get -q update && env DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends build-essential pkg-config golang-go git ca-certificates geoipupdate

ENV GOPATH=/go
RUN mkdir $GOPATH && \
    go get 0xacab.org/leap/getmyip && \
    strip $GOPATH/bin/getmyip

FROM debian:stable
RUN echo 'deb http://deb.debian.org/debian buster contrib' > /etc/apt/sources.list.d/contrib.list
RUN apt-get -q update && env DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends geoipupdate
COPY --from=build /usr/share/doc/geoipupdate/examples/GeoIP.conf.default /etc/GeoIP.conf
COPY --from=build /go/bin/getmyip /usr/local/bin/getmyip

ENTRYPOINT ["/usr/local/bin/getmyip", "-notls"]
