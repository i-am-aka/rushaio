globalThis.navigator = {};
// import jsdom from 'https://dev.jspm.io/jsdom';
// const dom = new jsdom.JSDOM(``, {
//   url: "https://www.yeezysupply.com/payment",
//   referrer: "https://www.yeezysupply.com/delivery",
//   contentType: "text/html",
//   includeNodeLocations: false,
// });
// // window.location = dom.window.location;
// // window.location.origin = dom.window.location.origin;
// window.document = dom.document;
// globalThis.document = dom.document;
// globalThis.navigator = dom.window.navigator;

// // import * as Storage from './storage.ts';
// globalThis.sessionStorage = {};
// window.sessionStorage = globalThis.sessionStorage;
// globalThis.navigator = window.navigator;

import './datastore.js';
const adyenPublicKey = window.DATA_STORE.app.config.adyenPublicKey;
// import Runtime from './runtime.js';
// Runtime();
// import './vendor.app.js';

// import initApp from './app.app.js';
// initApp(dom.window);
// import './vendors~chk-delivery~chk-payment~chk-payment-callback~frontend-c.js';
// import Adyen from './df.js';
// Adyen(dom.window, dom.window.document);


/*

type Card struct {
  Type string `json:"cardType"`
  Number string `json:"-"`
  Cvc string `json:"-"`
  Name string `json:"holder"`
  ExpMonth int `json:"expirationMonth"`
  ExpYear int `json:"expirationYear"`
  PaymentMethodId string `json:"paymentMethodId"`
  LastFour string `json:"lastFour"`
}
*/

