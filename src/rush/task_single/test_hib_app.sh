#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

PATH="$PATH:$PWD/../task/external"
PRODUCT_URL=${PRODUCT_URL:-"app://www.hibbett.com/jordan-maxin-200-black-rush-blue-white-mens-shoe/Q0773.html?utm_source=google&utm_medium=organic&utm_campaign=surfaces"}
# https://www.hibbett.com/nike-sportswear-big-kids-tech-black-white-fleece-pants/4924B.html}
# PRODUCT_URL="https://www.hibbett.com/nike-air-force-1-low-grade-school-white-kids-casual-shoe/5068ZW.html"
# PRODUCT_URL="https://www.hibbett.com/nike-air-force-1-react-qs-white-light-bone-sail-mens-shoe/0P400.html"
SIZE=${SIZE:-""}
# https://www.hibbett.com/jordan-1-retro-high-og-black-white-game-royal-mens-shoe/Q0851.html?dwvar_Q0851_color=0216&cgid=launch-calendar#start=8
PROXY=${PROXY:-""}
PROFILE=${PROFILE:-"profile_test.json"}
COUNT=${COUNT:-1}
export CERT="$HOME/drive/keys/charles-ssl-proxying-certificate.pem"
go run . -size "$SIZE" -productUrl $PRODUCT_URL -profile "$(cat "$PROFILE")" -count $COUNT $PROXY
