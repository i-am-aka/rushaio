#!/usr/bin/env python3

import appdirs
import json
import os
import subprocess
import tempfile
import traceback
from concurrent.futures import ThreadPoolExecutor
from rcfarm.config import *
from rcfarm.driver import start_browser

def launch_pool(n=os.cpu_count()):
  with ThreadPoolExecutor(n) as exc:
    for browser in exc.map(start_browser, [(FF_BINARY, GD_BINARY) for x in range(n)]):
      if browser:
        yield browser
  #       browsers.append(browser)
  # return browsers

def get_external():
  dirty = not os.path.exists(GD_BINARY)
  dirty = dirty or not os.path.exists(FF_BINARY)
  # dirty = dirty or subprocess.check_output(['sha256sum', GD_BINARY]).split()[0].decode() != GD_BINARY_SHA256
  # dirty = dirty or subprocess.check_output(['sha256sum', FF_BINARY]).split()[0].decode() != FF_BINARY_SHA256
  if not dirty:
    return


  # if os.name == 'posix':
  #   os.makedirs(FF_DIR, exist_ok=True)
  #   with tempfile.TemporaryDirectory() as d:
  #     os.makedirs(d, exist_ok=True)
  #     with tempfile.NamedTemporaryFile(prefix=d) as f:
  #       print(d)
  #       # if os.name == 'darwin':
  #       subprocess.check_call(['curl', 'https://download-installer.cdn.mozilla.net/pub/firefox/releases/83.0/mac/en-US/Firefox%2083.0.dmg',  '--output', f.name])
  #       mountpoint = os.path.join(d, "dmgmount")
  #       os.makedirs(mountpoint, exist_ok=True)
  #       try:
  #         subprocess.check_call(['hdiutil', 'attach', '-mountpoint', mountpoint, f.name])
  #         # print(subprocess.check_output(['ls', d]))
  #         subprocess.check_call(['cp', '-R', os.path.join(mountpoint, "Firefox.app"), FF_DIR])
  #       finally:
  #         subprocess.check_call(['hdiutil', 'detach', mountpoint])

  #       subprocess.check_call(['rm', f.name])
  #       subprocess.check_call(['curl',  'https://github.com/mozilla/geckodriver/releases/download/v0.28.0/geckodriver-v0.28.0-macos.tar.gz', '--output', f.name])
  #       subprocess.check_call(['tar', 'xfz', f.name], cwd=d)
  #       subprocess.check_call(['cp', os.path.join(d, "geckodriver"), FF_DIR])
  #       # else:
  #       #   response = requests.get('https://download-installer.cdn.mozilla.net/pub/firefox/releases/83.0/win64/en-US/Firefox%20Setup%2083.0.exe', stream=True, chunk_size=4096)
  #       #   with open(f, "wb") as handle:
  #       #     for data in tqdm(response.iter_content()):
  #       #         handle.write(data)

  assert os.path.exists(GD_BINARY)
  assert os.path.exists(FF_BINARY)

if __name__ == '__main__':
  get_external()
  browsers = launch_pool(1)
  print(browsers)


# TODO: Dress up
      # Integrate with cli
