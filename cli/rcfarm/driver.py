#!/usr/bin/env python3

import os
import time
import random
import selenium
import traceback
from selenium import webdriver
from selenium.webdriver.firefox.webdriver import WebDriver
from selenium.webdriver.firefox.options import Options as FfOptions
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from rcfarm.util import find_free_port

DEBUG =  os.getenv('DEBUG') == '1'
FFDEBUG = os.getenv('FFDEBUG') == '1'
PROXY_HOST = os.getenv('RCFARM_PROXY_HOST', '127.0.0.1')

def get_ff_opts(proxport):
  opts = FfOptions()
  opts.log.level = "trace" if FFDEBUG else "info"
  opts.set_preference('security.mixed_content.use_hstsc', False)
  opts.set_preference('network.stricttransportsecurity.preloadlist', False)
  opts.set_preference('network.proxy.type', 1)
  opts.set_preference('extensions.getAddons.cache.enabled', False)
  # if os.name == 'posix':
  #   opts.set_preference('browser.cache.disk.parent_directory', '/tmp') # TODO windows/static location across calls for each browser
  opts.set_preference("browser.cache.disk.enable", True)
  opts.set_preference("browser.cache.memory.enable", True)
  opts.set_preference("browser.cache.offline.enable", True)
  opts.set_preference("network.http.use-cache", True)
  opts.set_preference("toolkit.startup.max_resumed_crashes", "-1")
  opts.set_preference('browser.startup.homepage_override.mstone', 'ignore')
  opts.set_preference('browser.search.geoip.url', '')
  opts.set_preference('network.http.speculative-parallel-limit', 0)
  opts.set_preference('network.dns.disablePrefetch', True)
  opts.set_preference('network.prefetch-next', False)
  opts.set_preference('OCSP.enabled', False)
  opts.set_preference('security.ssl.enable_ocsp_stapling', False)
  opts.set_preference('app.normandy.enabled', False)
  opts.set_preference('browser.safebrowsing.downloads.remote.enabled', False)
  opts.set_preference('extensions.blocklist.enabled', False)
  opts.set_preference('network.captive-portal-service.enabled', False)
  opts.set_preference('browser.aboutHomeSnippets.updateUrl', '')
  opts.set_preference('network.proxy.http', PROXY_HOST)
  opts.set_preference('network.proxy.http_port', int(proxport))
  opts.set_preference('network.proxy.ssl', PROXY_HOST)
  opts.set_preference('network.proxy.ssl_port', int(proxport))
  # if os.name == 'posix' and FFDEBUG:
  #   opts.set_preference('webdriver.firefox.logfile', f'/tmp/log-{proxport}')
  #   opts.set_preference('webdriver.log.file', f'/tmp/log2-{proxport}')
  opts.set_preference('devtools.chrome.enabled', True)
  # opts.set_preference('dom.webdriver.enabled', False)
  # opts.set_preference('useAutomationExtension', False)
  opts.set_preference('webdriver.load.strategy', 'unstable')
  opts.set_preference('privacy.trackingprotection.enabled', True)
  opts.set_preference('privacy.trackingprotection.annotate_channels', True)
  opts.set_preference('privacy.trackingprotection.cryptomining.enabled', True)
  opts.set_preference('privacy.trackingprotection.fingerprinting.enabled', True)
  return opts

def start_browser(ext_binaries):
  try:
    ffpath, geckodriverpath = ext_binaries

    time.sleep(random.random())
    proxport = find_free_port()
    opts = get_ff_opts(proxport)

    caps = DesiredCapabilities.FIREFOX
    caps['acceptInsecureCerts'] = True
    if os.getenv('HEADFUL') != '1':
      os.environ['MOZ_HEADLESS'] = '1'
      os.environ['MOZ_HEADLESS_WIDTH'] = '1366'
      os.environ['MOZ_HEADLESS_HEIGHT'] = '768'
    kwargs = {}
    if os.name == 'posix':
      pass
      # kwargs['log_path'] =f'/tmp/log3-{proxport}'
      kwargs['service_log_path'] = os.path.devnull
    else:
      kwargs['service_log_path'] = os.path.join(os.path.dirname(geckodriverpath), 'geckodriver.log')
    driver = webdriver.Firefox(firefox_binary=ffpath, executable_path=geckodriverpath, desired_capabilities=caps, options=opts, **kwargs)
    time.sleep(1)
    driver.get('about:blank')

    return {'driver': driver, 'proxport': proxport, 'exec_url': driver.command_executor._url, 'session_id': driver.session_id}
  except:
    traceback.print_exc()
    raise


def attach_to_session(executor_url, session_id, caps, **kwargs):
  original_execute = WebDriver.execute
  def new_command_execute(self, command, params=None):
      if command == "newSession":
          # Mock the response
          return {'success': 0, 'value': None, 'sessionId': session_id}
      else:
          return original_execute(self, command, params)
  # Patch the function before creating the driver object
  WebDriver.execute = new_command_execute
  driver = webdriver.Remote(command_executor=executor_url, desired_capabilities=caps, **kwargs)
  driver.session_id = session_id
  # Replace the patched function with original function
  WebDriver.execute = original_execute
  return driver

if __name__ == '__main__':
  print(start_browser())
