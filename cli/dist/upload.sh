#!/usr/bin/env bash

aws s3 cp RushAIO.zip s3://rushaio/RushAIO-$BUILD.zip; aws s3 presign s3://rushaio/RushAIO-$BUILD.zip --expires-in 604800

aws s3 cp windows/RushAIO.exe s3://rushaio/RushAIO-$BUILD.exe; aws s3 presign s3://rushaio/RushAIO-$BUILD.exe --expires-in 604800

