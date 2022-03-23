var encode = function(e, t) {
    var p = new Array(8).fill(0);
    var l = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567".split("");
    var n = "string" != typeof e;
    return n && e.constructor === ArrayBuffer && (e = new Uint8Array(e)),
    n ? function(e) {
        for (var t, n, r, o, a, i = "", u = e.length, s = 0, c = 5 * parseInt(u / 5); s < c; )
            t = e[s++],
            n = e[s++],
            r = e[s++],
            o = e[s++],
            a = e[s++],
            i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[31 & (o << 3 | a >>> 5)] + l[31 & a];
        var p = u - c;
        return 1 === p ? (t = e[s],
        i += l[t >>> 3] + l[t << 2 & 31] + "======") : 2 === p ? (t = e[s++],
        n = e[s],
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[n << 4 & 31] + "====") : 3 === p ? (t = e[s++],
        n = e[s++],
        r = e[s],
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[r << 1 & 31] + "===") : 4 === p && (t = e[s++],
        n = e[s++],
        r = e[s++],
        o = e[s],
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[o << 3 & 31] + "="),
        i
    }(e) : t ? function(e) {
        for (var t, n, r, o, a, i = "", u = e.length, s = 0, c = 5 * parseInt(u / 5); s < c; )
            t = e.charCodeAt(s++),
            n = e.charCodeAt(s++),
            r = e.charCodeAt(s++),
            o = e.charCodeAt(s++),
            a = e.charCodeAt(s++),
            i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[31 & (o << 3 | a >>> 5)] + l[31 & a];
        var p = u - c;
        return 1 === p ? (t = e.charCodeAt(s),
        i += l[t >>> 3] + l[t << 2 & 31] + "======") : 2 === p ? (t = e.charCodeAt(s++),
        n = e.charCodeAt(s),
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[n << 4 & 31] + "====") : 3 === p ? (t = e.charCodeAt(s++),
        n = e.charCodeAt(s++),
        r = e.charCodeAt(s),
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[r << 1 & 31] + "===") : 4 === p && (t = e.charCodeAt(s++),
        n = e.charCodeAt(s++),
        r = e.charCodeAt(s++),
        o = e.charCodeAt(s),
        i += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[o << 3 & 31] + "="),
        i
    }(e) : function(e) {
        var t, n, r, o, a, i, u, s = !1, c = "", h = 0, d = 0, f = e.length;
        do {
            for (p[0] = p[5],
            p[1] = p[6],
            p[2] = p[7],
            u = d; h < f && u < 5; ++h)
                (i = e.charCodeAt(h)) < 128 ? p[u++] = i : i < 2048 ? (p[u++] = 192 | i >> 6,
                p[u++] = 128 | 63 & i) : i < 55296 || i >= 57344 ? (p[u++] = 224 | i >> 12,
                p[u++] = 128 | i >> 6 & 63,
                p[u++] = 128 | 63 & i) : (i = 65536 + ((1023 & i) << 10 | 1023 & e.charCodeAt(++h)),
                p[u++] = 240 | i >> 18,
                p[u++] = 128 | i >> 12 & 63,
                p[u++] = 128 | i >> 6 & 63,
                p[u++] = 128 | 63 & i);
            d = u - 5,
            h === f && ++h,
            h > f && u < 6 && (s = !0),
            t = p[0],
            u > 4 ? (n = p[1],
            r = p[2],
            o = p[3],
            a = p[4],
            c += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[31 & (o << 3 | a >>> 5)] + l[31 & a]) : 1 === u ? c += l[t >>> 3] + l[t << 2 & 31] + "======" : 2 === u ? (n = p[1],
            c += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[n << 4 & 31] + "====") : 3 === u ? (n = p[1],
            r = p[2],
            c += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[r << 1 & 31] + "===") : 4 === u && (n = p[1],
            r = p[2],
            o = p[3],
            c += l[t >>> 3] + l[31 & (t << 2 | n >>> 6)] + l[n >>> 1 & 31] + l[31 & (n << 4 | r >>> 4)] + l[31 & (r << 1 | o >>> 7)] + l[o >>> 2 & 31] + l[o << 3 & 31] + "=")
        } while (!s);return c
    }(e)
}
console.log(encode((new Date).toISOString().slice(11, 19).replace("T", ".")).replace(/[^a-zA-Z0-9]/g, "").toLowerCase())
