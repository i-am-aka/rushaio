#!/bin/bash
cd "$(dirname "$0")"
  # --worker-class gevent \
  # --worker-connections 2048 \
gunicorn \
  -w $(nproc) \
  --worker-class gevent \
  --worker-connections 512 \
  --error-logfile - \
  --access-logfile - \
  --access-logformat '%({X-Forwarded-For}i)s %(l)s %(u)s %(t)s "%(r)s" %(s)s %(b)s "%(f)s" "%(a)s"' \
  --forwarded-allow-ips '*' \
  -b 0.0.0.0:8081 \
  server:app
