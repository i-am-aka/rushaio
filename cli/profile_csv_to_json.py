#!/usr/bin/env python

import json
from rush import read_profiles_from_file
import sys
print(json.dumps(read_profiles_from_file(sys.argv[1])))

# read_profiles_from_file