var AdyenEnc = function(t, r) {
  var e = {exports:{}};
                ! function() {
                    try {
                        new Uint8Array(1), new Uint32Array(1), new Int32Array(1);
                        return
                    } catch (e) {}

                    function e(e, t) {
                        return this.slice(e, t)
                    }

                    function t(e, t) {
                        arguments.length < 2 && (t = 0);
                        for (var n = 0, o = e.length; n < o; ++n, ++t) this[t] = 255 & e[n]
                    }

                    function n(n) {
                        var o;
                        if ("number" == typeof n) {
                            o = new Array(n);
                            for (var r = 0; r < n; ++r) o[r] = 0
                        } else o = n.slice(0);
                        return o.subarray = e, o.buffer = o, o.byteLength = o.length, o.set = t, "object" == typeof n && n.buffer && (o.buffer = n.buffer), o
                    }
                    try {
                        window.Uint8Array = n
                    } catch (e) {}
                    try {
                        window.Uint32Array = n
                    } catch (e) {}
                    try {
                        window.Int32Array = n
                    } catch (e) {}
                }(),
                function() {
                    try {
                        if ("undefined" == typeof window) return;
                        if ("btoa" in window) return;
                        var e = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
                        window.btoa = function(t) {
                            var n, o, r = "";
                            for (n = 0, o = t.length; n < o; n += 3) {
                                var s = 255 & t.charCodeAt(n),
                                    i = 255 & t.charCodeAt(n + 1),
                                    a = 255 & t.charCodeAt(n + 2),
                                    u = (3 & s) << 4 | i >> 4,
                                    l = n + 1 < o ? (15 & i) << 2 | a >> 6 : 64,
                                    d = n + 2 < o ? 63 & a : 64;
                                r += e.charAt(s >> 2) + e.charAt(u) + e.charAt(l) + e.charAt(d)
                            }
                            return r
                        }
                    } catch (e) {}
                }();
                var s, i = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

                function a(e, t, n) {
                    null != e && ("number" == typeof e ? this.fromNumber(e, t, n) : null == t && "string" != typeof e ? this.fromString(e, 256) : this.fromString(e, t))
                }

                function u() {
                    return new a(null)
                }
                "Microsoft Internet Explorer" == navigator.appName ? (a.prototype.am = function(e, t, n, o, r, s) {
                    for (var i = 32767 & t, a = t >> 15; --s >= 0;) {
                        var u = 32767 & this[e],
                            l = this[e++] >> 15,
                            d = a * u + l * i;
                        r = ((u = i * u + ((32767 & d) << 15) + n[o] + (1073741823 & r)) >>> 30) + (d >>> 15) + a * l + (r >>> 30), n[o++] = 1073741823 & u
                    }
                    return r
                }, s = 30) : "Netscape" != navigator.appName ? (a.prototype.am = function(e, t, n, o, r, s) {
                    for (; --s >= 0;) {
                        var i = t * this[e++] + n[o] + r;
                        r = Math.floor(i / 67108864), n[o++] = 67108863 & i
                    }
                    return r
                }, s = 26) : (a.prototype.am = function(e, t, n, o, r, s) {
                    for (var i = 16383 & t, a = t >> 14; --s >= 0;) {
                        var u = 16383 & this[e],
                            l = this[e++] >> 14,
                            d = a * u + l * i;
                        r = ((u = i * u + ((16383 & d) << 14) + n[o] + r) >> 28) + (d >> 14) + a * l, n[o++] = 268435455 & u
                    }
                    return r
                }, s = 28), a.prototype.DB = s, a.prototype.DM = (1 << s) - 1, a.prototype.DV = 1 << s;
                a.prototype.FV = Math.pow(2, 52), a.prototype.F1 = 52 - s, a.prototype.F2 = 2 * s - 52;
                var l, d, c = new Array;
                for (l = "0".charCodeAt(0), d = 0; d <= 9; ++d) c[l++] = d;
                for (l = "a".charCodeAt(0), d = 10; d < 36; ++d) c[l++] = d;
                for (l = "A".charCodeAt(0), d = 10; d < 36; ++d) c[l++] = d;

                function h(e) {
                    return "0123456789abcdefghijklmnopqrstuvwxyz".charAt(e)
                }

                function f(e, t) {
                    var n = c[e.charCodeAt(t)];
                    return null == n ? -1 : n
                }

                function p(e) {
                    var t = u();
                    return t.fromInt(e), t
                }

                function m(e) {
                    var t, n = 1;
                    return 0 != (t = e >>> 16) && (e = t, n += 16), 0 != (t = e >> 8) && (e = t, n += 8), 0 != (t = e >> 4) && (e = t, n += 4), 0 != (t = e >> 2) && (e = t, n += 2), 0 != (t = e >> 1) && (e = t, n += 1), n
                }

                function _(e) {
                    this.m = e
                }

                function v(e) {
                    this.m = e, this.mp = e.invDigit(), this.mpl = 32767 & this.mp, this.mph = this.mp >> 15, this.um = (1 << e.DB - 15) - 1, this.mt2 = 2 * e.t
                }

                function y() {
                    this.i = 0, this.j = 0, this.S = new Array
                }
                _.prototype.convert = function(e) {
                    return e.s < 0 || e.compareTo(this.m) >= 0 ? e.mod(this.m) : e
                }, _.prototype.revert = function(e) {
                    return e
                }, _.prototype.reduce = function(e) {
                    e.divRemTo(this.m, null, e)
                }, _.prototype.mulTo = function(e, t, n) {
                    e.multiplyTo(t, n), this.reduce(n)
                }, _.prototype.sqrTo = function(e, t) {
                    e.squareTo(t), this.reduce(t)
                }, v.prototype.convert = function(e) {
                    var t = u();
                    return e.abs().dlShiftTo(this.m.t, t), t.divRemTo(this.m, null, t), e.s < 0 && t.compareTo(a.ZERO) > 0 && this.m.subTo(t, t), t
                }, v.prototype.revert = function(e) {
                    var t = u();
                    return e.copyTo(t), this.reduce(t), t
                }, v.prototype.reduce = function(e) {
                    for (; e.t <= this.mt2;) e[e.t++] = 0;
                    for (var t = 0; t < this.m.t; ++t) {
                        var n = 32767 & e[t],
                            o = n * this.mpl + ((n * this.mph + (e[t] >> 15) * this.mpl & this.um) << 15) & e.DM;
                        for (e[n = t + this.m.t] += this.m.am(0, o, e, t, 0, this.m.t); e[n] >= e.DV;) e[n] -= e.DV, e[++n]++
                    }
                    e.clamp(), e.drShiftTo(this.m.t, e), e.compareTo(this.m) >= 0 && e.subTo(this.m, e)
                }, v.prototype.mulTo = function(e, t, n) {
                    e.multiplyTo(t, n), this.reduce(n)
                }, v.prototype.sqrTo = function(e, t) {
                    e.squareTo(t), this.reduce(t)
                }, a.prototype.copyTo = function(e) {
                    for (var t = this.t - 1; t >= 0; --t) e[t] = this[t];
                    e.t = this.t, e.s = this.s
                }, a.prototype.fromInt = function(e) {
                    this.t = 1, this.s = e < 0 ? -1 : 0, e > 0 ? this[0] = e : e < -1 ? this[0] = e + this.DV : this.t = 0
                }, a.prototype.fromString = function(e, t) {
                    var n;
                    if (16 == t) n = 4;
                    else if (8 == t) n = 3;
                    else if (256 == t) n = 8;
                    else if (2 == t) n = 1;
                    else if (32 == t) n = 5;
                    else {
                        if (4 != t) return void this.fromRadix(e, t);
                        n = 2
                    }
                    this.t = 0, this.s = 0;
                    for (var o = e.length, r = !1, s = 0; --o >= 0;) {
                        var i = 8 == n ? 255 & e[o] : f(e, o);
                        i < 0 ? "-" == e.charAt(o) && (r = !0) : (r = !1, 0 == s ? this[this.t++] = i : s + n > this.DB ? (this[this.t - 1] |= (i & (1 << this.DB - s) - 1) << s, this[this.t++] = i >> this.DB - s) : this[this.t - 1] |= i << s, (s += n) >= this.DB && (s -= this.DB))
                    }
                    8 == n && 0 != (128 & e[0]) && (this.s = -1, s > 0 && (this[this.t - 1] |= (1 << this.DB - s) - 1 << s)), this.clamp(), r && a.ZERO.subTo(this, this)
                }, a.prototype.clamp = function() {
                    for (var e = this.s & this.DM; this.t > 0 && this[this.t - 1] == e;) --this.t
                }, a.prototype.dlShiftTo = function(e, t) {
                    var n;
                    for (n = this.t - 1; n >= 0; --n) t[n + e] = this[n];
                    for (n = e - 1; n >= 0; --n) t[n] = 0;
                    t.t = this.t + e, t.s = this.s
                }, a.prototype.drShiftTo = function(e, t) {
                    for (var n = e; n < this.t; ++n) t[n - e] = this[n];
                    t.t = Math.max(this.t - e, 0), t.s = this.s
                }, a.prototype.lShiftTo = function(e, t) {
                    var n, o = e % this.DB,
                        r = this.DB - o,
                        s = (1 << r) - 1,
                        i = Math.floor(e / this.DB),
                        a = this.s << o & this.DM;
                    for (n = this.t - 1; n >= 0; --n) t[n + i + 1] = this[n] >> r | a, a = (this[n] & s) << o;
                    for (n = i - 1; n >= 0; --n) t[n] = 0;
                    t[i] = a, t.t = this.t + i + 1, t.s = this.s, t.clamp()
                }, a.prototype.rShiftTo = function(e, t) {
                    t.s = this.s;
                    var n = Math.floor(e / this.DB);
                    if (n >= this.t) t.t = 0;
                    else {
                        var o = e % this.DB,
                            r = this.DB - o,
                            s = (1 << o) - 1;
                        t[0] = this[n] >> o;
                        for (var i = n + 1; i < this.t; ++i) t[i - n - 1] |= (this[i] & s) << r, t[i - n] = this[i] >> o;
                        o > 0 && (t[this.t - n - 1] |= (this.s & s) << r), t.t = this.t - n, t.clamp()
                    }
                }, a.prototype.subTo = function(e, t) {
                    for (var n = 0, o = 0, r = Math.min(e.t, this.t); n < r;) o += this[n] - e[n], t[n++] = o & this.DM, o >>= this.DB;
                    if (e.t < this.t) {
                        for (o -= e.s; n < this.t;) o += this[n], t[n++] = o & this.DM, o >>= this.DB;
                        o += this.s
                    } else {
                        for (o += this.s; n < e.t;) o -= e[n], t[n++] = o & this.DM, o >>= this.DB;
                        o -= e.s
                    }
                    t.s = o < 0 ? -1 : 0, o < -1 ? t[n++] = this.DV + o : o > 0 && (t[n++] = o), t.t = n, t.clamp()
                }, a.prototype.multiplyTo = function(e, t) {
                    var n = this.abs(),
                        o = e.abs(),
                        r = n.t;
                    for (t.t = r + o.t; --r >= 0;) t[r] = 0;
                    for (r = 0; r < o.t; ++r) t[r + n.t] = n.am(0, o[r], t, r, 0, n.t);
                    t.s = 0, t.clamp(), this.s != e.s && a.ZERO.subTo(t, t)
                }, a.prototype.squareTo = function(e) {
                    for (var t = this.abs(), n = e.t = 2 * t.t; --n >= 0;) e[n] = 0;
                    for (n = 0; n < t.t - 1; ++n) {
                        var o = t.am(n, t[n], e, 2 * n, 0, 1);
                        (e[n + t.t] += t.am(n + 1, 2 * t[n], e, 2 * n + 1, o, t.t - n - 1)) >= t.DV && (e[n + t.t] -= t.DV, e[n + t.t + 1] = 1)
                    }
                    e.t > 0 && (e[e.t - 1] += t.am(n, t[n], e, 2 * n, 0, 1)), e.s = 0, e.clamp()
                }, a.prototype.divRemTo = function(e, t, n) {
                    var o = e.abs();
                    if (!(o.t <= 0)) {
                        var r = this.abs();
                        if (r.t < o.t) return null != t && t.fromInt(0), void(null != n && this.copyTo(n));
                        null == n && (n = u());
                        var s = u(),
                            i = this.s,
                            l = e.s,
                            d = this.DB - m(o[o.t - 1]);
                        d > 0 ? (o.lShiftTo(d, s), r.lShiftTo(d, n)) : (o.copyTo(s), r.copyTo(n));
                        var c = s.t,
                            h = s[c - 1];
                        if (0 != h) {
                            var f = h * (1 << this.F1) + (c > 1 ? s[c - 2] >> this.F2 : 0),
                                p = this.FV / f,
                                _ = (1 << this.F1) / f,
                                v = 1 << this.F2,
                                y = n.t,
                                b = y - c,
                                j = null == t ? u() : t;
                            for (s.dlShiftTo(b, j), n.compareTo(j) >= 0 && (n[n.t++] = 1, n.subTo(j, n)), a.ONE.dlShiftTo(c, j), j.subTo(s, s); s.t < c;) s[s.t++] = 0;
                            for (; --b >= 0;) {
                                var g = n[--y] == h ? this.DM : Math.floor(n[y] * p + (n[y - 1] + v) * _);
                                if ((n[y] += s.am(0, g, n, b, 0, c)) < g)
                                    for (s.dlShiftTo(b, j), n.subTo(j, n); n[y] < --g;) n.subTo(j, n)
                            }
                            null != t && (n.drShiftTo(c, t), i != l && a.ZERO.subTo(t, t)), n.t = c, n.clamp(), d > 0 && n.rShiftTo(d, n), i < 0 && a.ZERO.subTo(n, n)
                        }
                    }
                }, a.prototype.invDigit = function() {
                    if (this.t < 1) return 0;
                    var e = this[0];
                    if (0 == (1 & e)) return 0;
                    var t = 3 & e;
                    return (t = (t = (t = (t = t * (2 - (15 & e) * t) & 15) * (2 - (255 & e) * t) & 255) * (2 - ((65535 & e) * t & 65535)) & 65535) * (2 - e * t % this.DV) % this.DV) > 0 ? this.DV - t : -t
                }, a.prototype.isEven = function() {
                    return 0 == (this.t > 0 ? 1 & this[0] : this.s)
                }, a.prototype.exp = function(e, t) {
                    if (e > 4294967295 || e < 1) return a.ONE;
                    var n = u(),
                        o = u(),
                        r = t.convert(this),
                        s = m(e) - 1;
                    for (r.copyTo(n); --s >= 0;)
                        if (t.sqrTo(n, o), (e & 1 << s) > 0) t.mulTo(o, r, n);
                        else {
                            var i = n;
                            n = o, o = i
                        } return t.revert(n)
                }, a.prototype.toString = function(e) {
                    if (this.s < 0) return "-" + this.negate().toString(e);
                    var t;
                    if (16 == e) t = 4;
                    else if (8 == e) t = 3;
                    else if (2 == e) t = 1;
                    else if (32 == e) t = 5;
                    else {
                        if (4 != e) return this.toRadix(e);
                        t = 2
                    }
                    var n, o = (1 << t) - 1,
                        r = !1,
                        s = "",
                        i = this.t,
                        a = this.DB - i * this.DB % t;
                    if (i-- > 0)
                        for (a < this.DB && (n = this[i] >> a) > 0 && (r = !0, s = h(n)); i >= 0;) a < t ? (n = (this[i] & (1 << a) - 1) << t - a, n |= this[--i] >> (a += this.DB - t)) : (n = this[i] >> (a -= t) & o, a <= 0 && (a += this.DB, --i)), n > 0 && (r = !0), r && (s += h(n));
                    return r ? s : "0"
                }, a.prototype.negate = function() {
                    var e = u();
                    return a.ZERO.subTo(this, e), e
                }, a.prototype.abs = function() {
                    return this.s < 0 ? this.negate() : this
                }, a.prototype.compareTo = function(e) {
                    var t = this.s - e.s;
                    if (0 != t) return t;
                    var n = this.t;
                    if (0 != (t = n - e.t)) return this.s < 0 ? -t : t;
                    for (; --n >= 0;)
                        if (0 != (t = this[n] - e[n])) return t;
                    return 0
                }, a.prototype.bitLength = function() {
                    return this.t <= 0 ? 0 : this.DB * (this.t - 1) + m(this[this.t - 1] ^ this.s & this.DM)
                }, a.prototype.mod = function(e) {
                    var t = u();
                    return this.abs().divRemTo(e, null, t), this.s < 0 && t.compareTo(a.ZERO) > 0 && e.subTo(t, t), t
                }, a.prototype.modPowInt = function(e, t) {
                    var n;
                    return n = e < 256 || t.isEven() ? new _(t) : new v(t), this.exp(e, n)
                }, a.ZERO = p(0), a.ONE = p(1), y.prototype.init = function(e) {
                    var t, n, o;
                    for (t = 0; t < 256; ++t) this.S[t] = t;
                    for (n = 0, t = 0; t < 256; ++t) n = n + this.S[t] + e[t % e.length] & 255, o = this.S[t], this.S[t] = this.S[n], this.S[n] = o;
                    this.i = 0, this.j = 0
                }, y.prototype.next = function() {
                    var e;
                    return this.i = this.i + 1 & 255, this.j = this.j + this.S[this.i] & 255, e = this.S[this.i], this.S[this.i] = this.S[this.j], this.S[this.j] = e, this.S[e + this.S[this.i] & 255]
                };
                var b, j, g;

                function x() {
                    var e;
                    e = (new Date).getTime(), j[g++] ^= 255 & e, j[g++] ^= e >> 8 & 255, j[g++] ^= e >> 16 & 255, j[g++] ^= e >> 24 & 255, g >= 256 && (g -= 256)
                }
                if (null == j) {
                    j = [], g = 0;
                    try {
                        if (window.crypto && window.crypto.getRandomValues) {
                            var w = new Uint8Array(32);
                            for (window.crypto.getRandomValues(w), D = 0; D < 32; ++D) j[g++] = w[D]
                        } else if (window.msCrypto && window.msCrypto.getRandomValues) {
                            w = new Uint8Array(32);
                            for (window.msCrypto.getRandomValues(w), D = 0; D < 32; ++D) j[g++] = w[D]
                        } else if (window.crypto && window.crypto.random) {
                            var F = window.crypto.random(32);
                            for (D = 0; D < F.length; ++D) j[g++] = 255 & F.charCodeAt(D)
                        }
                    } catch (e) {}
                    for (; g < 256;) D = Math.floor(65536 * Math.random()), j[g++] = D >>> 8, j[g++] = 255 & D;
                    g = 0, x()
                }

                function E() {
                    if (null == b) {
                        for (x(), (b = new y).init(j), g = 0; g < j.length; ++g) j[g] = 0;
                        g = 0
                    }
                    return b.next()
                }

                function C() {}

                function A() {
                    this.n = null, this.e = 0, this.d = null, this.p = null, this.q = null, this.dmp1 = null, this.dmq1 = null, this.coeff = null
                }

                function k(e) {
                    throw e
                }
                C.prototype.nextBytes = function(e) {
                    var t;
                    for (t = 0; t < e.length; ++t) e[t] = E()
                }, A.prototype.doPublic = function(e) {
                    return e.modPowInt(this.e, this.n)
                }, A.prototype.setPublic = function(e, t) {
                    null != e && null != t && e.length > 0 && t.length > 0 ? (this.n = function(e, t) {
                        return new a(e, t)
                    }(e, 16), this.e = parseInt(t, 16)) : alert("Invalid RSA public key")
                }, A.prototype.encrypt = function(e) {
                    var t = function(e, t) {
                        if (t < e.length + 11) return alert("Message too long for RSA"), null;
                        for (var n = new Array, o = e.length - 1; o >= 0 && t > 0;) n[--t] = e[o--];
                        n[--t] = 0;
                        for (var r = new C, s = new Array; t > 2;) {
                            for (s[0] = 0; 0 == s[0];) r.nextBytes(s);
                            n[--t] = s[0]
                        }
                        return n[--t] = 2, n[--t] = 0, new a(n)
                    }(e, this.n.bitLength() + 7 >> 3);
                    if (null == t) return null;
                    var n = this.doPublic(t);
                    if (null == n) return null;
                    var o = n.toString(16);
                    return 0 == (1 & o.length) ? o : "0" + o
                }, A.prototype.encrypt_b64 = function(e) {
                    var t = this.encrypt(e);
                    return t ? function(e) {
                        var t, n, o = "";
                        for (t = 0; t + 3 <= e.length; t += 3) n = parseInt(e.substring(t, t + 3), 16), o += i.charAt(n >> 6) + i.charAt(63 & n);
                        for (t + 1 == e.length ? (n = parseInt(e.substring(t, t + 1), 16), o += i.charAt(n << 2)) : t + 2 == e.length && (n = parseInt(e.substring(t, t + 2), 16), o += i.charAt(n >> 2) + i.charAt((3 & n) << 4));
                            (3 & o.length) > 0;) o += "=";
                        return o
                    }(t) : null
                };
                var O, S, T, D = void 0,
                    P = !1,
                    M = {
                        cipher: {},
                        hash: {},
                        keyexchange: {},
                        mode: {},
                        misc: {},
                        codec: {},
                        exception: {
                            corrupt: function(e) {
                                this.toString = function() {
                                    return "CORRUPT: " + this.message
                                }, this.message = e
                            },
                            invalid: function(e) {
                                this.toString = function() {
                                    return "INVALID: " + this.message
                                }, this.message = e
                            },
                            bug: function(e) {
                                this.toString = function() {
                                    return "BUG: " + this.message
                                }, this.message = e
                            },
                            notReady: function(e) {
                                this.toString = function() {
                                    return "NOT READY: " + this.message
                                }, this.message = e
                            }
                        }
                    };

                function R(e, t, n) {
                    4 !== t.length && k(new M.exception.invalid("invalid aes block size"));
                    var o = e.b[n],
                        r = t[0] ^ o[0],
                        s = t[n ? 3 : 1] ^ o[1],
                        i = t[2] ^ o[2];
                    t = t[n ? 1 : 3] ^ o[3];
                    var a, u, l, d, c = o.length / 4 - 2,
                        h = 4,
                        f = [0, 0, 0, 0];
                    e = (a = e.k[n])[0];
                    var p = a[1],
                        m = a[2],
                        _ = a[3],
                        v = a[4];
                    for (d = 0; d < c; d++) a = e[r >>> 24] ^ p[s >> 16 & 255] ^ m[i >> 8 & 255] ^ _[255 & t] ^ o[h], u = e[s >>> 24] ^ p[i >> 16 & 255] ^ m[t >> 8 & 255] ^ _[255 & r] ^ o[h + 1], l = e[i >>> 24] ^ p[t >> 16 & 255] ^ m[r >> 8 & 255] ^ _[255 & s] ^ o[h + 2], t = e[t >>> 24] ^ p[r >> 16 & 255] ^ m[s >> 8 & 255] ^ _[255 & i] ^ o[h + 3], h += 4, r = a, s = u, i = l;
                    for (d = 0; 4 > d; d++) f[n ? 3 & -d : d] = v[r >>> 24] << 24 ^ v[s >> 16 & 255] << 16 ^ v[i >> 8 & 255] << 8 ^ v[255 & t] ^ o[h++], a = r, r = s, s = i, i = t, t = a;
                    return f
                }

                function F(e, t) {
                    var n, o, r, s = t.slice(0),
                        i = e.r,
                        a = e.b,
                        u = i[0],
                        l = i[1],
                        d = i[2],
                        c = i[3],
                        h = i[4],
                        f = i[5],
                        p = i[6],
                        m = i[7];
                    for (n = 0; 64 > n; n++) 16 > n ? o = s[n] : (o = s[n + 1 & 15], r = s[n + 14 & 15], o = s[15 & n] = (o >>> 7 ^ o >>> 18 ^ o >>> 3 ^ o << 25 ^ o << 14) + (r >>> 17 ^ r >>> 19 ^ r >>> 10 ^ r << 15 ^ r << 13) + s[15 & n] + s[n + 9 & 15] | 0), o = o + m + (h >>> 6 ^ h >>> 11 ^ h >>> 25 ^ h << 26 ^ h << 21 ^ h << 7) + (p ^ h & (f ^ p)) + a[n], m = p, p = f, f = h, h = c + o | 0, c = d, d = l, u = o + ((l = u) & d ^ c & (l ^ d)) + (l >>> 2 ^ l >>> 13 ^ l >>> 22 ^ l << 30 ^ l << 19 ^ l << 10) | 0;
                    i[0] = i[0] + u | 0, i[1] = i[1] + l | 0, i[2] = i[2] + d | 0, i[3] = i[3] + c | 0, i[4] = i[4] + h | 0, i[5] = i[5] + f | 0, i[6] = i[6] + p | 0, i[7] = i[7] + m | 0
                }

                function I(e, t) {
                    var n, o = M.random.w[e],
                        r = [];
                    for (n in o) o.hasOwnProperty(n) && r.push(o[n]);
                    for (n = 0; n < r.length; n++) r[n](t)
                }

                function L(e) {
                    "undefined" != typeof window && window.performance && "function" == typeof window.performance.now ? M.random.addEntropy(window.performance.now(), e, "loadtime") : M.random.addEntropy((new Date).valueOf(), e, "loadtime")
                }

                function N(e) {
                    e.b = V(e).concat(V(e)), e.A = new M.cipher.aes(e.b)
                }

                function V(e) {
                    for (var t = 0; 4 > t && (e.f[t] = e.f[t] + 1 | 0, !e.f[t]); t++);
                    return e.A.encrypt(e.f)
                }

                function B(e, t) {
                    return function() {
                        t.apply(e, arguments)
                    }
                }
                e.exports && (e.exports = M), M.cipher.aes = function(e) {
                    this.k[0][0][0] || this.D();
                    var t, n, o, r, s = this.k[0][4],
                        i = this.k[1],
                        a = 1;
                    for (4 !== (t = e.length) && 6 !== t && 8 !== t && k(new M.exception.invalid("invalid aes key size")), this.b = [o = e.slice(0), r = []], e = t; e < 4 * t + 28; e++) n = o[e - 1], (0 == e % t || 8 === t && 4 == e % t) && (n = s[n >>> 24] << 24 ^ s[n >> 16 & 255] << 16 ^ s[n >> 8 & 255] << 8 ^ s[255 & n], 0 == e % t && (n = n << 8 ^ n >>> 24 ^ a << 24, a = a << 1 ^ 283 * (a >> 7))), o[e] = o[e - t] ^ n;
                    for (t = 0; e; t++, e--) n = o[3 & t ? e : e - 4], r[t] = 4 >= e || 4 > t ? n : i[0][s[n >>> 24]] ^ i[1][s[n >> 16 & 255]] ^ i[2][s[n >> 8 & 255]] ^ i[3][s[255 & n]]
                }, M.cipher.aes.prototype = {
                    encrypt: function(e) {
                        return R(this, e, 0)
                    },
                    decrypt: function(e) {
                        return R(this, e, 1)
                    },
                    k: [
                        [
                            [],
                            [],
                            [],
                            [],
                            []
                        ],
                        [
                            [],
                            [],
                            [],
                            [],
                            []
                        ]
                    ],
                    D: function() {
                        var e, t, n, o, r, s, i, a = this.k[0],
                            u = this.k[1],
                            l = a[4],
                            d = u[4],
                            c = [],
                            h = [];
                        for (e = 0; 256 > e; e++) h[(c[e] = e << 1 ^ 283 * (e >> 7)) ^ e] = e;
                        for (t = n = 0; !l[t]; t ^= o || 1, n = h[n] || 1)
                            for (s = (s = n ^ n << 1 ^ n << 2 ^ n << 3 ^ n << 4) >> 8 ^ 255 & s ^ 99, l[t] = s, d[s] = t, i = 16843009 * (r = c[e = c[o = c[t]]]) ^ 65537 * e ^ 257 * o ^ 16843008 * t, r = 257 * c[s] ^ 16843008 * s, e = 0; 4 > e; e++) a[e][t] = r = r << 24 ^ r >>> 8, u[e][s] = i = i << 24 ^ i >>> 8;
                        for (e = 0; 5 > e; e++) a[e] = a[e].slice(0), u[e] = u[e].slice(0)
                    }
                }, M.bitArray = {
                    bitSlice: function(e, t, n) {
                        return e = M.bitArray.P(e.slice(t / 32), 32 - (31 & t)).slice(1), n === D ? e : M.bitArray.clamp(e, n - t)
                    },
                    extract: function(e, t, n) {
                        var o = Math.floor(-t - n & 31);
                        return (-32 & (t + n - 1 ^ t) ? e[t / 32 | 0] << 32 - o ^ e[t / 32 + 1 | 0] >>> o : e[t / 32 | 0] >>> o) & (1 << n) - 1
                    },
                    concat: function(e, t) {
                        if (0 === e.length || 0 === t.length) return e.concat(t);
                        var n = e[e.length - 1],
                            o = M.bitArray.getPartial(n);
                        return 32 === o ? e.concat(t) : M.bitArray.P(t, o, 0 | n, e.slice(0, e.length - 1))
                    },
                    bitLength: function(e) {
                        var t = e.length;
                        return 0 === t ? 0 : 32 * (t - 1) + M.bitArray.getPartial(e[t - 1])
                    },
                    clamp: function(e, t) {
                        if (32 * e.length < t) return e;
                        var n = (e = e.slice(0, Math.ceil(t / 32))).length;
                        return t &= 31, 0 < n && t && (e[n - 1] = M.bitArray.partial(t, e[n - 1] & 2147483648 >> t - 1, 1)), e
                    },
                    partial: function(e, t, n) {
                        return 32 === e ? t : (n ? 0 | t : t << 32 - e) + 1099511627776 * e
                    },
                    getPartial: function(e) {
                        return Math.round(e / 1099511627776) || 32
                    },
                    equal: function(e, t) {
                        if (M.bitArray.bitLength(e) !== M.bitArray.bitLength(t)) return P;
                        var n, o = 0;
                        for (n = 0; n < e.length; n++) o |= e[n] ^ t[n];
                        return 0 === o
                    },
                    P: function(e, t, n, o) {
                        var r;
                        for (r = 0, o === D && (o = []); 32 <= t; t -= 32) o.push(n), n = 0;
                        if (0 === t) return o.concat(e);
                        for (r = 0; r < e.length; r++) o.push(n | e[r] >>> t), n = e[r] << 32 - t;
                        return r = e.length ? e[e.length - 1] : 0, e = M.bitArray.getPartial(r), o.push(M.bitArray.partial(t + e & 31, 32 < t + e ? n : o.pop(), 1)), o
                    },
                    l: function(e, t) {
                        return [e[0] ^ t[0], e[1] ^ t[1], e[2] ^ t[2], e[3] ^ t[3]]
                    },
                    byteswapM: function(e) {
                        var t, n;
                        for (t = 0; t < e.length; ++t) n = e[t], e[t] = n >>> 24 | n >>> 8 & 65280 | (65280 & n) << 8 | n << 24;
                        return e
                    }
                }, M.codec.utf8String = {
                    fromBits: function(e) {
                        var t, n, o = "",
                            r = M.bitArray.bitLength(e);
                        for (t = 0; t < r / 8; t++) 0 == (3 & t) && (n = e[t / 4]), o += String.fromCharCode(n >>> 24), n <<= 8;
                        return decodeURIComponent(escape(o))
                    },
                    toBits: function(e) {
                        e = unescape(encodeURIComponent(e));
                        var t, n = [],
                            o = 0;
                        for (t = 0; t < e.length; t++) o = o << 8 | e.charCodeAt(t), 3 == (3 & t) && (n.push(o), o = 0);
                        return 3 & t && n.push(M.bitArray.partial(8 * (3 & t), o)), n
                    }
                }, M.codec.hex = {
                    fromBits: function(e) {
                        var t, n = "";
                        for (t = 0; t < e.length; t++) n += (0xf00000000000 + (0 | e[t])).toString(16).substr(4);
                        return n.substr(0, M.bitArray.bitLength(e) / 4)
                    },
                    toBits: function(e) {
                        var t, n, o = [];
                        for (n = (e = e.replace(/\s|0x/g, "")).length, e += "00000000", t = 0; t < e.length; t += 8) o.push(0 ^ parseInt(e.substr(t, 8), 16));
                        return M.bitArray.clamp(o, 4 * n)
                    }
                }, M.codec.base64 = {
                    J: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
                    fromBits: function(e, t, n) {
                        var o = "",
                            r = 0,
                            s = M.codec.base64.J,
                            i = 0,
                            a = M.bitArray.bitLength(e);
                        for (n && (s = s.substr(0, 62) + "-_"), n = 0; 6 * o.length < a;) o += s.charAt((i ^ e[n] >>> r) >>> 26), 6 > r ? (i = e[n] << 6 - r, r += 26, n++) : (i <<= 6, r -= 6);
                        for (; 3 & o.length && !t;) o += "=";
                        return o
                    },
                    toBits: function(e, t) {
                        e = e.replace(/\s|=/g, "");
                        var n, o, r = [],
                            s = 0,
                            i = M.codec.base64.J,
                            a = 0;
                        for (t && (i = i.substr(0, 62) + "-_"), n = 0; n < e.length; n++) 0 > (o = i.indexOf(e.charAt(n))) && k(new M.exception.invalid("this isn't base64!")), 26 < s ? (s -= 26, r.push(a ^ o >>> s), a = o << 32 - s) : a ^= o << 32 - (s += 6);
                        return 56 & s && r.push(M.bitArray.partial(56 & s, a, 1)), r
                    }
                }, M.codec.base64url = {
                    fromBits: function(e) {
                        return M.codec.base64.fromBits(e, 1, 1)
                    },
                    toBits: function(e) {
                        return M.codec.base64.toBits(e, 1)
                    }
                }, M.hash.sha256 = function(e) {
                    this.b[0] || this.D(), e ? (this.r = e.r.slice(0), this.o = e.o.slice(0), this.h = e.h) : this.reset()
                }, M.hash.sha256.hash = function(e) {
                    return (new M.hash.sha256).update(e).finalize()
                }, M.hash.sha256.prototype = {
                    blockSize: 512,
                    reset: function() {
                        return this.r = this.N.slice(0), this.o = [], this.h = 0, this
                    },
                    update: function(e) {
                        "string" == typeof e && (e = M.codec.utf8String.toBits(e));
                        var t, n = this.o = M.bitArray.concat(this.o, e);
                        for (t = this.h, e = this.h = t + M.bitArray.bitLength(e), t = 512 + t & -512; t <= e; t += 512) F(this, n.splice(0, 16));
                        return this
                    },
                    finalize: function() {
                        var e, t = this.o,
                            n = this.r;
                        for (e = (t = M.bitArray.concat(t, [M.bitArray.partial(1, 1)])).length + 2; 15 & e; e++) t.push(0);
                        for (t.push(Math.floor(this.h / 4294967296)), t.push(0 | this.h); t.length;) F(this, t.splice(0, 16));
                        return this.reset(), n
                    },
                    N: [],
                    b: [],
                    D: function() {
                        function e(e) {
                            return 4294967296 * (e - Math.floor(e)) | 0
                        }
                        var t, n = 0,
                            o = 2;
                        e: for (; 64 > n; o++) {
                            for (t = 2; t * t <= o; t++)
                                if (0 == o % t) continue e;
                            8 > n && (this.N[n] = e(Math.pow(o, .5))), this.b[n] = e(Math.pow(o, 1 / 3)), n++
                        }
                    }
                }, M.mode.ccm = {
                    name: "ccm",
                    encrypt: function(e, t, n, o, r) {
                        var s, i = t.slice(0),
                            a = M.bitArray,
                            u = a.bitLength(n) / 8,
                            l = a.bitLength(i) / 8;
                        for (r = r || 64, o = o || [], 7 > u && k(new M.exception.invalid("ccm: iv must be at least 7 bytes")), s = 2; 4 > s && l >>> 8 * s; s++);
                        return s < 15 - u && (s = 15 - u), n = a.clamp(n, 8 * (15 - s)), t = M.mode.ccm.L(e, t, n, o, r, s), i = M.mode.ccm.p(e, i, n, t, r, s), a.concat(i.data, i.tag)
                    },
                    decrypt: function(e, t, n, o, r) {
                        r = r || 64, o = o || [];
                        var s = M.bitArray,
                            i = s.bitLength(n) / 8,
                            a = s.bitLength(t),
                            u = s.clamp(t, a - r),
                            l = s.bitSlice(t, a - r);
                        a = (a - r) / 8;
                        for (7 > i && k(new M.exception.invalid("ccm: iv must be at least 7 bytes")), t = 2; 4 > t && a >>> 8 * t; t++);
                        return t < 15 - i && (t = 15 - i), n = s.clamp(n, 8 * (15 - t)), u = M.mode.ccm.p(e, u, n, l, r, t), e = M.mode.ccm.L(e, u.data, n, o, r, t), s.equal(u.tag, e) || k(new M.exception.corrupt("ccm: tag doesn't match")), u.data
                    },
                    L: function(e, t, n, o, r, s) {
                        var i = [],
                            a = M.bitArray,
                            u = a.l;
                        if (((r /= 8) % 2 || 4 > r || 16 < r) && k(new M.exception.invalid("ccm: invalid tag length")), (4294967295 < o.length || 4294967295 < t.length) && k(new M.exception.bug("ccm: can't deal with 4GiB or more data")), s = [a.partial(8, (o.length ? 64 : 0) | r - 2 << 2 | s - 1)], (s = a.concat(s, n))[3] |= a.bitLength(t) / 8, s = e.encrypt(s), o.length)
                            for (65279 >= (n = a.bitLength(o) / 8) ? i = [a.partial(16, n)] : 4294967295 >= n && (i = a.concat([a.partial(16, 65534)], [n])), i = a.concat(i, o), o = 0; o < i.length; o += 4) s = e.encrypt(u(s, i.slice(o, o + 4).concat([0, 0, 0])));
                        for (o = 0; o < t.length; o += 4) s = e.encrypt(u(s, t.slice(o, o + 4).concat([0, 0, 0])));
                        return a.clamp(s, 8 * r)
                    },
                    p: function(e, t, n, o, r, s) {
                        var i, a = M.bitArray;
                        i = a.l;
                        var u = t.length,
                            l = a.bitLength(t);
                        if (n = a.concat([a.partial(8, s - 1)], n).concat([0, 0, 0]).slice(0, 4), o = a.bitSlice(i(o, e.encrypt(n)), 0, r), !u) return {
                            tag: o,
                            data: []
                        };
                        for (i = 0; i < u; i += 4) n[3]++, r = e.encrypt(n), t[i] ^= r[0], t[i + 1] ^= r[1], t[i + 2] ^= r[2], t[i + 3] ^= r[3];
                        return {
                            tag: o,
                            data: a.clamp(t, l)
                        }
                    }
                }, M.mode.ocb2 = {
                    name: "ocb2",
                    encrypt: function(e, t, n, o, r, s) {
                        128 !== M.bitArray.bitLength(n) && k(new M.exception.invalid("ocb iv must be 128 bits"));
                        var i, a = M.mode.ocb2.H,
                            u = M.bitArray,
                            l = u.l,
                            d = [0, 0, 0, 0];
                        n = a(e.encrypt(n));
                        var c, h = [];
                        for (o = o || [], r = r || 64, i = 0; i + 4 < t.length; i += 4) d = l(d, c = t.slice(i, i + 4)), h = h.concat(l(n, e.encrypt(l(n, c)))), n = a(n);
                        return c = t.slice(i), t = u.bitLength(c), i = e.encrypt(l(n, [0, 0, 0, t])), c = u.clamp(l(c.concat([0, 0, 0]), i), t), d = l(d, l(c.concat([0, 0, 0]), i)), d = e.encrypt(l(d, l(n, a(n)))), o.length && (d = l(d, s ? o : M.mode.ocb2.pmac(e, o))), h.concat(u.concat(c, u.clamp(d, r)))
                    },
                    decrypt: function(e, t, n, o, r, s) {
                        128 !== M.bitArray.bitLength(n) && k(new M.exception.invalid("ocb iv must be 128 bits")), r = r || 64;
                        var i, a, u = M.mode.ocb2.H,
                            l = M.bitArray,
                            d = l.l,
                            c = [0, 0, 0, 0],
                            h = u(e.encrypt(n)),
                            f = M.bitArray.bitLength(t) - r,
                            p = [];
                        for (o = o || [], n = 0; n + 4 < f / 32; n += 4) i = d(h, e.decrypt(d(h, t.slice(n, n + 4)))), c = d(c, i), p = p.concat(i), h = u(h);
                        return a = f - 32 * n, i = e.encrypt(d(h, [0, 0, 0, a])), i = d(i, l.clamp(t.slice(n), a).concat([0, 0, 0])), c = d(c, i), c = e.encrypt(d(c, d(h, u(h)))), o.length && (c = d(c, s ? o : M.mode.ocb2.pmac(e, o))), l.equal(l.clamp(c, r), l.bitSlice(t, f)) || k(new M.exception.corrupt("ocb: tag doesn't match")), p.concat(l.clamp(i, a))
                    },
                    pmac: function(e, t) {
                        var n, o = M.mode.ocb2.H,
                            r = M.bitArray,
                            s = r.l,
                            i = [0, 0, 0, 0],
                            a = s(a = e.encrypt([0, 0, 0, 0]), o(o(a)));
                        for (n = 0; n + 4 < t.length; n += 4) a = o(a), i = s(i, e.encrypt(s(a, t.slice(n, n + 4))));
                        return n = t.slice(n), 128 > r.bitLength(n) && (a = s(a, o(a)), n = r.concat(n, [-2147483648, 0, 0, 0])), i = s(i, n), e.encrypt(s(o(s(a, o(a))), i))
                    },
                    H: function(e) {
                        return [e[0] << 1 ^ e[1] >>> 31, e[1] << 1 ^ e[2] >>> 31, e[2] << 1 ^ e[3] >>> 31, e[3] << 1 ^ 135 * (e[0] >>> 31)]
                    }
                }, M.mode.gcm = {
                    name: "gcm",
                    encrypt: function(e, t, n, o, r) {
                        var s = t.slice(0);
                        return t = M.bitArray, o = o || [], e = M.mode.gcm.p(!0, e, s, o, n, r || 128), t.concat(e.data, e.tag)
                    },
                    decrypt: function(e, t, n, o, r) {
                        var s = t.slice(0),
                            i = M.bitArray,
                            a = i.bitLength(s);
                        return o = o || [], (r = r || 128) <= a ? (t = i.bitSlice(s, a - r), s = i.bitSlice(s, 0, a - r)) : (t = s, s = []), e = M.mode.gcm.p(P, e, s, o, n, r), i.equal(e.tag, t) || k(new M.exception.corrupt("gcm: tag doesn't match")), e.data
                    },
                    Z: function(e, t) {
                        var n, o, r, s, i, a = M.bitArray.l;
                        for (r = [0, 0, 0, 0], s = t.slice(0), n = 0; 128 > n; n++) {
                            for ((o = 0 != (e[Math.floor(n / 32)] & 1 << 31 - n % 32)) && (r = a(r, s)), i = 0 != (1 & s[3]), o = 3; 0 < o; o--) s[o] = s[o] >>> 1 | (1 & s[o - 1]) << 31;
                            s[0] >>>= 1, i && (s[0] ^= -520093696)
                        }
                        return r
                    },
                    g: function(e, t, n) {
                        var o, r = n.length;
                        for (t = t.slice(0), o = 0; o < r; o += 4) t[0] ^= 4294967295 & n[o], t[1] ^= 4294967295 & n[o + 1], t[2] ^= 4294967295 & n[o + 2], t[3] ^= 4294967295 & n[o + 3], t = M.mode.gcm.Z(t, e);
                        return t
                    },
                    p: function(e, t, n, o, r, s) {
                        var i, a, u, l, d, c, h, f, p = M.bitArray;
                        for (c = n.length, h = p.bitLength(n), f = p.bitLength(o), a = p.bitLength(r), i = t.encrypt([0, 0, 0, 0]), 96 === a ? (r = r.slice(0), r = p.concat(r, [1])) : (r = M.mode.gcm.g(i, [0, 0, 0, 0], r), r = M.mode.gcm.g(i, r, [0, 0, Math.floor(a / 4294967296), 4294967295 & a])), a = M.mode.gcm.g(i, [0, 0, 0, 0], o), d = r.slice(0), o = a.slice(0), e || (o = M.mode.gcm.g(i, a, n)), l = 0; l < c; l += 4) d[3]++, u = t.encrypt(d), n[l] ^= u[0], n[l + 1] ^= u[1], n[l + 2] ^= u[2], n[l + 3] ^= u[3];
                        return n = p.clamp(n, h), e && (o = M.mode.gcm.g(i, a, n)), e = [Math.floor(f / 4294967296), 4294967295 & f, Math.floor(h / 4294967296), 4294967295 & h], o = M.mode.gcm.g(i, o, e), u = t.encrypt(r), o[0] ^= u[0], o[1] ^= u[1], o[2] ^= u[2], o[3] ^= u[3], {
                            tag: p.bitSlice(o, 0, s),
                            data: n
                        }
                    }
                }, M.misc.hmac = function(e, t) {
                    this.M = t = t || M.hash.sha256;
                    var n, o = [
                            [],
                            []
                        ],
                        r = t.prototype.blockSize / 32;
                    for (this.n = [new t, new t], e.length > r && (e = t.hash(e)), n = 0; n < r; n++) o[0][n] = 909522486 ^ e[n], o[1][n] = 1549556828 ^ e[n];
                    this.n[0].update(o[0]), this.n[1].update(o[1]), this.G = new t(this.n[0])
                }, M.misc.hmac.prototype.encrypt = M.misc.hmac.prototype.mac = function(e) {
                    return this.Q && k(new M.exception.invalid("encrypt on already updated hmac called!")), this.update(e), this.digest(e)
                }, M.misc.hmac.prototype.reset = function() {
                    this.G = new this.M(this.n[0]), this.Q = P
                }, M.misc.hmac.prototype.update = function(e) {
                    this.Q = !0, this.G.update(e)
                }, M.misc.hmac.prototype.digest = function() {
                    var e = this.G.finalize();
                    e = new this.M(this.n[1]).update(e).finalize();
                    return this.reset(), e
                }, M.misc.pbkdf2 = function(e, t, n, o, r) {
                    n = n || 1e3, (0 > o || 0 > n) && k(M.exception.invalid("invalid params to pbkdf2")), "string" == typeof e && (e = M.codec.utf8String.toBits(e)), "string" == typeof t && (t = M.codec.utf8String.toBits(t)), e = new(r = r || M.misc.hmac)(e);
                    var s, i, a, u, l = [],
                        d = M.bitArray;
                    for (u = 1; 32 * l.length < (o || 1); u++) {
                        for (r = s = e.encrypt(d.concat(t, [u])), i = 1; i < n; i++)
                            for (s = e.encrypt(s), a = 0; a < s.length; a++) r[a] ^= s[a];
                        l = l.concat(r)
                    }
                    return o && (l = d.clamp(l, o)), l
                }, M.prng = function(e) {
                    this.c = [new M.hash.sha256], this.i = [0], this.F = 0, this.s = {}, this.C = 0, this.K = {}, this.O = this.d = this.j = this.W = 0, this.b = [0, 0, 0, 0, 0, 0, 0, 0], this.f = [0, 0, 0, 0], this.A = D, this.B = e, this.q = P, this.w = {
                        progress: {},
                        seeded: {}
                    }, this.m = this.V = 0, this.t = 1, this.u = 2, this.S = 65536, this.I = [0, 48, 64, 96, 128, 192, 256, 384, 512, 768, 1024], this.T = 3e4, this.R = 80
                }, M.prng.prototype = {
                    randomWords: function(e, t) {
                        var n, o, r = [];
                        if ((n = this.isReady(t)) === this.m && k(new M.exception.notReady("generator isn't seeded")), n & this.u) {
                            n = !(n & this.t), o = [];
                            var s, i = 0;
                            for (this.O = o[0] = (new Date).valueOf() + this.T, s = 0; 16 > s; s++) o.push(4294967296 * Math.random() | 0);
                            for (s = 0; s < this.c.length && (o = o.concat(this.c[s].finalize()), i += this.i[s], this.i[s] = 0, n || !(this.F & 1 << s)); s++);
                            for (this.F >= 1 << this.c.length && (this.c.push(new M.hash.sha256), this.i.push(0)), this.d -= i, i > this.j && (this.j = i), this.F++, this.b = M.hash.sha256.hash(this.b.concat(o)), this.A = new M.cipher.aes(this.b), n = 0; 4 > n && (this.f[n] = this.f[n] + 1 | 0, !this.f[n]); n++);
                        }
                        for (n = 0; n < e; n += 4) 0 == (n + 1) % this.S && N(this), o = V(this), r.push(o[0], o[1], o[2], o[3]);
                        return N(this), r.slice(0, e)
                    },
                    setDefaultParanoia: function(e, t) {
                        0 === e && "Setting paranoia=0 will ruin your security; use it only for testing" !== t && k("Setting paranoia=0 will ruin your security; use it only for testing"), this.B = e
                    },
                    addEntropy: function(e, t, n) {
                        n = n || "user";
                        var o, r, s = (new Date).valueOf(),
                            i = this.s[n],
                            a = this.isReady(),
                            u = 0;
                        switch ((o = this.K[n]) === D && (o = this.K[n] = this.W++), i === D && (i = this.s[n] = 0), this.s[n] = (this.s[n] + 1) % this.c.length, typeof e) {
                            case "number":
                                t === D && (t = 1), this.c[i].update([o, this.C++, 1, t, s, 1, 0 | e]);
                                break;
                            case "object":
                                if ("[object Uint32Array]" === (n = Object.prototype.toString.call(e))) {
                                    for (r = [], n = 0; n < e.length; n++) r.push(e[n]);
                                    e = r
                                } else
                                    for ("[object Array]" !== n && (u = 1), n = 0; n < e.length && !u; n++) "number" != typeof e[n] && (u = 1);
                                if (!u) {
                                    if (t === D)
                                        for (n = t = 0; n < e.length; n++)
                                            for (r = e[n]; 0 < r;) t++, r >>>= 1;
                                    this.c[i].update([o, this.C++, 2, t, s, e.length].concat(e))
                                }
                                break;
                            case "string":
                                t === D && (t = e.length), this.c[i].update([o, this.C++, 3, t, s, e.length]), this.c[i].update(e);
                                break;
                            default:
                                u = 1
                        }
                        u && k(new M.exception.bug("random: addEntropy only supports number, array of numbers or string")), this.i[i] += t, this.d += t, a === this.m && (this.isReady() !== this.m && I("seeded", Math.max(this.j, this.d)), I("progress", this.getProgress()))
                    },
                    isReady: function(e) {
                        return e = this.I[e !== D ? e : this.B], this.j && this.j >= e ? this.i[0] > this.R && (new Date).valueOf() > this.O ? this.u | this.t : this.t : this.d >= e ? this.u | this.m : this.m
                    },
                    getProgress: function(e) {
                        return e = this.I[e || this.B], this.j >= e ? 1 : this.d > e ? 1 : this.d / e
                    },
                    startCollectors: function() {
                        this.q || (this.a = {
                            loadTimeCollector: B(this, this.aa),
                            mouseCollector: B(this, this.ba),
                            keyboardCollector: B(this, this.$),
                            accelerometerCollector: B(this, this.U)
                        }, window.addEventListener ? (window.addEventListener("load", this.a.loadTimeCollector, P), window.addEventListener("mousemove", this.a.mouseCollector, P), window.addEventListener("keypress", this.a.keyboardCollector, P), window.addEventListener("devicemotion", this.a.accelerometerCollector, P)) : document.attachEvent ? (document.attachEvent("onload", this.a.loadTimeCollector), document.attachEvent("onmousemove", this.a.mouseCollector), document.attachEvent("keypress", this.a.keyboardCollector)) : k(new M.exception.bug("can't attach event")), this.q = !0)
                    },
                    stopCollectors: function() {
                        this.q && (window.removeEventListener ? (window.removeEventListener("load", this.a.loadTimeCollector, P), window.removeEventListener("mousemove", this.a.mouseCollector, P), window.removeEventListener("keypress", this.a.keyboardCollector, P), window.removeEventListener("devicemotion", this.a.accelerometerCollector, P)) : document.detachEvent && (document.detachEvent("onload", this.a.loadTimeCollector), document.detachEvent("onmousemove", this.a.mouseCollector), document.detachEvent("keypress", this.a.keyboardCollector)), this.q = P)
                    },
                    addEventListener: function(e, t) {
                        this.w[e][this.V++] = t
                    },
                    removeEventListener: function(e, t) {
                        var n, o, r = this.w[e],
                            s = [];
                        for (o in r) r.hasOwnProperty(o) && r[o] === t && s.push(o);
                        for (n = 0; n < s.length; n++) delete r[o = s[n]]
                    },
                    $: function() {
                        L(1)
                    },
                    ba: function(e) {
                        var t, n;
                        try {
                            t = e.x || e.clientX || e.offsetX || 0, n = e.y || e.clientY || e.offsetY || 0
                        } catch (e) {
                            n = t = 0
                        }
                        0 != t && 0 != n && M.random.addEntropy([t, n], 2, "mouse"), L(0)
                    },
                    aa: function() {
                        L(2)
                    },
                    U: function(e) {
                        if (e = (e.accelerationIncludingGravity || {}).x || (e.accelerationIncludingGravity || {}).y || (e.accelerationIncludingGravity || {}).z, window.orientation) {
                            var t = window.orientation;
                            "number" == typeof t && M.random.addEntropy(t, 1, "accelerometer")
                        }
                        e && M.random.addEntropy(e, 2, "accelerometer"), L(0)
                    }
                }, M.random = new M.prng(6);
                e: try {
                    var $, q, z, U;
                    if (U = void 0 !== e) {
                        var K;
                        if (K = e.exports) {
                            var G;
                            try {
                                G = n("./node_modules/node-libs-browser/mock/empty.js")
                            } catch (e) {
                                G = null
                            }
                            K = (q = G) && q.randomBytes
                        }
                        U = K
                    }
                    if (U) $ = q.randomBytes(128), $ = new Uint32Array(new Uint8Array($).buffer), M.random.addEntropy($, 1024, "crypto['randomBytes']");
                    else if ("undefined" != typeof window && "undefined" != typeof Uint32Array) {
                        if (z = new Uint32Array(32), window.crypto && window.crypto.getRandomValues) window.crypto.getRandomValues(z);
                        else {
                            if (!window.msCrypto || !window.msCrypto.getRandomValues) break e;
                            window.msCrypto.getRandomValues(z)
                        }
                        M.random.addEntropy(z, 1024, "crypto['getRandomValues']")
                    }
                } catch (e) {
                    "undefined" != typeof window && window.console && (console.log("There was an error collecting entropy from the browser:"), console.log(e))
                }
                M.json = {
                        defaults: {
                            v: 1,
                            iter: 1e3,
                            ks: 128,
                            ts: 64,
                            mode: "ccm",
                            adata: "",
                            cipher: "aes"
                        },
                        Y: function(e, t, n, o) {
                            n = n || {}, o = o || {};
                            var r, s = M.json,
                                i = s.e({
                                    iv: M.random.randomWords(4, 0)
                                }, s.defaults);
                            return s.e(i, n), n = i.adata, "string" == typeof i.salt && (i.salt = M.codec.base64.toBits(i.salt)), "string" == typeof i.iv && (i.iv = M.codec.base64.toBits(i.iv)), (!M.mode[i.mode] || !M.cipher[i.cipher] || "string" == typeof e && 100 >= i.iter || 64 !== i.ts && 96 !== i.ts && 128 !== i.ts || 128 !== i.ks && 192 !== i.ks && 256 !== i.ks || 2 > i.iv.length || 4 < i.iv.length) && k(new M.exception.invalid("json encrypt: invalid parameters")), "string" == typeof e ? (e = (r = M.misc.cachedPbkdf2(e, i)).key.slice(0, i.ks / 32), i.salt = r.salt) : M.ecc && e instanceof M.ecc.elGamal.publicKey && (r = e.kem(), i.kemtag = r.tag, e = r.key.slice(0, i.ks / 32)), "string" == typeof t && (t = M.codec.utf8String.toBits(t)), "string" == typeof n && (n = M.codec.utf8String.toBits(n)), r = new M.cipher[i.cipher](e), s.e(o, i), o.key = e, i.ct = M.mode[i.mode].encrypt(r, t, i.iv, n, i.ts), i
                        },
                        encrypt: function(e, t, n, o) {
                            var r = M.json,
                                s = r.Y.apply(r, arguments);
                            return r.encode(s)
                        },
                        X: function(e, t, n, o) {
                            n = n || {}, o = o || {};
                            var r, s, i = M.json;
                            return r = (t = i.e(i.e(i.e({}, i.defaults), t), n, !0)).adata, "string" == typeof t.salt && (t.salt = M.codec.base64.toBits(t.salt)), "string" == typeof t.iv && (t.iv = M.codec.base64.toBits(t.iv)), (!M.mode[t.mode] || !M.cipher[t.cipher] || "string" == typeof e && 100 >= t.iter || 64 !== t.ts && 96 !== t.ts && 128 !== t.ts || 128 !== t.ks && 192 !== t.ks && 256 !== t.ks || !t.iv || 2 > t.iv.length || 4 < t.iv.length) && k(new M.exception.invalid("json decrypt: invalid parameters")), "string" == typeof e ? (e = (s = M.misc.cachedPbkdf2(e, t)).key.slice(0, t.ks / 32), t.salt = s.salt) : M.ecc && e instanceof M.ecc.elGamal.secretKey && (e = e.unkem(M.codec.base64.toBits(t.kemtag)).slice(0, t.ks / 32)), "string" == typeof r && (r = M.codec.utf8String.toBits(r)), s = new M.cipher[t.cipher](e), r = M.mode[t.mode].decrypt(s, t.ct, t.iv, r, t.ts), i.e(o, t), o.key = e, 1 === n.raw ? r : M.codec.utf8String.fromBits(r)
                        },
                        decrypt: function(e, t, n, o) {
                            var r = M.json;
                            return r.X(e, r.decode(t), n, o)
                        },
                        encode: function(e) {
                            var t, n = "{",
                                o = "";
                            for (t in e)
                                if (e.hasOwnProperty(t)) switch (t.match(/^[a-z0-9]+$/i) || k(new M.exception.invalid("json encode: invalid property name")), n += o + '"' + t + '":', o = ",", typeof e[t]) {
                                    case "number":
                                    case "boolean":
                                        n += e[t];
                                        break;
                                    case "string":
                                        n += '"' + escape(e[t]) + '"';
                                        break;
                                    case "object":
                                        n += '"' + M.codec.base64.fromBits(e[t], 0) + '"';
                                        break;
                                    default:
                                        k(new M.exception.bug("json encode: unsupported type"))
                                }
                            return n + "}"
                        },
                        decode: function(e) {
                            (e = e.replace(/\s/g, "")).match(/^\{.*\}$/) || k(new M.exception.invalid("json decode: this isn't json!")), e = e.replace(/^\{|\}$/g, "").split(/,/);
                            var t, n, o = {};
                            for (t = 0; t < e.length; t++)(n = e[t].match(/^(?:(["']?)([a-z][a-z0-9]*)\1):(?:(\d+)|"([a-z0-9+\/%*_.@=\-]*)")$/i)) || k(new M.exception.invalid("json decode: this isn't json!")), o[n[2]] = n[3] ? parseInt(n[3], 10) : n[2].match(/^(ct|salt|iv)$/) ? M.codec.base64.toBits(n[4]) : unescape(n[4]);
                            return o
                        },
                        e: function(e, t, n) {
                            if (e === D && (e = {}), t === D) return e;
                            for (var o in t) t.hasOwnProperty(o) && (n && e[o] !== D && e[o] !== t[o] && k(new M.exception.invalid("required parameter overridden")), e[o] = t[o]);
                            return e
                        },
                        ea: function(e, t) {
                            var n, o = {};
                            for (n in e) e.hasOwnProperty(n) && e[n] !== t[n] && (o[n] = e[n]);
                            return o
                        },
                        da: function(e, t) {
                            var n, o = {};
                            for (n = 0; n < t.length; n++) e[t[n]] !== D && (o[t[n]] = e[t[n]]);
                            return o
                        }
                    }, M.encrypt = M.json.encrypt, M.decrypt = M.json.decrypt, M.misc.ca = {}, M.misc.cachedPbkdf2 = function(e, t) {
                        var n, o = M.misc.ca;
                        return n = (t = t || {}).iter || 1e3, (n = (o = o[e] = o[e] || {})[n] = o[n] || {
                            firstSalt: t.salt && t.salt.length ? t.salt.slice(0) : M.random.randomWords(2, 0)
                        })[o = t.salt === D ? n.firstSalt : t.salt] = n[o] || M.misc.pbkdf2(e, o, t.iter), {
                            key: n[o].slice(0),
                            salt: o.slice(0)
                        }
                    }, (S = (O = M).codec.bytes = O.codec.bytes || {}).fromBits = S.fromBits || function(e) {
                        var t, n, o = [],
                            r = O.bitArray.bitLength(e);
                        for (t = 0; t < r / 8; t++) 0 == (3 & t) && (n = e[t / 4]), o.push(n >>> 24), n <<= 8;
                        return o
                    }, S.toBits = S.toBits || function(e) {
                        var t, n = [],
                            o = 0;
                        for (t = 0; t < e.length; t++) o = o << 8 | e[t], 3 == (3 & t) && (n.push(o), o = 0);
                        return 3 & t && n.push(O.bitArray.partial(8 * (3 & t), o)), n
                    },
                    function() {
                        var e, t = (new Date).getTime();

                        function n(e, t, n, o) {
                            if ("function" == typeof e.addEventListener) e.addEventListener(t, n, o);
                            else {
                                if (!e.attachEvent) throw new Error(H.errors.UNABLETOBIND + ": Unable to bind " + t + "-event");
                                e.attachEvent("on" + t, n)
                            }
                        }
                        T = T || (e = {}, function(o, r, s) {
                            if ("bind" === o) return T(s + "Bind"), n(r, "change", (function(e) {
                                T(s + "FieldChangeCount"), T("log", s, "ch");
                                try {
                                    T("set", s + "FieldEvHa", function(e) {
                                        var t = function() {
                                            return {}
                                        };
                                        window.jQuery && "function" == typeof window.jQuery._data && (t = function(e) {
                                            return window.jQuery._data(e, "events")
                                        });
                                        for (var n = e, o = 0, r = [], s = ["onmousedown", "onmouseup", "onmouseover", "onmouseout", "onclick", "onmousemove", "ondblclick", "onerror", "onresize", "onscroll", "onkeydown", "onkeyup", "onkeypress", "onchange", "onsubmit"], i = s.length; n && n !== n.documentElement;) {
                                            for (var a, u = i, l = (n.nodeName || n.tagName || "").toUpperCase().substring(0, 3); u--;) c = s[u], n[name] && (o++, r[c = c + (n === e ? "Own" : "Par") + l] = r[c] || 0, r[c]++);
                                            var d = t(n);
                                            if ("object" == typeof d)
                                                for (var c in d) d.hasOwnProperty(c) && (a = d[c].length, r[c = c + (n === e ? "Own" : "Par") + l] = r[c] || 0, r[c] += a, o += a);
                                            if (!n.parentNode) break;
                                            n = n.parentNode
                                        }
                                        var h = ["total=" + o];
                                        for (var f in r) r.hasOwnProperty(f) && r[f] > 0 && h.push(f + "=" + r[f]);
                                        return h.join("&")
                                    }(r))
                                } catch (e) {
                                    T("set", s + "FieldEvHa", "Err")
                                }
                            }), !0), n(r, "click", (function() {
                                T(s + "FieldClickCount"), T("log", s, "cl")
                            }), !0), n(r, "focus", (function() {
                                T(s + "FieldFocusCount"), T("log", s, "fo")
                            }), !0), n(r, "blur", (function() {
                                T(s + "FieldBlurCount"), T("log", s, "bl")
                            }), !0), n(r, "touchstart", (function() {
                                T(s + "FieldTouchStartCount"), T("log", s, "Ts")
                            }), !0), n(r, "touchend", (function() {
                                T(s + "FieldTouchEndCount"), T("log", s, "Te")
                            }), !0), n(r, "touchcancel", (function() {
                                T(s + "FieldTouchCancelCount"), T("log", s, "Tc")
                            }), !0), n(r, "keyup", (function(e) {
                                16 == e.keyCode ? T("log", s, "Su") : 17 == e.keyCode ? T("log", s, "Cu") : 18 == e.keyCode && T("log", s, "Au")
                            })), void n(r, "keydown", (function(e) {
                                switch (T(s + "FieldKeyCount"), e && e.keyCode) {
                                    case 8:
                                        T("log", s, "Kb");
                                        break;
                                    case 16:
                                        T("log", s, "Sd");
                                        break;
                                    case 17:
                                        T("log", s, "Cd");
                                        break;
                                    case 18:
                                        T("log", s, "Ad");
                                        break;
                                    case 37:
                                        T("log", s, "Kl");
                                        break;
                                    case 39:
                                        T("log", s, "Kr");
                                        break;
                                    case 46:
                                        T("log", s, "Kd");
                                        break;
                                    case 32:
                                        T("log", s, "Ks");
                                        break;
                                    default:
                                        e.keyCode >= 48 && e.keyCode <= 57 || e.keyCode >= 96 && e.keyCode <= 105 ? T("log", s, "KN") : e.keyCode >= 65 && e.keyCode <= 90 ? T("log", s, "KL") : (T("log", s, "KU"), T("log", s + "UnkKeys", e.keyCode))
                                }
                            }), !0);
                            if ("set" !== o) {
                                if ("log" === o) {
                                    var i = r + "FieldLog",
                                        a = (new Date).getTime() - t;
                                    return a = Math.round(a / 100), e.hasOwnProperty(i) ? e[i] += "," + s + "@" + a : e[i] = s + "@" + a, void(e[i].length > 1500 && (e[i] = e[i].substring(e[i].length - 1500), e[i] = e[i].substring(e[i].indexOf(",") + 1)))
                                }
                                if ("extend" !== o) e.hasOwnProperty(o) ? e[o]++ : e[o] = 1;
                                else
                                    for (var u in e) "number" !== u && "expiryMonth" !== u && "expiryYear" !== u && "generationtime" !== u && "holderName" !== u && "cvc" !== u && e.hasOwnProperty(u) && (r[u] = "" + e[u])
                            } else e[r] = s
                        }), window && (window.attachEvent || window.addEventListener) && (n(window, "focus", (function() {
                            T("activate"), window.location && "string" == typeof window.location.href && T("set", "referrer", window.location.href)
                        })), n(window, "blur", (function() {
                            T("deactivate")
                        })))
                    }();
                var W = t.adyen = t.adyen || {},
                    H = W.encrypt = W.encrypt || {
                        createEncryption: function(e, t) {
                            return new J(e, t)
                        }
                    };
                 e.exports && (e.exports = H), H.errors = H.errors || {}, H.version = "0_1_21";
                var Z, Y = {};
                Y.luhnCheck = (Z = {}, function() {
                    var e = arguments,
                        t = arguments.length,
                        n = t > 0 ? e[0] : this.cardnumber;
                    if (isNaN(parseInt(n, 10))) return !1;
                    var o = n.length,
                        r = 1 & o,
                        s = 0;
                    if (void 0 === Z[n]) {
                        o >= 14 && T("luhnCount");
                        for (var i = 0; i < o; i++) {
                            var a = parseInt(n.charAt(i), 10);
                            1 & i ^ r || (a *= 2) > 9 && (a -= 9), s += a
                        }
                        s % 10 == 0 ? (T("luhnOkCount"), Z[n] = !0) : (T("luhnFailCount"), Z[n] = !1)
                    }
                    var u = 0;
                    for (var l in Z) Z.hasOwnProperty(l) && l.length === o && u++;
                    return T("set", "luhnSameLengthCount", u), Z[n]
                }), Y.numberCheck = function(e) {
                    return !(!(e || "").replace(/[^\d]/g, "").match(/^\d{10,20}$/) || !Y.luhnCheck(e))
                }, Y.cvcCheck = function(e) {
                    return !!(e && e.match && e.match(/^\d{3,4}$/))
                }, Y.yearCheck = function(e) {
                    if (!e || !e.match || !e.match(/^2\d{3}$/)) return !1;
                    var t = parseInt(e, 10),
                        n = (new Date).getFullYear();
                    return t >= n - 2 && t <= n + 15
                }, Y.monthCheck = function(e) {
                    var t = (e || "").replace(/^0(\d)$/, "$1");
                    return !!(t.match(/^([1-9]|10|11|12)$/) && parseInt(t, 10) >= 1 && parseInt(t, 10) <= 12)
                }, Y.holderNameCheck = function(e) {
                    return !!(e && e.match && e.match(/\S/))
                }, Y.generationTimeValidDate = function(e) {
                    if ("string" != typeof e) return !1;
                    var t = e.match(/^(\d{4})-?(\d{2})-?(\d{2})$/);
                    return !!(t && ("" + t[1]).match(/^20[1-9][0-9]$/) && ("" + t[2]).match(/^(12|11|10|0[1-9])$/) && ("" + t[3]).match(/^(31|30|20|10|[012][1-9])$/))
                }, Y.generationTimeValidTime = function(e) {
                    if ("string" != typeof e) return !1;
                    return e.replace(/(Z|[\+\-][012345][0-9]:?[012345][0-9])$/, "").replace(/\.\d+$/, "").match(/^[012345][0-9]:?[012345][0-9]:?[012345][0-9]$/)
                }, Y.generationTimeCheck = function(e) {
                    if ("string" != typeof e) return !1;
                    var t = e.split("T");
                    return !(2 !== t.length || !Y.generationTimeValidDate(t[0]) || !Y.generationTimeValidTime(t[1]))
                };
                var J = function(e, t) {
                  // t.randomBytes = function(n) {
                    // console.trace();
                    // return Uint32Array(n);
                  // }
                    try {
                        t.randomBytes && M.random.addEntropy(t.randomBytes, 1024, "crypto.randomBytes"), M.random.startCollectors()
                    } catch (e) {}
                    if (this.key = e, this.options = t || {}, void 0 === this.options.numberIgnoreNonNumeric && (this.options.numberIgnoreNonNumeric = !0), void 0 !== this.options.cvcIgnoreFornumber && delete this.options.cvcIgnoreFornumber, void 0 === this.options.fourDigitCvcForBins && (this.options.fourDigitCvcForBins = "34,37"), void 0 !== this.options.cvcLengthFornumber && delete this.options.cvcLengthFornumber, "string" == typeof this.options.cvcIgnoreBins) {
                        var n = [];
                        this.options.cvcIgnoreBins.replace(/\d+/g, (function(e) {
                            return e.length > 0 && !isNaN(parseInt(e, 10)) && n.push(e), e
                        })), n.length > 0 && (this.options.cvcIgnoreFornumber = new RegExp("^\\s*(" + n.join("|") + ")"))
                    } else void 0 !== this.options.cvcIgnoreBins && delete this.options.cvcIgnoreBins;
                    if ("string" == typeof this.options.fourDigitCvcForBins) {
                        var o = [];
                        this.options.fourDigitCvcForBins.replace(/\d+/g, (function(e) {
                            return e.length > 0 && !isNaN(parseInt(e, 10)) && o.push(e), e
                        })), o.length > 0 && (this.options.cvcLengthFornumber = {
                            matcher: new RegExp("^\\s*(" + o.join("|") + ")"),
                            requiredLength: 4
                        })
                    }
                    delete this.options.fourDigitCvcForBins, T("initializeCount")
                };
                J.prototype.createRSAKey = function() {
                    var e = this.key.split("|");
                    if (2 !== e.length) throw "Malformed public key";
                    var t = e[0],
                        n = e[1],
                        o = new A;
                    return o.setPublic(n, t), o
                }, J.prototype.createAESKey = function() {
                    return new Q
                }, J.prototype.encrypt = function(e) {
                    var t = {};
                    for (var n in e) e.hasOwnProperty(n) && (t[n] = e[n]);
                    var o, r, s, i, a, u = {};
                    if (void 0 !== t.number && (u.number = t.number), void 0 !== t.cvc && (u.cvc = t.cvc), void 0 !== t.expiryMonth && (u.month = t.expiryMonth), void 0 !== t.expiryYear && (u.year = t.expiryYear), void 0 !== t.holderName && (u.holderName = t.holderName), void 0 !== t.generationtime && (u.generationtime = t.generationtime), !1 !== this.options.enableValidations && !1 === this.validate(u).valid) return !1;
                    for (var l = 0; l < 11 && (M.random && M.random.isReady(l)); l++) T("set", "sjclStrength", l);
                    T("extend", t);
                    try {
                        t.dfValue = ""
                    } catch (e) {}
                    return o = this.createRSAKey(), s = (r = this.createAESKey()).encrypt(JSON.stringify(t)), i = M.codec.bytes.fromBits(r.key()), a = o.encrypt_b64(i), ["adyenjs_" + H.version + "$", a, "$", s].join("")
                }, J.prototype.validate = function(e) {
                    var t = {
                        valid: !0
                    };
                    if ("object" != typeof e) return t.valid = !1, t;
                    for (var n in e)
                        if (e.hasOwnProperty(n) && void 0 !== e[n]) {
                            var o = e[n];
                            if (this.options[n + "IgnoreNonNumeric"] && (o = o.replace(/\D/g, "")), !this.options[n + "SkipValidation"]) {
                                for (var r in e)
                                    if (e.hasOwnProperty(r)) {
                                        var s = this.options[n + "IgnoreFor" + r],
                                            i = this.options[n + "LengthFor" + r];
                                        if (s && e[r].match(s)) {
                                            t[n] = !0;
                                            continue
                                        }
                                        if (i && i.matcher && i.requiredLength && e[r].match(i.matcher) && o.length !== i.requiredLength) {
                                            t[n] = !1;
                                            continue
                                        }
                                    } if (t.hasOwnProperty(n)) t.valid = t.valid && t[n];
                                else switch (n) {
                                    case "number":
                                        t.number = Y.numberCheck(o), t.luhn = t.number, t.valid = t.valid && t.number;
                                        break;
                                    case "expiryYear":
                                    case "year":
                                        t.year = Y.yearCheck(o), t.expiryYear = t.year, t.valid = t.valid && t.year;
                                        break;
                                    case "cvc":
                                        t.cvc = Y.cvcCheck(o), t.valid = t.valid && t.cvc;
                                        break;
                                    case "expiryMonth":
                                    case "month":
                                        t.month = Y.monthCheck(o), t.expiryMonth = t.month, t.valid = t.valid && t.month;
                                        break;
                                    case "holderName":
                                        t.holderName = Y.holderNameCheck(o), t.valid = t.valid && t.holderName;
                                        break;
                                    case "generationtime":
                                        t.generationtime = Y.generationTimeCheck(o), t.valid = t.valid && t.generationtime;
                                        break;
                                    default:
                                        t.unknown = t.unknown || [], t.unknown.push(n), t.valid = !1
                                }
                            }
                        } return t
                }, J.prototype.monitor = function(e, t) {
                    if ("string" != typeof e || "number" !== e && "cvc" !== e && "holderName" !== e) throw new Error("invalid fieldname. Expected 'number', 'cvc' or 'holderName', but received '" + e + "'");
                    T("bind", t, e)
                };
                var Q = function() {};
                Q.prototype = {
                    constructor: Q,
                    key: function() {
                        return this._key = this._key || M.random.randomWords(8, 6), this._key
                    },
                    encrypt: function(e) {
                        return this.encryptWithIv(e, M.random.randomWords(3, 6))
                    },
                    encryptWithIv: function(e, t) {
                        var n, o, r, s;
                        return n = new M.cipher.aes(this.key()), o = M.codec.utf8String.toBits(e), r = M.mode.ccm.encrypt(n, o, t), s = M.bitArray.concat(t, r), M.codec.base64.fromBits(s)
                    }
                }
            }
var jj = {}
AdyenEnc(jj, null);
var card = JSON.parse(Deno.args[0]);
var mtime = Deno.args[1];
card.expirationMonth = card.expirationMonth.toString();
card.expirationYear = card.expirationYear.toString();
var cardEnc = jj.adyen.encrypt.createEncryption(adyenPublicKey, {}).encrypt({
    number: card.number,
    cvc: card.cvc,
    holderName: card.holder,
    expiryMonth: card.expirationMonth,
    expiryYear: card.expirationYear,
    generationtime: mtime,
    paymentMethodId: card.paymentMethodId,
    cardType: card.cardType
})
console.log(cardEnc);
