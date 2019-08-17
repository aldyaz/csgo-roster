# Copyright 2018 Core Services Team.

# Which architecture to build
ARCH ?= amd64

test:
	go test -v -cover -p 1 ./...

build:
	CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...

build-linux:
	GOOS=linux CGO_ENABLED=0 GOARCH=${ARCH} go install ./cmd/...
