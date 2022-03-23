#!/usr/bin/env python
import string
import os
import re
import zipfile
import secrets
import signal
import psutil
import tempfile
import atexit
import heapq
import sentry_sdk

import itertools
import pathlib
import sys

class Unbuffered(object):
   def __init__(self, stream):
       self.stream = stream
   def write(self, data):
       self.stream.write(data)
       self.stream.flush()
   def writelines(self, datas):
       self.stream.writelines(datas)
       self.stream.flush()
   def __getattr__(self, attr):
       return getattr(self.stream, attr)

sys.stdout.reconfigure(encoding='utf-8')
sys.stdout = Unbuffered(sys.stdout)

if os.name == 'nt':
  import win_unicode_console
  win_unicode_console.enable()
def start_rcfarm():


  if DEBUG:
    try:
      sys.stdout = sys.stderr = open('/Users/aka/drive/ysfc/cli/debug.log', 'a+')
    except:
      sys.stdout = sys.stderr = open('/dev/null', 'w')
  else:
    sys.stdout = sys.stderr = open('/dev/null', 'w')

  sys.stdin.close()


  start()
import csv
from enum import Enum
from cmd2 import ansi
import uuid
import threading
from threading import Thread
from datetime import datetime
from collections import defaultdict
# import timeago
import pprint
import subprocess
import math
import json
import requests
import tableformatter as tf
tf.TableColors.set_color_library('none')
import sqlite3
import sys
import traceback
import os
from urllib.parse import urlparse, parse_qs, urlunparse, urlencode
import random
import time
import io
from itertools import groupby
import concurrent
import argparse
import cmd2
import appdirs
from util import cc_type
from messaging import get_task_pub_socket, get_task_tail_socket

if hasattr(sys, "_MEIPASS"):
  DIR = sys._MEIPASS
else:
  DIR = os.path.dirname(sys.argv[0])

VERSION = '0.2.44'
DEBUG = os.getenv('DEBUG') is not None
if DEBUG:
  import logging
  logging.basicConfig(level=logging.DEBUG)

RC_POOL_EXE = 'rcpool' if os.name != 'nt' else 'rcpool.exe'
RC_LOGIN_EXE = 'rclogin' if os.name != 'nt' else 'rclogin.exe'
TASK_CONSUMER_EXE = 'task_consumer' if os.name != 'nt' else 'task_consumer-windows-8.1-amd64.exe'
APP_DIR = appdirs.user_data_dir("RushAIO", "RushAIO")
PIDFILE = os.path.join(APP_DIR, "tcpid")
os.makedirs(os.path.join(APP_DIR, 'chrome'), exist_ok=True)
os.makedirs(os.path.join(APP_DIR, 'config'), exist_ok=True)
RECAP_PROXIES_FN = os.path.join(APP_DIR, 'recap_proxies.json')
ZONEINFO_ZIP = os.path.join(DIR, 'external', 'zoneinfo.zip')

if os.name == "nt":
  CHROMIUM = os.path.join(APP_DIR, 'chrome', 'chrome.exe')
  CHROMIUM_ZIP = os.path.join(DIR, 'chrome', 'chromium_win.zip')
else:
  CHROMIUM = os.path.join(APP_DIR, 'chrome', 'Chromium.app/Contents/MacOS/Chromium')
  CHROMIUM_ZIP = os.path.join(DIR, 'chrome', 'chromium_mac.zip')

RUSH_DB = os.getenv('RUSH_DB', os.path.join(APP_DIR, 'rush.db'))

PROFILE_FIELDS=["profile_name", "email","ship_name","ship_phone","ship_address1","ship_address2","ship_city","ship_zip","ship_state","ship_country","card_name","card_address1","card_address2","card_city","card_zip","card_state","card_country","card_phone","card_number","card_cvc","card_expmonth","card_expyear","size"]
class TaskStatus(Enum):
  STOPPED = 'Stopped'
  STARTING = 'Starting'
  RUSHING = 'Rushing'
  FAIL = 'Failed'
  COOKED = 'Cooked'

class ProxyDistribution(Enum):
  RANDOM = 1
  SPECIFIC = 2

from zipfile import ZipFile, ZipInfo

class ZipFileWithPermissions(ZipFile):
  """ Custom ZipFile class handling file permissions. """
  def _extract_member(self, member, targetpath, pwd):
    if not isinstance(member, ZipInfo):
      member = self.getinfo(member)

    targetpath = super()._extract_member(member, targetpath, pwd)

    attr = member.external_attr >> 16
    if attr != 0:
      os.chmod(targetpath, attr)
    return targetpath

def row_to_dict(cursor, row):
  d = {}
  for idx, col in enumerate(cursor.description):
    d[col[0]] = row[idx]
  return d

def read_profiles_from_file(fn):
  try:
    with open(os.path.expanduser(fn), newline='', encoding='utf-8') as f:
      return read_profiles_from_buf(f)
  except Exception as e:
    with open(os.path.expanduser(fn), newline='', encoding='utf-8-sig') as f:
      return read_profiles_from_buf(f)


def read_profiles_from_buf(buf):
  r = csv.DictReader(buf)
  profiles = []

  for row in r:
    header = list(row.keys())
    if header != PROFILE_FIELDS:
      raise Exception(f'Invalid profile format: incorrect column names. Extra: {set(header) - set(PROFILE_FIELDS)} Missing: {set(PROFILE_FIELDS) - set(header)}')

    try:
      ship = {}
      bill = {}
      card = {}
      profile = {'Name': row['profile_name'], 'ShippingAddress': ship, 'BillingAddress': bill, 'Card': card, 'Size': None}
      profile['Email'] = row['email'].replace(' ', '')
      ship['address1'] = row['ship_address1']
      ship['address2'] = row['ship_address2']
      ship['city'] = row['ship_city']
      ship['zipcode'] = row['ship_zip']
      ship['stateCode'] = row['ship_state'].upper()
      if len(ship['stateCode']) != 2:
        raise Exception(f'Invalid ship_state: must be 2-letter state code, not "{ship["stateCode"]}"')
      ship['country'] = row['ship_country'].replace('United States', 'US').replace('USA', 'US').upper()
      if len(ship['country']) != 2:
        raise Exception(f'Invalid ship_country: must be 2-letter country code, not "{ship["country"]}"')


      ship['firstName'] = ' '.join(row['ship_name'].split()[:-1])
      ship['lastName'] = row['ship_name'].split()[-1]
      ship['phoneNumber'] = row['ship_phone']
      if len(ship['phoneNumber']) < 10:
        # todo nice list of required fields + errfmt
        raise Exception('Missing ship_phone (should be at least 10 digits)')

      if row['card_address1']:
        bill['address1'] = row.get('card_address1')
        bill['address2'] = row.get('card_address2')
        bill['city'] = row.get('card_city')
        bill['zipcode'] = row.get('card_zip')
        bill['stateCode'] = row.get('card_state')
        bill['country'] = row.get('card_country')
        bill['country'] = bill['country'].replace('United States', 'US').replace('USA', 'US')
        bill['firstName'] = ' '.join(row['ship_name'].split()[:-1])
        bill['lastName'] = row.get('card_name').split()[-1]
        bill['phoneNumber'] = row.get('card_phone')
      else:
        bill['address1'] = row.get('card_address1') or row['ship_address1']
        bill['address2'] = row.get('card_address2') or row['ship_address2']
        bill['city'] = row.get('card_city') or row['ship_city']
        bill['zipcode'] = row.get('card_zip') or row['ship_zip']
        bill['stateCode'] = row.get('card_state') or row['ship_state']
        bill['country'] = row.get('card_country') or row['ship_country']
        bill['country'] = bill['country'].replace('United States', 'US').replace('USA', 'US')
        bill['firstName'] = ' '.join((row.get('card_name') or row['ship_name']).split()[:-1])
        bill['lastName'] = (row.get('card_name') or row['ship_name']).split()[-1]
        bill['phoneNumber'] = row.get('card_phone') or row['ship_phone']

      if len(bill['stateCode']) != 2:
        raise Exception(f'Invalid card_state: must be 2-letter state code, not "{bill["stateCode"]}"')
      if len(bill['country']) != 2:
        raise Exception(f'Invalid card_country: must be 2-letter country code, not "{bill["country"]}"')
      if len(bill['phoneNumber']) < 10:
        # todo nice list of required fields + errfmt
        raise Exception('Invalid card_phone (should be at least 10 digits)')

      card['number'] = row['card_number']
      if len(card['number']) < 13 or not card['number'].isdigit():
        raise Exception(f'Invalid profile format: card number is {card["number"]}')
      card['cvc'] = row['card_cvc']
      card['cardType'] = cc_type(card['number'])
      if card['cardType'] is None:
        raise Exception(f'Could not determine card type (visa,master,amex,discover,...). Check card number: {card["number"]}')
      card['holder'] = row.get('card_name') or row.get('ship_name')
      card['expirationMonth'] = int(row['card_expmonth'])
      card['expirationYear'] = int(row['card_expyear'])
      card['paymentMethodId'] = "CREDIT_CARD"
      card['lastFour'] = ''.join(card['number'][-4:])

      profile['Size'] = row.get('size')

      for key, value in profile.items():
        if type(value) is str:
          profile[key] = value.strip()

      profiles.append(profile)
    except Exception as e:
      print('Warning: skipping profile, bad format', traceback.format_exc(1), row)

      continue

  return profiles

def profiles_to_csv(profs):
  b = io.StringIO()
  w = csv.DictWriter(b, fieldnames=PROFILE_FIELDS)

  w.writeheader()
  for p in profs:
    ship = p['ShippingAddress']
    bill = p['BillingAddress']
    card = p['Card']
    pp = {
      "profile_name": p.get('Name',  str(uuid.uuid4())[:8]),
      "email": p['Email'],
      "ship_name": ' '.join([ship['firstName'], ship['lastName']]),
      "ship_phone": ship['phoneNumber'],
      "ship_address1": ship['address1'],
      "ship_address2": ship['address2'],
      "ship_city": ship['city'],
      "ship_zip": ship['zipcode'],
      "ship_state": ship['stateCode'],
      "ship_country": ship['country'],
      "card_name": card['holder'],
      "card_phone": bill['phoneNumber'],
      "card_address1": bill['address1'],
      "card_address2": bill['address2'],
      "card_city": bill['city'],
      "card_zip": bill['zipcode'],
      "card_state": bill['stateCode'],
      "card_country": bill['country'],
      "card_number": card['number'],
      "card_cvc": card['cvc'],
      "card_expmonth": card['expirationMonth'],
      "card_expyear": card['expirationYear'],
      "size": p.get('Size', ""),
    }
    w.writerow(pp)
  return b.getvalue()

