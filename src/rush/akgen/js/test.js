import { gen } from './gen.js';

try {
  console.log(await gen(parseInt(Deno.args[0] || 0), Deno.args[1] || ""))
  Deno.exit(0);
} catch(e) {
  console.error(e);
  Deno.exit(1);
}
