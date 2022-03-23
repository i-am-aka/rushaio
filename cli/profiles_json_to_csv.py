#!/usr/bin/env pipenv run python

import json
from rush import profiles_to_csv
profs = []
import sys
with open(sys.argv[1]) as f:
  for p in f.read().split('\n'):
    p = p.strip()
    if not p:
      continue
    prof = json.loads(p)
    profs.append(prof)

print(profiles_to_csv(profs))
