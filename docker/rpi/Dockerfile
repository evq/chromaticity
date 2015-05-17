FROM jordi/golang-raspi
MAINTAINER Evey Quirk

RUN rm /etc/services
RUN apt-get install -y git

ADD . /go/src/github.com/evq/chromaticity

RUN cd /go/src/github.com/evq/chromaticity && go get
RUN cd /go/src/github.com/evq/chromaticity && go install

CMD /go/bin/chromaticity
VOLUME /root/.chromaticity

EXPOSE 80
