.PHONY: all deps test docker copy_templates publish-docker

VERSION?=$(shell git describe HEAD | sed s/^v//)
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
DOCKERTAG?=dkeis/eremetic:${VERSION}
LDFLAGS=-X main.Version '${VERSION}' -X main.BuildDate '${DATE}'
SRC=$(shell find . -name '*.go')

all: test

deps:
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/elazarl/go-bindata-assetfs/...
	go generate
	go get -t ./...

test: eremetic
	go test -v ./...

eremetic: deps
eremetic: ${SRC}
	go build -ldflags "${LDFLAGS}" -o $@

docker/eremetic: deps
docker/eremetic: ${SRC}
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -installsuffix cgo -o $@

docker: docker/eremetic
	docker build -t ${DOCKERTAG} docker

publish-docker: docker
	docker push ${DOCKERTAG}
