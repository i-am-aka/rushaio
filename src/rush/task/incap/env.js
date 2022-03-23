import jsdom from 'https://rushjspm.s3-us-west-1.amazonaws.com/jsdom.js';

const dom = new jsdom.JSDOM(``, {
  url: "https://shop.ccs.com/",
  // referrer: "https://secure.nordstrom.com/checkout",
  contentType: "text/html",
  includeNodeLocations: false,
  pretendToBeVisual: true,
});
// window.addEventListener = dom.window.document.addEventListener;
// globalThis.window = dom.window;
window.document = dom.window.document;

import Trace from './trace.js';
dom.window.Trace = Trace;


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
dom.window.navigator = navFp;
// globalThis.window = dom.window;
globalThis.document = dom.window.document;
globalThis.navigator = dom.window.navigator;
// globalThis.global = dom.window.document;

// import * as Storage from './storage.ts';
// globalThis.sessionStorage = new Storage.SessionStorage();
// window.sessionStorage = globalThis.sessionStorage;
globalThis.navigator = dom.window.navigator;
globalThis.webkitURL = dom.window.URL;
globalThis.global = new Proxy({}, {
  get: function() {
    throw new ReferenceError("global is not defined");
  }
});
// globalThis.global = undefined;
dom.window.WebGLRenderingContext = function(){}
globalThis.WebAssembly = {};


globalThis.chrome = {
  "app": {
    "isInstalled": false,
    "InstallState": {
      "DISABLED": "disabled",
      "INSTALLED": "installed",
      "NOT_INSTALLED": "not_installed"
    },
    "RunningState": {
      "CANNOT_RUN": "cannot_run",
      "READY_TO_RUN": "ready_to_run",
      "RUNNING": "running"
    }
  },
  "runtime": {
    "OnInstalledReason": {
      "CHROME_UPDATE": "chrome_update",
      "INSTALL": "install",
      "SHARED_MODULE_UPDATE": "shared_module_update",
      "UPDATE": "update"
    },
    "OnRestartRequiredReason": {
      "APP_UPDATE": "app_update",
      "OS_UPDATE": "os_update",
      "PERIODIC": "periodic"
    },
    "PlatformArch": {
      "ARM": "arm",
      "MIPS": "mips",
      "MIPS64": "mips64",
      "X86_32": "x86-32",
      "X86_64": "x86-64"
    },
    "PlatformNaclArch": {
      "ARM": "arm",
      "MIPS": "mips",
      "MIPS64": "mips64",
      "X86_32": "x86-32",
      "X86_64": "x86-64"
    },
    "PlatformOs": {
      "ANDROID": "android",
      "CROS": "cros",
      "LINUX": "linux",
      "MAC": "mac",
      "OPENBSD": "openbsd",
      "WIN": "win"
    },
    "RequestUpdateCheckStatus": {
      "NO_UPDATE": "no_update",
      "THROTTLED": "throttled",
      "UPDATE_AVAILABLE": "update_available"
    }
  }
};
// dom.window.window = dom.window;

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}

// var { XMLHttpRequest} = dom.window;
// globalThis.XMLHttpRequest = XMLHttpRequest;

// Object.defineProperty(dom.window, 'eval', {
//     writable: true,
//     value: Trace.getLoggableFunction(dom.window.eval, 'eval') });

Object.defineProperty(dom.window, 'eval', {
     writable: true,
     value: new Proxy(eval, {
      apply: function(target, thisArg, argumentsArr) {
        // console.log('eval("' + argumentsArr[0] + '")')
        if (argumentsArr[0] === "'vdf204bf4886cc5811a59a20751318eeb030107d232d9f8f0b8021fcd9d66b273'.toString()") {
          debugger;
        }
        if (argumentsArr[0] === "window.toString()") {
          return "[object Window]";
        } else if (argumentsArr[0] === "window.constructor.toString()") {
          return "function Window() { [native code] }";
        } else {
          var ret = Reflect.apply(...arguments);
          if (typeof ret === 'undefined') {
            return ret;
          } else {
            return new Proxy(ret, {
              get: function(obj, prop) {
                // console.trace();
                // console.log('get ' + prop);
                return Reflect.get(...arguments);
              }
            })
          }
        }
      }
     })
});


// Trace.addAccessLoggingToNamespace(dom.window.screen, 'window.screen');
// Trace.addAccessLoggingToNamespace(dom.window.Date, 'window.Date');
// Trace.addAccessLoggingToNamespace(dom.window.location, 'window.location');
// Trace.addAccessLoggingToNamespace(dom.window.navigator, 'navigator');
// Trace.addAccessLoggingToNamespace(dom.window.document, 'document');
// Trace.addAccessLoggingToNamespace(dom.window, 'window');

