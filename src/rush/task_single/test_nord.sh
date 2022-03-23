#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

export PATH="$PATH:$PWD/../task/external"
export CERT="$HOME/.mitmproxy/mitmproxy-ca.pem"
PRODUCT_URL=${PRODUCT_URL:-https://shop.nordstrom.com/s/cole-haan-generation-zerogrand-stitchlite-sneaker-men/5626093}
SIZE=${SIZE:-""}
PROFILE=${PROFILE:-"profile_test.json"}
PROXY=${PROXY:-"http://localhost:8084"}
COUNT=${COUNT:-1}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -proxy $PROXY -count $COUNT
