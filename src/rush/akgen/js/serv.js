import { gen } from './gen.js';

const decoder = new TextDecoder('utf-8');
const encoder = new TextEncoder('utf-8');


const filePath = Deno.cwd() + "/ak.sock";
try {
  // const listener = Deno.listen({
  //   path: filePath,
  //   transport: "unix"
  // });
  const listener = Deno.listen({
    // /port: parseInt(Deno.args[0] || 40401),
    port: 0,
    host: "127.0.0.1",
  });
  const {port} = listener.addr;

  console.log(`listening at 127.0.0.1:${port}`)
  var conns = [];

  for await (const conn of listener) {
    (async () => {
      try {
        const buf = new Uint8Array(8192);
        await conn.read(buf, 8192);
        const [tsOffset, abck] = decoder.decode(buf).split('~_~');
        const sd = await gen(parseInt(tsOffset || 0), abck || "");
        const sdenc = encoder.encode(sd);
        for (var i = 0; i < sdenc.length; i+=256) {
          await conn.write(sdenc.slice(i, i+256));
        }
        await conn.close();
      } catch(err) {
        console.warn('conn error', err)
      }
    })()
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