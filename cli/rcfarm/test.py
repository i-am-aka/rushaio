#!/usr/bin/env python3
import os
import signal
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from farm import simple_farm
from pool import get_external, launch_pool
from mitm import start_proxy, start_proxy_fast
from util import printdbg
get_external()
WORKERS = int(os.getenv('WORKERS', os.cpu_count()))
browsers = launch_pool(WORKERS)
mprocs = []
TOKENS = int(os.getenv('TOKENS', 1e4))
PROXY = 'http://mz:mz@45.95.2.42:65339'
# 'http://andyh:hury76tr5t4e@108.61.8.190:65322'
sitekey = '6LdyC2cUAAAAACGuDKpXeDorzUDWXmdqeg-xy696'
action = 'examples/v3scores'
url = 'https://recaptcha-demo.appspot.com'
enterprise = False
t0 = 0
cnt = 0

"""
TODO

any advantage to using browser profile?
"""

with ThreadPoolExecutor(len(browsers)) as exc:
  futs = [ exc.submit(start_proxy_fast, PROXY, b['proxport']) for b in browsers]
  for fut in as_completed(futs):
    mprocs.append(fut.result())
  try:
    futs = [exc.submit(simple_farm, browsers[x % len(browsers)], sitekey, action, url, enterprise=enterprise) for x in range(TOKENS)]
    for fut in as_completed(futs):
      if t0 == 0:
        t0 = time.time()
      tok = fut.result()
      if type(tok) is str:
        cnt += 1
        printdbg(tok, time.time())
        print()
        print('TOK/SEC: ', int(cnt / (time.time() - t0)))
        print()
  finally:
    def quit(b):
      b['driver'].quit()
    for _ in exc.map(quit, browsers):
      pass
    for m in mprocs:
      try:
        os.kill(mproc.pid, signal.SIGKILL)
      except:
        pass
