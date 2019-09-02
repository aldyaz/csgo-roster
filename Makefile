# Copyright 2019 Aldyaz.

IMAGE := csgo-roster
VERSION := $(shell git describe --tags --always --dirty)

test:
	go test -v -cover -p 1 ./...

build:
	CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...

build-linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...

docker-build: Dockerfile
    echo "Building the $(IMAGE) docker container.."
    docker build --label "version=$(VERSION)" -t $(IMAGE):$(VERSION) .

docker-run:
    docker run -it -p 8080:8080 --rm $(IMAGE):$(VERSION)