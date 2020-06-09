FROM golang:1.14 as builder
WORKDIR /go/src/github.com/awprice/k3s-openstack-controller
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY Gopkg.toml Gopkg.lock Makefile ./
RUN dep ensure -v -vendor-only
COPY cmd cmd
COPY internal internal
RUN make build-linux

FROM alpine:3
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/awprice/k3s-openstack-controller/k3s-openstack-controller /bin/k3s-openstack-controller
CMD [ "/bin/k3s-openstack-controller" ]