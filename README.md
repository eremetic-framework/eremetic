# Eremetic

Eremetic is a Mesos Framework to run one-shot tasks.

## Planned usage
Send a cURL to the eremetic framework with how much cpu and memory you need, what docker image to run and which command to run with that image.

    curl -H "Content-Type: application/json" \
          -X POST \
          -d '{"task_mem":22.0, "docker_image": "a_docker_container", "command": "rails", "task_cpus":1.0, "tasks_to_launch": 3}' \
          http://eremetic_server:8080/task

## Configuration
create /etc/eremetic/eremetic.yml with:

    address: 0.0.0.0
    port: 8080
    zookeeper: zk://<zookeeper_node1:port>,<zookeeper_node2:port>,(...)/mesos
