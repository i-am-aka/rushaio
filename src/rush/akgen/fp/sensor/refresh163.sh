#!/usr/bin/env bash

scp rush:./ysfc/src/rush/fp/collect/$1 . && cat $1 |  grep '^{' | jq -cr 'select(.akTel) | .akTel'  > $1.sensor && cat fpaug*sensor.txt fpseptsensor.txt fpseptsensor2.txt fprestocksensor.txt fpoct86.txt.sensor fpoct87.txt.sensor fpoct88.txt.sensor  fpoct89.txt.sensor $1.sensor > fpsensor.txt && FPFILE=fpsensor.txt go run .
gzip -f --keep sensors.gob
