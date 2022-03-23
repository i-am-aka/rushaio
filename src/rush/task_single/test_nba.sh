#!/usr/bin/env bash
PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"https://store.nba.com/los-angeles-lakers/mens-los-angeles-lakers-islide-black-tonal-pop-slide-sandals/t-36929646+p-8103113572872+z-9-2923626566?_ref=p-DLP:m-GRID:i-r0c0:po-0"}
# PRODUCT_URL=${PRODUCT_URL:-"https://store.nba.com/los-angeles-lakers/los-angeles-lakers-for-bare-feet-argyle-crew-socks/t-25697424+p-6804226635993+z-9-2707603240?_ref=p-DLP:m-GRID:i-r1c1:po-4"}
#PRODUCT_URL="https://store.nba.com/footwear/d-1294282497+z-957317-2506523281?pageSize=24&sortOption=NewestArrivals"
SIZE=${SIZE:-""}
# PROXY=""
PROFILE=${PROFILE:-"./profile_test_dc.json"}
PROXY=${PROXY:-""}
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat $PROFILE)" $PROXY #-keywords "Men's Phoenix Suns Devin Booker Jordan Brand Orange"
