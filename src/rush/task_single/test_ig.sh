#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

PATH="$PATH:$PWD/../task/external"

# sweats
# PRODUCT_URL=${PRODUCT_URL:-"ig://andyhaden:zzzzz@7307118110/2987970601227191"}
SIZE=${SIZE:-""}
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"profilep.json"}
COUNT=${COUNT:-1}
# CONFIG=${CONFIG:-"-config username=andyhaden,password=1RushAio"}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT $CONFIG $PROXY
