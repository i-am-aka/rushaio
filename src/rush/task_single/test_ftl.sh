#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR
export PUBLIC_WEBHOOK=1
PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"https://www.champssports.com/product/nike-club-pullover-hoodie-mens/Z4385016.html"}
SIZE=${SIZE:-""}
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"fake"}
COUNT=${COUNT:-1}
DELAY=${DELAY:-1000}
BINARY=${BINARY:-"go run ."}
if [ -n "$PROXY" ]; then
  DELAY=$DELAY $BINARY -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT -proxy "$PROXY"
else
  DELAY=$DELAY $BINARY -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT
fi
