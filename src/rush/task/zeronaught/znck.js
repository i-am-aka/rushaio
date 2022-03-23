import jsdom from 'https://dev.jspm.io/jsdom';
const dom = new jsdom.JSDOM(``, {
  url: "https://secure.nordstrom.com/checkout",
  referrer: "https://secure.nordstrom.com/checkout",
  contentType: "text/html",
  includeNodeLocations: false,
  pretendToBeVisual: true,
});
// window.addEventListener = dom.window.document.addEventListener;
// globalThis.window = dom.window;
window.document = dom.window.document;
// globalThis.window = dom.window;
window.domWindow = dom.window;
globalThis.document = dom.window.document;
globalThis.navigator = dom.window.navigator;
globalThis.global = dom.window.document;
import * as Storage from './storage.ts';
globalThis.sessionStorage = new Storage.SessionStorage();
window.sessionStorage = globalThis.sessionStorage;
globalThis.navigator = dom.window.navigator;
import Trace from './trace.js';

var { XMLHttpRequest} = dom.window;
globalThis.XMLHttpRequest = XMLHttpRequest;

var headerPromiseResolve, cookiePromiseResolve;
var znDataPromises = [new Promise(((resolve) => {
  headerPromiseResolve = resolve;
})), new Promise(((resolve) => {
  cookiePromiseResolve = resolve;
}))];


function r(s) {
      var q = String.fromCharCode.bind(String);
      return q(8238) + s + q(8237)
  }
var t = r("fZUcbnkvZ");
var u = r("sOtcztEKa");
Object.defineProperty(XMLHttpRequest.prototype, u, {
     writable: true,
     configurable: true,
     enumerable: false,
     value: true
 });
