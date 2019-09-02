# Copyright 2019 Aldyaz.

IMAGE := csgo-roster
VERSION := $(shell git describe --tags --always --dirty)

test:
	go test -v -cover -p 1 ./...

build:
	CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...

build-linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...

docker: Dockerfile
	echo "Building the $(IMAGE) container..."
	docker build --label "version=$(VERSION)" -t $(IMAGE):$(VERSION) .