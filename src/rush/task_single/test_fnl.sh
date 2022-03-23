#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

export PATH="$PATH:$PWD/../task/external"
export CERT="$HOME/.mitmproxy/mitmproxy-ca.pem"
PRODUCT_URL=${PRODUCT_URL:-app://www.finishline.com/store/product/nike-pro-hyperwarm-balaclava-hood/prod2793159?styleId=NHK63&colorId=058}
# PRODUCT_URL=${PRODUCT_URL:-https://www.finishline.com/store/product/mens-finish-line-6-pack-quarter-socks/prod734432?colorId=BLK&styleId=89947}
SIZE=${SIZE:-""}
PROFILE=${PROFILE:-"profile_test_dc.json"}
PROXY=${PROXY:-""} # -proxy http://proxy:WOO4sVE@199.188.93.98:8000
COUNT=${COUNT:-1}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT $PROXY
