var DEBUG = Deno.env.get("DEBUG") === "1";
export function logDebug() {
  if (DEBUG) {
    console.log(...arguments);
  }
}''
globalThis.logDebug = logDebug;
