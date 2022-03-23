var nt = "=";
var et = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
function n(t) {
    var e, n, i = "";
    for (e = 0; e + 3 <= t.length; e += 3)
        n = parseInt(t.substring(e, e + 3), 16),
        i += et.charAt(n >> 6) + et.charAt(63 & n);
    for (e + 1 == t.length ? (n = parseInt(t.substring(e, e + 1), 16),
    i += et.charAt(n << 2)) : e + 2 == t.length && (n = parseInt(t.substring(e, e + 2), 16),
    i += et.charAt(n >> 2) + et.charAt((3 & n) << 4)); (3 & i.length) > 0; )
        i += nt;
    return i
}
console.log(n(Deno.args[0].trim()));
