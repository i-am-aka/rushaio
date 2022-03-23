import { logDebug } from './log.js';


export function interceptXhr(dom, url) {
  var origin = url.split("/", 4).slice(0, 3).join("/"); // https://www.walmart.com/foo/bar
  logDebug(JSON.stringify({origin}));
  var { XMLHttpRequest} = dom.window;
  dom.window.XMLHttpRequest = XMLHttpRequest;

  XMLHttpRequest.prototype.open = function() {
    this.requestMethod = arguments[0].toUpperCase();
    this.requestUrl = arguments[1];
    if (this.requestUrl.indexOf('//') == 0) {
      this.requestUrl = 'https:' + this.requestUrl;
    }
    this.readyState = 1;
    logDebug(`XMLHttpRequest.open(${JSON.stringify(arguments)})`)
  }
  XMLHttpRequest.prototype.setRequestHeader = function() {
    logDebug(`XMLHttpRequest.setRequestHeader(${JSON.stringify(arguments)})`)
  }
  XMLHttpRequest.prototype.send = function(formData) {
    logDebug(`XMLHttpRequest.send(${JSON.stringify(arguments)})`)
    // debugger;
    setTimeout(function() {
      var headers = {
        "accept": "*/*",
        "accept-language": "en-us",
        "origin": origin,
        "referer": origin + "/",
        "user-agent": dom.window.navigator.userAgent,
        "accept-encoding": "gzip, deflate, br",
      }
      if (formData) {
        headers['content-type'] = "application/x-www-form-urlencoded";
      }
      if (this.requestUrl[0] == "/") {
        if (this.requestUrl[1] == "/") {
          this.requestUrl = "https:" + this.requestUrl;
        } else {
          this.requestUrl = origin + this.requestUrl;
        }
      }
      fetch(this.requestUrl, {
        "headers": headers,
        // "referrer": "https://secure.nordstrom.com/checkout",
        "referrerPolicy": "no-referrer-when-downgrade",
        "body": formData ? new dom.window.URLSearchParams(formData).toString() : null,
        "method": this.requestMethod, // todo real
        "mode": "cors"
      }).then(function(resp) {
        return resp.text()
      }).then(function(text) {
        this.responseText = text;
        this.response = text;
        this.responseType = "";
        this.responseURL = this.requestUrl;
        this.status = 200;
        this.statusText = "";
        this.readyState = 4;

        this.onload && this.onload.call(this, {
          type: "load",
          total: 0,
          timeStamp: Math.random() * 5000,
          target: this,
          srcElement: this,
          returnValue: true,
          path: [],
          loaded: text.length,
          lengthComputable: false,
          isTrusted: true,
          eventPhase: 2,
          defaultPrevented: false,
          currentTarget: this,
        });
      }.bind(this));
    }.bind(this), 25);
  }
}

const decoder = new TextDecoder('utf-8');
const encoder = new TextEncoder('utf-8');

