import { solveGeetest } from './geeTestSolver.js';
import { Buffer } from "http://deno.land/x/node_buffer/index.ts";
import { decode as b64decode } from 'https://deno.land/std@0.76.0/encoding/base64.ts';
globalThis.Buffer = Buffer;

var reqs = {};
var xhrProxy = async function(req) {
  try {
    var reqPromise = new Promise((resolve, reject) => {
      var tout = setTimeout(reject, 15000);
      reqs[req.url] = (r) => {
        clearTimeout(tout);
        resolve(r);
      }
    });
    setTimeout(() => self.postMessage(req), 10);
    return (await reqPromise);
    // console.log('XHR REQ', req)
  } catch(e) {
    console.error('worker xhrProxy', e);
    return {};
  }
}
globalThis.xhrProxy = xhrProxy;
self.onmessage = async (e) => {
  try {
    // console.log('worker onmessage', JSON.stringify(e.data));
    if (typeof e.data === 'string') {
      try {
      reqs = {};
      self.postMessage(await solveGeetest(e.data, xhrProxy));
      } catch(e) {
      self.postMessage({error: 1, message: e && e.message, stack: e && e.stack, str: e && e.toString()});
      }
    } else if (e.data && typeof e.data === 'object' && e.data.response) {
      var h = reqs[e.data.response.url];
      e.data.responseBuffer = b64decode(e.data.responseBuffer).buffer;
      // assert(e.data.responseBuffer);
      // console.log('ws to xhr prox: ', e.data);
      h && h(e.data);
      if (!h) {
        console.warn('no handler for resp: ', e.data.response.url);
        console.warn('active: ' + Array.from(Object.keys(reqs)).join("\n"));
      }
    }
  } catch(e) {
    console.error('worker onmessage err', e);
  }
}