#!/usr/bin/env python3

import json
import jwt
import time
import random
import os
from functools import wraps
from flask import Flask, jsonify, request, abort
from aka_mobile.gen import gen as mobile_gen
from config import JWT_KEY

app = Flask(__name__)
DEBUG = (os.getenv('DEBUG') == "1")
DIR = os.path.dirname(os.path.abspath(__file__))

def jwt_required(func):
  @wraps(func)
  def inner(*args, **kwargs):
    try:
      jwt.decode(request.headers['Authorization'].split()[1], JWT_KEY, algorithm='HS256')
    except:
      if not DEBUG:
        abort(401)
    return func(*args, **kwargs)
  return inner

@app.route('/hib_skus')
@jwt_required
def hib_skus():
  rootPid = request.args.get('rootPid')
  colorId = request.args.get('colorId')
  if not (rootPid and colorId):
    abort(400)

  hib_skus = [s for s in HIB_SKUS.get(rootPid, []) if s['color']['id'] == colorId]
  return jsonify([{
    "pid": s["id"],
    "color": s["color"]["label"],
    "colorId": colorId,
    "size": s["size"]
  } for s in hib_skus])

from aka_mobile.parse import parse as parse_mobile, serialize

@app.route('/mobile')
@jwt_required
def akamai_mobile():
  package = request.args.get('package', 'com.WinnersCircle')
  if not package or len(package) > 255:
    abort(400)
    return

  sensor_data = mobile_gen(package)

  # print(json.dumps(sensor_data[1], indent=2))
  # print(json.dumps(parse_mobile(serialize(sensor_data[1])), indent=2))
  return json.dumps([sensor_data[0], json.dumps(sensor_data[1])])

@app.route('/ay')
@jwt_required
def adyen():
  return random.choice(ADYEN_FPS)[0]

