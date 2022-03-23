#!/usr/bin/env bash
PATH="$PATH:$PWD/../task/external"
PRODUCT_URL="https://www.dickssportinggoods.com/p/adias-mens-adilette-aqua-slides-19adiadlttqblwhtxope/19adiadlttqblwhtxope"
SIZE=${SIZE:-"12.0"}

go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat ./profilep.json)" -keywords "fire,red,5" -proxy "http://localhost:8084" # -proxy "http://rushaio:rushaio@52.55.146.251:8084"
