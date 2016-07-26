.PHONY: all test test-server test-docker docker docker-clean publish-docker 

VERSION?=$(shell git describe HEAD | sed s/^v//)
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
DOCKERNAME?=alde/eremetic
DOCKERTAG?=${DOCKERNAME}:${VERSION}
LDFLAGS=-X main.Version '${VERSION}' -X main.BuildDate '${DATE}'
TOOLS=${GOPATH}/bin/go-bindata \
      ${GOPATH}/bin/go-bindata-assetfs \
      ${GOPATH}/bin/goconvey
SRC=$(shell find . -name '*.go')
STATIC=$(shell find static templates)

DOCKER_GO_SRC_PATH=/go/src/github.com/klarna/eremetic
DOCKER_GOLANG_RUN_CMD=docker run --rm -v "$(PWD)":$(DOCKER_GO_SRC_PATH) -w $(DOCKER_GO_SRC_PATH) golang:1.6 bash -c

all: test

${TOOLS}:
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/elazarl/go-bindata-assetfs/...
	go get github.com/smartystreets/goconvey

test: eremetic
	go test -v ./...

test-server: ${TOOLS}
	${GOPATH}/bin/goconvey

# Run tests cleanly in a docker container.
test-docker:
	$(DOCKER_GOLANG_RUN_CMD) "make test"

assets/assets.go: generate.go ${STATIC}
	go generate

eremetic: ${TOOLS} assets/assets.go
eremetic: ${SRC}
	go get -t ./...
	go build -ldflags "${LDFLAGS}" -o $@

docker/eremetic: ${TOOLS} assets/assets.go
docker/eremetic: ${SRC}
	go get -t ./...
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -installsuffix cgo -o $@

docker: docker/eremetic docker/Dockerfile docker/marathon.sh
	docker build -t ${DOCKERTAG} docker

docker-clean: docker/Dockerfile docker/marathon.sh
	# Create the docker/eremetic binary in the Docker container using the
	# golang docker image. This ensures a completely clean build.
	$(DOCKER_GOLANG_RUN_CMD) "make docker/eremetic"
	docker build -t ${DOCKERTAG} docker

publish-docker:
ifeq ($(strip $(shell docker images --format="{{.Repository}}:{{.Tag}}" $(DOCKERTAG))),)
	$(warning Docker tag does not exist:)
	$(warning ${DOCKERTAG})
	$(warning )
	$(error Cannot publish the docker image. Please run `make docker` or `make docker-clean` first.)
endif
	docker push ${DOCKERTAG}
	git describe HEAD --exact 2>/dev/null && \
		docker tag ${DOCKERTAG} ${DOCKERNAME}:latest && \
		docker push ${DOCKERNAME}:latest || true
