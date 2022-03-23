#!/usr/bin/env python3
import json
import math
import pprint
import os
import sys

DIR = os.path.dirname(__file__)

def parse(sd):
  parsed = {}
  sdp = sd.split('-1,2,-94,')
  parsed['version'] = sdp[0]
  devinfo = sdp[1].split(',')
  # _, parsed['test'] = (1,2)
  (_, __, __, ___,
  parsed['device_height'],
  parsed['device_width'],
  parsed['battery_receiver_flag'],
  parsed['firebase_analytics_level'],
  parsed['app_config_orientation'],
  parsed['language'], # M.d(language)
  parsed['build_version_release'], #M.d(BUILD.Version.Release)
  parsed['acceleromer_rotation'],
  parsed['build_model'], #M.d(BUILD.Model)
  parsed['build_bootloader'],
  parsed['build_hardware'],
  _____,
  parsed['package'],
  ______,
  _______,
  parsed['android_id'],
  ________,
  parsed['has_keyboard'],
  parsed['is_adb_enabled'],
  parsed['build_version_codename'], # M.d
  parsed['build_version_incremental'],  # M.d,
  parsed['build_version_sdk_int'],
  parsed['build_manufacturer'],
  parsed['build_product'],
  parsed['build_tags'],
  parsed['build_type'],
  parsed['build_user'],
  parsed['build_display'],
  parsed['build_board'],
  parsed['build_brand'],
  parsed['build_device'],
  parsed['build_fingerprint'],
  parsed['build_host'],
  parsed['build_id'],
  parsed['devinfo_hash'], # M.a,
  parsed['randint'],
  parsed['t0_wall_div2']
  ) = devinfo

  _, parsed['do_en'], parsed['dm_en'], parsed['t_en'] = sdp[2].split(',')

  parsed['touch_events'] = []
  touch_events_unp = [s.split(',') for s in sdp[5].split(';') if s]
  touch_events_unp[0] = touch_events_unp[0][1:]
  for tenp in touch_events_unp:
    if len(tenp) != 8:
      continue
    te = {}
    (te['type'],
     te['dt'],
     _,
     __,
     te['pointer_count'],
     ___,
     te['tool'],
     ____
    ) = tenp
    parsed['touch_events'].append(te)

  orient_events_unp = [s.split(',') for s in sdp[6].split(';') if s]
  orient_events_unp[0] = orient_events_unp[0][1:]
  parsed['orient_events'] = []
  for oenp in orient_events_unp:
    oe = {}
    oe['dt'], oe['yaw'], oe['pitch'], oe['roll'], oe['usually_one'] = [float(n) for n in oenp]
    parsed['orient_events'].append(oe)

  motion_events_unp = [s.split(',') for s in sdp[7].split(';') if s]
  motion_events_unp[0] = motion_events_unp[0][1:]
  parsed['motion_events'] = []
  for menp in motion_events_unp:
    me = {}
    (me['dt'],
      me['dax'],
      me['day'],
      me['daz'],
      me['ax'],
      me['ay'],
      me['az'],
      me['gax'],
      me['gay'],
      me['gaz'],
      me['usually_one']) = [float(n) for n in menp]
    parsed['motion_events'].append(me)


  # print(pprint.pformat(motion_events_unp))
  # print(pprint.pformat(parsed['motion_events']))
  # print(pprint.pformat(sdp[8]))

  # TODO:
  orient_quant = sdp[8].split(';')[1:]
  parsed['orient_quant'] = {
    'min': orient_quant[0], 'max': orient_quant[1],
    'sampleHash': orient_quant[2], # k.a(n.a(o.a(floatArr)))
    'quantStr': orient_quant[3] # n.a(o.a(floatArr, min, max))
  }

  orient_quant2 = sdp[9]
  orient_quants = orient_quant2.split(':')
  parsed['orient_quant2'] = [
    {'min': o[0], 'max': o[1],
    'sampleHash': o[2], # k.a(n.a(o.a(floatArr)))
    'quantStr': o[3]}
    for o in [oo.split(';')[1:] for oo in orient_quants]
  ]

  motion_quant = sdp[10].split(';')[1:]
  parsed['motion_quant'] = {
    'min': motion_quant[0], 'max': motion_quant[1],
    'sampleHash': motion_quant[2], # k.a(n.a(o.a(floatArr)))
    'quantStr': motion_quant[3] # n.a(o.a(floatArr, min, max))
  }

  motion_quant2 = sdp[11]
  motion_quants = motion_quant2.split(':')
  parsed['motion_quant2'] = [
    {'min': o[0], 'max': o[1],
    'sampleHash': o[2], # k.a(n.a(o.a(floatArr)))
    'quantStr': o[3]}
    for o in [oo.split(';')[1:] for oo in motion_quants]
  ]

  fields = sdp[12].split(',')[1:]
  # parsed['-115'] = {
  #   str(idx): value for idx, value in enumerate(fields)
  # }
  parsed['touch_dt_sum'] = fields[1]
  parsed['orient_quant_sum'] = fields[2]
  parsed['motion_quant_sum'] = fields[3]
  parsed['touch_orient_motion_sum'] = fields[4] # fields[1] + fields[2] + fields[3]
  parsed['sensor_dt'] = fields[5] # currentTimeMillis
  parsed['text_change_count'] = fields[6]
  parsed['v.p*1000'] = fields[10]
  parsed['sensor_uptime_dt'] = fields[11]
  parsed['has_orientation_listener'] = fields[12]

  parsed['evt_hash'] = fields[13]
  parsed['start_ts'] = fields[14]

  perf = sdp[15].split(',')[1:]
  for idx in range(len(perf)):
    parsed[f'perf_{idx}'] = perf[idx]

  lifecycle = [l.split(',') for l in sdp[16].split(';') if l]
  lifecycle[0] = lifecycle[0][1:]
  parsed['lifecycle'] = []
  for l in lifecycle:
    if len(l) >= 2:
      parsed['lifecycle'].append({'type': l[0], 'ts': l[1]})

  return parsed

