#!/usr/bin/env bash
DENO_DIR=./deno HTTPS_PROXY="http://localhost:8889" deno run --no-check  --unstable --allow-all --cert /Users/aka/drive/mitm/charles-ssl-proxying-certificate.pem geeTestSolver.js  2>&1
