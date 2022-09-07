.PHONY: all test test-server test-docker docker docker-clean publish-docker

REPO=github.com/rockerbox/eremetic
VERSION?=$(shell git describe HEAD | sed s/^v//)
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
DOCKERNAME?=rockerbox/eremetic
DOCKERTAG?=${DOCKERNAME}:${VERSION}
LDFLAGS=-X ${REPO}/version.Version=${VERSION} -X ${REPO}/version.BuildDate=${DATE}
TOOLS=${GOPATH}/bin/go-bindata \
      ${GOPATH}/bin/go-bindata-assetfs \
      ${GOPATH}/bin/goconvey
SRC=$(shell find . -name '*.go')
STATIC=$(shell find server/static server/templates)
TESTFLAGS="-v"

DOCKER_GO_SRC_PATH=/go/src/github.com/rockerbox/eremetic
DOCKER_GOLANG_RUN_CMD=docker run --rm -v "$(PWD)":/opt/eremetic -w /opt/eremetic golang:1.12 bash -c

PACKAGES=$(shell go list ./... | grep -v /vendor/)

all: test

deps: ${TOOLS}

${TOOLS}:
    curl https://github.com/kevinburke/go-bindata/releases/download/v3.22.0/go-bindata-linux-amd64 > $GOPATH/bin/go-bindata
	go get -u github.com/elazarl/go-bindata-assetfs@v1.0.0
	go get -u github.com/smartystreets/goconvey@v1.6.4

test: eremetic
	go test ${TESTFLAGS} ${PACKAGES}

test-server: ${TOOLS}
	${GOPATH}/bin/goconvey

# Run tests cleanly in a docker container.
test-docker:
	$(DOCKER_GOLANG_RUN_CMD) "make test"

vet:
	go vet ${PACKAGES}

lint:
	golint -set_exit_status $(shell go list ./... | grep -v /vendor/ | grep -v assets)

server/assets/assets.go: server/generate.go ${STATIC}
	go generate github.com/rockerbox/eremetic/server

eremetic: ${TOOLS} server/assets/assets.go
eremetic: ${SRC}
	go build -ldflags "${LDFLAGS}" -o $@ github.com/rockerbox/eremetic/cmd/eremetic

docker/eremetic: ${TOOLS} server/assets/assets.go
docker/eremetic: ${SRC}
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -installsuffix cgo -o $@ github.com/rockerbox/eremetic/cmd/eremetic

docker: docker/eremetic docker/Dockerfile docker/marathon.sh
	docker build --platform=linux/amd64 -t ${DOCKERTAG} docker

docker-clean: docker/Dockerfile docker/marathon.sh
	# Create the docker/eremetic binary in the Docker container using the
	# golang docker image. This ensures a completely clean build.
	$(DOCKER_GOLANG_RUN_CMD) "make docker/eremetic"
	docker build -t ${DOCKERTAG} docker

publish-docker:
#ifeq ($(strip $(shell docker images --format="{{.Repository}}:{{.Tag}}" $(DOCKERTAG))),)
#	$(warning Docker tag does not exist:)
#	$(warning ${DOCKERTAG})
#	$(warning )
#	$(error Cannot publish the docker image. Please run `make docker` or `make docker-clean` first.)
#endif
	docker push ${DOCKERTAG}
	git describe HEAD --exact 2>/dev/null && \
		docker tag ${DOCKERTAG} ${DOCKERNAME}:latest && \
		docker push ${DOCKERNAME}:latest || true