HIB_FP_BASE = [
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%206.1%3B%20WOW64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/80.0.3987.87%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%206.1%3B%20WOW64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/80.0.3987.87%20Safari/537.36;en-US;undefined;SITE;undefined;undefined;undefined;undefined;true;true;EPOCH_MILLIS;TZ_OFFSET_HOURS;6/7/2005%2C%209%3A33%3A44%20PM;1366;768;;;;;;;10;ABS_TZ_OFFSET_MINS;MINUS_SIXTY_ABS_TZ_OFFSET_MINS;URLENCODED_LOCAL_TIME;24;1366;768;0;0;;;;;;;;;;;;;;;;;;;18;"
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.0%3B%20SM-A520F%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.0%3B%20SM-A520F%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588858952672;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;8;480;420;07/05/2020%2C%2006%3A42%3A32;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20Redmi%20Note%205%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20Redmi%20Note%205%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;false;1588859082157;12;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;6;-780;-720;5/8/2020%2C%201%3A44%3A42%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20SM-F907B%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20SM-F907B%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;it-IT;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859204715;-8;7/6/2005%2C%2021%3A33%3A44;820;586;;;;;;;5;480;420;7/5/2020%2C%2006%3A46%3A44;24;820;586;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20SM-A530W%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20SM-A530W%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859264422;-8;6/7/2005%2C%209%3A33%3A44%20PM;740;360;;;;;;;5;480;420;5/7/2020%2C%206%3A47%3A44%20AM;24;740;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.0%3B%20LG-US996%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.0%3B%20LG-US996%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859349351;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;7;480;420;07/05/2020%2C%2006%3A49%3A09;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;false;false;1588859383503;-7;6/7/2005%2C%209%3A33%3A44%20PM;2752;1152;;;;;;;2;420;420;5/7/2020%2C%206%3A49%3A43%20AM;24;2752;1112;0;0;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20Moto%20Z%20%282%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20Moto%20Z%20%282%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859425095;-6;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;4;360;300;07/05/2020%2C%2008%3A50%3A25;24;640;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20moto%20g%286%29%20plus%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20moto%20g%286%29%20plus%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859482426;-8;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;7;480;420;5/7/2020%2C%206%3A51%3A22%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20LM-G810%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20LM-G810%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;ru-RU;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859549603;-8;07.06.2005%2C%2021%3A33%3A44;857;412;;;;;;;3;480;420;07.05.2020%2C%2006%3A52%3A29;24;857;412;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20moto%20g%286%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/76.0.3809.89%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv7l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20moto%20g%286%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/76.0.3809.89%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859640486;-8;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;8;480;420;5/7/2020%2C%206%3A54%3A00%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20SM-T820%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20SM-T820%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859727363;-8;6/7/2005%2C%209%3A33%3A44%20PM;1024;768;;;;;;;8;480;420;5/7/2020%2C%206%3A55%3A27%20AM;24;1024;768;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%206.0.1%3B%20D6503%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/77.0.3865.92%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv7l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%206.0.1%3B%20D6503%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/77.0.3865.92%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859782273;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;9;480;420;07/05/2020%2C%2006%3A56%3A22;24;640;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.1.1%3B%20ZTE%20A2017U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.1.1%3B%20ZTE%20A2017U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859911116;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;7;480;420;07/05/2020%2C%2006%3A58%3A31;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.1.1%3B%20E6653%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.1.1%3B%20E6653%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860030785;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;8;480;420;07/05/2020%2C%2007%3A00%3A30;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%2010%3B%20SM-F700U1%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%2010%3B%20SM-F700U1%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;false;false;1588860114721;9;6/7/2005%2C%209%3A33%3A44%20PM;879;360;;;;;;;3;-540;-540;5/7/2020%2C%2011%3A01%3A54%20PM;24;879;360;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%2010%3B%20SM-G965U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%2010%3B%20SM-G965U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860278674;-8;6/7/2005%2C%209%3A33%3A44%20PM;846;412;;;;;;;4;480;420;5/7/2020%2C%207%3A04%3A38%20AM;24;846;412;0;0;;;;;;;;;;;;;;;;;;;19;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_13_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_13_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860333560;-5;6/7/2005%2C%209%3A33%3A44%20PM;1440;900;;;;;;;3;300;240;5/7/2020%2C%2010%3A05%3A33%20AM;24;1440;900;0;0;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_4%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_4%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588895251049;-5;6/7/2005%2C%209%3A33%3A44%20PM;1680;1050;;;;;;;2;300;240;5/7/2020%2C%207%3A47%3A31%20PM;24;1680;1027;0;23;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_14_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_14_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588909325098;-5;6/7/2005%2C%209%3A33%3A44%20PM;1280;800;;;;;;;2;300;240;5/7/2020%2C%2011%3A42%3A05%20PM;24;1280;735;0;23;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588965502432;-5;6/7/2005%2C%209%3A33%3A44%20PM;1920;1080;;;;;;;2;300;240;5/8/2020%2C%203%3A18%3A22%20PM;24;1920;1040;0;0;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_0%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_0%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588853176855;-8;6/7/2005%2C%209%3A33%3A44%20PM;1280;960;;;;;;;3;480;420;5/7/2020%2C%205%3A06%3A16%20AM;24;1280;937;0;23;;;;;;;;;;;;;;;;;;;18;",
"TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_10_5%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_10_5%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588851116852;-8;6/7/2005%2C%209%3A33%3A44%20PM;1600;1200;;;;;;;3;480;420;5/7/2020%2C%204%3A31%3A56%20AM;24;1600;1173;0;23;;;;;;;;;;;;;;;;;;;18;"
]
@app.route('/hib_fp')
@jwt_required
def hib_fp():
  fp_base = random.choice(HIB_FP_BASE)
  fps = fp_base.split(';')
  fps[36] = "www.hibbett.com"
  fps[43] = str(int(time.time()*1000))
  fps[44] = "-8"
  fps[55] = "480"
  fps[56] = "420"
  return ";".join(fps)

  """
  dets := []string{
    // det1 := `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%206.1%3B%20WOW64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/80.0.3987.87%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%206.1%3B%20WOW64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/80.0.3987.87%20Safari/537.36;en-US;undefined;SITE;undefined;undefined;undefined;undefined;true;true;EPOCH_MILLIS;TZ_OFFSET_HOURS;6/7/2005%2C%209%3A33%3A44%20PM;1366;768;;;;;;;10;ABS_TZ_OFFSET_MINS;MINUS_SIXTY_ABS_TZ_OFFSET_MINS;URLENCODED_LOCAL_TIME;24;1366;768;0;0;;;;;;;;;;;;;;;;;;;18;`
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.0%3B%20SM-A520F%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.0%3B%20SM-A520F%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588858952672;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;8;480;420;07/05/2020%2C%2006%3A42%3A32;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20Redmi%20Note%205%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20Redmi%20Note%205%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;false;1588859082157;12;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;6;-780;-720;5/8/2020%2C%201%3A44%3A42%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20SM-F907B%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20SM-F907B%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;it-IT;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859204715;-8;7/6/2005%2C%2021%3A33%3A44;820;586;;;;;;;5;480;420;7/5/2020%2C%2006%3A46%3A44;24;820;586;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20SM-A530W%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20SM-A530W%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859264422;-8;6/7/2005%2C%209%3A33%3A44%20PM;740;360;;;;;;;5;480;420;5/7/2020%2C%206%3A47%3A44%20AM;24;740;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.0%3B%20LG-US996%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.0%3B%20LG-US996%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859349351;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;7;480;420;07/05/2020%2C%2006%3A49%3A09;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;false;false;1588859383503;-7;6/7/2005%2C%209%3A33%3A44%20PM;2752;1152;;;;;;;2;420;420;5/7/2020%2C%206%3A49%3A43%20AM;24;2752;1112;0;0;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20Moto%20Z%20%282%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20Moto%20Z%20%282%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859425095;-6;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;4;360;300;07/05/2020%2C%2008%3A50%3A25;24;640;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20moto%20g%286%29%20plus%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20moto%20g%286%29%20plus%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859482426;-8;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;7;480;420;5/7/2020%2C%206%3A51%3A22%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%209%3B%20LM-G810%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%209%3B%20LM-G810%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;ru-RU;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859549603;-8;07.06.2005%2C%2021%3A33%3A44;857;412;;;;;;;3;480;420;07.05.2020%2C%2006%3A52%3A29;24;857;412;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20moto%20g%286%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/76.0.3809.89%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv7l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20moto%20g%286%29%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/76.0.3809.89%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859640486;-8;6/7/2005%2C%209%3A33%3A44%20PM;720;360;;;;;;;8;480;420;5/7/2020%2C%206%3A54%3A00%20AM;24;720;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%208.0.0%3B%20SM-T820%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%208.0.0%3B%20SM-T820%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859727363;-8;6/7/2005%2C%209%3A33%3A44%20PM;1024;768;;;;;;;8;480;420;5/7/2020%2C%206%3A55%3A27%20AM;24;1024;768;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%206.0.1%3B%20D6503%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/77.0.3865.92%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv7l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%206.0.1%3B%20D6503%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/77.0.3865.92%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859782273;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;9;480;420;07/05/2020%2C%2006%3A56%3A22;24;640;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.1.1%3B%20ZTE%20A2017U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.1.1%3B%20ZTE%20A2017U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588859911116;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;7;480;420;07/05/2020%2C%2006%3A58%3A31;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%207.1.1%3B%20E6653%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%207.1.1%3B%20E6653%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-GB;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860030785;-8;07/06/2005%2C%2021%3A33%3A44;640;360;;;;;;;8;480;420;07/05/2020%2C%2007%3A00%3A30;24;640;360;0;0;;;;;;;;;;;;;;;;;;;22;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%2010%3B%20SM-F700U1%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%2010%3B%20SM-F700U1%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;false;false;1588860114721;9;6/7/2005%2C%209%3A33%3A44%20PM;879;360;;;;;;;3;-540;-540;5/7/2020%2C%2011%3A01%3A54%20PM;24;879;360;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Linux%3B%20Android%2010%3B%20SM-G965U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;20030107;undefined;true;;true;Linux%20armv8l;undefined;Mozilla/5.0%20%28Linux%3B%20Android%2010%3B%20SM-G965U%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.117%20Mobile%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860278674;-8;6/7/2005%2C%209%3A33%3A44%20PM;846;412;;;;;;;4;480;420;5/7/2020%2C%207%3A04%3A38%20AM;24;846;412;0;0;;;;;;;;;;;;;;;;;;;19;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_13_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_13_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588860333560;-5;6/7/2005%2C%209%3A33%3A44%20PM;1440;900;;;;;;;3;300;240;5/7/2020%2C%2010%3A05%3A33%20AM;24;1440;900;0;0;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_4%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_4%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588895251049;-5;6/7/2005%2C%209%3A33%3A44%20PM;1680;1050;;;;;;;2;300;240;5/7/2020%2C%207%3A47%3A31%20PM;24;1680;1027;0;23;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_14_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_14_6%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588909325098;-5;6/7/2005%2C%209%3A33%3A44%20PM;1280;800;;;;;;;2;300;240;5/7/2020%2C%2011%3A42%3A05%20PM;24;1280;735;0;23;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;20030107;undefined;true;;true;Win32;undefined;Mozilla/5.0%20%28Windows%20NT%2010.0%3B%20Win64%3B%20x64%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.138%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588965502432;-5;6/7/2005%2C%209%3A33%3A44%20PM;1920;1080;;;;;;;2;300;240;5/8/2020%2C%203%3A18%3A22%20PM;24;1920;1040;0;0;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_0%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_15_0%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588853176855;-8;6/7/2005%2C%209%3A33%3A44%20PM;1280;960;;;;;;;3;480;420;5/7/2020%2C%205%3A06%3A16%20AM;24;1280;937;0;23;;;;;;;;;;;;;;;;;;;18;`,
    `TF1;015;;;;;;;;;;;;;;;;;;;;;;Mozilla;Netscape;5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_10_5%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;20030107;undefined;true;;true;MacIntel;undefined;Mozilla/5.0%20%28Macintosh%3B%20Intel%20Mac%20OS%20X%2010_10_5%29%20AppleWebKit/537.36%20%28KHTML%2C%20like%20Gecko%29%20Chrome/81.0.4044.129%20Safari/537.36;en-US;undefined;fp.rushaio.co;undefined;undefined;undefined;undefined;true;true;1588851116852;-8;6/7/2005%2C%209%3A33%3A44%20PM;1600;1200;;;;;;;3;480;420;5/7/2020%2C%204%3A31%3A56%20AM;24;1600;1173;0;23;;;;;;;;;;;;;;;;;;;18;`,
  }
  det1 := dets[rand.Intn(len(dets))]
  fpSplit := strings.Split(det1, ";")
  fpSplit[36] = "www.hibbett.com"
  fpSplit[43] = fmt.Sprintf("%d", timeMillis())
  fpSplit[44] = "-8"
  fpSplit[55] = "480"
  fpSplit[56] = "420"
  pdt, err := time.LoadLocation("America/Los_Angeles")
  if err != nil {
    return err
  }
  fpSplit[57] = time.Now().In(pdt).Format("1/2/2006") + strings.Replace(url.QueryEscape(time.Now().In(pdt).Format(",3:04:05 PM")), "+", "%20", -1)
  det1 = strings.Join(fpSplit, ";")
  """
  dets = []

