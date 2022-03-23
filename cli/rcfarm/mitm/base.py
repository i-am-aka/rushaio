from urllib.parse import urlparse
import mitmproxy
from mitmproxy import ctx, http
from mitmproxy.net import http as nhttp
import traceback
import time
import random

proxy_data = None
auth = None
try:
  proxy_data = urlparse(PROXY)
  if proxy_data.username:
    auth = proxy_data.username + ":" + proxy_data.password
  else:
    auth = None
except:
  pass
  # traceback.print_exc()

def http_connect(flow: http.HTTPFlow):
  ctx.log.info(f"Connect {flow.request.host}")
  if flow.request.host in ('recaptcha-demo.appspot.com', 'www.finishline.com' , 'www.jdsports.com', 'www.yeezysupply.com' ,'score.rushaio.co' ,'rushaio.co'):
    flow.response = http.make_connect_response(flow.request.data.http_version)
    return
  # elif flow.request.host not in ('amiunique.org', 'www.google.com', 'www.gstatic.com', 'www.fonts.gstatic.com', 'fonts.gstatic.com', 'google.com', 'gstatic.com', 'ipinfo.io'):
  #   flow.live.send_error_response(404, "Unsupported Host")
  #   return

  # ctx.log.info(f"{flow.request}, {flow.request.method}, {flow}")
  # if proxy_data:
  #   if auth:
  #     ctx.options.upstream_auth = auth
  #   flow.live.change_upstream_proxy_server((proxy_data.hostname, proxy_data.port))
  #   if not flow.live.server_conn.connected():
  #     flow.live.connect()

  # ctx.log.info(f"{flow.request}, {flow.request.method}, {flow}")

  # flow.server_conn.address
  # if flow.live:
  #   ctx.log.info(flow.live)
def request(flow: http.HTTPFlow):

  ctx.log.info(f"Request {flow.request.host} {flow.live.server_conn if flow.live else None} {flow.live.server_conn.via if flow.live and flow.live.server_conn and flow.live.server_conn.via else None}")
  ctx.log.info(flow.request.host)
  if flow.request.host in ('recaptcha-demo.appspot.com', 'www.finishline.com', 'www.jdsports.com', 'www.yeezysupply.com' ,'score.rushaio.co' ,'rushaio.co'):
    if flow.request.method == 'GET':
      flow.response = http.HTTPResponse("HTTP/2", 200, b"", nhttp.Headers(Content_Type="text/html",Date="Fri, 02 Oct 2020 05:02:37 GMT",Server="AmazonS3"), b"<html><head></head><body></body></html>")

    # elif not (flow.request.path.startswith('/token')):
    #   flow.response = http.HTTPResponse("HTTP/2", 404, b"", nhttp.Headers(Content_Type="text/html",Date="Fri, 02 Oct 2020 05:02:37 GMT",Server="AmazonS3"), b"<html><head></head><body></body></html>")
    # flow.response.reason = b"Connection established"

#   ctx.log.info(flow.request, flow.request.method)
#   ctx.log.info(flow.request.host)
#   ctx.log.info(flow.request.port)
#   if flow.request.method == "CONNECT":
#     return

  # if not proxy_data:
  #   return

  # ctx.log.info(proxy_data.hostname, proxy_data.port, auth)
  # if flow.live:
  #   # if flow.live.se
  #   ctx.log.info(flow.live.server_conn, flow.live.server_conn.via, flow.live.server_conn.via.address)
  #   if auth:
  #     ctx.options.upstream_auth = auth
  #   flow.live.change_upstream_proxy_server((proxy_data.hostname, proxy_data.port))
  #   # flow.live.connect()

def response(flow: http.HTTPFlow):
  ctx.log.info(f'response {flow.request.host}')
  # if flow.request.host in ('www.finishline.com' ,'www.yeezysupply.com' ,'score.rushaio.co' ,'rushaio.co'):
  #   flow.response = http.HTTPResponse.make(200, b"Graduation", {"Content-Type": "text/html"})

def error(flow: http.HTTPFlow):
  ctx.log.info(f'error {flow.error} {flow.response}')
  # if flow.response.host in ('www.finishline.com' ,'www.yeezysupply.com' ,'score.rushaio.co' ,'rushaio.co'):
  #   flow.response = http.HTTPResponse.make(200, b"Graduation", {"Content-Type": "text/html"})

# def serverdisconnect(conn: mitmproxy.connections.ServerConnection):
#   ctx.log.info(f'serverdisconnect {conn.address}')
