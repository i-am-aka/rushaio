#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"https://www.walmart.com/ip/Black-Mountain-Products-Resistance-Band-Set-with-Door-Anchor-Ankle-Strap-Exercise-Chart-and-Resistance-Band-Carrying-Case/23471635?account=0"}
SIZE=${SIZE:-""}
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"profile_test_dc.json"}
COUNT=${COUNT:-1}

go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT $PROXY
