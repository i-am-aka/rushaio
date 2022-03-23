// TODO use worker pool
    //  wire up xhr proxy
    // impl xhr proxy in cap verify

import { serve } from "https://deno.land/std@0.75.0/http/server.ts";
import num_cpus from "https://deno.land/x/num_cpus/mod.ts";
import {
  acceptWebSocket,
  isWebSocketCloseEvent,
  isWebSocketPingEvent,
} from "https://deno.land/std@0.75.0/ws/mod.ts";
import { delay } from './util.js';


const encoder = new TextEncoder();
const decoder = new TextDecoder();

const DEBUG = Deno.env.get("DEBUG") === "1";
function logDebug(){
  if (DEBUG) {
    console.log(...arguments);
  }
}

const WORKERS = parseInt(Deno.env.get('WORKERS') || Math.max(1, num_cpus()*4));
const workers = new Array(WORKERS);
for (var i = 0; i < WORKERS; i++) {
  workers[i] = new Worker(new URL(Deno.env.get("WORKER_JS") || "worker.js", import.meta.url).href, { type: "module", deno: true });
}

async function handleWs(worker, sock, headers) {

  logDebug("socket connected!", headers, sock);
  var reqs = {};
  try {
    var url = headers.get('url');
    if (!url || url.length < 5) {
      return;
    }
    var xhrProxy = async (request) => {
      logDebug('xhrSend', request);
      request.body = decoder.decode(request.body);
      var reqPromise = new Promise((resolve, reject) => {
        reqs[request.url] = resolve;
        // setTimeout(resolve, 15000);;
      });
      await sock.send(JSON.stringify(request));
      const resp = await Promise.race([reqPromise, delay(10e3)]);
      if (!resp || sock.isClosed) {
        return null;
      }
      Array.from(Object.keys(resp.response.headers)).forEach(key => {
        resp.response.headers[key] = resp.response.headers[key][0].toLowerCase();
      });
      logDebug('RESP : ', resp)
      logDebug('RESP : ', resp)
      logDebug('len : ', resp.responseBuffer.length)
      return resp;
    };

    worker.postMessage(headers.get('url'));
    let setWorkerDone;
    let promise = new Promise((resolve, reject) => {
      setWorkerDone = resolve;
    })
    worker.onmessage = async (e) => {
      try {
        logDebug('RECV', e.data)
        if (!('error' in e.data) && !('url' in e.data)) {
          await sock.send(JSON.stringify(e.data));
          setWorkerDone();
        } else if ('url' in e.data)  {
          xhrProxy(e.data).then((resp) => worker.postMessage(resp)).catch(e => console.error(e))
        }
      } catch(e) {
        console.error(e);
      }
    }
    for await (const ev of sock) {
      logDebug(ev);
      if (typeof ev === "string") {
        var resp = JSON.parse(ev);
        reqs[resp.response.url] && reqs[resp.response.url](resp);
      }
    }
    await Promise.race([promise, delay(15e3)]);

    // await gen(headers.get('x-url'), headers.get('x-ref') || undefined, headers.get('x-cid') || undefined, xhrProxy, donePromise, headers.get('x-ua') || undefined);
  } catch (err) {
    logDebug(`failed to receive frame: ${err}`);
  } finally {
    try{
      worker.onmessage = null;
      workers.push(worker);
      if (!sock.isClosed) {
        await sock.close(1000).catch(console.error);
      }
    }catch(e)  {console.error(e)}
  }
}

async function main() {
  const port = Deno.args[0] || "1997";
  console.log(`websocket server is running on :${port}`);

  for await (const req of serve(`:${port}`)) {
    var c;
    try {
      const { conn, r: bufReader, w: bufWriter, headers } = req;
      console.log(headers)
      if (!headers.get('url') || headers.get('url') === '') {
        req.respond({status: 400, body: 'hi'});
        continue
      }
      logDebug(workers)
      var worker = workers.pop();
      logDebug('got worker', worker)
      if (!worker) {
        req.respond({status: 503, body: 'all workers busy'});
        continue
      }
      acceptWebSocket({
        conn,
        bufReader,
        bufWriter,
        headers,
      })
        .then((sock) => handleWs(worker, sock, headers))
        .catch(async (err) => {
          console.error(`failed to accept websocket: ${err}`);
          await req.respond({ status: 400 });
        });
      } catch(e) {
        console.error(e);
      } finally {

      }
  }
  Deno.exit(0);
}

main();