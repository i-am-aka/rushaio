#!/usr/bin/env bash
set -ex
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
mkdir -p bin
rm bin/* || true
export CGO_ENABLED=1
GODEBUG=x509ignoreCN=1 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/task_consumer
if [ -z "$NO_WINDOWS" ]; then
  GODEBUG=x509ignoreCN=1 xgo -ldflags="-s -w" -dest bin/ -targets windows-8.1/amd64 ./
fi

go install .
