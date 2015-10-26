.PHONY: all docker

DOCKERTAG?=eremetic
SRC=$(shell find . -name '*.go')

all: eremetic

eremetic: ${SRC}
	go build -o $@

docker/eremetic: ${SRC}
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $@

docker: docker/eremetic
	docker build -t ${DOCKERTAG} docker
