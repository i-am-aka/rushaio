#!/usr/bin/env python3
import json
import random
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
from selenium.webdriver.common.touch_actions import TouchActions
from PIL import Image
from io import BytesIO
from PIL import ImageChops

import base64
from rcfarm.config import POOL_JSON, DEBUG
from rcfarm.driver import get_ff_opts, attach_to_session
from rcfarm.util import printdbg

TIMEOUT = 6

from selenium.webdriver.common.actions.pointer_input import PointerInput
origptrmov = PointerInput.create_pointer_move

def geets_farm(browser, url, cid):
  driver = browser['driver']
  driver.maximize_window()
  driver.delete_all_cookies()
  driver.get(url)
  box = WebDriverWait(driver, 10).until(
    EC.element_to_be_clickable((By.CLASS_NAME, 'geetest_radar_btn')))

  time.sleep(1)

  clickact = ActionChains(driver)
  clickact.move_to_element_with_offset(box, int(random.random() * 25), int(random.random()*25))
  clickact.click()
  clickact.perform()


  silo = WebDriverWait(driver, 10).until(
    EC.visibility_of_element_located((By.CLASS_NAME, 'geetest_canvas_bg'))
  )
  btn = WebDriverWait(driver, 10).until(
    EC.visibility_of_element_located((By.CLASS_NAME, 'geetest_slider_button'))
  )

  png_silo = s = driver.execute_script('''
  return document.getElementsByClassName('geetest_canvas_bg geetest_absolute')[0].toDataURL('image/png');
  ''').replace('data:image/png;base64,', '')
  t0 = time.time()
  while s == png_silo and (time.time() - t0) < 1:
    png_silo = driver.execute_script('''
  return document.getElementsByClassName('geetest_canvas_bg geetest_absolute')[0].toDataURL('image/png');
  ''').replace('data:image/png;base64,', '')

  png_nosilo = driver.execute_script('''
  return document.getElementsByClassName('geetest_canvas_fullbg geetest_fade geetest_absolute')[0].toDataURL('image/png');
  ''').replace('data:image/png;base64,', '')

  print()
  print(png_silo)
  print(png_nosilo)
  print()
  im_silo = Image.open(BytesIO(base64.b64decode(png_silo))).convert('RGB')
  im_nosilo = Image.open(BytesIO(base64.b64decode(png_nosilo))).convert('RGB')

  diff = ImageChops.difference(im_silo, im_nosilo)
  dd = diff.getdata()

  bounds = [9999,9999,0,0]
  for y in range(diff.height):
    for x in range(diff.width):
      p = diff.getpixel((x,y))
      if sum(p)/len(p) > 50:
        print((x,y), sum(p)/len(p))
        if x < bounds[0]:
          bounds[0] = x
        if y < bounds[1]:
          bounds[1] = y
        if x > bounds[2]:
          bounds[2] = x
        if y > bounds[3]:
          bounds[3] = y
  print(bounds)

  startx = 20 #(bounds[2] - bounds[0])//2
  starty = bounds[1] + (bounds[3] - bounds[1])//2
  targetx = bounds[0] - 20
  print('startx', startx)
  print('starty', starty)
  print('targetx', targetx)


  # drag = TouchActions(driver)
  # silox = silo.location['x']
  # siloy = silo.location['y']
  # drag.tap_and_hold(silox + startx, siloy + starty)
  # # drag.pause(0.25)
  # for x in range(0, targetx, 10):
  #   drag.move(silox + 10*(x+1), siloy + starty + int(-10 + random.random(20)))

  # # drag.pause(0.1)
  # drag.move(silox + targetx, siloy + starty)
  # # drag.pause(0.25)
  # drag.release()
  # drag.perform()


  def fastmove(*args, **kwargs):
    return origptrmov(*args, duration=int(4*random.random()), **kwargs)
  PointerInput.create_pointer_move = fastmove

  drag = ActionChains(driver)
  drag.move_to_element_with_offset(silo, startx, starty)
  drag.click_and_hold()
  drag.perform()
  driver.implicitly_wait(1)

  x = startx
  while x < targetx + 10:
    drag = ActionChains(driver)
    dx = int(4*random.random())
    x += dx
    drag.click_and_hold()
    drag.move_by_offset(dx, 0)
    drag.perform()
    if random.random() < 0.1:
      time.sleep(0.1)
    driver.implicitly_wait(1)

  time.sleep(random.random())

  drag = ActionChains(driver)
  drag.click_and_hold()
  drag.move_by_offset(targetx - x, 0)
  drag.release()
  drag.perform()




  # for idx, x in enumerate(dd):
  #   col = idx // diff.width
  #   row = diff.height % idx if idx > 0 else 0

  #   if any(z > 5 for z in x):
  #     print(row, col, x)
  # print(dir(diff))
  # print(diff)
  # print(list(diff.getdata()))

  # print(diff.getbbox())


