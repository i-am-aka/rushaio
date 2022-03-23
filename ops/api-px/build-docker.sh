#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR/../../src

docker build --no-cache -f $DIR/Dockerfile -t rush-px-api .
