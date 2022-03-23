#!/usr/bin/env python3
import os
from datetime import datetime, timezone
import random
import json
import requests
import selenium
import time
import traceback
from selenium import webdriver
from selenium.webdriver.firefox.webdriver import WebDriver
from selenium.webdriver.firefox.options import Options as FfOptions
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver import ActionChains

import threading

from rcfarm.config import POOL_JSON, DEBUG
from rcfarm.driver import get_ff_opts, attach_to_session
from rcfarm.util import printdbg
from rcfarm.pool import launch_pool
from rcfarm.mitm import start_proxy_fast

TIMEOUT = 5

import boto3
import io
from PIL import Image, ImageDraw, ExifTags, ImageColor
import secrets
def grid_scrape(browser, site_key, url, enterprise=False):
  solvid = secrets.token_hex(8)
  t0 = time.time()
  url += '/'
  driver = browser['driver']

  try:
    driver.maximize_window()
    driver.delete_all_cookies()
    recap_inited = driver.execute_script("return (typeof window.grecaptcha !== undefined)")
    driver.get(url)

    sc = r"""
    var rcbox = document.createElement('div');
    rcbox.className = 'g-recaptcha';
    rcbox.id = 'recaptcha-demo';
    rcbox.dataset.sitekey = 'SITEKEY';
    document.body.appendChild(rcbox);
    var s = document.createElement('script');
    s.src = "https://www.google.com/recaptcha/enterprise.js?hl=en";
    document.head.appendChild(s);
    s.addEventListener('load', function(e) {
      console.log('load', window.grecaptcha, e.target.src)
    });
    """.replace('SITEKEY', site_key)
    if not enterprise:
      sc = sc.replace('grecaptcha.enterprise', 'grecaptcha').replace('enterprise.js', 'api.js')
    printdbg(sc)
    driver.execute_script(sc)

    rciframe = WebDriverWait(driver, 5).until(
      EC.element_to_be_clickable((By.TAG_NAME, 'iframe')))

    driver.switch_to.frame(rciframe)
    rcbox = WebDriverWait(driver, 5).until(
      EC.element_to_be_clickable((By.CLASS_NAME, 'recaptcha-checkbox-border')))

    clickact = ActionChains(driver)
    clickact.move_to_element_with_offset(rcbox, int(random.random() * 25), int(random.random()*25))
    clickact.click()
    clickact.perform()

    driver.switch_to.default_content()

    gridframe = WebDriverWait(driver, 5).until(
      EC.element_to_be_clickable((By.XPATH, "//iframe[@title='recaptcha challenge']")))

    driver.switch_to.frame(gridframe)
    # rciframe = WebDriverWait(driver, 5).until(
    #   EC.element_to_be_clickable((By.TAG_NAME, 'iframe')))
    # print(rciframe)
    # driver.switch_to.frame(rciframe)

      # driver.switch_to.frame(rciframe)
    time.sleep(30033)
    last=time.time()
    for x in range(int(1e4)):
#      for img in os.listdir('cap-{solvid}')
      try:
        reloadbtn = WebDriverWait(driver, 5).until(
          EC.element_to_be_clickable((By.CLASS_NAME, 'reload-button-holder')))
        last=time.time()
        reloadact = ActionChains(driver)
        reloadact.move_to_element_with_offset(reloadbtn, int(random.random() * 15), int(random.random()*15))
        reloadact.click()
        reloadact.perform()

        print(threading.get_native_id(), datetime.now(timezone.utc).isoformat())

        driver.delete_all_cookies()
        driver.execute_script('localStorage.clear();')
        driver.execute_script('sessionStorage.clear();')
      except selenium.common.exceptions.StaleElementReferenceException:
        print('stale', time.time()-last)
        pass

      time.sleep(random.random())

  except:
    traceback.print_exc()
  finally:
    try:
      driver.switch_to.default_content()
      seq = driver.find_elements_by_tag_name('iframe')
      for idx in range(len(seq)):
        driver.switch_to.default_content()
        try:
          iframe = seq[idx]
          printdbg(iframe)
          driver.switch_to.frame(iframe)
          # driver.implicitly_wait(1)
          driver.delete_all_cookies()
          driver.execute_script('localStorage.clear();')
          driver.execute_script('sessionStorage.clear();')
        except:
          continue
    finally:
      try:
        driver.switch_to.default_content()
        pass
        # driver.quit()
      except:
        pass
      finally:
        pass
        #driver.quit()
        #driver.quit()

if __name__ == '__main__':
  browsers = list(launch_pool(1))
  start_proxy_fast(os.getenv('PROXY').replace('-proxy ',''), browsers[0]['proxport'])
  grid_scrape(browsers[0], '6LfW6wATAAAAAHLqO2pb8bDBahxlMxNdo9g947u9', 'https://recaptcha-demo.appspot.com/recaptcha-v2-checkbox.php')
  time.sleep(60)

  # with open(POOL_JSON) as f:
  #   browsers = json.load(f)
  #   for x in range(4):
  #     print(simple_farm(browsers[0], '6Ld_R8wZAAAAADxLG5Hzi2HyFyVGLDuEcGyoVV5M', 'home_verif', 'https://rushaio.co'))
