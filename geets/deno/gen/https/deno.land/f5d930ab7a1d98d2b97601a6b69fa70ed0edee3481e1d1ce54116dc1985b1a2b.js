// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/**
 * Test whether or not the given path exists by checking with the file system
 */ export async function exists(filePath) {
    try {
        await Deno.lstat(filePath);
        return true;
    } catch (err) {
        if (err instanceof Deno.errors.NotFound) {
            return false;
        }
        throw err;
    }
}
/**
 * Test whether or not the given path exists by checking with the file system
 */ export function existsSync(filePath) {
    try {
        Deno.lstatSync(filePath);
        return true;
    } catch (err) {
        if (err instanceof Deno.errors.NotFound) {
            return false;
        }
        throw err;
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL2ZzL2V4aXN0cy50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cbi8qKlxuICogVGVzdCB3aGV0aGVyIG9yIG5vdCB0aGUgZ2l2ZW4gcGF0aCBleGlzdHMgYnkgY2hlY2tpbmcgd2l0aCB0aGUgZmlsZSBzeXN0ZW1cbiAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIGV4aXN0cyhmaWxlUGF0aDogc3RyaW5nKTogUHJvbWlzZTxib29sZWFuPiB7XG4gIHRyeSB7XG4gICAgYXdhaXQgRGVuby5sc3RhdChmaWxlUGF0aCk7XG4gICAgcmV0dXJuIHRydWU7XG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGlmIChlcnIgaW5zdGFuY2VvZiBEZW5vLmVycm9ycy5Ob3RGb3VuZCkge1xuICAgICAgcmV0dXJuIGZhbHNlO1xuICAgIH1cblxuICAgIHRocm93IGVycjtcbiAgfVxufVxuXG4vKipcbiAqIFRlc3Qgd2hldGhlciBvciBub3QgdGhlIGdpdmVuIHBhdGggZXhpc3RzIGJ5IGNoZWNraW5nIHdpdGggdGhlIGZpbGUgc3lzdGVtXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBleGlzdHNTeW5jKGZpbGVQYXRoOiBzdHJpbmcpOiBib29sZWFuIHtcbiAgdHJ5IHtcbiAgICBEZW5vLmxzdGF0U3luYyhmaWxlUGF0aCk7XG4gICAgcmV0dXJuIHRydWU7XG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGlmIChlcnIgaW5zdGFuY2VvZiBEZW5vLmVycm9ycy5Ob3RGb3VuZCkge1xuICAgICAgcmV0dXJuIGZhbHNlO1xuICAgIH1cbiAgICB0aHJvdyBlcnI7XG4gIH1cbn1cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxFQUFBLHdFQUFBO0FBQ0EsRUFFQSxBQUZBLGlGQUVBLEFBRkEsRUFFQSx1QkFDQSxNQUFBLENBQUEsUUFBQTs7Y0FFQSxJQUFBLENBQUEsS0FBQSxDQUFBLFFBQUE7ZUFDQSxJQUFBO2FBQ0EsR0FBQTtZQUNBLEdBQUEsWUFBQSxJQUFBLENBQUEsTUFBQSxDQUFBLFFBQUE7bUJBQ0EsS0FBQTs7Y0FHQSxHQUFBOzs7QUFJQSxFQUVBLEFBRkEsaUZBRUEsQUFGQSxFQUVBLGlCQUNBLFVBQUEsQ0FBQSxRQUFBOztBQUVBLFlBQUEsQ0FBQSxTQUFBLENBQUEsUUFBQTtlQUNBLElBQUE7YUFDQSxHQUFBO1lBQ0EsR0FBQSxZQUFBLElBQUEsQ0FBQSxNQUFBLENBQUEsUUFBQTttQkFDQSxLQUFBOztjQUVBLEdBQUEifQ==