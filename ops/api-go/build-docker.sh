#!/usr/bin/env bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR/../../src

docker build --no-cache -f $DIR/Dockerfile -t rush-api-build .
docker run -v $DIR/out:/out rush-api-build
ls -tral $DIR/out
$DIR/runtime/build-docker.sh

