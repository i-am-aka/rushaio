#!/usr/bin/env bash
set -ex
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
GOOS=darwin GOARCH=amd64 go build -o bin/task
GOOS=windows GOARCH=amd64 go build -o bin/task.exe
go install .

cp bin/task bin/task.exe ../cli/bin

mkdir -p ../cli/adyen
cp -R adyen/* ../cli/adyen

mkdir -p ../cli/external
cp -R external/* ../cli/external