XMLHttpRequest.prototype.open = function() {
  this.requestUrl = arguments[1];
  this[t] = {method: arguments[0].toLowerCase(), url: arguments[1]};
  // console.log(`open(${JSON.stringify(arguments)})`)
}
XMLHttpRequest.prototype.send = function() {
  // console.log(`send(${JSON.stringify(arguments)})`)
  setTimeout(function() {
    fetch(this.requestUrl, {
      "headers": {
        "accept": "*/*",
        "accept-language": "en-US,en;q=0.9",
        "cache-control": "no-cache",
        "pragma": "no-cache",
        "sec-fetch-dest": "empty",
        "sec-fetch-mode": "cors",
        "sec-fetch-site": "cross-site",
        "user-agent": "user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36",
      },
      "referrer": "https://secure.nordstrom.com/checkout",
      "referrerPolicy": "no-referrer-when-downgrade",
      "body": null,
      "method": "GET",
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

      this.onload.call(this, {
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
      cookiePromiseResolve({[document.cookie.split('=')[0]]: document.cookie.split('=')[1]});
    }.bind(this));
  }.bind(this), 25);
}

// for (var x in XMLHttpRequest.prototype) {
//   XMLHttpRequest.prototype[x] = new Proxy()
// }

async function afterReadyCb() {
    // console.log('afterReady')
    // Trace.addLoggingToNamespace(window.document);
    // Trace.addLoggingToNamespace(window.domWindow);
    var xhr = new XMLHttpRequest();
    xhr.open("get", window.domWindow.location.href, true);
    xhr.send("");
}
window.domWindow.addEventListener('afterReady', afterReadyCb, false)

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}

const o = await import('./fp.js');
o.TEXT_MEASUREMENTS.__default__ = o.TEXT_MEASUREMENTS["0-_{w."]["50px 'sans-serif', sans-serif"];
var navFp = {
  "vendorSub": "",
  "productSub": "20030107",
  "vendor": "Google Inc.",
  "maxTouchPoints": 0,
  "hardwareConcurrency": 12,
  "cookieEnabled": true,
  "appCodeName": "Mozilla",
  "appName": "Netscape",
  "appVersion": "5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36",
  "platform": "MacIntel",
  "product": "Gecko",
  "userAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36",
  "language": "en-US",
  "languages": {
    "0": "en-US",
    "1": "en"
  },
  "onLine": true,
  "connection": {
    "effectiveType": "4g",
    "rtt": 50,
    "downlink": 10,
    "saveData": false
  },
  "plugins": {
    "0": {
      "0": {
        "type": "application/x-google-chrome-pdf",
        "suffixes": "pdf",
        "description": "Portable Document Format"
      },
      "name": "Chrome PDF Plugin",
      "filename": "internal-pdf-viewer",
      "description": "Portable Document Format",
      "length": 1
    },
    "1": {
      "0": {
        "type": "application/pdf",
        "suffixes": "pdf",
        "description": ""
      },
      "name": "Chrome PDF Viewer",
      "filename": "mhjfbmdgcfjbbpaeojofohoefgiehjai",
      "description": "",
      "length": 1
    },
    "2": {
      "0": {
        "type": "application/x-nacl",
        "suffixes": "",
        "description": "Native Client Executable"
      },
      "1": {
        "type": "application/x-pnacl",
        "suffixes": "",
        "description": "Portable Native Client Executable"
      },
      "name": "Native Client",
      "filename": "internal-nacl-plugin",
      "description": "",
      "length": 2
    },
    "length": 3
  },
  "mimeTypes": {
    "0": {
      "type": "application/pdf",
      "suffixes": "pdf",
      "description": ""
    },
    "1": {
      "type": "application/x-google-chrome-pdf",
      "suffixes": "pdf",
      "description": "Portable Document Format"
    },
    "2": {
      "type": "application/x-nacl",
      "suffixes": "",
      "description": "Native Client Executable"
    },
    "3": {
      "type": "application/x-pnacl",
      "suffixes": "",
      "description": "Portable Native Client Executable"
    },
    "length": 4
  },
  "userActivation": {
    "hasBeenActive": false,
    "isActive": false
  },
  "mediaSession": {
    "playbackState": "none"
  },
  "deviceMemory": 8
};
window.domWindow.navigator = navFp;


var origCr = window.document.createElement;
var noop = function(){}.bind(null);
Object.defineProperty(window.document, 'createElement', {
     writable: true,
     value: new Proxy(origCr, {
      apply: function(target, thisArg, argumentsArr) {
        if (argumentsArr[0] === "canvas") {
          debugger;
          return Canvas();
        } else {
          return Reflect.apply(target, thisArg, argumentsArr);
        }
      }
     })
});

var Canvas = function() {
  var cid = (parseInt(Math.random()*1e9).toString()).substring(0,6);
  var x = {
    getContext: function() {
      console.trace();
      return new Proxy({}, { get: function(target, prop, receiver) {
        console.log('get Canvas-' + cid + '.' + JSON.stringify(prop));
        // if (prop === "measureText") {
        //   return function(args) {
        //     return {
        //       "width":30.576171875,
        //       "actualBoundingBoxLeft":-0.3125,
        //       "actualBoundingBoxRight":29.8583984375,
        //       "actualBoundingBoxAscent":7.2998046875,
        //       "actualBoundingBoxDescent":2.041015625
        //     };
        //   }
        if (typeof target[prop] !== 'undefined') {
          return target[prop];
        } else {
          var fn = noop;
          if (prop === 'measureText') {
            fn = function(text) {
              var measurement = o.TEXT_MEASUREMENTS[text][target.font.replace(/'/g, '"')] || o.TEXT_MEASUREMENTS[text][target.font.replace(/'/g, '')];
              return measurement || o.TEXT_MEASUREMENTS.__default__;
            }
          }
          if (prop[0].toUpperCase() === prop[0]) {
            // const
            return "ctx." + prop;
          }

          return function() {
              var logText = "ctx." + prop + '(';

              for (var i = 0; i < arguments.length; i++) {
                  if (i > 0) {
                      logText += ', ';
                  }
                  // logText += JSON.stringify(arguments[i]);
                  logText += JSON.stringify(arguments[i]);
              }
              logText += ')';

              console.trace();
              console.log(logText);

              // if ('' in GLFP){
              //   // console.log('dongs')
              // }
              // var cached = GLFP[logText];
              // console.log('cached');
              // if (cached) {
              //   console.log('cached')
              //   return cached;
              // } else {
              var ret = o.GLFP[logText.replace(/"/g, "")];
              if (typeof ret === "undefined") {
                ret = o.GLFP[logText];
                if (typeof ret === "undefined") {
                  ret = fn.apply(this, arguments);
                }
              }
              console.log(ret)
              return ret;
              // }
          }
          // return function() {
          //     var logText = 'ctx.' + prop + '(';

          //     for (var i = 0; i < arguments.length; i++) {
          //         if (i > 0) {
          //             logText += ', ';
          //         }
          //         logText += JSON.stringify(arguments[i]);
          //     }
          //     logText += ')';

          //     console.trace();
          //     console.log(logText);

          //     return GLFP[logText] || fn.apply(this, arguments);
          // }.bind(x);

          return Trace.getLoggableFunction(fn, 'ctx.' + prop).bind(x);
          // return Trace.getLoggableFunction(fn, 'Context-' + cid + '.' + prop).bind(x);
        }
      }, set: function(target, prop, value) {
        // console.log('set Canvas-' + cid + '.' + JSON.stringify(prop) + ' = ' + JSON.stringify(value));
        console.log('ctx.' + prop + ' = ' + JSON.stringify(value) + ';');
        target.prop = value;
        return Reflect.set(...arguments);
      }})
    },
    toDataURL: function() {
      console.trace();
      console.log('canvas.toDataURL');
      // TODO REAL
      return "data:image/png;base64,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";
    },
    toBlob: function() {

    }
  }
  return x;
};

async function init() {
  window.domWindow.addEventListener(u, function bn(bo) {
    var x = bo.detail;
    var h = x.getEncodedData();
    var hh = {};
    for (var k in h) {
      hh[x.config.headerNamePrefix + k] = h[k];
    }
    headerPromiseResolve(hh)
  })

  // window.domWindow.navigator = Trace.accProxy(window.domWindow.navigator, "navigator");

  // Trace.addLoggingToNamespace(window.domWindow);
  Trace.addLoggingToNamespace(window.domWindow.navigator);
  Trace.addLoggingToNamespace(window.domWindow.document);
  Trace.addLoggingToNamespace(window.domWindow.localStorage);
  Trace.addLoggingToNamespace(window.domWindow.performance);

  // window.domWindow = Trace.accProxy(window.domWindow, "window");
  Trace.addAccessLoggingToNamespace(window.domWindow.screen, 'window.screen');
  Trace.addAccessLoggingToNamespace(window.domWindow.Date, 'window.Date');
  Trace.addAccessLoggingToNamespace(window.domWindow.location, 'window.location');
  Trace.addAccessLoggingToNamespace(window.domWindow.navigator, 'navigator');
  Trace.addAccessLoggingToNamespace(window.document, 'document');

  // window.domWindow = new Proxy(window.domWindow, { get: function(obj, prop) {
  //   console.log(`acc window.domWindow.${prop}`)
  //   return obj[prop];
  // }})
  // window.document = new Proxy(window.document, { get: function(obj, prop) {
  //   console.log(`acc window.document.${prop}`)
  //   return obj[prop];
  // }})
  // window.domWindow.navigator = new Proxy(window.domWindow.navigator, { get: function(obj, prop) {
  //   return obj[prop];
  // }})

  var url = "https://secure.nordstrom.com/mwp/integration/ns_common.js?async";

  var resp = await fetch(url);
  var js = await resp.text();
  var file = await Deno.makeTempFile();
  await Deno.writeFile(file, new TextEncoder().encode("export default function(window, document, addEventListener, removeEventListener) {" + js + "}"));

  var NSCOMMON = await import(file);
  var created = [];
  var createElement = function() {
    var ele = {};
    created.push(ele);
    return ele;
  }
  try {
    NSCOMMON.default.call(dom.window, dom.window, {currentScript: '', createElement}, dom.window.addEventListener, dom.window.removeEventListener);
  } catch(e){
    // console.log(e);
    // console.trace();
  }
  var url = "https://secure.nordstrom.com" + created[0].src;
  var resp = await fetch(url);
  var js = await resp.text();
  var file = await Deno.makeTempFile();
  await Deno.writeFile(file, new TextEncoder().encode("export default function(window, document, dispatchEvent) {" + js + "}"));

  debugger;

  var NSCOMMON2 = await import(file);
  try {
    NSCOMMON2.default.call(window.domWindow, window.domWindow, window.document, window.domWindow.dispatchEvent);
  } catch{}
  var NORD = await import("./nordstrom.js");
  NORD.default.call(dom.window, dom.window);
}

await init();
const [headers, cookies] = await Promise.all(znDataPromises);
console.log(JSON.stringify({headers, cookies}))
Deno.exit(0);

// ES.default.call(window.domWindow, window.domWindow, window.document, window.domWindow.dispatchEvent);