def validate_profile(p):
  pass

SIZES = [
    "3K",
    "4K",
    "4.5K",
    "5K",
    "5.5K",
    "6K",
    "6.5K",
    "7K",
    "7.5K",
    "8K",
    "8.5K",
    "9K",
    "9.5K",
    "10K",
    "10.5K",
    "11K",
    "11.5K",
    "12K",
    "12.5K",
    "13K",
    "13.5K",
    "1",
    "1.5",
    "2",
    "2.5",
    "3",
    "3.5",
    "4",
    "4.5",
    "5",
    "5.5",
    "6",
    "6.5",
    "7",
    "7.5",
    "8",
    "8.5",
    "9",
    "9.5",
    "10",
    "10.5",
    "11",
    "11.5",
    "12",
    "12.5",
    "13",
    "13.5",
    "14",
    "14.5",
    "15",
    "15.5",
    "16",
    "16.5",
    "17",
]

class ResourceShare:
  def __init__(self, key=lambda resource: resource):
    self.pq = []
    self.entry_finder = {}
    self.REMOVED = '<removed-resource>'
    self.counter = itertools.count()
    self.key_fn = key
    """
    trie of profile names +
    profile name -> profile dict +

    """

  def add(self, resource, priority=0):
      'Add a new resource or update the priority of an existing resource'
      if self.key_fn(resource) in self.entry_finder:
          self.remove(resource)
      count = next(self.counter)
      entry = [priority, count, resource]
      self.entry_finder[self.key_fn(resource)] = entry
      heapq.heappush(self.pq, entry)

  def has(self, resource):
    return self.key_fn(resource) in self.entry_finder

  def remove(self, resource):
      'Mark an existing resource as REMOVED.  Raise KeyError if not found.'
      entry = self.entry_finder.pop(self.key_fn(resource))
      entry[-1] = self.REMOVED

  def pop(self):
    'Remove and return the lowest priority resource. Raise KeyError if empty.'
    while self.pq:
      priority, count, resource = heapq.heappop(self.pq)
      if resource != self.REMOVED:
        del self.entry_finder[self.key_fn(resource)]
        return resource
    raise KeyError('pop from an empty priority queue')

  def popincr(self, testfn=lambda x: True):
    #
    s = time.time()
    cnt = 0
    to_re_add = []
    ll = len(self.pq)
    while self.pq and cnt <= ll:
      cnt += 1
      priority, count, resource = heapq.heappop(self.pq)
      if resource != self.REMOVED:
        del self.entry_finder[self.key_fn(resource)]
        if testfn(resource):
          # print('tested positive', resource)
          self.add(resource, priority=priority+1)
          for pri, rr in to_re_add:
            self.add(rr, priority=pri)
          return resource
        else:
          to_re_add.append((priority, resource))
          # self.add(resource, priority=priority)
    for pri, resource in to_re_add:
      self.add(resource, priority=pri)
    raise KeyError('pop from an empty priority queue')

    # pqbak = self.pq.copy()

    # cnt = 0
    # ll = len(self.pq)
    # while self.pq and cnt <= ll:
    #   cnt += 1
    #   priority, count, resource = heapq.heappop(self.pq)
    #   if resource != self.REMOVED:
    #     if testfn(resource):
    #       self.pq = pqbak
    #       self.add(resource, priority=priority+1)
    #       return resource
    # self.pq = pqbak
    # raise KeyError('pop from an empty priority queue')

  def decr(self, resource):
    if resource in self.entry_finder:
      priority, count, resource = self.entry_finder[self.key_fn(resource)]
      self.add(resource, priority=max(0,priority-1))

  def all_priority(self):
    h = self.pq.copy()
    hp = [heapq.heappop(h) for i in range(len(h))]
    return [(h[0], h[2]) for h in hp if h[2] != self.REMOVED]

  def all(self):
    h = self.pq.copy()
    hp = [heapq.heappop(h) for i in range(len(h))]
    return [h[2] for h in hp if h[2] != self.REMOVED]
    # for _, __, resource in self.pq.copy():
    #   if resource is not self.REMOVED:

def unescape_backslash(str_):
  if os.name == 'posix':
    return str_.replace('\\', '')
  else:
    return str_

parser = argparse.ArgumentParser()
subp = parser.add_subparsers(title='rush', help='rush help')

key_parser = subp.add_parser('key')
key_cmds = key_parser.add_subparsers(title='license key commands', help='Activate License')
key_activate = key_cmds.add_parser('activate')
key_activate.add_argument('key')

prof_parser = subp.add_parser('profile', help='profile help')
prof_cmds = prof_parser.add_subparsers(title='profile commands', help='Import profiles')

prof_imp = prof_cmds.add_parser('import')
prof_imp.add_argument('csv_file', nargs='+')

prof_remove = prof_cmds.add_parser('remove')
prof_remove.add_argument('profile_name')

prof_remove_all = prof_cmds.add_parser('remove-all')
prof_remove_all.add_argument('--name')

prof_list = prof_cmds.add_parser('list')

prof_show = prof_cmds.add_parser('show')
prof_show.add_argument('name')

proxy_parser = subp.add_parser('proxy', help='proxy help')
proxy_cmds = proxy_parser.add_subparsers(title='proxy commands')

proxy_imp = proxy_cmds.add_parser('import')
proxy_imp.add_argument('proxy_txt_file', nargs='+')
proxy_imp.add_argument('--group', default='default')
proxy_add = proxy_cmds.add_parser('add')
proxy_add.add_argument('proxy')
proxy_add.add_argument('--group', default='default')
proxy_list = proxy_cmds.add_parser('list')
proxy_list.add_argument('--group')
proxy_remove = proxy_cmds.add_parser('remove')
proxy_remove.add_argument('proxy')
proxy_remove.add_argument('--group')
proxy_remove_all_p = proxy_cmds.add_parser('remove-all')
proxy_remove_all_p.add_argument('--group')
# TODO
# proxy_test = proxy_cmds.add_parser('test')
# proxy_test.add_argument('proxy')

vault_parser = subp.add_parser('vault', help='vault help')
vault_cmds = vault_parser.add_subparsers(title='vault commands')

vault_add = vault_cmds.add_parser('add')
vault_add.add_argument('site')
vault_add.add_argument('email')
vault_add.add_argument('password')

vault_remove = vault_cmds.add_parser('remove')
vault_remove.add_argument('site')
vault_remove.add_argument('email')

vault_list = vault_cmds.add_parser('list')

task_parser = subp.add_parser('task', help='task')
task_cmds = task_parser.add_subparsers(title='task commands')

task_tail = task_cmds.add_parser('tail')

task_list = task_cmds.add_parser('list')
task_list.add_argument('--status', help='Task status', nargs='+', choices=[t.value for t in list(TaskStatus)])

task_status = task_cmds.add_parser('status')
task_status.add_argument('--watch', action='store_true', default=False)

task_watch = task_cmds.add_parser('watch')
task_watch.add_argument('--watch', action='store_true', default=True)

task_data = task_cmds.add_parser('data')
task_data.add_argument('--watch', action='store_true', default=True)

task_edit = task_cmds.add_parser('edit')
task_edit.add_argument('--url')
task_edit.add_argument('--new-url')


task_add = task_cmds.add_parser('add')
task_add.add_argument('--url',  help='URL of product or URL of page for keyword-based product lookup', required=True)
task_add.add_argument('--profile', help='(Optional) Profile name. Defaults to random profile (recommended for task count > 1), prioritizing profiles assigned to fewest tasks')
task_add.add_argument('--size', help=f'(Optional) Size (overrides profile size) ({",".join(SIZES)})')
task_add.add_argument('--count', default=1, type=int, help='(Optional, default 1) Task Count. If set to >1, uses random profile, overriding --profile option')
task_add_prox = task_add.add_mutually_exclusive_group()
task_add_prox.add_argument('--proxy', help='Proxy string. Defaults to random proxy. Specify "none" to disable proxy')
task_add_prox.add_argument('--proxy-group', help='Proxy group name. Defaults to "default"', default='default')
task_add.add_argument('--keywords', nargs='+', help='Keywords that must exist in product name')
task_add.add_argument('--start', action='store_true', default=False)
task_add.add_argument('--delay', type=int)

task_start = task_cmds.add_parser('start')
task_start.add_argument('task_id', nargs='+')

task_start_all = task_cmds.add_parser('start-all')
task_start_all.add_argument('--site')
task_start_all.add_argument('--url')
task_start_all.add_argument('--limit', type=int)
task_start_all.add_argument('--profile')
task_start_all.add_argument('--proxy-group')

task_stop = task_cmds.add_parser('stop')
task_stop.add_argument('task_id', nargs='+')

task_stop_all = task_cmds.add_parser('stop-all')
task_stop_all.add_argument('--state')
task_stop_all.add_argument('--status')
task_stop_all.add_argument('--site')
task_stop_all.add_argument('--limit', type=int)
task_stop_all.add_argument('--profile')
task_stop_all.add_argument('--proxy-group')

task_remove = task_cmds.add_parser('remove')
task_remove.add_argument('task_id')

task_remove_all = task_cmds.add_parser('remove-all')
task_remove_all.add_argument('--site')
task_remove_all.add_argument('--status')
task_remove_all.add_argument('--profile')
task_remove_all.add_argument('--proxy-group')

# task_success = task_cmds.add_parser('success')

# recap_parser = subp.add_parser('recap', help='recap')
# recap_cmds = recap_parser.add_subparsers(title='Recap Commands')

# recap_proxy = recap_cmds.add_parser('proxy')
# recap_proxy_cmds = recap_proxy.add_subparsers(title='recap proxy commands')
# recap_proxy_imp = recap_proxy_cmds.add_parser('import')
# recap_proxy_imp.add_argument('proxy_txt_file', nargs='+')
# recap_proxy_add = recap_proxy_cmds.add_parser('add')
# recap_proxy_add.add_argument('proxy')
# recap_proxy_list = recap_proxy_cmds.add_parser('list')
# recap_proxy_remove = recap_proxy_cmds.add_parser('remove')
# recap_proxy_remove.add_argument('proxy')
# recap_proxy_remove_all_p = recap_proxy_cmds.add_parser('remove-all')

