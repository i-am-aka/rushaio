#!/usr/bin/env bash
PRODUCT=${PRODUCT:-https://www.yeezysupply.com/product/FX0496}
# http://proxy:dADGGU7@199.188.92.176:8000
# PROXY=""
go run . -productUrl $PRODUCT -profile "$(cat profile_test.json)" #-proxy "$PROXY"
