import { logDebug } from './log.js';
const decoder = new TextDecoder('utf-8');
const encoder = new TextEncoder('utf-8');

async function handle(conn) {
   try {
      const lbuf = new Uint8Array(4);
      await conn.read(lbuf, 4);
      logDebug('lbuf', lbuf);
      var len = (lbuf[3] << 24) + (lbuf[2] << 16) + (lbuf[1] << 8) + lbuf[0];
      if (len > 1e6)  {
        throw new Error('max len exceeed');
      }
      logDebug('len', len);
      const buf = new Uint8Array(len);
      await conn.read(buf, len);
      const str = decoder.decode(buf);
      logDebug('reqstr', str);
      const req = JSON.parse(str);
      logDebug('req', req);
      const worker = new Worker(new URL("worker.js", import.meta.url).href, { type: "module", deno: true });
      worker.postMessage(req);
      let setWorkerDone;
      let promise = new Promise((resolve, reject) => {
        setWorkerDone = resolve;
      })
      worker.onmessage = async (e) => {
        console.log('RECV', e.data)
        if (typeof e.data === 'string') {
          const cenc = encoder.encode(e.data);
          for (var i = 0; i < cenc.length; i+=256) {
            await conn.write(cenc.slice(i, i+256));
          }
          logDebug('wrote all');
        }
        setWorkerDone();
      }
      await Promise.race([promise, delay(30000)]);
    } catch(err) {
      console.warn('conn error', err);
    } finally {
      logDebug('closing');
      await conn.close();
    }
    logDebug('conn done');
}

const filePath = Deno.cwd() + "/px.sock";
try {
  const listener = Deno.listen({ port: 40400, host: "127.0.0.1",
    // path: filePath,
    // transport: "unix"
  });
  console.log(`listening at ${filePath}`)
  var conns = [];

  for await (const conn of listener) {
    logDebug('conn');
    handle(conn);
  }
} finally {
  try {
    await Deno.remove(filePath);
  } catch{}
}

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}
