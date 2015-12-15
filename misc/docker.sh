#!/bin/bash

set -e

test -n "$DOCKER_USERNAME" -a "$PUBLISH_VERSION" == "$TRAVIS_GO_VERSION"
grep -E 'master|v[0-9.]+' > /dev/null <<< "$TRAVIS_BRANCH"

docker login -e="$DOCKER_EMAIL" -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
make publish-docker
