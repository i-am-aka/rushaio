#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

PATH="$PATH:$PWD/../task/external"
#PRODUCT_URL=${PRODUCT_URL:-https://www.eblens.com/nike/men-s-air-force-1-07-85365}
PRODUCT_URL=${PRODUCT_URL:-https://www.eblens.com/nike/performance-cushioned-crew-training-socks-62803}
SIZE=${SIZE:-""}
# PROXY=${PROXY:-"http://sp329576880:189XbRbOYdi2HlVF_country-UnitedStates@p.stellarproxies.com:32000"}
# PROXY="http://localhost:8084"
PROFILE=${PROFILE:-"profile_test.json"}
# go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat ./profile_coach.json)" -proxy $PROXY # "http://proxy:WOO4sVE@199.188.93.98:8000" # -proxy "http://localhost:8084" # -proxy "http://proxy:WOO4sVE@199.188.93.98:8000" #-proxy "http://localhost:8084" # -proxy "http://rushaio:rushaio@52.55.146.251:8084"
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")"  -proxy $PROXY