async function importWrappedUrl(url, cookie) {
  var resp = await fetch(url, {
    "headers": {
      "accept": "*/*",
      "accept-language": "en-US,en;q=0.9",
      "cache-control": "no-cache",
      "pragma": "no-cache",
      "sec-fetch-dest": "script",
      "sec-fetch-mode": "no-cors",
      "sec-fetch-site": "same-origin",
      "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.137 Safari/537.36",
      "cookie": cookie
    },
    "referrer": "https://shop.ccs.com/",
    "referrerPolicy": "no-referrer-when-downgrade",
    "body": null,
    "method": "GET",
    "mode": "cors"
  });

  var js = await resp.text();
  js = js.replace(' return z;})());})();', 'return z; }');
  js = js.replace('eval((function(){', '');
  js = js.replace('(function() { ', 'export default function() {');
  var file = await Deno.makeTempFile();

  await Deno.writeFile(file, new TextEncoder().encode(js));
  var unWrapFn = await import(file);
  var unwrappedJs = unWrapFn.default();
  unwrappedJs = unwrappedJs.replace(/;}}\(/g, ';}}.bind(this)(');
  unwrappedJs = unwrappedJs.replace(/;}\(/g, ';}.bind(this)(');

  var file = await Deno.makeTempFile();
  await Deno.writeFile(file, new TextEncoder().encode("export default function(window, document, navigator, WebGLRenderingContext) {" + unwrappedJs + "}" ));

   // "export default function(window, document, navigator, WebGLRenderingContext) {" + js + "}"));
  return importWrapped(file);
}

async function importWrapped(file) {
  var wrapped = await import(file);
  return wrapped.default.bind(dom.window);
}


// async function jsdomImport(file) {
//   var scriptEl = dom.window.document.createElement('script');
//   scriptEl.textContent = await Deno.readFile(file);
//   console.log(scriptel.textContent)
//   dom.window.document.head.appendChild(scriptEl);
// }

// import { readFileStr } from 'https://deno.land/std/fs/mod.ts';

// var incap = await importWrapped('http://shop.ccs.com/_Incapsula_Resource?SWJIYLWA=5074a744e2e3d891814e9a2dace20bd4,719d34d31c8e3a6e6fffd425f7e032f3');
// var incap = await importWrapped('./incap_unwrap.js');
// const decoder = new TextDecoder('utf-8');
// eval(decoder.decode(await Deno.readFile('./incap_unwrap.js')));
// var {incap} = await import('./incap_unwrap.js')

var incap = await importWrappedUrl(
  Deno.args[0] || "https://shop.ccs.com/_Incapsula_Resource?SWJIYLWA=719d34d31c8e3a6e6fffd425f7e032f3&ns=1&cb=446181089",
  Deno.args[1]);

// incap = incap.bind(dom.window);
// var origCrEle = dom.window.document.createElement;
var eles = [];
dom.window.document.createElement = new Proxy(dom.window.document.createElement, {
  apply: function() {
    var ele = Reflect.apply(...arguments);
    eles.push(ele)
    return ele;
  }
});

function getCookie(cookies, name) {
  const value = `; ${cookies}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
}

var seshCookie = Deno.args[1].split('; ').filter(s => s.indexOf('incap_ses')===0)[0];
if (!seshCookie) {
  console.log('Error: no session cookie provided in second arg')
  Deno.exit(1);
}
dom.window.document.cookie = seshCookie;

// debugger;
// Trace.addLoggingToNamespace(dom.window);
// Trace.addLoggingToNamespace(dom.window._globalProxy);
// Trace.addLoggingToNamespace(dom.window.MutationObserver);
// Trace.addLoggingToNamespace(dom.window.navigator);
// Trace.addLoggingToNamespace(dom.window.document);
// Trace.addLoggingToNamespace(dom.window.localStorage);
// Trace.addLoggingToNamespace(dom.window.performance);

// dom.window.navigator = Trace.accProxy(dom.window.navigator, "navigator");
// dom.window.document.location = Trace.accProxy(dom.window.document.location, "location");
// dom.window = Trace.accProxy(dom.window, "window");
// dom.document = Trace.accProxy(dom.window.document, "document");

incap(dom.window, dom.window.document, dom.window.navigator, globalThis.WebGLRenderingContext);
// jsdomImport('./incap_unwrap.js');

// await delay(5000);
console.log(JSON.stringify({
  cookie: getCookie(dom.window.document.cookie, '___utmvc'),
  endpoint: eles[0].src
}))

// incap(dom.window, dom.window.document, null);
