#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"https://www.champssports.com/product/~/15122001.html"}
SIZE=${SIZE:-""}
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"profile_test_dc.json"}
COUNT=${COUNT:-1}
DELAY=${DELAY:-1000}
BINARY=${BINARY:-"go run ."}
if [ -n "$PROXY" ]; then
  DELAY=$DELAY $BINARY -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT -proxy "$PROXY"
else
  DELAY=$DELAY $BINARY -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT
fi
