.PHONY: all deps test docker copy_templates

DOCKERTAG?=eremetic
SRC=$(shell find . -name '*.go')

all: test

deps:
	go get -t ./...

test: eremetic
	go test -v ./...

eremetic: deps
eremetic: ${SRC}
	go build -o $@

docker/eremetic: ${SRC}
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@

copy_templates:
	cp -r static docker
	cp -r templates docker

docker: docker/eremetic copy_templates
	docker build -t ${DOCKERTAG} docker
