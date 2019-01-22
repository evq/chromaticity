FROM golang:1.11 as builder
MAINTAINER eV Quirk
ARG DOCKER_TARGET
ARG GOARCH
ARG GOARM

WORKDIR /chromaticity/
RUN go get -u github.com/jteeuwen/go-bindata/...
COPY . ./
RUN make
RUN GOARCH=${GOARCH:-arm} GOARM=${GOARM:-5} go build

FROM ${DOCKER_TARGET:-arm32v6}/alpine:3.6
COPY --from=builder /chromaticity/chromaticity /bin/
EXPOSE 80
VOLUME /root/.chromaticity
CMD ["/bin/chromaticity"]
