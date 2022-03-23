#!/usr/bin/env bash
PATH="$PATH:$PWD/../task/external"
PROXY=${PROXY:-""}
PRODUCT=${PRODUCT:-https://www.yeezysupply.com/product/F36640}
PROFILE=${PROFILE:-"profile_test.json"}
COUNT=${COUNT:-1}
# PROXY=""
#go run . -size "$SIZE" -product $PRODUCT -profile "$(cat ./profilep.json)" #-proxy "$PROXY"
COUNT=$COUNT WEBHOOK=$WEBHOOK NO_WEBHOOK=$NO_WEBHOOK go run . -size "$SIZE" -productUrl $PRODUCT -count $COUNT -profile "$(cat ./profile_test.json)" $PROXY # -proxy "$PROXY"