export async function interceptXhrProxy(url, dom, sock, cb) {
  var origin = url.split("/", 4).slice(0, 3).join("/"); // https://www.walmart.com/foo/bar
  logDebug(JSON.stringify({origin}));

  var proxyReqsInFlight = {};

  setTimeout(async function() {
    try {
      while(1) {
        await delay(25);
        const respbuf = new Uint8Array(32768); // todo more?
        logDebug('xhrread');
        while(respbuf[0] === 0) { // todo less jank
          if (sock.closed) {
            throw new Error('socket closed');
          }
          await sock.read(respbuf, 32768);
        }
        var mlen = 0;
        for(let i = respbuf.length - 1; i > 0; i--) {
          if (respbuf[i] !== 0) {
            mlen = i + 1;
            break
          }
        }
        var respmsg = respbuf.slice(0, mlen);
        const text = decoder.decode(respmsg);
        var resp = JSON.parse(text);
        logDebug('recvmsg', resp);

        var resolve = proxyReqsInFlight[resp.url];
        if (resolve) {
          resolve(resp.text);
        } else {
          logDebug("no inflight resolve for " + resp.url);
          logDebug("available: " + Object.keys(proxyReqsInFlight).join("\n"));
        }
      }
    } catch(e) {
      logDebug(e);
    }
  }, 0);

  var { XMLHttpRequest} = dom.window;
  // globalThis.XMLHttpRequest = XMLHttpRequest;

  XMLHttpRequest.prototype.open = function() {
    this._headers = {};
    this.requestMethod = arguments[0].toUpperCase();
    this.requestUrl = arguments[1];
    if (this.requestUrl.indexOf('//') == 0) {
      this.requestUrl = 'https:' + this.requestUrl;
    }
    this.readyState = 1;
    logDebug(`XMLHttpRequest.open(${JSON.stringify(arguments)})`)
  }
  XMLHttpRequest.prototype.setRequestHeader = function() {
    logDebug(`XMLHttpRequest.setRequestHeader(${JSON.stringify(arguments)})`)
    this._headers[arguments[0].toLowerCase()] = arguments[1];
  }
  XMLHttpRequest.prototype.send = function(formData) {
    logDebug(`XMLHttpRequest.send(${JSON.stringify(arguments)})`)
    // debugger;


    setTimeout(async function() {
      try {
        var headers = {
          "accept": "*/*",
          "accept-language": "en-us",
          "origin": origin,
          "referer": origin + "/",
          "user-agent": dom.window.navigator.userAgent,
          "accept-encoding": "gzip, deflate, br",
        }
        if (formData && this._headers['content-type'] === undefined) {
          headers['content-type'] = "application/x-www-form-urlencoded";
        }
        headers = {...headers, ...this._headers};
        if (this.requestUrl[0] == "/") {
          if (this.requestUrl.indexOf('ocaptcha') !== -1 ){
            // hacky af wtf
            this.requestUrl = 'https://collector-PXu6b0qd2S.perimeterx.net/api/v2/collector/ocaptcha';
          } else {
            if (this.requestUrl[1] == "/") {
              this.requestUrl = "https:" + this.requestUrl;
            } else {
              this.requestUrl = origin + this.requestUrl;
            }
          }
        }
        var request = {
          "headers": headers,
          // "referrer": "https://secure.nordstrom.com/checkout",
          "referrerPolicy": "no-referrer-when-downgrade",
          "body": formData && headers['content-type'] === 'application/x-www-form-urlencoded' ? new dom.window.URLSearchParams(formData).toString() : (formData ? formData : null),
          "method": this.requestMethod, // todo real
          "mode": "cors",
          "url": this.requestUrl,
        };

        const rj = JSON.stringify(request);
        const reqenc = encoder.encode(rj);
        var presolve;
        var p = new Promise((resolve, reject) => {
          presolve = resolve;
        });
        proxyReqsInFlight[this.requestUrl] = presolve;

        logDebug('xhrwrite');
        for(var i = 0; i < reqenc.length; i+=32768) {
          if (sock.closed) {
            throw new Error('socket closed');
          }
          await sock.write(reqenc.slice(i, i+32768));
        }

        var text = await Promise.race([p, delay(10000)]);
        if (!text) {
          logDebug(`TIMED OUT (${request})`);
          this.onerror(new Error());
          return;
        }
        text = text.trim();
        logDebug('got text', text, text.length);

        if (this.requestUrl === 'https://collector-PXu6b0qd2S.perimeterx.net/api/v2/collector/ocaptcha') {
          var ckies = JSON.parse(text).cookies;
          if (ckies.length) {
            logDebug('set cookie', `${ckies[0].split('|')[1]}=${ckies[0].split('|')[3]}`);
            dom.window.document.cookie = `${ckies[0].split('|')[1]}=${ckies[0].split('|')[3]}`;
          }
        }

        cb && cb(this.requestUrl);
        this.responseText = text;
        this.response = text;
        this.responseType = "";
        this.responseURL = this.requestUrl;
        this.status = 200;
        this.statusText = "";
        this.readyState = 4;

        this.onload && this.onload.call(this, {
          type: "load",
          total: 0,
          timeStamp: Math.random() * 5000,
          target: this,
          srcElement: this,
          returnValue: true,
          path: [],
          loaded: text.length,
          lengthComputable: false,
          isTrusted: true,
          eventPhase: 2,
          defaultPrevented: false,
          currentTarget: this,
        });
      } catch(e) {
        logDebug(e);
      }
    }.bind(this), 5);
  }
}


function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}