def formatNumber(num):
  if num % 1 == 0:
    if num == 0.0 or num == 0:
      return "-0"
    else:
      return math.copysign(int(num), num)
  else:
    return num

def serialize(parsed):
  SEP = "-1,2,-94,"

  devinfo = [
    "-1",
    "uaend",
    "-1",
    parsed['device_height'],
    parsed['device_width'],
    parsed['battery_receiver_flag'],
    parsed['firebase_analytics_level'],
    parsed['app_config_orientation'],
    parsed['language'], # M.d(language)
    parsed['build_version_release'], #M.d(BUILD.Version.Release)
    parsed['acceleromer_rotation'],
    parsed['build_model'], #M.d(BUILD.Model)
    parsed['build_bootloader'],
    parsed['build_hardware'],
    "-1",
    parsed['package'],
    "-1",
    "-1",
    parsed['android_id'],
    "-1",
    parsed['has_keyboard'],
    parsed['is_adb_enabled'],
    parsed['build_version_codename'], # M.d
    parsed['build_version_incremental'],  # M.d,
    parsed['build_version_sdk_int'],
    parsed['build_manufacturer'],
    parsed['build_product'],
    parsed['build_tags'],
    parsed['build_type'],
    parsed['build_user'],
    parsed['build_display'],
    parsed['build_board'],
    parsed['build_brand'],
    parsed['build_device'],
    parsed['build_fingerprint'],
    parsed['build_host'],
    parsed['build_id']
  ]

  devinfo_hash = str(str_hash_sum(','.join(devinfo)))
  print(','.join(devinfo), devinfo_hash)

  return (
    "2.2.2" + SEP +
    f"-100," +
    ",".join(devinfo + [
        devinfo_hash,
        parsed['randint'],
        str(parsed['t0_wall_div2'])
      ]) + SEP + "-101,"
    + ",".join([parsed['do_en'], parsed['dm_en'], parsed['t_en']])
    + SEP + "-102," + SEP + "-108," + SEP + "-117,"
    + "".join([",".join([
        te['type'],
        int(te['dt']),
        "0",
        "0",
        te['pointer_count'],
        "1",
        te['tool'],
        "-1"
      ]) + ";" for te in parsed['touch_events']])
    + SEP + "-111," +
     "".join([",".join(
       map(str, [int(oe['dt']), formatNumber(round(oe['yaw'], 2)), formatNumber(round(oe['pitch'], 2)), formatNumber(round(oe['roll'], 2)), int(oe['usually_one'])])
      ) + ";" for oe in parsed['orient_events']])
    + SEP + "-109," +
    "".join([",".join(map(str, [
      int(me['dt']),
      formatNumber(round(me['dax'], 2)),
      formatNumber(round(me['day'], 2)),
      formatNumber(round(me['daz'], 2)),
      formatNumber(round(me['ax'], 2)),
      formatNumber(round(me['ay'], 2)),
      formatNumber(round(me['az'], 2)),
      formatNumber(round(me['gax'], 2)),
      formatNumber(round(me['gay'], 2)),
      formatNumber(round(me['gaz'], 2)),
      int(me['usually_one'])
     ])) + ";" for me in parsed['motion_events']])
    + SEP + "-144,"
    + ":".join([
      ";".join([
        "2",
        parsed['orient_quant']['min'],
        parsed['orient_quant']['max'],
        parsed['orient_quant']['sampleHash'],
        parsed['orient_quant']['quantStr']
      ])
    ]) + SEP + "-142,"
    + ":".join([
      ";".join([
        "2",
        oq['min'],
        oq['max'],
        oq['sampleHash'],
        oq['quantStr']
      ])
      for oq in parsed['orient_quant2']
    ]) + SEP + "-145,"
    + ":".join([
      ";".join([
        "2",
        parsed['motion_quant']['min'],
        parsed['motion_quant']['max'],
        parsed['motion_quant']['sampleHash'],
        parsed['motion_quant']['quantStr']
      ])
    ]) + SEP + "-143,"
    + ":".join([
      ";".join([
        "2",
        mq['min'],
        mq['max'],
        mq['sampleHash'],
        mq['quantStr']
      ])
      for mq in parsed['motion_quant2']
    ]) + SEP + "-115,"
    + ",".join([
      "0",
      parsed['touch_dt_sum'],
      str(parsed['orient_quant_sum']),
      str(parsed['motion_quant_sum']),
      str(parsed['touch_orient_motion_sum']),
      parsed['sensor_dt'],
      parsed['text_change_count'],
      str(len(parsed['motion_events'])),
      str(len(parsed['orient_events'])),
      str(len(parsed['orient_events'])),
      parsed['v.p*1000'],
      parsed['sensor_uptime_dt'],
      parsed['has_orientation_listener'],
      parsed['evt_hash'],
      str(parsed['start_ts']),
      "0"
    ]) + SEP + "-106,-1,0" + SEP + "-120," + SEP + "-112,"
    + ",".join([parsed[f'perf_{idx}'] for idx in range(9)])
    + SEP + "-103," +
    "".join([",".join([evt['type'], str(evt['ts'])]) + ";" for evt in parsed['lifecycle']])

  )

def str_hash_sum(str_):
  s = 0
  for c in str_:
    if ord(c) < 128:
      s += ord(c)
  return s
"""
public static int a(String str) {
        if (str == null || str.trim().equalsIgnoreCase("")) {
            return -1;
        }
        int i2 = 0;
        int i3 = 0;
        while (i2 < str.length()) {
            try {
                char charAt = str.charAt(i2);
                if (charAt < 128) {
                    i3 += charAt;
                }
                i2++;
            } catch (Exception unused) {
                return -2;
            }
        }
        return i3;
    }

"""

if __name__ == '__main__':
  with open(DIR + "/sensor_trace6.txt") as f:
    sensor = f.read().strip()
    parsed = parse(sensor)
    # print(pprint.pformat(parsed))
    print(json.dumps(parsed))
    print(serialize(parsed))
    print(serialize(parsed) == sensor)