def simple_farm(browser, site_key, action, url, enterprise=False):
  t0 = time.time()
  url += '/'
  driver = browser['driver']

  try:
    driver.maximize_window()
    driver.delete_all_cookies()
    if enterprise:
      try:
        recap_inited = driver.execute_script("return ((typeof window.grecaptcha !== undefined) && (typeof window.grecaptcha.enterprise !== undefined))")
      except selenium.common.exceptions.JavascriptException:
        recap_inited = False
    else:
      recap_inited = driver.execute_script("return ((typeof window.grecaptcha !== undefined))")
    printdbg('CURURL', driver.current_url, url, recap_inited)
    if driver.current_url != url or not recap_inited:
      driver.get(url)
      printdbg('got', driver.current_url)
      sc = r"""
      var $ = function(){return {val: function(e){window.token=e}}};
      var FL = {pdp: {Cart: {completeAddToCart: function(){}}}};
        var s = document.createElement('script');
        s.src = "https://www.google.com/recaptcha/enterprise.js?render=SITEKEY";
        document.getElementsByTagName("head")[0].appendChild(s);
        s.addEventListener('load', function(e) {
          console.log('load', window.grecaptcha, e.target.src)
          window.grecaptcha.enterprise.ready(function(e) {
            window.grecaptcha.enterprise.execute('SITEKEY', {action: 'ACTION'}).then(function(e){$("#addToCartExecutionToken").val(e),FL.pdp.Cart.completeAddToCart()})
          });
        })
      """.replace('SITEKEY', site_key).replace('ACTION', action)
      if not enterprise:
        sc = sc.replace('grecaptcha.enterprise', 'grecaptcha').replace('enterprise.js', 'api.js')
      printdbg(sc)
      driver.execute_script(sc)
    else:
      printdbg('already in', driver.current_url)
      sc = r"""
      var $ = function(){return {val: function(e){window.token=e}}};
      var FL = {pdp: {Cart: {completeAddToCart: function(){}}}};
      try {
        window.token = window.tokenerr = undefined;
        window.grecaptcha.enterprise.execute('SITEKEY', {action: 'ACTION'})
          .then(function(e){$("#addToCartExecutionToken").val(e),FL.pdp.Cart.completeAddToCart()})
          .catch(function(e){window.tokenerr=e.toString()})
        return true;
      } catch(e) {
        return e.toString();
      }
      """.replace('SITEKEY', site_key).replace('ACTION', action)
      if not enterprise:
        sc = sc.replace('grecaptcha.enterprise', 'grecaptcha').replace('enterprise.js', 'api.js')
      printdbg(sc)
      scret = driver.execute_script(sc)
      if scret != True:
        raise Exception('Top-level script failed')


    exect0 = time.time()
    token = driver.execute_script("return window.token || window.tokenerr;")
    while token is None and (time.time() - exect0) < TIMEOUT:
      token = driver.execute_script("return window.token || window.tokenerr;")
      time.sleep(0.2)
    if token is None:
      return token

    printdbg()
    printdbg('TOKEN', token)
    printdbg()
    printdbg(f'SOLVETIME={round(time.time() - t0, 3)} EXECTIME={round(time.time() - exect0, 3)}')
    printdbg()

    if DEBUG:
      import threading
      def scorelookup():
        if 'score.rushaio.co' in url:
          resp = requests.post("https://score.rushaio.co/token", json={"token": token}).json()
          printdbg(resp)
        elif 'rushaio.co' in url:
          resp = requests.post("https://score.rushaio.co/tokenhome", json={"token": token}).json()
          del resp['remoteip']
          printdbg(resp)
        elif 'ghostaio' in url:
          resp =  requests.post("https://recaptcha-test.ghostaio.com/api/verify", json={"token": token}).json()['score']
          printdbg(resp)
        elif 'recaptcha-demo.appspot.com' in url:
          resp = requests.post('https://recaptcha-demo.appspot.com/recaptcha-v3-verify.php?action=examples/v3scores&token=' + token).json()['score']
          printdbg(resp)
      t=threading.Thread(target=scorelookup)
      t.start()

    return token
  except:
    printdbg(traceback.format_exc())
  finally:
    try:
      driver.execute_script('window.token = undefined; window.tokenerr = undefined;')
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

if __name__ == '__main__':
  pass
  # with open(POOL_JSON) as f:
  #   browsers = json.load(f)
  #   for x in range(4):
  #     print(simple_farm(browsers[0], '6Ld_R8wZAAAAADxLG5Hzi2HyFyVGLDuEcGyoVV5M', 'home_verif', 'https://rushaio.co'))