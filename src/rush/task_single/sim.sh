#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

trap 'kill $(jobs -p)' EXIT

set -e
if ! lsof -Pi:5001; then
  pushd ../../../queue
    PORT=5001 nohup ./sim.py >> sim.log 2>&1 &
    SIM_PID=$!
  popd
fi
#if ! lsof -Pi:5000; then
#  pushd ../../../
#    nohup ./server.py >> server.log 2>&1&
#    API_PID=$!
#  popd
#fi
if ! lsof -Pi:8084 | grep -i listen; then
  pushd ../../../mitm
    mitmdump --listen-host 0.0.0.0 -p 8084 --proxyauth 'foo:bar' -k -v --set termlog_verbosity=debug --set console_eventlog_verbosity="debug" --set upstream_cert=false --set ssl_insecure=true --set allow_remote=true -s rewrite.py -w $(date +%s).dump >> mitm.log 2>&1 &
    MITM_PID=$!
  popd
fi
#if ! lsof -Pi:42424; then
#  pushd ../rcpool
#    go run . >> rcpool.log 2>&1 &
#    RCPOOL_PID=$!
#  popd
#fi
while ! lsof -Pi:5001 && ! lsof -Pi:8084 ; do
  sleep 0.1
done
sleep 2
#for i in $(seq 1 1000); do
# tail -f ../rcpool/rcpool.log ../mitm/mitm.log ../server.log ../queue/sim.log &

[ -z $NO_TASK ] && DEBUG=1 LOG=1 WEBHOOK=0 go run . -productUrl https://www.yeezysupply.com/product/H68771 -profile "$(cat profile_test.json)" -proxy 'http://foo:bar@127.0.0.1:8084' >> sim.log 2>&1 &
tail -f sim.log ../rcpool/rcpool.log ../../../server.log ../../../queue/sim.log # ../../../mitm/mitm.log

#  ./task -product H67799 -profile "$(cat profile_test.json)" -size 7 -proxy 'http://foo:bar@localhost:8084'
#done
sleep 3600
