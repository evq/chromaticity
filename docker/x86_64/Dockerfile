FROM golang
MAINTAINER Evey Quirk

ADD . /go/src/github.com/evq/chromaticity

RUN go get -u github.com/jteeuwen/go-bindata/...

RUN cd /go/src/github.com/evq/chromaticity && make
RUN cd /go/src/github.com/evq/chromaticity && go get
RUN cd /go/src/github.com/evq/chromaticity && go install

CMD /go/bin/chromaticity
VOLUME /root/.chromaticity

EXPOSE 80
