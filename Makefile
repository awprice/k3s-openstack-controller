.PHONY: build run

build:
	go build -o k3s-openstack-controller cmd/main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o k3s-openstack-controller cmd/main.go

run: build
	./k3s-openstack-controller