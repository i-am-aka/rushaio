#!/usr/bin/env bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd $DIR

if [ -z $SKIP_TASK ]; then
  ../src/rush/task_consumer/build.sh
  cp ../src/rush/task_consumer/bin/* bin/
  [ ! -L adyen ] && ln -s ../src/rush/task/adyen adyen
  [ ! -L adyen2 ] && ln -s ../src/rush/task/adyen2 adyen2
  # [ ! -L zeronaught ] && ln -s ../src/rush/task/zeronaught zeronaught
  [ ! -L eprotect ] && ln -s ../src/rush/task/eprotect eprotect
  [ ! -L incap ] && ln -s ../src/rush/task/incap incap
  [ ! -L wmt ] && ln -s ../src/rush/task/wmt wmt
  [ ! -L external ] && ln -s ../src/rush/task/external external
  [ ! -L paypal ] && ln -s ../src/rush/task/paypal paypal
  [ ! -L datadome ] && ln -s ../src/rush/task/datadome datadome
fi

pyinstaller --clean -y  --onefile --noconfirm rush.osx.spec
mv dist/RushAIO.app/Contents/MacOS/rush dist/RushAIO.app/Contents/MacOS/rush_cli
cp app/rush_init.sh dist/RushAIO.app/Contents/MacOS/rush_init
cat << 'EOF' > dist/RushAIO.app/Contents/MacOS/rush
#!/usr/bin/osascript
set initScriptPath to POSIX path of ((path to me as text) & "::" & "rush_init")
tell app "Terminal" to activate
tell app "Terminal" to do script initScriptPath
EOF
chmod +x dist/RushAIO.app/Contents/MacOS/rush

zip -r dist/RushAIO.zip dist/RushAIO.app

[ -z $NO_WINDOWS ] && ./build_win.sh

if [ -n $BUILD ]; then
  cd dist && BUILD=$BUILD ./upload.sh 
fi
