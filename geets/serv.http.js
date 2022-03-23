import { serve } from "https://deno.land/std@0.79.0/http/server.ts";
import { delay } from './util.js';
import num_cpus from "https://deno.land/x/num_cpus/mod.ts";


// TODO: use task proxy

const WORKERS = parseInt(Deno.env.get('WORKERS') || Math.max(1, num_cpus() / 2));
const workers = new Array(WORKERS);
for (var i = 0; i < WORKERS; i++) {
  workers[i] = new Worker(new URL(Deno.env.get("WORKER_JS", "worker.bundle.js"), import.meta.url).href, { type: "module", deno: true });
}
const s = serve({ port: 1996, host: '127.0.0.1' });
console.log(`${WORKERS} workers online at http://localhost:1996/`);

for await (const req of s) {
  (async () => {
    var url = req.headers.get('url', '');
    if (!url || url.length < 6) {
      req.respond({status: 400, body: 'invalid url'});
      return
    }
    console.log(workers)
    var worker = workers.pop();
    if (!worker) {
      req.respond({status: 503, body: 'all workers busy'});
      return
    }
    try {
      worker.postMessage(req.headers.get('url'));
      let setWorkerDone;
      let promise = new Promise((resolve, reject) => {
        setWorkerDone = resolve;
      })
      worker.onmessage = async (e) => {
        console.log('RECV', e.data)
        if (!('error' in e.data)) {
          req.respond({body: JSON.stringify(e.data)});
        }
        setWorkerDone();
      }
      await Promise.race([promise, delay(10e3)]);
    } catch(e) {
      req.respond({status: 500, body: e.toString()});
    } finally {
      worker.onmessage = null;
      workers.push(worker);
    }
  })()
}

// import { serve } from "https://deno.land/std@0.75.0/http/server.ts";
// import {
//   acceptWebSocket,
//   isWebSocketCloseEvent,
//   isWebSocketPingEvent,
// } from "https://deno.land/std@0.75.0/ws/mod.ts";
// import { gen } from './gen.js';

// const encoder = new TextEncoder();
// const decoder = new TextDecoder();
// function typedArrayToBuffer(array)  {
//     return array.buffer.slice(array.byteOffset, array.byteLength + array.byteOffset)
// }

// const DEBUG = Deno.env.get("DEBUG") === "1";
// function logDebug(){
//   if (DEBUG) {
//     console.log(...arguments);
//   }
// }
// async function handleWs(sock, headers) {
//   logDebug("socket connected!", headers, sock);
//   logDebug(headers.get('x-url'));
//   logDebug(headers.get('x-ref'));
//   logDebug(headers.get('x-cid'));
//   var xhrRespResolve, xhrRespReject, donePromiseResolve;
//   var donePromise = new Promise((resolve) => donePromiseResolve = resolve);
//   try {
//     var xhrProxy = async (request) => {
//       logDebug('xhrSend', request);
//       request.body = decoder.decode(request.body);
//       await sock.send(JSON.stringify(request));
//       await delay(1000);
//       throw new Error('Halt');
//     };
//     await gen(headers.get('x-url'), headers.get('x-ref') || undefined, headers.get('x-cid') || undefined, xhrProxy, donePromise, headers.get('x-ua') || undefined);
//   } catch (err) {
//     logDebug(`failed to receive frame: ${err}`);
//   } finally {
//     donePromiseResolve();
//     if (!sock.isClosed) {
//       await sock.close(1000).catch(console.error);
//     }
//   }
// }

// async function main() {
//   const port = Deno.args[0] || "25144";
//   console.log(`websocket server is running on :${port}`);
//   for await (const req of serve(`127.0.0.1:${port}`)) {
//     var c;
//     try {
//       const { conn, r: bufReader, w: bufWriter, headers } = req;
//       acceptWebSocket({
//         conn,
//         bufReader,
//         bufWriter,
//         headers,
//       })
//         .then((sock) => handleWs(sock, headers))
//         .catch(async (err) => {
//           console.error(`failed to accept websocket: ${err}`);
//           await req.respond({ status: 400 });
//         });
//       } catch(e) {
//         console.error(e);
//       } finally {
//       }
//   }
//   Deno.exit(0);
// }

// main();