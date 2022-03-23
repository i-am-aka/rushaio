import appdirs
import sys
import os

if hasattr(sys, "_MEIPASS"):
  APP_DIR = sys._MEIPASS
else:
  APP_DIR = os.path.dirname(sys.argv[0])

APP_DIR = appdirs.user_data_dir("RushAIO", "RushAIO")

DEBUG = os.getenv('DEBUG') == '1'

POOL_JSON = os.path.join(APP_DIR, 'firefox', 'pool.json')

FF_DIR = os.path.join(APP_DIR, 'firefox')
if os.name == 'nt':
  GD_BINARY = os.path.join(FF_DIR, 'geckodriver.exe')
  FF_BINARY = os.path.join(FF_DIR, 'firefox.exe')
else:
  GD_BINARY = os.path.join(FF_DIR, 'geckodriver')
  GD_BINARY_SHA256 = "e30150c8547eed9886929e531cb81545aff8d8403201de7efd37e3946452db9b"
  FF_BINARY = os.path.join(FF_DIR, 'Firefox.app/Contents/MacOS/firefox')
  FF_BINARY_SHA256 = "96d81da6c376fa7afd25543d3968951249dcc01ce5ba985b81fc13efbd647824"