# recap_proxy_litport_rotate = recap_proxy_cmds.add_parser('litport-rotate')
# recap_proxy_litport_rotate.add_argument('--url', required=True)
# recap_proxy_litport_rotate.add_argument('--count', required=True)


# recap_login = recap_cmds.add_parser('login')
# recap_login.add_argument('--proxy')
# recap_logout = recap_cmds.add_parser('logout')
# recap_logout.add_argument('email')

# recap_logout_all = recap_cmds.add_parser('logout-all')

# recap_list = recap_cmds.add_parser('list')
# recap_list_logins = recap_cmds.add_parser('list-logins')

# recap_score = recap_cmds.add_parser('score')
# recap_score.add_argument('--proxy')
# recap_score.add_argument('--email')
# recap_score.add_argument('--loop', action='store_true', default=False)
# recap_score.add_argument('--delay', type=int, default=0)
# recap_score.add_argument('--sources', nargs='+')

config_parser = subp.add_parser('config', help='RushAIO configuration')
config_cmds = config_parser.add_subparsers(title='Configuration')
config_set_webhook = config_cmds.add_parser('set-webhook')
config_set_webhook.add_argument('webhook')

def normalize_proxy(proxy):
  parsed = urlparse(proxy)
  if parsed.scheme not in ('http', 'socks5'):
    parsed = urlparse('http://' + proxy)
  hostname = parsed.hostname
  if hostname and '::' in hostname:
    hostname = '[' + hostname + ']'
  try:
    if parsed.scheme not in ('http', 'socks5'):
      raise ValueError()
    credentials = f'{parsed.username}{":"+ parsed.password if parsed.password else ""}@' if parsed.username else ''
    return f'{parsed.scheme}://{credentials}{hostname}{":" + str(parsed.port) if parsed.port else ""}'
  except ValueError:
    try:
      host, port, user, password = proxy.split(":")
      return f'http://{user}:{password}@{host}:{port}'
    except Exception:
      return None




def uniq_psort(seq, key=None):
  if key is None:
    key = lambda x: x

  seen = set()
  seen_add = seen.add
  return [x for x in seq if not (key(x) in seen or seen_add(key(x)))]

def find_dupes(seq, key=None):
  if key is None:
    key = lambda x: x

  seen = set()
  dupes = []
  for x in seq:
    if key(x) in seen:
      dupes.append(x)
    else:
      seen.add(key(x))
  return dupes

class TaskRow():
  def __init__(self, task):
    self.task = task
    for k,v in task.items():
      setattr(self, k, v)


  def __len__(self):
    return len(self.task.keys())

  @property
  def level_(self):
    return int(self._level_) if hasattr(self, '_level_') and (type(self._level_) is int or (type(self._level_) is str and self._level_.isdigit())) else 0

  def get_product(self):
    return self.product if hasattr(self, 'product') else '/'.join([s for s in self.url.split('/')[-2:] if s not in ('product', 'products')])

  def get_site(self):
    return self.url
    # up = urlparse(self.url)
    # return up.scheme + "://" + up.hostname

  def get_size(self):
    has_size_override = self.size not in ("",None)
    if has_size_override:
      return self.size
    elif self.size == "":
      return "Random"
    else:
      profile = self.profile or getattr(self, 'profileInUse', None)
      if type(profile) is str:
        try:
          profile = json.loads(profile)
        except Exception:
          pass

      if profile:
        try:
          return profile['Size']
        except Exception:
          return f'Random'
      elif getattr(self, 'profileInUse', None):
        return self.profileInUse['Size'] or 'Random'
      else:
        return 'Random'

  def get_proxy(self):
    proxy_distribution = getattr(self, 'proxy_distribution', ProxyDistribution.RANDOM.value if self.proxy is None else ProxyDistribution.SPECIFIC.value)
    proxy_group = getattr(self, 'proxy_group', '') or ''
    if proxy_group:
      proxy_group = f'({proxy_group}) '
    active_proxy = f'({self.proxy})' if self.proxy else ''

    return f'Random {proxy_group}{active_proxy}' if proxy_distribution == ProxyDistribution.RANDOM.value else self.proxy # self.proxy or f"Random{' (' + self.proxyInUse + ')' if getattr(self, 'proxyInUse', None) else ''}"

  def get_profile_name(self):
    if self.profile:
      try:
        return json.loads(self.profile)['Name']
      except Exception:
        return f"Random ({self.profile})"
    else:
      return f"Random {'(' + self.profileInUse['Name'] + ')' if getattr(self, 'profileInUse', None) else ''}"

  # def get_last_updated(self):
  #   if getattr(self, 'updated_at', None):
  #     return timeago.format(self.updated_at, datetime.now())
  #   else:
  #     return ''

from queue import Queue
es_queue = Queue()
def es_log():
  pass


