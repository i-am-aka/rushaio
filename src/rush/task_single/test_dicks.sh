#!/usr/bin/env bash
#set -x
PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"https://www.dickssportinggoods.com/p/nike-mens-dri-fit-challenger-9-running-shorts-18nikmmnkchllgrshapbb/18nikmmnkchllgrshapbb"}
SIZE=${SIZE:-""}
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"./profile_test_dc.json"}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat $PROFILE)" $PROXY # -keywords "fire,red,5"
