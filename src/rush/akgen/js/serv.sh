#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
DENO_DIR=$DIR/deno deno run --allow-read --allow-write --unstable ./serv.js $@
