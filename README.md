# <img src="server/static/images/eremiteLOGO_02.png" width="400px" alt="Eremetic">

[![Build Status][travis-image]](https://travis-ci.org/klarna/eremetic)
[![Coverage Status][coveralls-image]](https://coveralls.io/r/klarna/eremetic?branch=master)

## Purpose
Eremetic is a Mesos Framework to run one-shot tasks. The vision is to provide a
bridge between Applications that need to run tasks and Mesos. That way a developer
creating an application that needs to schedule tasks (such as cron) wouldn't need
to connect to Mesos directly.

## Usage
Send a cURL to the eremetic framework with how much cpu and memory you need, what docker image to run and which command to run with that image.

```bash
curl -H "Content-Type: application/json" \
     -X POST \
     -d '{"task_mem":22.0, "task_cpus":1.0, "docker_image": "a_docker_container", "command": "rails"}' \
     http://eremetic_server:8080/task
```

These basic fields are required but you can also specify volumes, ports, environment
variables, docker parameters, and URIs for the mesos fetcher to download. See
[examples.md](examples.md) for more examples on how to use eremetic.

JSON format:

```javascript
{
  // Float64, fractions of a CPU to request
  "task_cpus":      1.0,
  // Float64, memory to use (MiB)
  "task_mem":       22.0,
  // String, full tag or hash of container to run
  "docker_image":   "busybox",
  // Boolean, if set to true, docker image will be pulled before each task launch.
  "force_pull_image": false,
  // String, command to run in the docker container
  "command": "echo date",
  // Array of Objects, volumes to mount in the container
  "volumes": [
    {
      "container_path": "/var/run/docker.sock",
      "host_path": "/var/run/docker.sock"
    }
  ],
  // Array of Objects, ports to forward to the container
  "ports": [
    {
      "container_port": 80,
      "protocol": "tcp"
    }
  ],
  // Object, Environment variables to pass to the container
  "env": {
    "KEY": "value"
  },
  // Object, Will be merged to `env` when passed to Mesos, but masked when doing a GET.
  // See Clarification of the Masked Env field below for more information
  "masked_env": {
    "KEY": "value"
  },
  // Object, Parameters to pass to docker when starting.
  // This allows passing things like --volume-driver to the container executor
  "parameters": {
    "KEY": "value"
  },
  // URIs of resource to download
  "uris": [
    "http://server.local/resource"
  ],
  // URIs and attributes of resource to download
  // Please note that `uris` auto-extract archive files based on their extension
  // with `fetch`, you need to explicitly define `"extract"` to unarchive files.
  "fetch": [
    {
      "uri" : "http://server.local/another_resource",
      "extract": false,
      "executable": false,
      "cache": false
    }
  ],
  // Constraints for which slave the task can run on (beyond cpu/memory).
  // Matching is strict and only attributes are currently supported. If
  // multiple constraints exist, they are evaluated using AND (ie: all or none).
  "slave_constraints": [
      {
          "attribute_name": "aws-region",
          "attribute_value": "us-west-2"
      }
  ],
  // String, URL to post a callback to. Callback message has format:
  // {"time":1451398320,"status":"TASK_FAILED","task_id":"eremetic-task.79feb50d-3d36-47cf-98ff-a52ef2bc0eb5"}
  "callback_uri": "http://callback.local"
}
```

### Note
Most of this meta-data will not remain after a full restart of Eremetic.

### Clarification of the Masked Env field
The purpose of the field is to provide a way to pass along environment variables that you don't want to have exposed in a subsequent GET call.
It is not intended to provide full security, as someone with access to either the machine running Eremetic or the Mesos Slave that the task is being run on will still be able to view these values.
These values are not encrypted, but simply masked when retrieved back via the API.

For security purposes, ensure TLS (https) is being used for the Eremetic communication and that access to any machines is properly restricted.


## Configuration
create /etc/eremetic/eremetic.yml with:

    address: 0.0.0.0
    port: 8080
    master: zk://<zookeeper_node1:port>,<zookeeper_node2:port>,(...)/mesos
    messenger_address: <callback address for mesos>
    messenger_port: <port for mesos to communicate on>
    loglevel: DEBUG
    logformat: json

## Database
Eremetic uses a database to store task information. The driver can be configured
by setting the `database_driver` value.

Allowed values are: `zk`, `boltdb`

The location of the database can be configured by setting the `database` value.

### BoltDB
The default database that will be used unless anything is configured.

The default value of the `database` field is `db/eremetic.db`

### ZooKeeper
If you use `zk` as a database driver, the `database` field must be provided as a
complete zk-uri (zk://zk1:1234,zk2:1234/my/database).

## Authentication
To enable mesos framework authentication add the location of credential file to your configuration:

    credential_file: /var/mesos_secret

The file should contain the Principal to authenticate and the secret separated by white space like so:

    principal    secret_key

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

## Running with minimesos
Using [minimesos](https://www.minimesos.org/) is a very simple way to test and play with eremetic.

```bash
docker run -e MASTER=$MINIMESOS_ZOOKEEPER -e HOST=0.0.0.0 -e DATABASE_DRIVER=zk -e PORT=8000 -p 8000:8000 alde/eremetic:latest
```

## Contributors

These are the fine folks who helped build eremetic

- Rickard Dybeck
- David Keijser
- Aidan McGinley
- William Strucke
- Charles G.
- Clément Laforet
- Marcus Olsson

## Acknowledgements
Thanks to Sebastian Norde for the awesome logo!

## Licensing
Apache-2

[travis-image]: https://img.shields.io/travis/klarna/eremetic.svg?style=flat
[coveralls-image]: https://img.shields.io/coveralls/klarna/eremetic.svg?style=flat
