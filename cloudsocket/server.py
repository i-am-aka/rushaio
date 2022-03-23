#!/usr/bin/env python

"""
CLOUDSOCKET TODO
- Logging to ES
"""

import time
import aiomysql
import asyncio
import datetime
import hashlib
import json
import jwt
import http
import websockets
import traceback
from functools import partial
from urllib.parse import parse_qs
from cloudsocket.config import JWT_KEY
from cloudsocket.db import get_engine, License

import logging
logname = 'auth.log'
logging.basicConfig(filename=logname,
                            filemode='a',
                            format='%(asctime)s,%(msecs)d %(name)s %(levelname)s %(message)s',
                            datefmt='%H:%M:%S',
                            level=logging.INFO)

logger = logging.getLogger(__name__)

def now(): return (datetime.datetime.utcnow() - datetime.datetime.utcfromtimestamp(0)).total_seconds()

async def update_db_pingtime(db_engine, license_key, pt):
  async with db_engine.acquire() as conn:
    await conn.execute(License.update().where(License.c.license_key==license_key).values({'ws_ping_time': pt}))
    await conn._commit_impl()

async def auth_task_coro(db_engine, ws):
  ip = ws.request_headers.get('X-Forwarded-For', ws.remote_address[0])
  ua = ws.request_headers.get('User-Agent', 'unset')

  key = parse_qs(ws.request_headers['Cookie'])['key'][0]
  async with db_engine.acquire() as conn:
    lic = await conn.execute(License.select().where(License.c.license_key==key))
    row = await lic.first()
    if not row:
      logger.info(json.dumps({"event": "unlicensed", "key": key, "ip": ip, "ua": ua, "time": time.time()}))
      print(f'{key} unlicensed ~~~ ip={ip} ua={ua}',)
      await ws.send(json.dumps({'jwt': '', 'error': 'Unlicensed'}))
      # No license
      return
    if row[License.c.ws_ping_time] > (now() - 60):
      logger.info(json.dumps({"event": "instance_active", "key": key, "ip": ip, "ua": ua, "time": time.time()}))
      print(f'{key} instance active ~~~ ip={ip} ua={ua}')
      await ws.send(json.dumps({'jwt': '', 'error': 'Another Instance Active'}))
      # Instance active
      return

  print(f'valid conn for {key} ~~~ ip={ip} ua={ua}')
  logger.info(json.dumps({"event": "conn_established", "key": key, "ip": ip, "ua": ua, "time": time.time()})) 

  key_hash = hashlib.sha256(key.encode()).hexdigest()

  try:
    while 1:
      await update_db_pingtime(db_engine, key, int(now()))
      token = jwt.encode({'exp': datetime.datetime.utcnow() + datetime.timedelta(seconds=180), 'rush_id': key_hash}, JWT_KEY, algorithm='HS256')
      await ws.send(json.dumps({'jwt': token.decode()}))
      # TODO: update db non blocking
      await asyncio.wait({asyncio.sleep(30), ws.wait_closed()}, return_when=asyncio.FIRST_COMPLETED)
  except Exception as e:
    print(f'conn closed for {key} ({type(e).__name__} {e}) ~~~ ip={ip}')
  finally:
    logger.info(json.dumps({"event": "conn_closed", "key": key, "ip": ip, "ua": ua, "time": time.time()})) 

    await update_db_pingtime(db_engine, key, int(now()) - 60)

async def handle_conn(db_engine, ws, path):
  auth_task = asyncio.ensure_future(
    auth_task_coro(db_engine, ws))
  done, pending = await asyncio.wait(
    [auth_task],
    return_when=asyncio.FIRST_COMPLETED,
  )
  for task in pending:
    task.cancel()

class Protocol(websockets.server.WebSocketServerProtocol):
  async def process_request(self, path, request_headers):
    try:
      cookie = parse_qs(request_headers['Cookie'])
      if len(cookie['key'][0]) != 32:
        return (http.HTTPStatus.BAD_REQUEST, websockets.http.Headers(), "Bad Request")

    except (KeyError, TypeError):
      return (http.HTTPStatus.BAD_REQUEST, websockets.http.Headers(), "Bad Request")

async def go():
  db_engine = await get_engine()
  await websockets.serve(partial(handle_conn, db_engine), "0.0.0.0", "8099", create_protocol=Protocol)

if __name__ == '__main__':
  asyncio.get_event_loop().run_until_complete(go())
  asyncio.get_event_loop().run_forever()
