FROM golang:1.11 as builder
MAINTAINER eV Quirk

WORKDIR /chromaticity/
RUN go get -u github.com/jteeuwen/go-bindata/...
COPY . ./
RUN make
RUN go build

FROM alpine:3.6
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /chromaticity/chromaticity /bin/
EXPOSE 80
VOLUME /root/.chromaticity
CMD ["/bin/chromaticity"]
