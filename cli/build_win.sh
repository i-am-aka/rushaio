#!/usr/bin/env bash
set -ex
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

# pipenv run pip freeze > requirements.txt
cat << EOF > requirements.txt
wcwidth==0.1.9
win-unicode-console==0.5
sentry-sdk==0.19.4
tableformatter==0.1.4
pycryptodome==3.9.7
timeago==1.0.13
cmd2==1.0.1
appdirs==1.4.3
requests==2.23.0
psutil==5.7.0
pyzmq==19.0.0
selenium==3.141.0
flask
tqdm
pywin32
EOF

docker run --rm -v "$(PWD):/src/" \
  -v $(readlink -f $PWD/external):/src/external \
  -v $(readlink -f $PWD/adyen):/src/adyen \
  -v $(readlink -f $PWD/adyen2):/src/adyen2 \
  -v $(readlink -f $PWD/eprotect):/src/eprotect \
  -v $(readlink -f $PWD/incap):/src/incap \
  -v $(readlink -f $PWD/wmt):/src/wmt \
  -v $(readlink -f $PWD/paypal):/src/paypal \
  -v $(readlink -f $PWD/rcfarm):/src/rcfarm \
  -v $(readlink -f $PWD/datadome):/src/datadome \
  cdrx/pyinstaller-windows:python3 \
  'bash -c ". /root/.bashrc; pip install -r requirements.txt;
  cp /wine/drive_c/Python37/Lib/site-packages/zmq/libzmq.cp37-win_amd64.pyd /wine/drive_c/Python37/Lib/site-packages/zmq/backend/cython/;
  pyinstaller --uac-admin --clean -y --onefile --dist ./dist/windows --workpath /tmp --icon=rush.ico rush.win.spec"'
