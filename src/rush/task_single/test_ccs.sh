#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

export PATH="$PATH:$PWD/../task/external"
export CERT="$HOME/.mitmproxy/mitmproxy-ca.pem"
PRODUCT_URL=${PRODUCT_URL:-https://shop.ccs.com/dc-og-legacy-shoes-pastel-11-0}
SIZE=${SIZE:-""}
PROFILE=${PROFILE:-"profile_test.json"}
PROXY=${PROXY:-"http://127.0.0.1:8084"}
COUNT=${COUNT:-1}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" # -proxy $PROXY -count $COUNT