@app.route('/rd')
def rd():
  # TODO gen rd fp
  return ""

@app.route('/config')
@jwt_required
def config():
  return jsonify({
    "qbp": {
      "akavpau_CS": "1592663148~id=619c8a09febd004fe8d9a6f1e8996797",
      "akavpau_FA": "1592663148~id=d6123bd5d616d3c79f25b7c2d1c6f065",
      "akavpau_EB": "1592663148~id=86d9c6d3cf463b55591ee81f77dba24f",
      "akavpau_FL": "1592663148~id=a39bb8ca6839ca79e12fdd1a24dbc688",
    },
    "slimEndpoint": "__queue/yzysply",
    "rcAction": "yeezy_supply_wr_pageview_",
    "rcApiKey": "6Lcoe9wUAAAAAOtgfc4c6rnvgptxiBZwDBX3Tqvl",
    "rcCookie": "xhwUqgFqfW88H50",
    "sizeLookup": {
      "4K": "230",
      "5K": "240",
      "5.5K": "250",
      "6K": "260",
      "6.5K": "270",
      "7K": "280",
      "7.5K": "290",
      "8K": "310",
      "8.5K": "320",
      "9K": "330",
      "9.5K": "340",
      "10K": "350",
      "10.5K": "370",
      "11K": "380",
      "11.5K": "390",
      "12K": "410",
      "12.5K": "420",
      "13K": "430",
      "13.5K": "440",
      "1": "450",
      "1.5": "470",
      "2": "480",
      "2.5": "490",
      "3": "510",
      "3.5": "520",
      "4": "530",
      "4.5": "540",
      "5": "550",
      "5.5": "560",
      "6": "570",
      "6.5": "580",
      "7": "590",
      "7.5": "600",
      "8": "610",
      "8.5": "620",
      "9": "630",
      "9.5": "640",
      "10": "650",
      "10.5": "660",
      "11": "670",
      "11.5": "680",
      "12": "690",
      "12.5": "700",
      "13": "710",
      "13.5": "720",
      "14": "730",
      "14.5": "740",
      "15": "750",
      "16": "760",
      "17": "780",
   }
  })

if __name__ == '__main__':
  app.run(host='0.0.0.0',port=8081,debug=True,passthrough_errors=True,threaded=False,processes=1)
