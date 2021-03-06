import { prepare } from "https://deno.land/x/plugin_prepare/mod.ts";
export const PLUGIN_VERSION = "v0.1.1";
const releaseUrl = `https://github.com/justjavac/deno_plugin_num_cpus/releases/download/${PLUGIN_VERSION}`;
const pluginOptions = {
    name: "deno_plugin_num_cpus",
    urls: {
        linux: `${releaseUrl}/libdeno_plugin_num_cpus.so`,
        darwin: `${releaseUrl}/libdeno_plugin_num_cpus.dylib`,
        windows: `${releaseUrl}/deno_plugin_num_cpus.dll`
    }
};
let pluginId = null;
/**
 * Load the plugin
 */ async function load() {
    unload();
    pluginId = await prepare(pluginOptions);
}
/**
 * Free the plugin resource
 */ function unload() {
    if (pluginId !== null) Deno.close(pluginId);
    pluginId = null;
}
export default function num_cpus() {
    // deno-lint-ignore ban-ts-comment
    //@ts-ignore
    const { op_num_cpus  } = Deno.core.ops();
    // deno-lint-ignore ban-ts-comment
    //@ts-ignore
    const response = Deno.core.dispatch(op_num_cpus);
    return response[0];
};
await load();
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC94L251bV9jcHVzQHYwLjEuNy9tb2QudHM+Il0sInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7XG4gIHByZXBhcmUsXG4gIFBlcnBhcmVPcHRpb25zLFxufSBmcm9tIFwiaHR0cHM6Ly9kZW5vLmxhbmQveC9wbHVnaW5fcHJlcGFyZS9tb2QudHNcIjtcblxuZXhwb3J0IGNvbnN0IFBMVUdJTl9WRVJTSU9OID0gXCJ2MC4xLjFcIjtcbmNvbnN0IHJlbGVhc2VVcmwgPVxuICBgaHR0cHM6Ly9naXRodWIuY29tL2p1c3RqYXZhYy9kZW5vX3BsdWdpbl9udW1fY3B1cy9yZWxlYXNlcy9kb3dubG9hZC8ke1BMVUdJTl9WRVJTSU9OfWA7XG5cbmNvbnN0IHBsdWdpbk9wdGlvbnM6IFBlcnBhcmVPcHRpb25zID0ge1xuICBuYW1lOiBcImRlbm9fcGx1Z2luX251bV9jcHVzXCIsXG4gIHVybHM6IHtcbiAgICBsaW51eDogYCR7cmVsZWFzZVVybH0vbGliZGVub19wbHVnaW5fbnVtX2NwdXMuc29gLFxuICAgIGRhcndpbjogYCR7cmVsZWFzZVVybH0vbGliZGVub19wbHVnaW5fbnVtX2NwdXMuZHlsaWJgLFxuICAgIHdpbmRvd3M6IGAke3JlbGVhc2VVcmx9L2Rlbm9fcGx1Z2luX251bV9jcHVzLmRsbGAsXG4gIH0sXG59O1xuXG5sZXQgcGx1Z2luSWQ6IG51bWJlciB8IG51bGwgPSBudWxsO1xuXG4vKipcbiAqIExvYWQgdGhlIHBsdWdpblxuICovXG5hc3luYyBmdW5jdGlvbiBsb2FkKCkge1xuICB1bmxvYWQoKTtcbiAgcGx1Z2luSWQgPSBhd2FpdCBwcmVwYXJlKHBsdWdpbk9wdGlvbnMpO1xufVxuXG4vKipcbiAqIEZyZWUgdGhlIHBsdWdpbiByZXNvdXJjZVxuICovXG5mdW5jdGlvbiB1bmxvYWQoKTogdm9pZCB7XG4gIGlmIChwbHVnaW5JZCAhPT0gbnVsbCkgRGVuby5jbG9zZShwbHVnaW5JZCk7XG4gIHBsdWdpbklkID0gbnVsbDtcbn1cblxuLyoqXG4gKiBHZXQgdGhlIG51bWJlciBvZiBDUFVzIGF2YWlsYWJsZSBvbiB0aGUgY3VycmVudCBzeXN0ZW0uXG4gKiBcbiAqICMjIEV4YW1wbGVcbiAqIFxuICogYGBgdHNcbiAqIGBgYFxuICovXG5leHBvcnQgZGVmYXVsdCBmdW5jdGlvbiBudW1fY3B1cygpOiBudW1iZXIge1xuICAvLyBkZW5vLWxpbnQtaWdub3JlIGJhbi10cy1jb21tZW50XG4gIC8vQHRzLWlnbm9yZVxuICBjb25zdCB7IG9wX251bV9jcHVzIH0gPSBEZW5vLmNvcmUub3BzKCk7XG4gIC8vIGRlbm8tbGludC1pZ25vcmUgYmFuLXRzLWNvbW1lbnRcbiAgLy9AdHMtaWdub3JlXG4gIGNvbnN0IHJlc3BvbnNlOiBVaW50OEFycmF5ID0gRGVuby5jb3JlLmRpc3BhdGNoKG9wX251bV9jcHVzKSE7XG4gIHJldHVybiByZXNwb25zZVswXTtcbn1cblxuYXdhaXQgbG9hZCgpO1xuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJTQUNBLE9BQUEsU0FFQSx5Q0FBQTthQUVBLGNBQUEsSUFBQSxNQUFBO01BQ0EsVUFBQSxJQUNBLG9FQUFBLEVBQUEsY0FBQTtNQUVBLGFBQUE7QUFDQSxRQUFBLEdBQUEsb0JBQUE7QUFDQSxRQUFBO0FBQ0EsYUFBQSxLQUFBLFVBQUEsQ0FBQSwyQkFBQTtBQUNBLGNBQUEsS0FBQSxVQUFBLENBQUEsOEJBQUE7QUFDQSxlQUFBLEtBQUEsVUFBQSxDQUFBLHlCQUFBOzs7SUFJQSxRQUFBLEdBQUEsSUFBQTtBQUVBLEVBRUEsQUFGQSxzQkFFQSxBQUZBLEVBRUEsZ0JBQ0EsSUFBQTtBQUNBLFVBQUE7QUFDQSxZQUFBLFNBQUEsT0FBQSxDQUFBLGFBQUE7O0FBR0EsRUFFQSxBQUZBLCtCQUVBLEFBRkEsRUFFQSxVQUNBLE1BQUE7UUFDQSxRQUFBLEtBQUEsSUFBQSxFQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsUUFBQTtBQUNBLFlBQUEsR0FBQSxJQUFBOzt3QkFXQSxRQUFBO0FBQ0EsTUFBQSxnQ0FBQTtBQUNBLE1BQUEsVUFBQTtZQUNBLFdBQUEsTUFBQSxJQUFBLENBQUEsSUFBQSxDQUFBLEdBQUE7QUFDQSxNQUFBLGdDQUFBO0FBQ0EsTUFBQSxVQUFBO1VBQ0EsUUFBQSxHQUFBLElBQUEsQ0FBQSxJQUFBLENBQUEsUUFBQSxDQUFBLFdBQUE7V0FDQSxRQUFBLENBQUEsQ0FBQTs7TUFHQSxJQUFBIn0=