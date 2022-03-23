#!/usr/bin/env bash
set -ex

#scp rush:./ysfc/src/rush/fp/collect/fpmobile.txt .
# scp rush:./ysfc/src/rush/fp/collect/fpaugust.txt .
cat fpaugust2.txt fpmobile.txt fpaugust.txt > fpall.txt
jq -cr 'select(.akTel) | .akTel' fpall.txt > sensor.txt
go run .