class App(cmd2.Cmd):
  prompt = "(Rush) "
  intro = "Type 'help' to explore commands"

  def __init__(self):
    cmd2.Cmd.__init__(self, allow_cli_args=False, persistent_history_file=os.path.join(APP_DIR, 'rush_history'))
    # To hide commands from displaying in the help menu, add their function name to
    # the exclude_from_help list
    # self.hidden_commands.append('py')

    # To remove built-in commands entirely, delete their "do_*" function from the
    # cmd2.Cmd class
    del cmd2.Cmd.do_alias
    del cmd2.Cmd.do_edit
    del cmd2.Cmd.do_macro
    del cmd2.Cmd.do_py
    del cmd2.Cmd.do_run_pyscript
    del cmd2.Cmd.do_run_script
    # del cmd2.Cmd.do_set
    del cmd2.Cmd.do_shell
    del cmd2.Cmd.do_shortcuts

    if DEBUG:
      self.poutput(f'on main thread: {threading.current_thread() is threading.main_thread()} ({threading.current_thread()})')

    self.stop_task_consumer()

    if DEBUG:
      self.poutput('set up task db')
    c = db_conn.cursor()
    c.execute('''DROP TABLE IF EXISTS tasks;''')
    c.execute('''CREATE TABLE IF NOT EXISTS tasks2(
      id varchar(32) primary key not null,
      url text,
      profile text,
      proxy text,
      size text,
      state text,
      status text,
      keywords text,
      status_level short);
    ''')
    c.execute('''CREATE INDEX IF NOT EXISTS tasks2_proxy_state ON tasks2(proxy, state)''')
    c.execute('''CREATE TABLE IF NOT EXISTS proxy(
      url varchar(256) primary key not null
    );''')
    c.execute('''CREATE TABLE IF NOT EXISTS profile(
      name varchar(256) primary key not null,
      blob text
    );''')
    c.execute('''UPDATE tasks2 SET state = 'Stopped';''')
    c.execute('PRAGMA case_sensitive_like=OFF;')
    # c.close()
    if DEBUG:
      self.poutput('commit')
    db_conn.commit()
    if DEBUG:
      self.poutput('alter table')

    #mig1
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE tasks2 ADD proxy_distribution int''');
      c.execute('''UPDATE tasks2 SET proxy_distribution=? WHERE length(proxy) == 0 or proxy is null;''', (ProxyDistribution.RANDOM.value,));
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    #mig2
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE proxy ADD pgroup varchar(256);''');
      c.execute('''CREATE INDEX IF NOT EXISTS proxy_pgroup ON proxy(pgroup);''')
      c.execute('''UPDATE proxy SET pgroup='default' WHERE pgroup is null OR length(pgroup) = 0;''');
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    try:
      c = db_conn.cursor()
      c.execute('''UPDATE proxy SET pgroup='default' WHERE pgroup is null OR length(pgroup) = 0;''');
    except Exception:
      if DEBUG:
        traceback.print_exc()

    #mig3
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE tasks2 ADD proxy_group varchar(256);''');
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    c = db_conn.cursor()
    c.execute('''UPDATE tasks2 SET proxy_group='default' WHERE proxy_group is null OR length(proxy_group) = 0;''');
    db_conn.commit()

    #mig4
    try:
      c = db_conn.cursor()
      c.execute('''PRAGMA table_info(proxy);''');
      row = row_to_dict(c, c.fetchone())
      if row['name'] != 'id':
        c.execute('''ALTER TABLE proxy RENAME TO proxybak;''');
        c.execute('''CREATE TABLE IF NOT EXISTS proxy(
          id INTEGER PRIMARY KEY,
          url varchar(256) not null,
          pgroup varchar(256),
          UNIQUE(pgroup,url)
        );''')
        c.execute('''INSERT INTO proxy (url,pgroup) SELECT url, pgroup from proxybak;''')
        db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    #mig5
    c = db_conn.cursor()
    c.execute('''CREATE INDEX IF NOT EXISTS proxy_pgroup ON proxy(pgroup)''')
    db_conn.commit()

    # mig6
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE tasks2 ADD profile_in_use text;''');
      db_conn.commit()
      c = db_conn.cursor()
      c.execute('''SELECT id,profile FROM tasks2;''')
      for tid, profile in c.fetchall():
        if profile and profile[0] == '{':
          try:
            pname = json.loads(profile)['Name']
            c.execute('''UPDATE tasks2 SET profile_in_use=? WHERE id = ?''', (pname, tid))
          except Exception:
            if DEBUG:
              traceback.print_exc()
            pass
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    c = db_conn.cursor()
    c.execute('''CREATE INDEX IF NOT EXISTS tasks2_prof_state ON tasks2(profile_in_use, state)''')
    c.execute('''CREATE INDEX IF NOT EXISTS tasks2_state_url_prof ON tasks2(state, url, profile_in_use, profile)''')
    db_conn.commit()

    #mig7
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE tasks2 ADD delay int;''')
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise

    #mig8
    c = db_conn.cursor()
    c.execute('''CREATE INDEX IF NOT EXISTS tasks2_state_url_prof_pgroup ON tasks2(state, url, profile_in_use, profile, proxy_group)''')
    # c.execute('''DROP INDEX tasks2_state_url_prof ON tasks2(state, url, profile_in_use, profile, proxy_group)''')
    db_conn.commit()

    #mig9
    c = db_conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS vault(
      key varchar(256) primary key not null,
      secret varchar(256)
    );''')
    db_conn.commit()

    #mig10
    #mig7
    try:
      c = db_conn.cursor()
      c.execute('''ALTER TABLE tasks2 ADD modified int;''')
      db_conn.commit()
    except sqlite3.OperationalError as e:
      if "duplicate column name" not in str(e):
        raise


    if DEBUG:
      self.poutput('import json')
    self.success_log = []
    try:
      with open(os.path.join(APP_DIR, 'success.json')) as f:
        self.success_log = json.load(f)
    except FileNotFoundError:
      pass
    except json.decoder.JSONDecodeError:
      pass

    try:
      with open(os.path.join(APP_DIR, 'profiles.json')) as f:
        for p in json.load(f):
          try:
            c = db_conn.cursor()
            c.execute("INSERT INTO profile (name, blob) VALUES (?, ?)", (p['Name'],json.dumps(p)))
            db_conn.commit()
          except sqlite3.IntegrityError:
            pass
      os.unlink(os.path.join(APP_DIR, 'profiles.json'))
    except FileNotFoundError:
      pass
    except json.decoder.JSONDecodeError:
      os.unlink(os.path.join(APP_DIR, 'profiles.json'))

    if DEBUG:
      self.poutput('import proxies')
    try:
      with open(os.path.join(APP_DIR, 'proxies.json')) as f:
        for p in json.load(f):
          try:
            c = db_conn.cursor()
            c.execute("INSERT INTO proxy (url) VALUES (?)", (p,))
            db_conn.commit()
          except sqlite3.IntegrityError:
            pass
      os.unlink(os.path.join(APP_DIR, 'proxies.json'))
    except FileNotFoundError:
      pass
    except json.decoder.JSONDecodeError:
      os.unlink(os.path.join(APP_DIR, 'proxies.json'))


    self.key = ''
    try:
      with open(os.path.join(APP_DIR, 'key')) as f:
        self.key = f.read()
    except FileNotFoundError:
      pass

    if not self.key:
      self.pwarning('UNACTIVATED - activate with `key activate`')

    self.litport_rotations = []
    self.started_task_manager = False
    try:
      self.task_socket, self.task_socket_init_evt = get_task_pub_socket()
      if os.getenv('NOCONSUMER') != '1':
        self.start_task_consumer()
    except Exception:
      if DEBUG:
        traceback.print_exc()
      self.pwarning('Warning: Read-Only Instance (Another Instance Active On This Computer)')

    # if os.name == 'posix':
    #   subprocess.call("ps aux | grep firef | grep RushAIO | awk '{print $2}' | xargs kill", shell=True)
    #   self.poutput('Launching harvesters...')

    #   from multiprocessing import Process

    #   def handler():
    #     pass

    #   self.rcfarm = Process(target=start_rcfarm)
    #   self.rcfarm.start()
      # signal.signal(signal.SIGINT, original_sigint_handler)
      # signal.signal(signal.SIGTERM, handler)
      # with
      # pid = os.fork()
      # if pid == 0:
      #   if os.getenv('RCDBG') == '1':
      #     f = open('/Users/aka/drive/ysfc/cli/debug.log', 'a+')
      #   else:
      #     f = open(os.devnull, 'w')

      #   sys.stdout = f
      #   sys.stderr = f
      #   start_rcfarm()


    self.es_thread = Thread(target=es_log, daemon=True)
    self.es_thread.start()

    # if os.name == 'nt':
    """
    Recap settings
      Login
      List
      Logout
      Set-workers
    """
    self.allow_style = ansi.STYLE_ALWAYS

    atexit.register(self.kill)

    if DEBUG:
      self.poutput('done with main')

  def kill(self):
    if DEBUG:
      self.poutput('kill')
    if os.name == 'nt':
      subprocess.call(['taskkill', '/F', '/T', '/PID', os.getpid()], creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)
      return
    else:
      try:
        os.kill(self.task_consumer_proc.pid,signal.SIGTERM)
        os.kill(self.task_consumer_proc.pid,signal.SIGINT)
        os.kill(self.task_consumer_proc.pid,signal.SIGKILL)
      except Exception:
        if DEBUG:
          traceback.print_exc()

    try:
      consumer_pid = self.task_consumer_proc.pid
      os.kill(consumer_pid,signal.SIGINT)
      os.kill(consumer_pid,signal.SIGKILL)
      self.task_consumer_proc.terminate()
      time.sleep(1)
    except Exception:
      if DEBUG:
        traceback.print_exc()
    finally:
      pass
      # try:
      #   # os.kill(-os.getpgid(os.getpid()),signal.SIGINT)
      # except:
      #   if DEBUG:
      #     traceback.print_exc()
  # def sigint_handler(self, *args, **kwargs):
  #   if self._cur_pipe_proc_reader is not None:
  #       # Pass the SIGINT to the current pipe process
  #       self._cur_pipe_proc_reader.send_sigint()

  def add_profiles(self, profiles):
    c = db_conn.cursor()
    c.execute("select name from profile;")
    orig_profnames = [r[0] for r in c.fetchall()]
    orig_len = len(orig_profnames)
    namechange_cnt = 0
    for p in profiles:
      if any(p['Name'] == name for name in orig_profnames):
        p['Name'] = (p['Name'] + '-'  + str(uuid.uuid4())[:4]).strip()
        namechange_cnt += 1
      try:
        c = db_conn.cursor()
        c.execute("INSERT INTO profile (name, blob) VALUES (?, ?)", (p['Name'], json.dumps(p)))
        db_conn.commit()
      except sqlite3.IntegrityError:
        p['Name'] = (p['Name'] + '-'  + str(uuid.uuid4())[:4]).strip()
        namechange_cnt += 1
        c = db_conn.cursor()
        c.execute("INSERT INTO profile (name, blob) VALUES (?, ?)", (p['Name'], json.dumps(p)))
        db_conn.commit()

    return namechange_cnt, len(profiles), 0

  def profile_remove_all(self, args):
    c = db_conn.cursor()
    if args.name:
      c.execute('delete from profile where name like ?;', (args.name.replace('*', '%'),))
    else:
      c.execute('delete from profile;')
    c.execute('select changes();')
    db_conn.commit()
    olen = c.fetchone()[0]
    self.poutput(f'Removed {olen} profiles')

  def ppaged(self, content, *args, length=None, **kwargs):
    if length is not None:
      content = (f"0/{length}"if length else "") + "\n" + content
      end = "\n" + (f"{length}/{length}"if length else "")
    else:
      end = ""
    if os.name == 'nt':
      self.poutput(content)
      self.poutput(end)
    else:
      super().ppaged(content, *args, **kwargs, end=end + "\n\n(press 'q' to exit)")

  def profile_list(self, args):
    c = db_conn.cursor()
    c.execute('select name, blob from profile;')
    profiles = [json.loads(p[1]) for p in c.fetchall()]

    pp = [list(csv.reader(io.StringIO(profiles_to_csv([p]))))[1] for p in profiles]
    cols = [
      tf.Column(f) for f in PROFILE_FIELDS
    ]

    table = tf.generate_table(rows=pp, columns=cols)
    self.ppaged(table, chop=True, length=len(pp))

  def add_proxies(self, proxies, group='default'):
    new = 0
    dups = 0
    for proxy in proxies:
      nproxy = normalize_proxy(proxy)
      if nproxy:
        try:
          c = db_conn.cursor()
          c.execute("INSERT INTO proxy (url, pgroup) VALUES (?, ?)", (nproxy, group))
          db_conn.commit()
          new += 1
        except sqlite3.IntegrityError:
          dups += 1
      else:
        self.pwarning(f'Warning: skipping {proxy}, invalid format')

    return new, dups

  def profile_import(self, args):
    csv_file = os.path.expanduser(unescape_backslash(' '.join(args.csv_file)))
    try:
      namechange_cnt, new_cnt, dup_cnt = self.add_profiles(read_profiles_from_file(csv_file))
      self.poutput(f'Imported {new_cnt} new profiles ' + (f'({dup_cnt} already imported)' if dup_cnt > 0 else ''))
      if namechange_cnt > 0:
        self.pwarning(f'Warning: changed {namechange_cnt} profile names due to conflicts')
    except FileNotFoundError:
      self.perror(f'Error: {os.path.expanduser(csv_file)} not found')

  def proxy_import(self, args):
    proxy_txt_file = os.path.expanduser(unescape_backslash(' '.join(args.proxy_txt_file)))
    try:
      with open(proxy_txt_file) as f:
        lines = [l.strip() for l in f.read().split('\n') if l.strip()]
        lines = [''.join(filter(lambda x: x in string.printable, l)) for l in lines]
        new_cnt, dup_cnt = self.add_proxies(lines, group=args.group)
        self.poutput(f'Imported {new_cnt} new proxies ' + (f'({dup_cnt} already imported)' if dup_cnt > 0 else ''))
    except FileNotFoundError:
      self.perror(f'Error: {os.path.expanduser(proxy_txt_file)} not found')

  def cmd_proxy_list(self, args):
    c = db_conn.cursor()
    c.execute('SELECT proxy.pgroup, count(1) pcount from proxy group by proxy.pgroup ORDER BY pcount DESC;')
    pgroups = ['PROXY GROUPS:'] + [f'{group}: {gcnt} proxies in group' for group, gcnt in c.fetchall()]


    if args.group:
      c.execute("SELECT proxy.url, proxy.pgroup, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state='Rushing') FROM proxy WHERE proxy.pgroup LIKE ?;", (args.group,))
    else:
      c.execute("SELECT proxy.url, proxy.pgroup, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state='Rushing') FROM proxy ORDER BY proxy.pgroup;")
    prox = c.fetchall()
    self.ppaged('\n'.join([''] + pgroups + [''] + [f'({p[1]}) {p[0]} (tasks using: {p[2]})' for p in prox] + [''] + pgroups + ['']), chop=False, length=len(prox))

  def proxy_remove_all(self, args):
    c = db_conn.cursor()
    if args.group:
      c.execute("DELETE FROM proxy WHERE proxy.pgroup LIKE ?;", (args.group,))
    else:
      c.execute("DELETE FROM proxy;")
    c.execute('select changes();')
    db_conn.commit()
    olen = c.fetchone()[0]
    self.poutput(f'Removed {olen} proxies')

  def proxy_add_fn(self, args):
    np = normalize_proxy(args.proxy)
    if np:
      new_cnt, dup_cnt = self.add_proxies([np], group=args.group or 'default')
      if new_cnt > 0:
        self.poutput(f'Added proxy {np}')
      else:
        self.poutput(f'Proxy already added {np}')

  def proxy_remove_fn(self, args):
    c = db_conn.cursor()
    np = normalize_proxy(args.proxy)
    if args.group:
      c.execute("DELETE FROM proxy WHERE url = ? AND pgroup LIKE ?;", (np, args.group))
    else:
      c.execute("DELETE FROM proxy WHERE url = ?;", (np,))
    c.execute("select changes();")
    db_conn.commit()
    changes = c.fetchone()[0]
    if changes > 0:
      self.poutput(f'Removed {args.proxy}')
    else:
      self.poutput(f'No such proxy {args.proxy}')

  def profile_remove_fn(self, args):
    c = db_conn.cursor()
    c.execute('''DELETE FROM profile WHERE name like ?''', (args.profile_name,))
    db_conn.commit()
    self.poutput(f'Removed {args.profile_name}')

  def profile_show(self, args):
    try:
      c = db_conn.cursor()
      c.execute('''SELECT blob FROM profile WHERE name like ?''', (args.name,))
      profile = c.fetchone()[0]
      pp = [list(csv.reader(io.StringIO(profiles_to_csv([p]))))[1] for p in [json.loads(profile)]]
      cols = [
        tf.Column(f) for f in PROFILE_FIELDS
      ]

      table = tf.generate_table(rows=pp, columns=cols)
      self.ppaged(table, chop=True, length=len(pp))
    except Exception:
      self.poutput(f'No profile named {args.name}')
      pass

  prof_imp.set_defaults(func=profile_import)
  prof_remove.set_defaults(func=profile_remove_fn)
  prof_remove_all.set_defaults(func=profile_remove_all)
  prof_list.set_defaults(func=profile_list)
  prof_show.set_defaults(func=profile_show)

  proxy_imp.set_defaults(func=proxy_import)
  proxy_list.set_defaults(func=cmd_proxy_list)
  proxy_add.set_defaults(func=proxy_add_fn)
  proxy_remove.set_defaults(func=proxy_remove_fn)
  proxy_remove_all_p.set_defaults(func=proxy_remove_all)

  def task_add_fn(self, args):
    count = args.count
    url = args.url
    urlp = None
    if url:
      urlp = urlparse(url)
      if not (urlp.hostname and urlp.scheme and urlp.path):
        self.perror(f'Invalid URL {url}')
        return

    nt = []
    proxy_group = args.proxy_group or None
    for x in range(count):
      proxy = args.proxy or None
      if args.proxy and args.proxy != 'none':
        try:
          proxy = normalize_proxy(args.proxy)
          if proxy == ('http://' + args.proxy) and urlparse(proxy).port == None:
            proxy = None
            if not proxy_group or proxy_group == 'default':
              proxy_group = args.proxy
          elif not proxy:
            self.perror(f'Error: Invalid proxy format')
            return
        except Exception:
            self.perror(f'Error: Invalid proxy format')

      profile = None
      try:
        if args.profile:
          if '*' in args.profile:
            profile = args.profile
          else:
            c = db_conn.cursor()
            c.execute('''select blob from profile where name like ?''', (args.profile,))
            profile = json.loads(c.fetchone()[0])
      except (StopIteration, TypeError, json.decoder.JSONDecodeError):
        self.perror(f'Profile {args.profile} not found')
        return

      size = args.size.upper() if args.size is not None else (profile.get('Size') if type(profile) is dict else None) #  or 'Using Profile Size' # profile['Size']
      nt.append({
        'id': secrets.token_hex(8),
        'state': TaskStatus.STOPPED.value,
        'status': '',
        'delay': args.delay,
        'profile': profile,
        'size': size,
        'proxy': proxy,
        'proxy_group': proxy_group,
        'proxy_distribution': ProxyDistribution.RANDOM.value if not proxy else ProxyDistribution.SPECIFIC.value,
        'keywords': ','.join(args.keywords or []),
        'url': url
      })
      # print(nt[-1])


    c = db_conn.cursor()
    c.executemany('INSERT INTO tasks2(id, url, state, proxy, proxy_group, proxy_distribution, profile, profile_in_use, size, keywords, delay) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)',
      [(t['id'], t['url'], t['state'], t['proxy'], t['proxy_group'], t['proxy_distribution'],
        json.dumps(t['profile']) if type(t['profile']) is dict else (t['profile'] if type(t['profile']) is str else None), t['profile']['Name'] if type(t['profile']) is dict else None, t['size'], t['keywords'], args.delay) for t in nt])
    # c.close()
    db_conn.commit()

    self.poutput(f"Added {count} tasks. IDs: {' '.join([t['id'] for t in nt])}")
    if args.start:
      for t in nt:
        self._start_task(t)

  def task_list_fn(self, args):
    status = args.status
    sv = 1 if os.name == 'nt' else -1
    c = db_conn.cursor()
    c.execute('PRAGMA read_uncommitted = true;')
    c.execute('select * from tasks2;')
    tasks = [row_to_dict(c, row) for row in c.fetchall()]
    # c.close()
    rows = sorted([
      TaskRow(task)
      for task in tasks
      if not status or task['status'].lower() == status.lower()
    ], key=lambda t: (sv*t.level_, sv*getattr(t, 'updated_at', 0)))

    columns = [
      tf.Column('ID', attrib='id', width=8, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('State', attrib='state', width=12, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Status', attrib='status', width=12, header_halign=tf.ColumnAlignment.AlignCenter),
      # tf.Column('LastUpdated', attrib='get_last_updated', width=12, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Site', attrib='get_site', width=32, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Keywords', attrib='keywords', width=24, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Size', attrib='get_size', width=7, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Profile', attrib='get_profile_name', width=12, header_halign=tf.ColumnAlignment.AlignCenter),
      tf.Column('Proxy', attrib='get_proxy', width=10, header_halign=tf.ColumnAlignment.AlignCenter)
    ]
    table = tf.generate_table(rows=rows, columns=columns, grid_style=tf.FancyGrid())
    self.ppaged(table, chop=True, length=len(rows))

  def _pop_proxy(self, group=None):
    # t0 = time.time()
    c = db_conn.cursor()
    # TODO: refactor to select all proxies, perform random sort preserving count order, then create tasks
    if group:
      c.execute("SELECT proxy.url, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state in ('Rushing', 'Starting')) cnt FROM proxy WHERE pgroup LIKE ? ORDER BY cnt LIMIT 1;", (group,))
    else:
      c.execute("SELECT proxy.url, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.proxy=proxy.url AND tasks2.state in ('Rushing', 'Starting')) cnt FROM proxy ORDER BY cnt LIMIT 1;")
    proxies_counts = sorted(c.fetchall(), key=lambda x: (x[1], random.random()))
    # dt = time.time() - t0
    # print('popped proxy in', dt)
    if len(proxies_counts) == 0:
      return
    return proxies_counts[0][0]

  def _pop_profile(self, group=None):
    # t0 = time.time()
    c = db_conn.cursor()
    # TODO: refactor to select all proxies, perform random sort preserving count order, then create tasks
    if group:
      group = group.replace('*', '%')
      c.execute("SELECT profile.blob, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.profile_in_use=profile.name AND tasks2.state in ('Rushing', 'Starting')) cnt FROM profile WHERE profile.name LIKE ? ORDER BY cnt LIMIT 1;", (group,))
    else:
      c.execute("SELECT profile.blob, (SELECT COUNT(1) FROM tasks2 WHERE tasks2.profile_in_use=profile.name AND tasks2.state in ('Rushing', 'Starting')) cnt FROM profile ORDER BY cnt LIMIT 1;")
    proxies_counts = sorted(c.fetchall(), key=lambda x: (x[1], random.random()))
    # dt = time.time() - t0
    # print('popped proxy in', dt)
    return json.loads(proxies_counts[0][0])

  def _start_task(self, task):
    tprof = task.get('profile')
    try:
      profile = json.loads(tprof)
    except Exception:
      try:
        profile = self._pop_profile(group=tprof if (tprof and '*' in tprof) else None)
      except Exception:
        # if DEBUG:
        traceback.print_exc()
        self.perror(f'No profiles matching {tprof}')
        return

    if profile is None:
      self.perror('Profile Not Found: Please Import Profiles')
      return

    task['profileInUse'] = profile

    if task['proxy_distribution'] == ProxyDistribution.RANDOM.value:
      proxy = self._pop_proxy(group=task['proxy_group'])
      if not proxy:
        self.perror(f"No proxies available in group ({task['proxy_group']})")
        return
    else:
      proxy = task.get('proxy', 'none')

    c = db_conn.cursor()
    task['proxyInUse'] = proxy
    c.execute("UPDATE tasks2 SET proxy=?, state='Starting', status='', profile_in_use=? WHERE id=?", (proxy, profile['Name'], task['id']))
    db_conn.commit()

    if proxy == 'none':
      proxy = ''

    size = task['size'] if task['size'] is not None else (profile['Size'] if profile.get('Size') else "")
    t0 = time.time()

    return b"start " + json.dumps({
      **task,
      "size": size,
      "profile": profile,
      "proxy": proxy,
      "proxy_group": task.get('proxy_group') or '',
      "keywords": [k for k in task['keywords'].split(",") if k],
      "config": task.get('config', {})
    }).encode('utf-8')

    # self.task_socket.send(b"start " + json.dumps({
    #   **task,
    #   "size": size,
    #   "profile": profile,
    #   "proxy": proxy,
    #   "keywords": [k for k in task['keywords'].split(",") if k],
    #   "config": task.get('config', {})
    # }).encode('utf-8'))

  def task_starting_with(self, task_id_prefix):
    c = db_conn.cursor()
    c.execute('PRAGMA read_uncommitted = true;')
    c.execute('select * from tasks2 where id like ?', (f'{task_id_prefix}%',))
    rows = c.fetchall()
    dd = None
    if len(rows) > 1:
      self.perror(f'Error: multiple tasks start with {task_id_prefix}, be more specific')
    elif len(rows) == 0:
      self.perror(f'Error: task "{task_id_prefix}" not found')
    else:
      dd = row_to_dict(c, rows[0])
    # c.close()
    return dd

  def task_start_fn(self, args):
    started = 0
    for task in args.task_id:
      task = self.task_starting_with(task)
      if not task:
        continue
      if task['state'] != TaskStatus.STOPPED.value:
        self.perror(f'Error: task is already running. Stop with "task stop {task["id"]}" and try again')
        continue
      cmd = self._start_task(task)
      if cmd:
        started += 1
        self.task_socket.send(cmd)
        self.poutput(f'Started task {task["id"]}')
    if started > 0:
      self.poutput('Track with "task watch"')

  def task_start_all_fn(self, args):
    if args.profile:
      args.profile = args.profile.replace('*', '%')

    limit = args.limit or 1e6
    c = db_conn.cursor()
    args.site = args.site.lower() if args.site else ''
    c.execute('PRAGMA read_uncommitted = true;')
    c.execute('select * from tasks2 where state == ? and (? or url like ?) and (? or profile_in_use like ? or profile like ?) and (? or proxy_group like ?) limit ?', (TaskStatus.STOPPED.value,
      not args.site, '%'+(args.site if args.site else '__NOTEXISTS__')+'%',
      not args.profile, (args.profile if args.profile else '__NOTEXISTS__'), (args.profile if args.profile else '__NOTEXISTS__'),
      not args.proxy_group, (args.proxy_group if args.proxy_group else '__NOTEXISTS__'),
      limit))
    tt = 0
    tasks = [row_to_dict(c, taskrow) for taskrow in c.fetchall()]
    # c.close()
    cmdtosend = []
    ten = int(len(tasks)/10)
    for idx, task in enumerate(tasks): # TODO tqdm causes process to restart, change output pipe ?
      if idx > 0 and ten > 0 and idx % ten == 0:
        self.poutput(f"Started {idx}/{len(tasks)} tasks")
      up = urlparse(task['url'])
      cmd = self._start_task(task)
      if cmd:
        self.task_socket.send(cmd)
        tt += 1
    # for c in tqdm(cmdtosend):
    #   self.task_socket.send(c)
    self.poutput(f'Started {tt} tasks')

  def _task_stop_all(self, state=None, status=None, site=None, limit=None, profile=None, proxy_group=None):
    if profile:
      profile = profile.replace('*', '%')

    limit = limit or 1e6
    c = db_conn.cursor()
    site = site.lower() if site else ''
    c.execute('PRAGMA read_uncommitted = true;')
    if status is not None:
      c.execute('select id,url,state from tasks2 where state != ? and status like ? and (? or state like ?) and (? or url like ?) and (? or profile_in_use like ? or profile like ?)  and (? or proxy_group like ?) limit ?', (
        TaskStatus.STOPPED.value,
        '%' + status + '%',
        not state, '%'+(state if state else '__NOTEXISTS__')+'%',
        not site, '%'+(site if site else '__NOTEXISTS__')+'%',
        not profile, (profile if profile else '__NOTEXISTS__'),  (profile if profile else '__NOTEXISTS__'),
        not proxy_group, (proxy_group if proxy_group else '__NOTEXISTS__'),
        limit))
    else:
      c.execute('select id,url,state from tasks2 where state != ? and (? or state like ?) and (? or url like ?) and (? or profile_in_use like ? or profile like ?) and (? or proxy_group like ?) limit ?', (
        TaskStatus.STOPPED.value,
        not state, '%'+(state if state else '__NOTEXISTS__')+'%',
        not site, '%'+(site if site else '__NOTEXISTS__')+'%',
        not profile, (profile if profile else '__NOTEXISTS__'), (profile if profile else '__NOTEXISTS__'),
        not proxy_group, (proxy_group if proxy_group else '__NOTEXISTS__'),
        limit, ))
    tt = 0
    rows = c.fetchall()
    # c.close()
    for task_id,url,state in rows:
      up = urlparse(url)
      self.task_socket.send(b"stop " + json.dumps({"id": task_id}).encode('utf-8'))
      if state != TaskStatus.STOPPED.value:
        tt += 1
    return tt

  def task_stop_all_fn(self, args):
    tt = self._task_stop_all(state=args.state, status=args.status, site=args.site, limit=args.limit, profile=args.profile, proxy_group=args.proxy_group)

    self.poutput(f'Stopped {tt} tasks')

  def task_remove_fn(self, args):
    task = self.task_starting_with(args.task_id)
    if not task:
      return

    c = db_conn.cursor()
    self.task_socket.send(b"stop " + json.dumps({"id": task['id']}).encode('utf-8'))
    c.execute('delete from tasks2 where id = ?', (task['id'],))
    # c.close()
    db_conn.commit()

  def task_remove_all_fn(self, args):
    if args.profile:
      args.profile = args.profile.replace('*', '%')
    proxy_group = args.proxy_group

    self._task_stop_all(site=args.site, status=args.status, profile=args.profile, proxy_group=args.proxy_group)

    c = db_conn.cursor()
    if args.status:
      c.execute('delete from tasks2 where (? or url like ?) and status like ? and (? or profile_in_use like ? or profile like ?) and (? or proxy_group like ?);',
          ((not args.site, '%'+(args.site or '__NOTEXISTS__')+'%', '%'+args.status+'%',
           not args.profile, (args.profile if args.profile else '__NOTEXISTS__'), (args.profile if args.profile else '__NOTEXISTS__')),
           not proxy_group, (proxy_group if proxy_group else '__NOTEXISTS__'),
          ))
    else:
      c.execute('delete from tasks2 where (? or url like ?) and (? or profile_in_use like ? or profile like ?) and (? or proxy_group like ?);',
        (not args.site, '%'+(args.site or '__NOTEXISTS__')+'%', not args.profile, (args.profile if args.profile else '__NOTEXISTS__'), (args.profile if args.profile else '__NOTEXISTS__'), not proxy_group, (proxy_group if proxy_group else '__NOTEXISTS__')
          ))
    c.execute('select changes();')
    # c.close()
    db_conn.commit()

    self.poutput(f'Removed {c.fetchone()[0]} tasks')

  def task_stop_fn(self, args):
    for task in args.task_id:
      task = self.task_starting_with(task)
      if not task:
        continue

      self.task_socket.send(b"stop " + json.dumps({"id": task['id']}).encode('utf-8'))
      self.poutput(f'Stopped {task["id"]}')

  def task_status_fn(self, args):
    # end_evt = threading.Event()
    # original_sigint_handler = signal.getsignal(signal.SIGINT)
    # def handler(*args):
    #   end_evt.set()
    #   if DEBUG:
    #     print('set end_evt')
    # signal.signal(signal.SIGINT, handler)


    # with self.sigint_protection:
    try:
      while 1:
        self.poutput(datetime.now().isoformat() + '\n')
        # c.execute('select state, count(1) from tasks2 group by state order by count(1);')
        # cols = [tf.Column("count", attrib='count'), tf.Column("state", attrib='state')]
        # table = tf.generate_table(rows=[{"count": row[1], "state": row[0]} for row in c.fetchall()], columns=cols)
        # self.poutput(table)
        try:
          c = db_conn.cursor()
          c.execute('PRAGMA read_uncommitted = true;')
          c.execute("select status,state,url, count(1) from tasks2 group by status,state,url order by state,url,status;")

          rows = c.fetchall()
          # c.close()
          cols = [tf.Column("count", attrib='count'), tf.Column("state", attrib='state'), tf.Column("status", attrib='status'), tf.Column("url", attrib='url'), tf.Column("pid", attrib='pid')]
          table = tf.generate_table(rows=[{"count": row[3], "state": row[1], "status": row[0], "url": urlparse(row[2]).scheme + "://" + urlparse(row[2]).hostname, "pid": urlparse(row[2]).path.split('/')[-1]} for row in rows], columns=cols)

          self.poutput(table)
        except Exception as e:
          self.pwarning(f'status update unavailable, please wait ({e})')
        finally:
          if args.watch:
            self.poutput('ctrl-c to stop watching\n')
            time.sleep(1)
          else:
            break
    except Exception:
      if DEBUG:
        traceback.print_exc()
      pass

    # signal.signal(signal.SIGINT, original_sigint_handler)

  def task_success_fn(self, args):
    self.ppaged('\n'.join(self.success_log))

  def task_tail_fn(self, args):
    self.pwarning('task tail has been temporarily disabled')
    return
    # end_evt = threading.Event()
    # original_sigint_handler = signal.getsignal(signal.SIGINT)
    # def handler(*args):
    #   end_evt.set()
    # signal.signal(signal.SIGINT, handler)

    poller = get_task_tail_socket()
    while True:
      evts = poller.poll(100)
      if evts and evts[0][1] == 1:
        msg = evts[0][0].recv()
        self.poutput(json.loads(msg))
      time.sleep(0.01)

    # signal.signal(signal.SIGINT, original_sigint_handler)

  def vault_add_fn(self, args):
    if not args.site.startswith('www.'):
      args.site = 'www.' + args.site
    try:
      c = db_conn.cursor()
      c.execute('insert into vault (key, secret) values (?, ?);', (args.site + '_' + args.email, args.password))
      db_conn.commit()
    except Exception:
      self.perror('Login already in vault, use vault remove')

  def vault_list_fn(self, args):
    c = db_conn.cursor()
    c.execute('select * from vault order by key;')
    for r in c.fetchall():
      self.poutput(f'{" | ".join(r[0].split("_", 1))} | {r[1]}')


  def vault_remove_fn(self, args):
    c = db_conn.cursor()
    c.execute('delete from vault where key=?;', (args.site + '_' + args.email,))
    db_conn.commit()

  def task_edit_fn(self, args):
    c = db_conn.cursor()
    c.execute('update tasks2 set url=? where url like ?', (args.new_url, args.url,))
    c.execute("select changes();")
    db_conn.commit()
    changes = c.fetchone()[0]
    self.poutput(f'Edited {changes} tasks')

  vault_add.set_defaults(func=vault_add_fn)
  vault_remove.set_defaults(func=vault_remove_fn)
  vault_list.set_defaults(func=vault_list_fn)

  task_add.set_defaults(func=task_add_fn)
  task_edit.set_defaults(func=task_edit_fn)
  task_tail.set_defaults(func=task_tail_fn)
  task_list.set_defaults(func=task_list_fn)
  task_status.set_defaults(func=task_status_fn)
  task_watch.set_defaults(func=task_status_fn)
  task_start.set_defaults(func=task_start_fn)
  task_start_all.set_defaults(func=task_start_all_fn)
  task_stop_all.set_defaults(func=task_stop_all_fn)
  task_remove.set_defaults(func=task_remove_fn)
  task_remove_all.set_defaults(func=task_remove_all_fn)
  task_stop.set_defaults(func=task_stop_fn)

  def recap_persist(self):
    with open(os.path.join(APP_DIR, "recap_sessions.json"), 'w') as f:
      json.dump(self.recap_sessions, f)

  def recap_list_fn(self, args):
    cols = [
      tf.Column("idx", attrib='idx'),
      tf.Column("email", attrib='email')
    ]

    table = tf.generate_table(rows=[{"idx": idx, "email": s['email']} for idx,s in enumerate(self.recap_sessions)], columns=cols)
    self.ppaged(table, chop=True, length=len(self.recap_sessions))

  def get_token(self, token):
    s = requests.Session()
    # s.proxies={'http': token[1], 'https': token[1]}
    resp = s.post("https://score.rushaio.co/token", json={"token": token[0]}).json()
    if DEBUG:
      self.poutput(resp)
    return resp['score']

  def recap_score_fn(self, args):
    self.poutput('(ctrl-c to halt)')
    start = time.time()
    sources = [
      # {"name": "google", "url": "https://recaptcha-demo.appspot.com/recaptcha-v3-request-scores.php", "action": "examples/v3scores", "apiKey": "6LdyC2cUAAAAACGuDKpXeDorzUDWXmdqeg-xy696",
      #   "test_fn": lambda token: requests.get('https://recaptcha-demo.appspot.com/recaptcha-v3-verify.php',
      #     # proxies={'http': token[1], 'https': token[1]},
      #      params={'action': "examples/v3scores", 'token': token[0]}).json()['score']},
      {"name": "rush", "url": "https://score.rushaio.co/", "action": "sc_pageview", "apiKey": "6LcI0-MUAAAAAGdTwaIHO-VfOrpwlaqMNeBpGR1E",
        "test_fn": lambda token: self.get_token(token)
      },
      {"name": "ghost", "url": "https://recaptcha-test.ghostaio.com/v3-recaptcha-test", "action": "test_action", "apiKey": "6LdTy_AUAAAAAPP3tBG300bdKQ2kXleDurXPr1_z",
      "test_fn": lambda token: requests.post("https://recaptcha-test.ghostaio.com/api/verify", json={"token": token[0]}).json()['score']},
      {"name": "kodai", "url": "https://essentials.kodai.io/recaptcha/test/v3", "action": "score_test", "apiKey": "6LcVZb0UAAAAANGRKthG23T_1grv4Jsx6drqCTsj",
       "test_fn": lambda token: requests.post('https://essentials.kodai.io/api/recaptcha/verifyScore', json={'token': token[0]}).json()['score']},
      {"name": "antcpt", "url": "https://antcpt.com/eng/information/demo-form/recaptcha-3-test-score.html",
        "action": "homepage", "apiKey": "6LcR_okUAAAAAPYrPe-HK_0RULO1aZM15ENyM-Mf",
        "test_fn": lambda token: requests.post('https://ar1n.xyz/recaptcha3ScoreTest', json={"g-recaptcha-reponse": token[0]}).json()['score']
      },
      {"name": "popfindr", "url": "https://www.popfindr.com/reCaptchaV3_Check",
        "action": "homepage", "apiKey": "6LcviskUAAAAADLr4B5NBd-vblJli41seFfuUCMC",
        "test_fn": lambda token: re.search(
          r'Score : (.+?)<',
          requests.post('https://www.popfindr.com/reCaptchaV3_Check', data={'reCaptchaToken': token[0],'ipAddress': ''}).text
        ).group(1)
      },
    ]
    if args.sources:
      sources = [s for s in sources if s['name'] in [sn.lower() for sn in args.sources]]
    if args.email:
      try:
        session = next(s for s in self.recap_sessions if s['email'].lower() == args.email.lower())
      except Exception:
        self.perror(f'Error: email {args.email} is not logged in')
        return
    elif args.email is not None:
      try:
        session = random.choice(self.recap_sessions)
      except Exception:
        session = {}
    else:
      session = {}


    while sources:
      proxy = "none"
      if args.proxy:
        proxy = normalize_proxy(args.proxy)
      elif args.proxy != "" and len(self.recap_proxies.all())>0:
        proxy = random.choice(self.recap_proxies.all())

      self.poutput(f" +({round(time.time()-start, 2)}s) using proxy: {proxy}")
      self.poutput(f" +({round(time.time()-start, 2)}s) using email: {session.get('email', 'none')}")

      end_evt = threading.Event()
      original_sigint_handler = signal.getsignal(signal.SIGINT)
      def handler(*args):
        end_evt.set()
      signal.signal(signal.SIGINT, handler)

      def do_run(proxy, end_evt):
        for source in sources:
          if end_evt.is_set():
            break
          self.poutput(f' +({round(time.time()-start, 2)}s) {source["url"]}')
          try:
            token = requests.post('http://localhost:42424', json={
              "url": source['url'],
              "proxy": proxy,
              "action": source['action'],
              "apiKey": source['apiKey'],
              "session": session
            }).text
            score = source['test_fn']((token, proxy))
            if end_evt.is_set():
              break
            self.poutput(f' +({round(time.time()-start, 2)}s) {score} ({source["url"]})')
          except Exception as e:
            if not end_evt.is_set():
              self.perror(f' +({round(time.time()-start, 2)}s) error on {source["url"]}: {e}')

      t= Thread(target=do_run, args=(proxy, end_evt))
      t.start()

      while not end_evt.is_set() and t.isAlive():
        time.sleep(0.1)

      signal.signal(signal.SIGINT, original_sigint_handler)

      # if self.terminal_lock.acquire(blocking=False):
      #   try:
      #   finally:
      #     self.terminal_lock.release()
      #     running = False

      if end_evt.is_set() or not args.loop:
        return
      else:
        self.poutput('')
        if args.delay:
          time.sleep(args.delay)

  def recap_login_fn(self, args):
    try:
      cmd = [os.path.join(DIR, "bin", RC_LOGIN_EXE)]
      if args.proxy:
        cmd += ["-proxy", args.proxy]
      session = subprocess.check_output(cmd, env={**os.environ, "CHROMIUM": CHROMIUM}, stderr=subprocess.DEVNULL)
      session = json.loads(session)
      for c in session["cookies"]:
        del c['priority']
      if DEBUG:
        self.poutput(session)
      self.recap_sessions.append(session)
      self.recap_persist()
      self.start_rcpool()
      self.poutput('Added recap session\n')
    except Exception:
      if DEBUG:
        traceback.print_exc()

  def recap_logout_all_fn(self, args):
    self.recap_sessions = []
    self.recap_persist()
    self.start_rcpool()

  def stop_task_consumer(self):
    if getattr(self, 'task_consumer_proc', None):
      if DEBUG:
        self.poutput('terminating task consumer')
      self.task_consumer_proc.terminate()

    # if os.path.exists(PIDFILE):
    #   with open(PIDFILE, 'r') as f:
    #     pid = f.read()
    #     if pid.isdigit():
    #       if os.name == 'nt':
    #         try:
    #           subprocess.check_output(['taskkill', '/F', '/T', '/PID', pid], stderr=subprocess.DEVNULL, creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)
    #         except:
    #           if DEBUG:
    #             traceback.print_exc()
    #         # todo check if process alive in a windows way
    #         time.sleep(3)
    #       else:
    #         for x in range(12):
    #           try:
    #             os.kill(int(pid), signal.SIGKILL)
    #             time.sleep(1)
    #           except ProcessLookupError:
    #             break

  def start_task_consumer(self):
    # DEBUG = True
    if self.started_task_manager:
      return

    self.started_task_manager = True

    def _do_start():
      backoff = 1
      while 1:
        backoff = backoff * 2
        backoff = max(60, backoff)
        try:
          self.stop_task_consumer()
          cmd=[os.path.join(DIR, "bin", TASK_CONSUMER_EXE)]
          if DEBUG:
            self.poutput(' '.join(cmd))

          envflags = {}
          if os.name=='nt':
            envflags['creationflags'] = subprocess.CREATE_NEW_PROCESS_GROUP
          # else:
          #   envflags['preexec_fn'] = os.setpgrp

          # out = subprocess.PIPE if DEBUG else subprocess.DEVNULL
          out = sys.stdout if DEBUG else subprocess.DEVNULL
          self.task_consumer_proc = subprocess.Popen(
              cmd,
              close_fds=True,
              stdout=out,
              cwd=DIR,
              env={**os.environ, "VERSION": VERSION, "ZONEINFO": ZONEINFO_ZIP, "RUSH_CSHOST": os.getenv('RUSH_CSHOST', 'wss://auth.rushaio.co/'), "RUSH_APPDIR": APP_DIR, "RUSH_DB": RUSH_DB, "DENO_DIR": os.getenv('DENO_DIR', os.path.join(DIR, 'adyen', 'deno'))},
              stderr=out, stdin=subprocess.DEVNULL,
              universal_newlines=True,
              start_new_session=True,
              **envflags)
          with open(PIDFILE, 'w') as f:
            f.write(str(self.task_consumer_proc.pid))
          # if DEBUG:
          #   for line in itertools.chain(iter(self.task_consumer_proc.stdout.readline, ""), iter(self.task_consumer_proc.stderr.readline, "")):
          #     print('TASK_CONSUMER_OUT', line)
          #   self.task_consumer_proc.stdout.close()
          #   self.task_consumer_proc.stderr.close()
          exitcode = self.task_consumer_proc.wait()
          if DEBUG:
            print(f'TASK_CONSUMER EXITED {exitcode}')
            print(f'WAITING {exitcode}')
          self.task_consumer_proc = None
        except KeyboardInterrupt:
          raise
        except Exception:
          # TODO sentry
          pass
        finally:
          time.sleep(backoff)

    self.task_consumer_manager_thread = threading.Thread(target=_do_start)
    self.task_consumer_manager_thread.start()

    # self.poutput('Waiting for task engine to start...')
    s = time.time()
    while not self.task_socket_init_evt.is_set():
      if (time.time() - s) > 30:
        self.kill()
        return
      time.sleep(0.25)

  def start_rcpool(self):
    if getattr(self, 'rcpool_proc', None):
      if DEBUG:
        self.poutput('terminating rcpool')
      self.rcpool_proc.terminate()

    if os.name == 'nt':
      zf = ZipFile
    else:
      zf = ZipFileWithPermissions

    if not os.path.exists(CHROMIUM):
      self.poutput('Finishing Install...')
      with zf(CHROMIUM_ZIP, 'r') as zz:
        zz.extractall(os.path.join(APP_DIR, 'chrome'))

    out = subprocess.DEVNULL if not DEBUG else sys.stdout
    f = tempfile.NamedTemporaryFile(mode='w', delete=False)
    sesh = self.recap_sessions.copy()
    random.shuffle(sesh)
    json.dump(sesh, f)
    f.flush()
    f.close()

    cmd=[
      os.path.join(DIR, "bin", RC_POOL_EXE),
      "-workers", str(self.recap_workers),
      "-sessionsJson", f.name,
      "-proxiesJson", RECAP_PROXIES_FN,
    ]
    if DEBUG:
      self.poutput(' '.join(cmd))

    envflags = {}
    if os.name=='nt':
      envflags['creationflags'] = subprocess.CREATE_NEW_PROCESS_GROUP
    else:
      envflags['preexec_fn'] = os.setpgrp

    self.rcpool_proc = subprocess.Popen(
        cmd,
        close_fds=True, stdout=out,
        env={**os.environ, "CHROMIUM": CHROMIUM},
        stderr=out, stdin=subprocess.DEVNULL,
        **envflags)
    while 1:
      try:
        requests.get('http://localhost:42424')
        break
      except Exception:
        continue

  def recap_logout_fn(self, args):
    self.recap_sessions = [s for s in self.recap_sessions if s['email'].lower() != args.email.lower()]
    self.recap_persist()
    self.start_rcpool()

  def save_recap_proxies(self):
    with open(RECAP_PROXIES_FN, 'w') as f:
      json.dump(self.recap_proxies.all(), f)

  def recap_add_proxies(self, proxies):
    orig_len = len(self.recap_proxies.all())
    np = []
    for proxy in proxies:
      nproxy = normalize_proxy(proxy)
      if nproxy:
        if not self.recap_proxies.has(nproxy):
          np.append(nproxy)
          self.recap_proxies.add(nproxy)
      else:
        self.pwarning(f'Warning: skipping {proxy}, invalid format')

    self.save_recap_proxies()
    self.start_rcpool()
    return len(self.recap_proxies.all()) - orig_len, (orig_len + len(np) - len(self.recap_proxies.all()))

  def recap_proxy_import(self, args):
    proxy_txt_file = os.path.expanduser(' '.join(args.proxy_txt_file))
    try:
      with open(proxy_txt_file) as f:
        lines = [l.strip() for l in f.read().split('\n') if l.strip()]
        new_cnt, dup_cnt = self.recap_add_proxies(lines)
        self.poutput(f'Imported {new_cnt} new proxies for recap ' + (f'({dup_cnt} already imported)' if dup_cnt > 0 else ''))
    except FileNotFoundError:
      self.perror(f'Error: {os.path.expanduser(proxy_txt_file)} not found')

  def recap_cmd_proxy_list(self, args):
    pp = self.recap_proxies.all_priority()
    self.ppaged('\n'.join([f'{p[1]}' for p in pp]), chop=False, length=len(pp))

  def recap_proxy_remove_all(self, args):
    olen = len(self.recap_proxies.pq)
    self.recap_proxies = ResourceShare()
    self.save_recap_proxies()
    self.start_rcpool()
    self.poutput(f'Removed {olen} proxies')

  def recap_proxy_add_fn(self, args):
    np = normalize_proxy(args.proxy)
    if np:
      new_cnt, dup_cnt = self.recap_add_proxies([np])
      if new_cnt > 0:
        self.poutput(f'Added proxy {np}')

  def recap_proxy_remove_fn(self, args):
    np = normalize_proxy(args.proxy)
    self.recap_proxies.remove(np)
    self.save_recap_proxies()
    self.start_rcpool()
    self.poutput(f'Removed {args.proxy}')

  def recap_proxy_litport_rotate_fn(self, args):
    key = parse_qs(urlparse(args.url).query)['key'][0]
    for rot in self.litport_rotations:
      if rot['key'] == key:
        self.perror(f'Error: already rotating litport {key}')
        return

    end_evt = threading.Event()
    def litport_rotate(url, count, end_evt):
      urlp = urlparse(url)
      qsp = parse_qs(urlp.query)
      num = 0
      while not end_evt.is_set():
        urlp = urlp._replace(query=f'key={qsp["key"][0]}&num={num}')
        requests.get(urlunparse(urlp))
        end_evt.wait(10)
        if num < (count - 1):
          num += 1
        else:
          num = 0

    t = threading.Thread(target=litport_rotate, args=(args.url, int(args.count), end_evt))
    t.start()

    print('Started litport rotation')

    self.litport_rotations.append({'key': key, 'end_evt': end_evt})

  # recap_login.set_defaults(func=recap_login_fn)
  # recap_list.set_defaults(func=recap_list_fn)
  # recap_list_logins.set_defaults(func=recap_list_fn)
  # recap_score.set_defaults(func=recap_score_fn)
  # recap_logout.set_defaults(func=recap_logout_fn)
  # recap_logout_all.set_defaults(func=recap_logout_all_fn)
  # recap_proxy_imp.set_defaults(func=recap_proxy_import)
  # recap_proxy_list.set_defaults(func=recap_cmd_proxy_list)
  # recap_proxy_add.set_defaults(func=recap_proxy_add_fn)
  # recap_proxy_remove.set_defaults(func=recap_proxy_remove_fn)
  # recap_proxy_remove_all_p.set_defaults(func=recap_proxy_remove_all)
  # recap_proxy_litport_rotate.set_defaults(func=recap_proxy_litport_rotate_fn)

  def key_activate_fn(self, args):
    key = args.key.strip()
    if len(key) != 32:
      self.poutput(f'Invalid key {key} {len(key)}')
      return

    self.key = key
    with open(os.path.join(APP_DIR, 'key'), 'w') as f:
      f.write(key)

    self.start_task_consumer()
    self.poutput(f'Activated key {key}')
    # TODO key verify with api here + error msg?

  key_activate.set_defaults(func=key_activate_fn)

  def config_set_webhook_fn(self, args):
    with open(os.path.join(APP_DIR, 'config', 'webhook'), 'w') as f:
      f.write(args.webhook.strip())
    self.poutput(f'Set webhook: {args.webhook.strip()}. Restart Rush to activate')

  config_set_webhook.set_defaults(func=config_set_webhook_fn)

  @cmd2.with_argparser(config_parser)
  def do_config(self, args):
    """Config"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('config')

  @cmd2.with_argparser(key_parser)
  def do_key(self, args):
    """Key"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('key')

  @cmd2.with_argparser(prof_parser)
  def do_profile(self, args):
    """Profile"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('profile')

  @cmd2.with_argparser(proxy_parser)
  def do_proxy(self, args):
    """Proxy"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('proxy')

  @cmd2.with_argparser(task_parser)
  def do_task(self, args):
    """Task"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('task')

  @cmd2.with_argparser(vault_parser)
  def do_vault(self, args):
    """Task"""
    func = getattr(args, 'func', None)
    if func is not None:
      # Call whatever subcommand function was selected
      func(self, args)
    else:
      # No subcommand was provided, so call help
      self.do_help('vault')

  # @cmd2.with_argparser(recap_parser)
  # def do_recap(self, args):
  #   """Task"""
  #   func = getattr(args, 'func', None)
  #   if func is not None:
  #     # Call whatever subcommand function was selected
  #     func(self, args)
  #   else:
  #     # No subcommand was provided, so call help
  #     self.do_help('recap')

icon = open(os.path.join(DIR, 'icon.txt')).read()

if __name__ == '__main__':
  # print(MOTD)
  def exit_gracefully():
    app.kill()
    if DEBUG:
      print('exit_gracefully.postkill')
    try:
      os.kill(-os.getpgid(os.getpid()),signal.SIGKILL)
    except:
      if DEBUG:
        traceback.print_exc()

  try:

    signal.signal(signal.SIGHUP, exit_gracefully)
  except:
    pass
  try:
    signal.signal(signal.SIGQUIT, exit_gracefully)
  except:
    pass
  try:
    signal.signal(signal.SIGTERM, exit_gracefully)
  except:
    pass
  # try:
  #   signal.signal(signal.SIGINT, exit_gracefully)
  # except:
  #   pass
  is_harvest = False
  try:
    is_harvest = open(os.path.join(DIR, 'is_harvest'), 'r').read().strip() == '1'
  except:
    pass
  if is_harvest or os.getenv('HARVEST') == '1':
    import sys
    print(icon)
    print()
    print(f'RushAIO FARM {VERSION}')
    print()
    if os.name == 'posix':
      subprocess.call("ps aux | grep firef | grep folders | awk '{print $2}' | xargs kill", shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
      subprocess.call("pkill mitm", shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    else:
      subprocess.call(['taskkill', '/F', '/IM', 'firefox.exe'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)
      subprocess.call(['taskkill', '/F', '/IM', 'mitm.exe'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, creationflags=subprocess.CREATE_NEW_PROCESS_GROUP)

    print('Preparing for winter harvest...')
    print()
    import selenium
    from rcfarm.serv import start as rcstart
    # import shutil


    # if os.path.isdir(os.path.join(APP_DIR, 'firefox')):
    #   shutil.rmtree(os.path.join(APP_DIR, 'firefox'), ignore_errors=True)
    # if os.path.isdir(os.path.join(DIR, 'firefox')):
    #   shutil.copytree(os.path.join(DIR, 'firefox'), os.path.join(APP_DIR, 'firefox'), ignore_errors=True)



    rcstart()
    sys.exit(0)

  db_conn = sqlite3.connect('file:' + RUSH_DB + "?_busy_timeout=5000&_journal_mode=WAL&cached=shared", uri=True)
  if DEBUG:
    db_conn.set_trace_callback(print)

  c = db_conn.cursor()
  c.execute('PRAGMA journal_mode=WAL;')
  c.fetchall()

  import sys
  print(icon)
  print()
  print(f'RushAIO {VERSION}')
  print()
  app = App()

  try:
    if len(sys.argv) > 1:
      app.onecmd(" ".join(sys.argv[1:]))
    res = app.cmdloop()
    if DEBUG:
      print('exit', res)
  finally:
    exit_gracefully()
