#!/usr/bin/env python
import datetime
import sys
from cloudsocket.config import JWT_KEY
import jwt
print(jwt.encode({'exp': datetime.datetime.utcnow() + datetime.timedelta(seconds=86400*60), sys.argv[1]: '1'}, JWT_KEY, algorithm='HS256').decode())

