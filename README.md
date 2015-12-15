# <img src="static/images/eremiteLOGO_02.png" width="400px" alt="Eremetic">[![Build Status][travis-image]](https://travis-ci.org/alde/eremetic)

Eremetic is a Mesos Framework to run one-shot tasks.

## Usage
Send a cURL to the eremetic framework with how much cpu and memory you need, what docker image to run and which command to run with that image.

    curl -H "Content-Type: application/json" \
         -X POST \
         -d '{"task_mem":22.0, "task_cpus":1.0, "docker_image": "a_docker_container", "command": "rails"}' \
         http://eremetic_server:8080/task

## Configuration
create /etc/eremetic/eremetic.yml with:

    address: 0.0.0.0
    port: 8080
    master: zk://<zookeeper_node1:port>,<zookeeper_node2:port>,(...)/mesos
    messenger_address: <callback address for mesos>
    messenger_port: <port for mesos to communicate on>

## Building

### Environment
Clone the repository into `$GOCODE/src/github.com/klarna/eremetic`.
This is needed because of internal package dependencies

### Install dependencies
First you need to install dependencies. Parts of the eremetic code is auto-generated (assets and templates for the HTML view are compiled). In order for go generate to work, `go-bindata` and `go-bindata-assetfs` needs to be manually installed.

    go get github.com/jteeuwen/go-bindata/...
    go get github.com/elazarl/go-bindata-assetfs/...
    go generate
    go get -t ./...

### Creating the docker image
To build a docker image with eremetic, simply run

    make docker

## Running on mesos

Eremetic can itself by run on mesos using e.g marathon. An
[example configuration](misc/eremetic.json) for marathon is provided that is
ready to be submitted through the api.

```bash
curl -X POST -H 'Content-Type: application/json' $MARATHON/v2/apps -d@misc/eremetic.json
```

## Running tests
The tests rely on [GoConvey](http://goconvey.co/), and can be run either by running `goconvey` or `go test ./...` from the project root.

## Contributors
- Rickard Dybeck
- David Keijser

## Acknowledgements
Thanks to Sebastian Norde for the awesome logo!

## Licensing
Apache-2

[travis-image]: https://travis-ci.org/alde/eremetic.svg
