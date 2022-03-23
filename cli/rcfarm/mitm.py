import subprocess
import os
import sys
import requests
import time
from urllib.parse import urlparse
from rcfarm.config import DEBUG

if hasattr(sys, "_MEIPASS"):
  DIR = sys._MEIPASS
else:
  DIR = os.path.dirname(sys.argv[0])

MITMDEBUG = os.getenv('MITMDEBUG') == '1'

def start_proxy(upstream, proxport):
  scfn = os.path.join(DIR, 'mitm', 'base.py')
  pp = urlparse(upstream)

  mpargs = (f'mitmdump -s {scfn} --flow-detail 3 --listen-host 0.0.0.0 -p {proxport} -k -v --set termlog_verbosity=debug --set console_eventlog_verbosity=debug --set upstream_cert=false --set ssl_insecure=true --set allow_remote=true').split()
  if upstream:
    mpargs = mpargs[:3] + (f'--mode upstream:{pp.hostname}:{pp.port} --upstream-auth {pp.username}:{pp.password}'.split()) + mpargs[3:]
  mproc = subprocess.Popen(mpargs, stdout=sys.stderr if MITMDEBUG else subprocess.DEVNULL, stderr=sys.stderr if MITMDEBUG else subprocess.DEVNULL)
  t0 = time.time()
  while 1:
    if time.time() - t0 > 10:
      raise Exception(f'failed to init proxy 127.0.0.1:{proxport}')
    try:
      requests.get(f'http://127.0.0.1:{proxport}', timeout=0.1)
      break
    except:
      time.sleep(0.01)
  return mproc

def start_proxy_fast(upstream, proxport):
  if DEBUG and os.name != 'nt'  :
    mpargs = ['go', 'run', '.', '-upstream', upstream or '', '-proxport', str(proxport)]
  elif os.name == 'nt':
    mpargs = [os.path.join(DIR, 'mitm', 'mitm.exe'), '-upstream', upstream or '', '-proxport', str(proxport)]
  else:
    mpargs = [os.path.join(DIR, 'mitm', 'mitm'), '-upstream', upstream or '', '-proxport', str(proxport)]
  mproc = subprocess.Popen(mpargs, stdout=sys.stderr if MITMDEBUG else subprocess.DEVNULL, stderr=sys.stderr if MITMDEBUG else subprocess.DEVNULL)
  t0 = time.time()
  while 1:
    if time.time() - t0 > 10:
      raise Exception(f'failed to init proxy 127.0.0.1:{proxport}')
    try:
      requests.get(f'http://127.0.0.1:{proxport}', timeout=0.1)
      break
    except:
      time.sleep(0.01)
  return mproc