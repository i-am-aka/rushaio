#!/usr/bin/env bash
set -ex
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/rclogin
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o bin/rclogin.exe
go install .


