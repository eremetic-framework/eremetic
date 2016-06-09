# Examples

## Running a simple health check task

This runs the `dkeis/dht-probe` image with the default command, it will check
if we can find peers on the network with dht.

```json
{
    "docker_image": "dkeis/dht-probe",
    "task_cpus": 0.1,
    "task_mem": 100.0
}
```

## Interfacing with docker on the host

By mounting the docker socket into the container the application is able to
interface with the docker api. Here we use
[cibox](https://github.com/keis/cibox) to run tests in a dynamically created
docker container.

Beware security implications of mounting the docker socket as it gives full
access to the *host* system.

```json
{
    "command": "/cibox https://git@github.com/keis/cibox.git --matrix-id 0",
    "docker_image": "dkeis/cibox",
    "task_cpus": 0.1,
    "task_mem": 100.0,
    "volumes": [
        {
            "container_path": "/var/run/docker.sock",
            "host_path": "/var/run/docker.sock"
        }
    ]
}
```

## Building a docker image

The docker api also enables us to build docker images and by combining this
with a list of `uris` that will be downloaded by mesos we can build images from
arbitrary dockerfiles by url.

```json
{
    "command": "docker build -t dkeis/golang /tmp/build && docker login -u $USER -p $PASSWORD && docker push dkeis/golang",
    "docker_image": "docker:1.8",
    "env": {
        "USER": "dkeis"
    },
    "masked_env": {
        "PASSWORD": "myactualpassword"
    },
    "task_cpus": 0.1,
    "task_mem": 100.0,
    "uris": [
        "https://raw.githubusercontent.com/docker-library/golang/3cdd85183c0f3f6608588166410d24260cd8cb2f/1.6/alpine/Dockerfile"
    ],
    "volumes": [
        {
            "container_path": "/var/run/docker.sock",
            "host_path": "/var/run/docker.sock"
        },
        {
            "container_path": "/tmp/build/Dockerfile",
            "host_path": "Dockerfile"
        }
    ]
}
```

## Running a task with certain attributes

This configures a task to run the `busybox` image with a basic loop outputting
the time on any Mesos Slave with the attribute "role" set to "build".

```json
{
    "docker_image": "busybox",
    "command": "for i in $(seq 1 5); do echo \"`date` $i\"; sleep 5; done",
    "task_cpus": 0.1,
    "task_mem": 100.0,
    "slave_constraints": [
        {
            "attribute_name": "role",
            "attribute_value": "build"
        }
    ]
}
```

Mesos slaves can be configured with arbitrary attributes. See the
[documentation](https://open.mesosphere.com/reference/mesos-slave/) for more
information on how to configure attributes.
