export default function(window, navigator, jQuery) {
    var VantiveProtectPpStatsReporter = function() {
        var b = 0;
        var c = 3;
        var a = 0;
        return {
            reportMethodInvocation: function(e) {
                b++;
                if (b > c) {
                    return
                }
                e = encodeURIComponent(e);
                var d = this.getUriEncodedStack(new Error());
                var f = "https://request.eprotect.vantivcnp.com";
                var g = "errorHandler=PUBLIC_API_CALL&errorStack=" + d + "&errorMessage=" + e;
                setTimeout(function() {
                    try {
                        jQuery.getJSON(f + "/eProtect/ppstats?" + g + "&jsoncallback=?", function(i) {})
                    } catch (h) {}
                }, 0)
            },
            report3rdPartyError: function(d) {
                this.reportError(d, "3RD_PARTY_ERROR")
            },
            reportIframeClientError: function(d) {
                this.reportError(d, "CLIENT_ERROR")
            },
            reportError: function(h, d) {
                a++;
                if (a > c) {
                    return
                }
                var f = encodeURIComponent(h);
                f = this.removeNonStandardCharacters(f);
                var e = "https://request.eprotect.vantivcnp.com";
                var i = "errorHandler=" + d;
                var g = this.getUriEncodedStack(h);
                f += encodeURIComponent(" ") + g;
                if (f.length > 3000) {
                    f = f.substr(0, 3000)
                }
                i += "&errorStack=" + f;
                setTimeout(function() {
                    try {
                        jQuery.getJSON(e + "/eProtect/ppstats?" + i + "&jsoncallback=?", function(k) {})
                    } catch (j) {}
                }, 0)
            },
            removeNonStandardCharacters: function(e) {
                var d = decodeURIComponent(e);
                d = d.replace(/(\r\n|\n|\r)/g, " ");
                return encodeURIComponent(d)
            },
            executeActionAndReportError: function(f) {
                try {
                    f()
                } catch (e) {
                    try {
                        this.reportIframeClientError(e)
                    } catch (d) {}
                    throw e
                }
            },
            getUriEncodedStack: function(f) {
                var d = "";
                if (f.stack) {
                    d = f.stack;
                    if (d.length > 2000) {
                        d = d.substr(0, 2000)
                    }
                    try {
                        var g = encodeURIComponent(d);
                        g = this.removeNonStandardCharacters(d);
                        d = g
                    } catch (e) {}
                }
                return d
            }
        }
    };
    var myVantivEProtectReporterForPpStats = new VantiveProtectPpStatsReporter();
    window.eProtect = function() {
        var b = {
            modulus: "af4917a55f4ba0918309e972e8a6ddb0f67624e1c3541555c81ceb584040fe5f6c374daa313e05100fce9bf27797e3853f14b94e30f15fb3fa8402b57b38627527755e356a9099acfc8cee716de850e26773d56222ec96dfe3bafb6ae2c2d515c5daa31bb380fc4e6e90eb0c042ed2ce73c0812788c38bb922c6cf1f01d0f5d8395bc0f4028a5eb3982392eeddf53a4a6bf4349366ff8e9df552df9305ddefd12e3c7a585587a4cbecd49b4e3410fd217b4d50b72cf61980d58cef1f648af568775c01ddf44f52a2703e018b209d57699529667db089ebf6c499d2568d4cc74e6380fed99b3264c36459f7e4c5d2934025903f878977f08a6d2c6fd7cc7f0a07",
            exponent: "10001",
            keyId: "27223200042",
            visaCheckoutApiKey: "WPKSQ4SVKENZW38OQJ0S14Vg8m6kpdvHrRPzD-Ch4tXQ429zs"
        };
        var a = {
            primaryUrl: "https://request.eprotect.vantivcnp.com",
            secondaryUrl: "https://secondary.eprotect.vantivcnp.com",
            primaryTimeout: 5000
        };
        return {
            getUrl: function() {
                return a.primaryUrl
            },
            sendToEprotect: function(aM, ap, aj, ah, M, ar) {
                function a9(bo) {
                    var bh = 0;
                    var bf;
                    var bs;
                    var br;
                    var bm;
                    var bn;
                    try {
                        if (window.crypto && window.crypto.getRandomValues) {
                            bf = new Int8Array(bo.length);
                            window.crypto.getRandomValues(bf);
                            for (bs = 0; bs < bf.length; ++bs) {
                                while (bf[bs] == 0) {
                                    bm = new Int8Array(1);
                                    window.crypto.getRandomValues(bm);
                                    bf[bs] = bm[0]
                                }
                                bo[bh++] = bf[bs]
                            }
                        } else {
                            if (window.msCrypto && window.msCrypto.getRandomValues) {
                                bf = new Int8Array(bo.length);
                                window.msCrypto.getRandomValues(bf);
                                for (bs = 0; bs < bf.length; ++bs) {
                                    while (bf[bs] == 0) {
                                        bm = new Int8Array(1);
                                        window.msCrypto.getRandomValues(bm);
                                        bf[bs] = bm[0]
                                    }
                                    bo[bh++] = bf[bs]
                                }
                            } else {
                                br = sjcl.random.randomWords((bo.length / 4) + 1, 0);
                                var bk = 0;
                                while (bh < br.length) {
                                    var bi = br[bh++];
                                    var bp = bi >> 0 & 255;
                                    var bq = bi >> 8 & 255;
                                    var be = bi >> 16 & 255;
                                    var bj = bi >> 24 & 255;
                                    while (bp == 0 || bq == 0 || be == 0 || bj == 0) {
                                        bn = new Array();
                                        bn = sjcl.random.randomWords(1, 0);
                                        bi = bn[0];
                                        bp = bi >> 0 & 255;
                                        bq = bi >> 8 & 255;
                                        be = bi >> 16 & 255;
                                        bj = bi >> 24 & 255
                                    }
                                    if (bk < bo.length) {
                                        bo[bk++] = bp
                                    }
                                    if (bk < bo.length) {
                                        bo[bk++] = bq
                                    }
                                    if (bk < bo.length) {
                                        bo[bk++] = be
                                    }
                                    if (bk < bo.length) {
                                        bo[bk++] = bj
                                    }
                                }
                            }
                        }
                    } catch (bl) {
                        for (bs = 0; bs < bo.length; ++bs) {
                            var bg = Math.floor((Math.random() * 255) + 1);
                            while (bg == 0) {
                                bg = Math.floor((Math.random() * 255) + 1)
                            }
                            bo[bh++] = bg
                        }
                    }
                    return 1
                }
                /*
           * Copyright (c) 2003-2005  Tom Wu
           * All Rights Reserved.
           *
           * Permission is hereby granted, free of charge, to any person obtaining
           * a copy of this software and associated documentation files (the
           * "Software"), to deal in the Software without restriction, including
           * without limitation the rights to use, copy, modify, merge, publish,
           * distribute, sublicense, and/or sell copies of the Software, and to
           * permit persons to whom the Software is furnished to do so, subject to
           * the following conditions:
           *
           * The above copyright notice and this permission notice shall be
           * included in all copies or substantial portions of the Software.
           *
           * THE SOFTWARE IS PROVIDED "AS-IS" AND WITHOUT WARRANTY OF ANY KIND,
           * EXPRESS, IMPLIED OR OTHERWISE, INCLUDING WITHOUT LIMITATION, ANY
           * WARRANTY OF MERCHANTABILITY OR FITNESS FOR A PARTICULAR PURPOSE.
           *
           * IN NO EVENT SHALL TOM WU BE LIABLE FOR ANY SPECIAL, INCIDENTAL,
           * INDIRECT OR CONSEQUENTIAL DAMAGES OF ANY KIND, OR ANY DAMAGES WHATSOEVER
           * RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER OR NOT ADVISED OF
           * THE POSSIBILITY OF DAMAGE, AND ON ANY THEORY OF LIABILITY, ARISING OUT
           * OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
           *
           * In addition, the following condition applies:
           *
           * All redistributions must retain an intact copy of this copyright notice
           * and disclaimer.
           */
                function h(bf, be) {
                    return new a1(bf,be)
                }
                function aP(bg, bh) {
                    var be = "";
                    var bf = 0;
                    while (bf + bh < bg.length) {
                        be += bg.substring(bf, bf + bh) + "\n";
                        bf += bh
                    }
                    return be + bg.substring(bf, bg.length)
                }
                function x(be) {
                    if (be < 16) {
                        return "0" + be.toString(16)
                    } else {
                        return be.toString(16)
                    }
                }
                function aL(bh, bk) {
                    if (bk < bh.length + 11) {
                        throw "Message too long for RSA"
                    }
                    var bj = new Array();
                    var bg = bh.length - 1;
                    while (bg >= 0 && bk > 0) {
                        var bi = bh.charCodeAt(bg--);
                        if (bi < 128) {
                            bj[--bk] = bi
                        } else {
                            if ((bi > 127) && (bi < 2048)) {
                                bj[--bk] = (bi & 63) | 128;
                                bj[--bk] = (bi >> 6) | 192
                            } else {
                                bj[--bk] = (bi & 63) | 128;
                                bj[--bk] = ((bi >> 6) & 63) | 128;
                                bj[--bk] = (bi >> 12) | 224
                            }
                        }
                    }
                    bj[--bk] = 0;
                    var bf = new aF();
                    var be = new Array(bk - 2);
                    bf.nextBytes(be);
                    bg = 0;
                    while (bk > 2) {
                        bj[--bk] = be[bg];
                        bg++
                    }
                    bj[--bk] = 2;
                    bj[--bk] = 0;
                    return new a1(bj)
                }
                function X() {
                    this.n = null;
                    this.e = 0;
                    this.d = null;
                    this.p = null;
                    this.q = null;
                    this.dmp1 = null;
                    this.dmq1 = null;
                    this.coeff = null
                }
                function s(bf, be) {
                    if (bf != null && be != null && bf.length > 0 && be.length > 0) {
                        this.n = h(bf, 16);
                        this.e = parseInt(be, 16)
                    } else {
                        throw "Error setting public key"
                    }
                }
                function at(be) {
                    return be.modPowInt(this.e, this.n)
                }
                function u(bg) {
                    var be = aL(bg, (this.n.bitLength() + 7) >> 3);
                    if (be == null) {
                        return null
                    }
                    var bh = this.doPublic(be);
                    if (bh == null) {
                        return null
                    }
                    var bf = bh.toString(16);
                    if ((bf.length & 1) == 0) {
                        return bf
                    } else {
                        return "0" + bf
                    }
                }
                X.prototype.doPublic = at;
                X.prototype.setPublic = s;
                X.prototype.encrypt = u;
                var ba;
                var aQ = 244837814094590;
                var ax = ((aQ & 16777215) == 15715070);
                function a1(bf, be, bg) {
                    if (bf != null) {
                        if ("number" == typeof bf) {
                            this.fromNumber(bf, be, bg)
                        } else {
                            if (be == null && "string" != typeof bf) {
                                this.fromString(bf, 256)
                            } else {
                                this.fromString(bf, be)
                            }
                        }
                    }
                }
                function k() {
                    return new a1(null)
                }
                function c(bi, be, bf, bh, bk, bj) {
                    while (--bj >= 0) {
                        var bg = be * this[bi++] + bf[bh] + bk;
                        bk = Math.floor(bg / 67108864);
                        bf[bh++] = bg & 67108863
                    }
                    return bk
                }
                function bc(bi, bn, bo, bh, bl, be) {
                    var bk = bn & 32767
                      , bm = bn >> 15;
                    while (--be >= 0) {
                        var bg = this[bi] & 32767;
                        var bj = this[bi++] >> 15;
                        var bf = bm * bg + bj * bk;
                        bg = bk * bg + ((bf & 32767) << 15) + bo[bh] + (bl & 1073741823);
                        bl = (bg >>> 30) + (bf >>> 15) + bm * bj + (bl >>> 30);
                        bo[bh++] = bg & 1073741823
                    }
                    return bl
                }
                function bb(bi, bn, bo, bh, bl, be) {
                    var bk = bn & 16383
                      , bm = bn >> 14;
                    while (--be >= 0) {
                        var bg = this[bi] & 16383;
                        var bj = this[bi++] >> 14;
                        var bf = bm * bg + bj * bk;
                        bg = bk * bg + ((bf & 16383) << 14) + bo[bh] + bl;
                        bl = (bg >> 28) + (bf >> 14) + bm * bj;
                        bo[bh++] = bg & 268435455
                    }
                    return bl
                }
                if (ax && (navigator.appName == "Microsoft Internet Explorer")) {
                    a1.prototype.am = bc;
                    ba = 30
                } else {
                    if (ax && (navigator.appName != "Netscape")) {
                        a1.prototype.am = c;
                        ba = 26
                    } else {
                        a1.prototype.am = bb;
                        ba = 28
                    }
                }
                a1.prototype.DB = ba;
                a1.prototype.DM = ((1 << ba) - 1);
                a1.prototype.DV = (1 << ba);
                var aA = 52;
                a1.prototype.FV = Math.pow(2, aA);
                a1.prototype.F1 = aA - ba;
                a1.prototype.F2 = 2 * ba - aA;
                var aI = "0123456789abcdefghijklmnopqrstuvwxyz";
                var aO = new Array();
                var aX, E;
                aX = "0".charCodeAt(0);
                for (E = 0; E <= 9; ++E) {
                    aO[aX++] = E
                }
                aX = "a".charCodeAt(0);
                for (E = 10; E < 36; ++E) {
                    aO[aX++] = E
                }
                aX = "A".charCodeAt(0);
                for (E = 10; E < 36; ++E) {
                    aO[aX++] = E
                }
                function bd(be) {
                    return aI.charAt(be)
                }
                function I(bf, be) {
                    var bg = aO[bf.charCodeAt(be)];
                    return (bg == null) ? -1 : bg
                }
                function aw(bf) {
                    for (var be = this.t - 1; be >= 0; --be) {
                        bf[be] = this[be]
                    }
                    bf.t = this.t;
                    bf.s = this.s
                }
                function r(be) {
                    this.t = 1;
                    this.s = (be < 0) ? -1 : 0;
                    if (be > 0) {
                        this[0] = be
                    } else {
                        if (be < -1) {
                            this[0] = be + DV
                        } else {
                            this.t = 0
                        }
                    }
                }
                function f(be) {
                    var bf = k();
                    bf.fromInt(be);
                    return bf
                }
                function F(bk, bf) {
                    var bh;
                    if (bf == 16) {
                        bh = 4
                    } else {
                        if (bf == 8) {
                            bh = 3
                        } else {
                            if (bf == 256) {
                                bh = 8
                            } else {
                                if (bf == 2) {
                                    bh = 1
                                } else {
                                    if (bf == 32) {
                                        bh = 5
                                    } else {
                                        if (bf == 4) {
                                            bh = 2
                                        } else {
                                            this.fromRadix(bk, bf);
                                            return
                                        }
                                    }
                                }
                            }
                        }
                    }
                    this.t = 0;
                    this.s = 0;
                    var bj = bk.length
                      , bg = false
                      , bi = 0;
                    while (--bj >= 0) {
                        var be = (bh == 8) ? bk[bj] & 255 : I(bk, bj);
                        if (be < 0) {
                            if (bk.charAt(bj) == "-") {
                                bg = true
                            }
                            continue
                        }
                        bg = false;
                        if (bi == 0) {
                            this[this.t++] = be
                        } else {
                            if (bi + bh > this.DB) {
                                this[this.t - 1] |= (be & ((1 << (this.DB - bi)) - 1)) << bi;
                                this[this.t++] = (be >> (this.DB - bi))
                            } else {
                                this[this.t - 1] |= be << bi
                            }
                        }
                        bi += bh;
                        if (bi >= this.DB) {
                            bi -= this.DB
                        }
                    }
                    if (bh == 8 && (bk[0] & 128) != 0) {
                        this.s = -1;
                        if (bi > 0) {
                            this[this.t - 1] |= ((1 << (this.DB - bi)) - 1) << bi
                        }
                    }
                    this.clamp();
                    if (bg) {
                        a1.ZERO.subTo(this, this)
                    }
                }
                function Z() {
                    var be = this.s & this.DM;
                    while (this.t > 0 && this[this.t - 1] == be) {
                        --this.t
                    }
                }
                function w(bf) {
                    if (this.s < 0) {
                        return "-" + this.negate().toString(bf)
                    }
                    var bg;
                    if (bf == 16) {
                        bg = 4
                    } else {
                        if (bf == 8) {
                            bg = 3
                        } else {
                            if (bf == 2) {
                                bg = 1
                            } else {
                                if (bf == 32) {
                                    bg = 5
                                } else {
                                    if (bf == 4) {
                                        bg = 2
                                    } else {
                                        return this.toRadix(bf)
                                    }
                                }
                            }
                        }
                    }
                    var bi = (1 << bg) - 1, bl, be = false, bj = "", bh = this.t;
                    var bk = this.DB - (bh * this.DB) % bg;
                    if (bh-- > 0) {
                        if (bk < this.DB && (bl = this[bh] >> bk) > 0) {
                            be = true;
                            bj = bd(bl)
                        }
                        while (bh >= 0) {
                            if (bk < bg) {
                                bl = (this[bh] & ((1 << bk) - 1)) << (bg - bk);
                                bl |= this[--bh] >> (bk += this.DB - bg)
                            } else {
                                bl = (this[bh] >> (bk -= bg)) & bi;
                                if (bk <= 0) {
                                    bk += this.DB;
                                    --bh
                                }
                            }
                            if (bl > 0) {
                                be = true
                            }
                            if (be) {
                                bj += bd(bl)
                            }
                        }
                    }
                    return be ? bj : "0"
                }
                function af() {
                    var be = k();
                    a1.ZERO.subTo(this, be);
                    return be
                }
                function aU() {
                    return (this.s < 0) ? this.negate() : this
                }
                function R(be) {
                    var bg = this.s - be.s;
                    if (bg != 0) {
                        return bg
                    }
                    var bf = this.t;
                    bg = bf - be.t;
                    if (bg != 0) {
                        return bg
                    }
                    while (--bf >= 0) {
                        if ((bg = this[bf] - be[bf]) != 0) {
                            return bg
                        }
                    }
                    return 0
                }
                function m(be) {
                    var bg = 1, bf;
                    if ((bf = be >>> 16) != 0) {
                        be = bf;
                        bg += 16
                    }
                    if ((bf = be >> 8) != 0) {
                        be = bf;
                        bg += 8
                    }
                    if ((bf = be >> 4) != 0) {
                        be = bf;
                        bg += 4
                    }
                    if ((bf = be >> 2) != 0) {
                        be = bf;
                        bg += 2
                    }
                    if ((bf = be >> 1) != 0) {
                        be = bf;
                        bg += 1
                    }
                    return bg
                }
                function C() {
                    if (this.t <= 0) {
                        return 0
                    }
                    return this.DB * (this.t - 1) + m(this[this.t - 1] ^ (this.s & this.DM))
                }
                function aZ(bg, bf) {
                    var be;
                    for (be = this.t - 1; be >= 0; --be) {
                        bf[be + bg] = this[be]
                    }
                    for (be = bg - 1; be >= 0; --be) {
                        bf[be] = 0
                    }
                    bf.t = this.t + bg;
                    bf.s = this.s
                }
                function av(bg, bf) {
                    for (var be = bg; be < this.t; ++be) {
                        bf[be - bg] = this[be]
                    }
                    bf.t = Math.max(this.t - bg, 0);
                    bf.s = this.s
                }
                function B(bl, bh) {
                    var bf = bl % this.DB;
                    var be = this.DB - bf;
                    var bj = (1 << be) - 1;
                    var bi = Math.floor(bl / this.DB), bk = (this.s << bf) & this.DM, bg;
                    for (bg = this.t - 1; bg >= 0; --bg) {
                        bh[bg + bi + 1] = (this[bg] >> be) | bk;
                        bk = (this[bg] & bj) << bf
                    }
                    for (bg = bi - 1; bg >= 0; --bg) {
                        bh[bg] = 0
                    }
                    bh[bi] = bk;
                    bh.t = this.t + bi + 1;
                    bh.s = this.s;
                    bh.clamp()
                }
                function n(bk, bh) {
                    bh.s = this.s;
                    var bi = Math.floor(bk / this.DB);
                    if (bi >= this.t) {
                        bh.t = 0;
                        return
                    }
                    var bf = bk % this.DB;
                    var be = this.DB - bf;
                    var bj = (1 << bf) - 1;
                    bh[0] = this[bi] >> bf;
                    for (var bg = bi + 1; bg < this.t; ++bg) {
                        bh[bg - bi - 1] |= (this[bg] & bj) << be;
                        bh[bg - bi] = this[bg] >> bf
                    }
                    if (bf > 0) {
                        bh[this.t - bi - 1] |= (this.s & bj) << be
                    }
                    bh.t = this.t - bi;
                    bh.clamp()
                }
                function aB(bf, bh) {
                    var bg = 0
                      , bi = 0
                      , be = Math.min(bf.t, this.t);
                    while (bg < be) {
                        bi += this[bg] - bf[bg];
                        bh[bg++] = bi & this.DM;
                        bi >>= this.DB
                    }
                    if (bf.t < this.t) {
                        bi -= bf.s;
                        while (bg < this.t) {
                            bi += this[bg];
                            bh[bg++] = bi & this.DM;
                            bi >>= this.DB
                        }
                        bi += this.s
                    } else {
                        bi += this.s;
                        while (bg < bf.t) {
                            bi -= bf[bg];
                            bh[bg++] = bi & this.DM;
                            bi >>= this.DB
                        }
                        bi -= bf.s
                    }
                    bh.s = (bi < 0) ? -1 : 0;
                    if (bi < -1) {
                        bh[bg++] = this.DV + bi
                    } else {
                        if (bi > 0) {
                            bh[bg++] = bi
                        }
                    }
                    bh.t = bg;
                    bh.clamp()
                }
                function O(bf, bh) {
                    var be = this.abs()
                      , bi = bf.abs();
                    var bg = be.t;
                    bh.t = bg + bi.t;
                    while (--bg >= 0) {
                        bh[bg] = 0
                    }
                    for (bg = 0; bg < bi.t; ++bg) {
                        bh[bg + be.t] = be.am(0, bi[bg], bh, bg, 0, be.t)
                    }
                    bh.s = 0;
                    bh.clamp();
                    if (this.s != bf.s) {
                        a1.ZERO.subTo(bh, bh)
                    }
                }
                function ab(bg) {
                    var be = this.abs();
                    var bf = bg.t = 2 * be.t;
                    while (--bf >= 0) {
                        bg[bf] = 0
                    }
                    for (bf = 0; bf < be.t - 1; ++bf) {
                        var bh = be.am(bf, be[bf], bg, 2 * bf, 0, 1);
                        if ((bg[bf + be.t] += be.am(bf + 1, 2 * be[bf], bg, 2 * bf + 1, bh, be.t - bf - 1)) >= be.DV) {
                            bg[bf + be.t] -= be.DV;
                            bg[bf + be.t + 1] = 1
                        }
                    }
                    if (bg.t > 0) {
                        bg[bg.t - 1] += be.am(bf, be[bf], bg, 2 * bf, 0, 1)
                    }
                    bg.s = 0;
                    bg.clamp()
                }
                function P(bn, bk, bj) {
                    var bt = bn.abs();
                    if (bt.t <= 0) {
                        return
                    }
                    var bl = this.abs();
                    if (bl.t < bt.t) {
                        if (bk != null) {
                            bk.fromInt(0)
                        }
                        if (bj != null) {
                            this.copyTo(bj)
                        }
                        return
                    }
                    if (bj == null) {
                        bj = k()
                    }
                    var bh = k()
                      , be = this.s
                      , bm = bn.s;
                    var bs = this.DB - m(bt[bt.t - 1]);
                    if (bs > 0) {
                        bt.lShiftTo(bs, bh);
                        bl.lShiftTo(bs, bj)
                    } else {
                        bt.copyTo(bh);
                        bl.copyTo(bj)
                    }
                    var bp = bh.t;
                    var bf = bh[bp - 1];
                    if (bf == 0) {
                        return
                    }
                    var bo = bf * (1 << this.F1) + ((bp > 1) ? bh[bp - 2] >> this.F2 : 0);
                    var bw = this.FV / bo
                      , bv = (1 << this.F1) / bo
                      , bu = 1 << this.F2;
                    var br = bj.t
                      , bq = br - bp
                      , bi = (bk == null) ? k() : bk;
                    bh.dlShiftTo(bq, bi);
                    if (bj.compareTo(bi) >= 0) {
                        bj[bj.t++] = 1;
                        bj.subTo(bi, bj)
                    }
                    a1.ONE.dlShiftTo(bp, bi);
                    bi.subTo(bh, bh);
                    while (bh.t < bp) {
                        bh[bh.t++] = 0
                    }
                    while (--bq >= 0) {
                        var bg = (bj[--br] == bf) ? this.DM : Math.floor(bj[br] * bw + (bj[br - 1] + bu) * bv);
                        if ((bj[br] += bh.am(0, bg, bj, bq, 0, bp)) < bg) {
                            bh.dlShiftTo(bq, bi);
                            bj.subTo(bi, bj);
                            while (bj[br] < --bg) {
                                bj.subTo(bi, bj)
                            }
                        }
                    }
                    if (bk != null) {
                        bj.drShiftTo(bp, bk);
                        if (be != bm) {
                            a1.ZERO.subTo(bk, bk)
                        }
                    }
                    bj.t = bp;
                    bj.clamp();
                    if (bs > 0) {
                        bj.rShiftTo(bs, bj)
                    }
                    if (be < 0) {
                        a1.ZERO.subTo(bj, bj)
                    }
                }
                function Y(be) {
                    var bf = k();
                    this.abs().divRemTo(be, null, bf);
                    if (this.s < 0 && bf.compareTo(a1.ZERO) > 0) {
                        be.subTo(bf, bf)
                    }
                    return bf
                }
                function V(be) {
                    this.m = be
                }
                function am(be) {
                    if (be.s < 0 || be.compareTo(this.m) >= 0) {
                        return be.mod(this.m)
                    } else {
                        return be
                    }
                }
                function aT(be) {
                    return be
                }
                function U(be) {
                    be.divRemTo(this.m, null, be)
                }
                function S(be, bg, bf) {
                    be.multiplyTo(bg, bf);
                    this.reduce(bf)
                }
                function a5(be, bf) {
                    be.squareTo(bf);
                    this.reduce(bf)
                }
                V.prototype.convert = am;
                V.prototype.revert = aT;
                V.prototype.reduce = U;
                V.prototype.mulTo = S;
                V.prototype.sqrTo = a5;
                function K() {
                    if (this.t < 1) {
                        return 0
                    }
                    var be = this[0];
                    if ((be & 1) == 0) {
                        return 0
                    }
                    var bf = be & 3;
                    bf = (bf * (2 - (be & 15) * bf)) & 15;
                    bf = (bf * (2 - (be & 255) * bf)) & 255;
                    bf = (bf * (2 - (((be & 65535) * bf) & 65535))) & 65535;
                    bf = (bf * (2 - be * bf % this.DV)) % this.DV;
                    return (bf > 0) ? this.DV - bf : -bf
                }
                function g(be) {
                    this.m = be;
                    this.mp = be.invDigit();
                    this.mpl = this.mp & 32767;
                    this.mph = this.mp >> 15;
                    this.um = (1 << (be.DB - 15)) - 1;
                    this.mt2 = 2 * be.t
                }
                function aS(be) {
                    var bf = k();
                    be.abs().dlShiftTo(this.m.t, bf);
                    bf.divRemTo(this.m, null, bf);
                    if (be.s < 0 && bf.compareTo(a1.ZERO) > 0) {
                        this.m.subTo(bf, bf)
                    }
                    return bf
                }
                function a3(be) {
                    var bf = k();
                    be.copyTo(bf);
                    this.reduce(bf);
                    return bf
                }
                function aa(be) {
                    while (be.t <= this.mt2) {
                        be[be.t++] = 0
                    }
                    for (var bg = 0; bg < this.m.t; ++bg) {
                        var bf = be[bg] & 32767;
                        var bh = (bf * this.mpl + (((bf * this.mph + (be[bg] >> 15) * this.mpl) & this.um) << 15)) & be.DM;
                        bf = bg + this.m.t;
                        be[bf] += this.m.am(0, bh, be, bg, 0, this.m.t);
                        while (be[bf] >= be.DV) {
                            be[bf] -= be.DV;
                            be[++bf]++
                        }
                    }
                    be.clamp();
                    be.drShiftTo(this.m.t, be);
                    if (be.compareTo(this.m) >= 0) {
                        be.subTo(this.m, be)
                    }
                }
                function aV(be, bf) {
                    be.squareTo(bf);
                    this.reduce(bf)
                }
                function H(be, bg, bf) {
                    be.multiplyTo(bg, bf);
                    this.reduce(bf)
                }
                g.prototype.convert = aS;
                g.prototype.revert = a3;
                g.prototype.reduce = aa;
                g.prototype.mulTo = H;
                g.prototype.sqrTo = aV;
                function l() {
                    return ((this.t > 0) ? (this[0] & 1) : this.s) == 0
                }
                function G(bj, bk) {
                    if (bj > 4294967295 || bj < 1) {
                        return a1.ONE
                    }
                    var bi = k()
                      , be = k()
                      , bh = bk.convert(this)
                      , bg = m(bj) - 1;
                    bh.copyTo(bi);
                    while (--bg >= 0) {
                        bk.sqrTo(bi, be);
                        if ((bj & (1 << bg)) > 0) {
                            bk.mulTo(be, bh, bi)
                        } else {
                            var bf = bi;
                            bi = be;
                            be = bf
                        }
                    }
                    return bk.revert(bi)
                }
                function aW(bf, be) {
                    var bg;
                    if (bf < 256 || be.isEven()) {
                        bg = new V(be)
                    } else {
                        bg = new g(be)
                    }
                    return this.exp(bf, bg)
                }
                a1.prototype.copyTo = aw;
                a1.prototype.fromInt = r;
                a1.prototype.fromString = F;
                a1.prototype.clamp = Z;
                a1.prototype.dlShiftTo = aZ;
                a1.prototype.drShiftTo = av;
                a1.prototype.lShiftTo = B;
                a1.prototype.rShiftTo = n;
                a1.prototype.subTo = aB;
                a1.prototype.multiplyTo = O;
                a1.prototype.squareTo = ab;
                a1.prototype.divRemTo = P;
                a1.prototype.invDigit = K;
                a1.prototype.isEven = l;
                a1.prototype.exp = G;
                a1.prototype.toString = w;
                a1.prototype.negate = af;
                a1.prototype.abs = aU;
                a1.prototype.compareTo = R;
                a1.prototype.bitLength = C;
                a1.prototype.mod = Y;
                a1.prototype.modPowInt = aW;
                a1.ZERO = f(0);
                a1.ONE = f(1);
                function aF() {}
                aF.prototype.nextBytes = a9;
                var az = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
                var aq = "=";
                function aG(bg) {
                    var bf;
                    var bh;
                    var be = "";
                    for (bf = 0; bf + 3 <= bg.length; bf += 3) {
                        bh = parseInt(bg.substring(bf, bf + 3), 16);
                        be += az.charAt(bh >> 6) + az.charAt(bh & 63)
                    }
                    if (bf + 1 == bg.length) {
                        bh = parseInt(bg.substring(bf, bf + 1), 16);
                        be += az.charAt(bh << 2)
                    } else {
                        if (bf + 2 == bg.length) {
                            bh = parseInt(bg.substring(bf, bf + 2), 16);
                            be += az.charAt(bh >> 2) + az.charAt((bh & 3) << 4)
                        }
                    }
                    while ((be.length & 3) > 0) {
                        be += aq
                    }
                    return be
                }
                function d(bi) {
                    var bg = "";
                    var bh;
                    var be = 0;
                    var bf;
                    for (bh = 0; bh < bi.length; ++bh) {
                        if (bi.charAt(bh) == aq) {
                            break
                        }
                        v = az.indexOf(bi.charAt(bh));
                        if (v < 0) {
                            continue
                        }
                        if (be == 0) {
                            bg += bd(v >> 2);
                            bf = v & 3;
                            be = 1
                        } else {
                            if (be == 1) {
                                bg += bd((bf << 2) | (v >> 4));
                                bf = v & 15;
                                be = 2
                            } else {
                                if (be == 2) {
                                    bg += bd(bf);
                                    bg += bd(v >> 2);
                                    bf = v & 3;
                                    be = 3
                                } else {
                                    bg += bd((bf << 2) | (v >> 4));
                                    bg += bd(v & 15);
                                    be = 0
                                }
                            }
                        }
                    }
                    if (be == 1) {
                        bg += bd(bf << 2)
                    }
                    return bg
                }
                function al(bh) {
                    var bg = d(bh);
                    var bf;
                    var be = new Array();
                    for (bf = 0; 2 * bf < bg.length; ++bf) {
                        be[bf] = parseInt(bg.substring(2 * bf, 2 * bf + 2), 16)
                    }
                    return be
                }
                var q = null;
                var p = null;
                var j = true;
                var aR = false;
                var aY = {
                    ACCOUNT_NUM: 0,
                    APPLE_PAY: 1,
                    VISA_CHECKOUT: 2,
                    CHECKOUT_ID: 3
                };
                try {
                    var A = new Date().getTime();
                    var aN = 0;
                    var ad = 0;
                    var i = null;
                    var W = null;
                    var ao = o(ar);
                    y(aM);
                    setTimeout(T, 10);
                    t();
                    var ae = null;
                    var ay = null;
                    var aE = null;
                    var Q = null;
                    var a8;
                    var z;
                    var ag
                } catch (aD) {
                    aJ(aD)
                }
                function ac(be) {
                    var bf;
                    if (i.length > 5000) {
                        throw new Error("Request URI is too long.  Length is " + i.length + " characters")
                    }
                    if (be) {
                        bf = a.primaryUrl + "/eProtect/paypage?" + i + "&targetServer=primary&jsoncallback=?";
                        jQuery.getJSON(bf, function(bg) {
                            p = bg
                        })
                    } else {
                        bf = a.secondaryUrl + "/eProtect/paypage?" + i + "&targetServer=secondary&jsoncallback=?";
                        jQuery.getJSON(bf, function(bg) {
                            q = bg
                        })
                    }
                }
                function ai(bh, bs, br, bn, bi) {
                    try {
                        var bo = new Date();
                        var bp = bo.getTime();
                        var bm = bp - bh;
                        var bu = 0;
                        if (br) {
                            bu = br
                        }
                        var bl = encodeURIComponent(bn.paypageId);
                        var bq = encodeURIComponent(bn.reportGroup);
                        var bg = encodeURIComponent(bn.orderId);
                        var bj = encodeURIComponent(bn.id);
                        var bf;
                        var bt = "secondary";
                        if (bi) {
                            bf = a.primaryUrl;
                            bt = "primary"
                        } else {
                            bf = a.secondaryUrl
                        }
                        var be = "paypageId=" + bl + "&responseTime=" + bm + "&responseCode=" + bs + "&tzOffset=" + bo.getTimezoneOffset() + "&timeout=" + bu;
                        be += "&reportGroup=" + bq + "&txnId=" + bj + "&orderId=" + bg + "&startTime=" + A + "&targetServer=" + bt;
                        setTimeout(function() {
                            try {
                                jQuery.getJSON(bf + "/eProtect/ppstats?" + be + "&jsoncallback=?", function(bw) {})
                            } catch (bv) {}
                        }, 0)
                    } catch (bk) {}
                }
                function y(be) {
                    if (be !== undefined && be.url != undefined && be.url != null && be.url.length > 0) {
                        a.primaryUrl = be.url;
                        if (be.secondaryUrl != undefined && be.secondaryUrl != null && be.secondaryUrl.length > 0) {
                            a.secondaryUrl = be.secondaryUrl
                        }
                    }
                    j = true;
                    if ((ao > 0 && ao <= a.primaryTimeout) || (a.secondaryUrl == undefined || a.secondaryUrl == null || a.secondaryUrl.length == 0)) {
                        j = false
                    }
                }
                function T() {
                    try {
                        aN = new Date().getTime() - A;
                        if (W != null) {
                            aH(W);
                            return
                        }
                        if (ao > 0 && aN > ao) {
                            aR = true
                        }
                        if (j) {
                            switch (ad) {
                            case 0:
                                break;
                            case 1:
                                if (p != null) {
                                    ai(A, p.response, ao, aM, true);
                                    if (p.response == "889") {
                                        ad = 3;
                                        ac(false)
                                    } else {
                                        a0(p);
                                        return
                                    }
                                } else {
                                    if (aN > a.primaryTimeout) {
                                        ad = 2;
                                        ac(false)
                                    }
                                }
                                break;
                            case 2:
                                if (p != null) {
                                    ai(A, p.response, ao, aM, true);
                                    if (p.response == "889") {
                                        ad = 3
                                    } else {
                                        a0(p);
                                        return
                                    }
                                } else {
                                    if (q != null) {
                                        ai(A, q.response, ao, aM, false);
                                        if (q.response != "887" && q.response != "889") {
                                            a0(q);
                                            return
                                        } else {
                                            ad = 4
                                        }
                                    }
                                }
                                break;
                            case 3:
                                if (q != null) {
                                    ai(A, q.response, ao, aM, false);
                                    if (q.response == "887") {
                                        a0(p)
                                    } else {
                                        a0(q)
                                    }
                                    return
                                } else {
                                    if (aR) {
                                        a0(p);
                                        return
                                    }
                                }
                                break;
                            case 4:
                                if (p != null) {
                                    ai(A, p.response, ao, aM, true);
                                    a0(p);
                                    return
                                } else {
                                    if (aR) {
                                        if (q.response == "887") {
                                            aC()
                                        } else {
                                            a0(q)
                                        }
                                        return
                                    }
                                }
                                break;
                            default:
                                break
                            }
                        } else {
                            if (p != null) {
                                ai(A, p.response, ar, aM, true);
                                a0(p);
                                return
                            }
                        }
                        if (aR) {
                            if (ad == 0) {
                                ai(A, "900", ao, aM, true)
                            } else {
                                ai(A, "901", ao, aM, true)
                            }
                            aC()
                        } else {
                            setTimeout(T, 10)
                        }
                    } catch (be) {
                        aJ(be);
                        if (W != null) {
                            a0(W);
                            return
                        }
                    }
                }
                function aJ(bj) {
                    try {
                        var bl = encodeURIComponent("GLOBAL_TRY_CATCH");
                        var bi = encodeURIComponent(0);
                        var bf = encodeURIComponent(0);
                        var bm = encodeURIComponent("A");
                        var br = encodeURIComponent("NOT_A_STRING");
                        if (typeof bj === "object") {
                            try {
                                if (typeof bj.message === "undefined") {
                                    bm = "undefined"
                                } else {
                                    if (typeof bj.message === "string") {
                                        bm = bj.message;
                                        if (bm.length > 1024) {
                                            bm = bm.substr(0, 1024)
                                        }
                                    } else {
                                        bm = "NOT_A_STRING"
                                    }
                                }
                            } catch (bn) {
                                bm = "UNABLE_TO_GET_ERROR_FROM_OBJECT"
                            } finally {
                                bm = encodeURIComponent(bm)
                            }
                            try {
                                if (typeof bj.stack === "undefined") {
                                    bi = encodeURIComponent("UNDEFINED");
                                    bf = encodeURIComponent("UNDEFINED");
                                    br = encodeURIComponent("UNDEFINED")
                                } else {
                                    if (typeof bj.stack === "string") {
                                        br = bj.stack;
                                        if (br.length > 3072) {
                                            br = br.substr(0, 3072)
                                        }
                                        var bk = /.*?eProtect-api.*?\.js:(\d+):(\d+)/.exec(br);
                                        br = encodeURIComponent(br);
                                        if (!/^\d+$/.test(bk[1])) {
                                            bi = "NaN"
                                        } else {
                                            if (!/^\d{0,6}$/.test(bk[1])) {
                                                bi = "TOO_BIG"
                                            } else {
                                                bi = bk[1]
                                            }
                                        }
                                        bi = encodeURIComponent(bi);
                                        if (!/^\d+$/.test(bk[2])) {
                                            bf = "NaN"
                                        } else {
                                            if (!/^\d{0,6}$/.test(bk[2])) {
                                                bf = "TOO_BIG"
                                            } else {
                                                bf = bk[2]
                                            }
                                        }
                                        bf = encodeURIComponent(bf)
                                    }
                                }
                            } catch (bn) {
                                bi = encodeURIComponent("EXCEPTION");
                                bf = encodeURIComponent("EXCEPTION");
                                if (br.length > 2000) {
                                    br = br.substr(0, 2000)
                                }
                                br = encodeURIComponent(br)
                            }
                        } else {
                            if (typeof bj === "string") {
                                if (bj.length > 1024) {
                                    bj = bj.substr(0, 1024)
                                }
                                bm = encodeURIComponent(bj)
                            }
                        }
                        if (typeof aM === "object") {
                            try {
                                var bp = "undefined";
                                if (typeof aM.paypageId === "undefined") {
                                    bp = "undefined"
                                } else {
                                    if (typeof aM.paypageId === "string") {
                                        bp = aM.paypageId;
                                        if (bp.length > 50) {
                                            bp = bp.substr(0, 50)
                                        }
                                    } else {
                                        bp = "NOT_A_STRING"
                                    }
                                }
                            } catch (bn) {
                                bp = "UNABLE_TO_GET_PAYPAGE_ID"
                            } finally {
                                bp = encodeURIComponent(bp)
                            }
                            var bh = "undefined";
                            try {
                                if (typeof aM.orderId === "undefined") {
                                    bh = "undefined"
                                } else {
                                    if (typeof aM.orderId === "string") {
                                        bh = aM.orderId;
                                        if (bh.length > 32) {
                                            bh = bh.substr(0, 32)
                                        }
                                    } else {
                                        bh = "NOT_A_STRING"
                                    }
                                }
                            } catch (bn) {
                                bh = "UNABLE_TO_GET_ORDER_ID"
                            } finally {
                                bh = encodeURIComponent(bh)
                            }
                            var bg = "undefined";
                            try {
                                if (typeof aM.reportGroup === "undefined") {
                                    bg = "undefined"
                                } else {
                                    if (typeof aM.reportGroup === "string") {
                                        bg = aM.reportGroup;
                                        if (bg.length > 32) {
                                            bg = bg.substr(0, 32)
                                        }
                                    } else {
                                        bg = "NOT_A_STRING"
                                    }
                                }
                            } catch (bn) {
                                bg = "UNABLE_TO_GET_REPORT_GROUP"
                            } finally {
                                bg = encodeURIComponent(bg)
                            }
                        }
                        bm = myVantivEProtectReporterForPpStats.removeNonStandardCharacters(bm);
                        br = myVantivEProtectReporterForPpStats.removeNonStandardCharacters(br);
                        var be = "errorHandler=" + bl + "&columnNumber=" + bf + "&errorMessage=" + bm + "&lineNumber=" + bi + "&paypageId=" + bp + "&orderId=" + bh + "&reportGroup=" + bg + "&errorStack=" + br;
                        var bq = "https://request.eprotect.vantivcnp.com/eProtect/ppstats?" + be + "&jsoncallback=?";
                        setTimeout(function() {
                            try {
                                jQuery.getJSON(bq, function(bt) {})
                            } catch (bs) {}
                        }, 0)
                    } catch (bo) {} finally {
                        e("889", bj)
                    }
                }
                function t() {
                    if (ap === undefined) {
                        return e("889", "Missing litleFormFields")
                    }
                    if (aM === undefined) {
                        return e("889", "Missing litleRequest")
                    }
                    if (ap.paypageRegistrationId) {
                        ap.paypageRegistrationId.value = ""
                    }
                    if (ap.bin) {
                        ap.bin.value = ""
                    }
                    z = aY.ACCOUNT_NUM;
                    if (aM.applepay !== undefined) {
                        if (aM.applepay.data !== undefined && aM.applepay.signature !== undefined && aM.applepay.version !== undefined && aM.applepay.header.ephemeralPublicKey !== undefined && aM.applepay.header.publicKeyHash !== undefined && aM.applepay.header.transactionId !== undefined) {
                            z = aY.APPLE_PAY
                        } else {
                            return e("889", "Missing ApplePay elements")
                        }
                    } else {
                        if (aM.checkoutIdMode !== undefined && aM.checkoutIdMode) {
                            z = aY.CHECKOUT_ID
                        } else {
                            if (aM.visaCheckout !== undefined) {
                                if (aM.visaCheckout.vInitRequest.apikey !== undefined && aM.visaCheckout.encPaymentData !== undefined && aM.visaCheckout.encKey !== undefined && aM.visaCheckout.callid !== undefined) {
                                    z = aY.VISA_CHECKOUT
                                } else {
                                    return e("889", "Missing VisaCheckout elements")
                                }
                            }
                        }
                    }
                    if (z === aY.CHECKOUT_ID) {
                        if (ap.checkoutId) {
                            ap.checkoutId.value = ""
                        }
                        ag = (jQuery(ap.cvv2).length > 0);
                        if (ag) {
                            if (ap.cvv2.value === undefined) {
                                throw "Parameter cvv2.value is undefined"
                            }
                            var bi = ap.cvv2.value;
                            bi = an(bi)
                        } else {
                            return e("889", "Missing cvv2")
                        }
                        br = au(bi);
                        if (br != "870") {
                            return e(br)
                        }
                        try {
                            var bj = new X();
                            var bs = bj.setPublic(b.modulus, b.exponent);
                            var bm = bj.encrypt(bi)
                        } catch (bv) {
                            myVantivEProtectReporterForPpStats.report3rdPartyError(bv);
                            return e("875")
                        }
                        if (bm) {
                            var bn = aG(bm);
                            var bl = encodeURIComponent(bn)
                        } else {
                            return e("875")
                        }
                    } else {
                        if (z === aY.ACCOUNT_NUM) {
                            try {
                                a4("accountNum", ap.accountNum, aK, N)
                            } catch (bv) {
                                return e("889", bv)
                            }
                            if (ap.accountNum.value === undefined) {
                                throw "Parameter accountNum.value is undefined"
                            }
                            a8 = ap.accountNum.value;
                            a8 = an(a8);
                            ag = (ap.cvv2.value.length > 0);
                            if (ag) {
                                if (ap.cvv2.value === undefined) {
                                    throw "Parameter cvv2.value is undefined"
                                }
                                var bi = ap.cvv2.value;
                                bi = an(bi)
                            }
                            if (aM.pciNonSensitive === undefined) {
                                aM.pciNonSensitive = false
                            }
                            var br = ak(a8, aM.pciNonSensitive);
                            if (br != "870") {
                                return e(br)
                            }
                            if (ag) {
                                br = au(bi);
                                if (br != "870") {
                                    return e(br)
                                }
                            }
                            var bk = (ap.expYear && ap.expYear.value.length > 0);
                            var be = "";
                            if (bk) {
                                var bu = ap.expYear;
                                be = bu.options[bu.selectedIndex].text;
                                br = a6(be);
                                if (br != "870") {
                                    return e(br)
                                }
                            }
                            var bh = (ap.expMonth && ap.expMonth.value.length > 0);
                            if (bh) {
                                var bq = ap.expMonth;
                                var bf = bq.options[bq.selectedIndex].value;
                                br = a2(bf, be);
                                if (br != "870") {
                                    return e(br)
                                }
                            }
                            try {
                                var bj = new X();
                                var bs = bj.setPublic(b.modulus, b.exponent);
                                var bg = bj.encrypt(a8);
                                if (ag) {
                                    var bm = bj.encrypt(bi)
                                }
                            } catch (bv) {
                                myVantivEProtectReporterForPpStats.report3rdPartyError(bv);
                                return e("875")
                            }
                            if (bg) {
                                var bp = aG(bg);
                                var bt = encodeURIComponent(bp);
                                if (ag) {
                                    var bn = aG(bm);
                                    var bl = encodeURIComponent(bn)
                                }
                            } else {
                                return e("875")
                            }
                        }
                    }
                    try {
                        a4("paypageId", aM.paypageId, aK, N, a7);
                        a4("reportGroup", aM.reportGroup, aK, N);
                        a4("id", aM.id, aK, N)
                    } catch (bv) {
                        return e("889", bv)
                    }
                    ae = encodeURIComponent(aM.paypageId);
                    ay = encodeURIComponent(aM.reportGroup);
                    aE = encodeURIComponent(aM.orderId);
                    Q = encodeURIComponent(aM.id);
                    i = "paypageId=" + ae + "&reportGroup=" + ay + "&id=" + Q + "&orderId=" + aE;
                    if (z === aY.ACCOUNT_NUM) {
                        var bo = encodeURIComponent(aM.pciNonSensitive);
                        i += "&encryptedAccount=" + bt + "&publicKeyId=" + b.keyId + "&pciNonSensitive=" + bo
                    }
                    if (z === aY.APPLE_PAY) {
                        i += "&applepay.data=" + encodeURIComponent(aM.applepay.data);
                        i += "&applepay.signature=" + encodeURIComponent(aM.applepay.signature);
                        i += "&applepay.version=" + encodeURIComponent(aM.applepay.version);
                        i += "&applepay.header.ephemeralPublicKey=" + encodeURIComponent(aM.applepay.header.ephemeralPublicKey);
                        i += "&applepay.header.publicKeyHash=" + encodeURIComponent(aM.applepay.header.publicKeyHash);
                        i += "&applepay.header.transactionId=" + encodeURIComponent(aM.applepay.header.transactionId);
                        if (aM.applepay.header.applicationData !== undefined) {
                            i += "&applepay.header.applicationData=" + encodeURIComponent(aM.applepay.header.applicationData)
                        }
                    }
                    if (z === aY.VISA_CHECKOUT) {
                        i += "&visaCheckout.encPaymentData=" + encodeURIComponent(aM.visaCheckout.encPaymentData);
                        i += "&visaCheckout.encKey=" + encodeURIComponent(aM.visaCheckout.encKey);
                        i += "&visaCheckout.apiKey=" + encodeURIComponent(aM.visaCheckout.vInitRequest.apikey);
                        i += "&visaCheckout.callid=" + encodeURIComponent(aM.visaCheckout.callid)
                    }
                    if (ag) {
                        i += "&encryptedCvv=" + bl
                    }
                    if (z === aY.CHECKOUT_ID) {
                        i += "&checkoutIdMode=true&publicKeyId=" + b.keyId
                    }
                    i += "&requestType=eProtect";
                    ad = 1;
                    ac(true)
                }
                function o(bf) {
                    if (bf != undefined) {
                        if (typeof bf == "number") {
                            return bf
                        } else {
                            if (typeof bf == "string") {
                                var be = /^[0-9]+$/.test(bf);
                                if (be) {
                                    return parseInt(bf)
                                }
                                return 15000
                            }
                        }
                    }
                    return 0
                }
                function L(bg) {
                    if (z === aY.CHECKOUT_ID) {
                        if (ap.checkoutId) {
                            ap.checkoutId.value = bg.checkoutId
                        }
                    } else {
                        if (z === aY.ACCOUNT_NUM) {
                            var bh = an(a8);
                            ap.accountNum.value = D(a8);
                            bg.firstSix = bh.substring(0, 6);
                            bg.lastFour = bh.substring(bh.length - 4, bh.length);
                            if (ap.extraAccountNums) {
                                for (var bf in ap.extraAccountNums) {
                                    var be = ap.extraAccountNums[bf];
                                    if (be.value === undefined) {
                                        throw "Parameter extraAccountNums[" + bf + "].value is undefined"
                                    }
                                    be.value = D(an(be.value))
                                }
                            }
                        }
                        if (ap.bin) {
                            ap.bin.value = bg.bin
                        }
                        if (ap.paypageRegistrationId) {
                            ap.paypageRegistrationId.value = bg.paypageRegistrationId
                        }
                    }
                    if (z === aY.ACCOUNT_NUM || z === aY.CHECKOUT_ID) {
                        if (ag) {
                            ap.cvv2.value = "000"
                        }
                    }
                    if (aj === undefined) {
                        throw "successCallback undefined"
                    }
                    if (typeof aj !== "function") {
                        throw "successCallback not a function"
                    }
                    aj(bg)
                }
                function aH(be) {
                    if (ah === undefined) {
                        throw "errorCallback undefined"
                    }
                    if (typeof ah !== "function") {
                        throw "errorCallback not a function"
                    }
                    ah(be)
                }
                function aC() {
                    M()
                }
                function a0(be) {
                    if (be.response == "870") {
                        L(be)
                    } else {
                        aH(be)
                    }
                    return
                }
                function D(be) {
                    if (!be) {
                        return be
                    }
                    be = be.substring(0, be.length - 4).replace(/./g, "X").concat(be.substring(be.length - 4));
                    return be
                }
                function J(be) {
                    be = (be + "").split("").reverse();
                    if (!be.length) {
                        return false
                    }
                    var bg = 0, bf;
                    for (bf = 0; bf < be.length; bf++) {
                        be[bf] = parseInt(be[bf]);
                        bg += bf % 2 ? 2 * be[bf] - (be[bf] > 4 ? 9 : 0) : be[bf]
                    }
                    return (bg % 10) == 0
                }
                function ak(bf, be) {
                    if (bf.length < 13) {
                        return "872"
                    } else {
                        if (bf.length > 19) {
                            return "873"
                        } else {
                            if (!bf.match(/^[0-9]{13,19}$/)) {
                                return "874"
                            } else {
                                if (!be && !J(bf)) {
                                    return "871"
                                } else {
                                    return "870"
                                }
                            }
                        }
                    }
                }
                function au(be) {
                    if (be.length < 3) {
                        return "882"
                    } else {
                        if (be.length > 4) {
                            return "883"
                        } else {
                            if (!be.match(/^\d\d\d\d?$/)) {
                                return "881"
                            } else {
                                return "870"
                            }
                        }
                    }
                }
                function a6(bg) {
                    if (bg == "") {
                        return "886"
                    }
                    var be = new Date();
                    var bf = be.getFullYear();
                    if (parseInt(bg) < bf) {
                        return "886"
                    } else {
                        return "870"
                    }
                }
                function a2(bi, bh) {
                    if (bi == "" || bh == "") {
                        return "886"
                    }
                    var be = new Date();
                    var bf = be.getFullYear();
                    var bg = be.getMonth() + 1;
                    if (bf == parseInt(bh) && bg > parseInt(bi)) {
                        return "886"
                    } else {
                        return "870"
                    }
                }
                function a4() {
                    var bf = arguments[0];
                    var bg = arguments[1];
                    if (bg === undefined) {
                        throw "Parameter " + bf + " is undefined"
                    }
                    for (var be = 2; be < arguments.length; be++) {
                        arguments[be](bf, bg)
                    }
                }
                function aK(be, bf) {
                    if (bf.length == 0) {
                        throw "Parameter " + be + " is required"
                    }
                }
                function N(be, bf) {
                    if (bf.length > 1024) {
                        throw "Parameter " + be + " is too long.  Length is " + bf.length
                    }
                }
                function a7(be, bf) {
                    if (!bf.match(/^[0-9a-zA-Z]+$/)) {
                        throw "Parameter " + be + " with value " + bf + " is not alphanumeric"
                    }
                }
                function e(bi, bh) {
                    var bj = {
                        response: null,
                        message: null
                    };
                    var bf = {
                        "870": "Success",
                        "871": "Account number not mod10",
                        "872": "Account number too short",
                        "873": "Account number too long",
                        "874": "Account number not numeric",
                        "875": "Unable to encrypt field",
                        "876": "Account number invalid",
                        "877": "Invalid paypage registration id",
                        "878": "Expired paypage registration id",
                        "879": "Merchant is not authorized for Paypage",
                        "880": "Report Group Invalid",
                        "881": "Card validation num not numeric",
                        "882": "Card validation num too short",
                        "883": "Card validation num too long",
                        "884": "PayFrame HTML failed to load",
                        "885": "PayFrame CSS failed to load",
                        "886": "Expiration date invalid",
                        "887": "Secondary PayPage request for non-Vantiv token merchant",
                        "888": "Paypage Signature Verification Failed",
                        "889": "Failure"
                    };
                    function be(bk) {
                        return bk < 10 ? "0" + bk : bk
                    }
                    function bg(bk) {
                        return bk.getUTCFullYear() + "-" + be(bk.getUTCMonth() + 1) + "-" + be(bk.getUTCDate()) + "T" + be(bk.getUTCHours()) + ":" + be(bk.getUTCMinutes()) + ":" + be(bk.getUTCSeconds())
                    }
                    bj.response = bi;
                    if (bh == undefined) {
                        bj.message = bf[bi]
                    } else {
                        bj.message = bh
                    }
                    bj.responseTime = bg(new Date());
                    if (aM !== undefined) {
                        bj.reportGroup = aM.reportGroup;
                        bj.id = aM.id;
                        bj.orderId = aM.orderId
                    }
                    W = bj
                }
                function an(be) {
                    return be.replace(/[ -]/gi, "")
                }
            },
            getVisaCheckoutApiKey: function() {
                return b.visaCheckoutApiKey
            }
        }
    };
}
