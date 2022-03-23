import jsdom from 'https://dev.jspm.io/jsdom';
import { logDebug } from './log.js';
import docFields from './Document.js';

export function createDom(url, userAgent) {
  userAgent = userAgent || "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.67 Safari/537.36";

  var FuncNativeCode = function(name) {
    return "function " + (name || "") + "() { [native code] }";
  }
  var nativeProxy = function(fn, name) {
    console.log('nativeProxy (' + name + ")")
    const n = name;
    return new Proxy(fn, {
      get: () => {
        if (arguments[0] === "toString") {
          console.log(n);
          return FuncNativeCode.bind(null, n || "")
        }
        return Reflect.get(...arguments);
      },
    })
  }


  const dom = new jsdom.JSDOM(`<html lang="en">
<body>
    <div>
        <div data-active="0" id="sign-in-widget">
            <div class="sign-in-widget">
                <div class="re-captcha">
                    <p class="bot-message"> Help us keep your account safe by clicking on the checkbox below.</p>
                    <div id="px-captcha" style="margin:32px; align-content:center;"></div>
                </div>
            </div>
        </div>
</body>

</html>
`, {
    url: url,
    referer: "",
    contentType: "text/html",
    includeNodeLocations: false,
    pretendToBeVisual: true,
  });

  dom.window.document.elementFromPoint = function(){};
  dom.Document = function(){}
  docFields.forEach(field => { dom.Document.prototype[field] = null })

  dom.window.history.pushState({}, "");
  dom.window.screen = {
    width: 1366,
    height: 768,
    availWidth: 1366,
    availHeight: 768-23,
    colorDepth: 24,
    pixelDepth: 24,
  }
  dom.window.outerHeight = dom.window.screen.availHeight - 23 - parseInt(Math.random() * 600);
  dom.window.outerWidth = dom.window.screen.availWidth - parseInt(Math.random() * 500);
  dom.window.innerHeight = dom.window.outerHeight - 141;
  dom.window.innerWidth = dom.window.outerWidth;

  dom.window.PaymentInstruments = function(){}
  Object.defineProperty(dom.window.PaymentInstruments, "toString", {enumerable: false, writable: true});
  dom.window.PaymentInstruments.toString = FuncNativeCode.bind(null, "PaymentInstruments");

  dom.window.PaymentManager = function(){}
  Object.defineProperty(dom.window.PaymentManager, "toString", {enumerable: false, writable: true});
  dom.window.PaymentManager.toString = FuncNativeCode.bind(null, "PaymentManager");

  dom.window.PaymentRequest = function(){}
  Object.defineProperty(dom.window.PaymentRequest, "toString", {enumerable: false, writable: true});
  dom.window.PaymentRequest.toString = FuncNativeCode.bind(null, "PaymentRequest");

  dom.window.PaymentResponse = function(){}
  Object.defineProperty(dom.window.PaymentResponse, "toString", {enumerable: false, writable: true});
  dom.window.PaymentResponse.toString = FuncNativeCode.bind(null, "PaymentResponse");

  dom.window.PaymentAddress = function(){}
  Object.defineProperty(dom.window.PaymentAddress, "toString", {enumerable: false, writable: true});
  dom.window.PaymentAddress.toString = FuncNativeCode.bind(null, "PaymentAddress");

  dom.window.PaymentRequestUpdateEvent = function(){}
  Object.defineProperty(dom.window.PaymentRequestUpdateEvent, "toString", {enumerable: false, writable: true});
  dom.window.PaymentRequestUpdateEvent.toString = FuncNativeCode.bind(null, "PaymentRequestUpdateEvent");
  dom.window.PaymentRequestUpdateEvent.NONE = 0
  dom.window.PaymentRequestUpdateEvent.CAPTURING_PHASE = 1
  dom.window.PaymentRequestUpdateEvent.AT_TARGET = 2
  dom.window.PaymentRequestUpdateEvent.BUBBLING_PHASE = 3









  dom.window.Notification = function(){}
  Object.defineProperty(dom.window.Notification, "toString", {enumerable: false, writable: true});
  dom.window.Notification.permission = "default";
  dom.window.Notification.maxActions = 2;
  dom.window.Notification.requestPermission = FuncNativeCode();
  dom.window.Notification.toString = FuncNativeCode.bind(null, "Notification");

  dom.window.Worklet = {};
  dom.window.AudioWorklet = {};
  dom.window.AudioWorkletNode = {};
  dom.window.caches = {};
  dom.window.offscreenBuffering = true;
  dom.window.caches.toString = function() { return "[object CacheStorage]" }
  dom.window.applicationCache = {};
  dom.window.applicationCache.toString = function() { return "[object ApplicationCache]" }
  dom.window.BatteryManager = {};
  dom.window.BatteryManager.toString = function() {
    return "{ [native code] }";
  }
  dom.window.trustedTypes = {};
  dom.window.trustedTypes.toString = function() {
    return "[object TrustedTypePolicyFactory]";
  }
  dom.window.ondevicemotion = null;
  dom.window.ondeviceorientation = null;
  dom.window.ondragexit = undefined;
  delete dom.window.ondragexit;
  dom.window.document.ondragexit = undefined;
  delete dom.window.document.ondragexit;

  dom.window.document.location.ancestorOrigins = {length: 0};
  dom.window.document.location.ancestorOrigins.toString = function() { return "[object DOMStringList]" };
  dom.window.location.fragmentDirective = {};
  dom.window.location.fragmentDirective.toString = function() { return "[object FragmentDirective]" };
  var origHasOP = dom.window.location.hasOwnProperty.bind(dom.window.location);
  dom.window.location.hasOwnProperty = function() {
    if (arguments[0] === "fragmentDirective") {
      return false;
    } else {
      return origHasOP(...arguments);
    }
  }
  dom.window.console.log = function() { }
  dom.window.console.log.toString = function() { return "function log() { [native code] }"}


  // f.contentDocument.appendChild('iframe')
  // debugger;

  // window.addEventListener = dom.window.document.addEventListener;
  // globalThis.window = dom.window;
  globalThis.document = dom.window.document;
  globalThis.window = dom.window;

  /*

  */
  /*
  PX44: 4640
  PX45: 2488
  PX46: 11
  PX47: undefined
  PX48: undefined
  PX49: undefined
  PX50: undefined
  PX51: undefined
  PX52: 1017
  PX53: 810
  PX54: 257
  */

/*
PX44: 2399
PX45: 1366
PX46: 1
PX47: undefined
PX48: undefined
PX49: undefined
PX50: undefined
PX51: undefined
PX52: 344
PX53: 619
PX54: 18
*/

/*
PX44: 2609
PX45: 1528
PX46: 1
PX47: undefined
PX48: undefined
PX49: undefined
PX50: undefined
PX51: undefined
PX52: 318
PX53: 671
PX54: 41

connectEnd: 1597464769780
connectStart: 1597464769780
domComplete: 1597464772301
domContentLoadedEventEnd: 1597464770785
domContentLoadedEventStart: 1597464770774
domInteractive: 1597464770773
domLoading: 1597464770067
domainLookupEnd: 1597464769780
domainLookupStart: 1597464769780
fetchStart: 1597464769780
loadEventEnd: 1597464772388
loadEventStart: 1597464772347
navigationStart: 1597464769779
redirectEnd: 0
redirectStart: 0
requestStart: 1597464769784
responseEnd: 1597464770102
responseStart: 1597464770050
secureConnectionStart: 0
unloadEventEnd: 1597464770056
unloadEventStart: 1597464770056

  PX44: js.timing.loadEventEnd - js.timing.navigationStart || void 0,
  PX45: js.timing.domComplete - js.timing.domInteractive || void 0,
  PX46: js.timing.fetchStart - js.timing.navigationStart || void 0,
  PX47: js.timing.redirectEnd - js.timing.redirectStart || void 0,
  PX48: js.timing.domainLookupStart - js.timing.fetchStart || void 0,
  PX49: js.timing.unloadEventEnd - js.timing.unloadEventStart || void 0,
  PX50: js.timing.domainLookupEnd - js.timing.domainLookupStart || void 0,
  PX51: js.timing.connectEnd - js.timing.connectStart || void 0,
  PX52: js.timing.responseEnd - js.timing.requestStart || void 0,
  PX53: js.timing.domInteractive - js.timing.responseEnd || void 0,
  PX54: js.timing.loadEventEnd - js.timing.loadEventStart || void 0,

*/
  var now = Date.now();
  var t0 = now - parseInt(Math.random() * 3000);
  var respEnd = t0 + parseInt(Math.random()*300);
  var domInteractive = respEnd + 500 + parseInt(Math.random() * 500);
  var loadStart = t0 + 2000 + parseInt(Math.random()*3000);
  var loadEnd = loadStart + parseInt(Math.random()*5000);
  var loadAlmostEnd = loadEnd - parseInt(Math.random()*50);
  var fetchStart = t0 + 1 + parseInt(Math.random() * 15);
  var perfFp = {
    "timeOrigin": Date.now() + .453,
    "timing": {
        "connectStart": now,
        "navigationStart": t0,
        "loadEventEnd": loadStart + parseInt(Math.random()*300),
        "domLoading": 1596256167447,
        "secureConnectionStart": 0,
        "fetchStart": fetchStart,
        "domContentLoadedEventStart": loadAlmostEnd + parseInt(Math.random()*100),
        "responseStart": now,
        "responseEnd": respEnd,
        "domInteractive":  domInteractive,
        "domainLookupEnd": fetchStart,
        "redirectStart": 0,
        "requestStart": respEnd - (300 + parseInt(Math.random() * 1000)),
        "unloadEventEnd": 0,
        "unloadEventStart": 0,
        "domComplete": loadStart - parseInt(Math.random()*100),
        "domainLookupStart": fetchStart,
        "loadEventStart": loadStart,
        "domContentLoadedEventEnd": loadAlmostEnd + parseInt(Math.random()*100),
        "redirectEnd": 0,
        "connectEnd": now
      },

    "navigation": {
      "type": 0,
      "redirectCount": 0
    },
    "memory": {"totalJSHeapSize":10966782,"usedJSHeapSize":9231994,"jsHeapSizeLimit":4294705152},
  };
  var js = perfFp;
  logDebug({  PX44: js.timing.loadEventEnd - js.timing.navigationStart || void 0,
  PX45: js.timing.domComplete - js.timing.domInteractive || void 0,
  PX46: js.timing.fetchStart - js.timing.navigationStart || void 0,
  PX47: js.timing.redirectEnd - js.timing.redirectStart || void 0,
  PX48: js.timing.domainLookupStart - js.timing.fetchStart || void 0,
  PX49: js.timing.unloadEventEnd - js.timing.unloadEventStart || void 0,
  PX50: js.timing.domainLookupEnd - js.timing.domainLookupStart || void 0,
  PX51: js.timing.connectEnd - js.timing.connectStart || void 0,
  PX52: js.timing.responseEnd - js.timing.requestStart || void 0,
  PX53: js.timing.domInteractive - js.timing.responseEnd || void 0,
  PX54: js.timing.loadEventEnd - js.timing.loadEventStart || void 0});
  globalThis.performance = dom.window.performance;
  for (var x in perfFp) {
    try {
      dom.window.performance[x] = perfFp[x];
    }catch(e) {
      logDebug('cannot set performance.'+x)
    }
  }

  dom.window.assert = function(){};
  globalThis.assert = dom.window.assert;

  var origNav = dom.window.navigator;
  var navFp = {
    "vendorSub": "",
    "productSub": "20030107",
    "vendor": "Google Inc.",
    "maxTouchPoints": 0,
    "doNotTrack": true,
    "hardwareConcurrency": [4,6,8,12,16][Math.floor(Math.random() * 5)],
    "cookieEnabled": true,
    "appCodeName": "Mozilla",
    "appName": "Netscape",
    "appVersion": userAgent.replace('Mozilla/', ''), // todo this is mayb crhome only
    "platform": "MacIntel",
    "product": "Gecko",
    "userAgent": userAgent,
    "language": "en-US",
    "languages": [
      "en-US",
      "en"
    ],
    "onLine": true,
    "connection": {
      "effectiveType": "4g",
      "rtt": 90,
      "downlink": (1 + Math.random(10)).toFixed(1),
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
      "length": 3,
      "toString": function() { return "[object PluginArray]"; },
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
      "toString": function() { return "[object MimeTypeArray]" },
      "length": 4
    },
    "userActivation": {
      "hasBeenActive": false,
      "isActive": false
    },
    "mediaSession": {
      "playbackState": "none"
    },
    "deviceMemory": 8,
    "geolocation": "[object Geolocation]",
    "getBattery": function() {

    },
  };
  var stringFields = {
  clipboard: "[object Clipboard]",
  credentials: "[object CredentialsContainer]",
  keyboard: "[object Keyboard]",
  locks: "[object LockManager]",
  mediaDevices: "[object MediaDevices]",
  serviceWorker: "[object ServiceWorkerContainer]",
  storage: "[object StorageManager]",
  presentation: "[object Presentation]",
  bluetooth: "[object Bluetooth]",
  usb: "[object USB]",
  xr: "[object XRSystem]",
  "setAppBadge": "function () { [native code] }",
  "clearAppBadge": "function () { [native code] }",
  "getInstalledRelatedApps":"function () { [native code] }",
  "getUserMedia": "function () { [native code] }",
  "webkitGetUserMedia": "function () { [native code] }",
  "requestMIDIAccess": "function () { [native code] }",
  wakeLock: "[object WakeLock]"
  };
  for (var field in stringFields) {
    navFp[field] = {};
    navFp[field].toString = function(f) { return stringFields[f] }.bind(null, field);
  }
  // var origHasOP = navFp.hasOwnProperty;
  navFp.hasOwnProperty = function() {
    return false;
  }
  // navFp.toString = function() {
  //   return "[object Object]function Object() { [native code] }loadTimesfunction () { [native code] }csifunction () { [native code] }app[object Object]runtime[object Object]";
  // }
  navFp.getBattery.toString = function() {
    return "{ [native code] }";
  }

  dom.window.navigator = navFp;
  // for (var x in navFp)  {
  //   try {
  //     dom.window.navigator[x] = navFp[x];
  //   }catch(e) {
  //     logDebug('cannot set navigator.'+x)
  //   }
  // }
  // dom.window.navigator.plugins = navFp.plugins;
  // globalThis.window = dom.window;
  globalThis.MediaSource = {isTypeSupported: function() { return true }};
  dom.window.MediaSource = globalThis.MediaSource;
  dom.window.isSecureContext = true;
  globalThis.document = dom.window.document;
  globalThis.navigator = dom.window.navigator;
  // globalThis.global = dom.window.document;

  // import * as Storage from './storage.ts';
  // globalThis.sessionStorage = new Storage.SessionStorage();
  // window.sessionStorage = globalThis.sessionStorage;
  globalThis.navigator = dom.window.navigator;
  globalThis.webkitURL = dom.window.URL;
  globalThis.Element = dom.window.Element;
  globalThis.HTMLIFrameElement = dom.window.HTMLIFrameElement;
  globalThis.HTMLElement = dom.window.HTMLElement;
  globalThis.HTMLDocument = dom.window.HTMLDocument;
  globalThis.screen = dom.window.screen;
  globalThis.Audio = function() {
    this.addEventListener = function(){}
  }
  globalThis.location = dom.window.location || dom.window.document.location;

  globalThis.global = new Proxy({}, {
    get: function() {
      throw new ReferenceError("global is not defined");
    }
  });
  // globalThis.global = undefined;
  dom.window.WebGLRenderingContext = function(){}
  globalThis.WebAssembly = {};
  dom.window.WebAssembly = globalThis.WebAssembly;


  globalThis.chrome = {
    loadTimes: function(){},
    csi: function(){},
    "app": {
      "isInstalled": false,
      getDetails: function(){},
      getIsInstalled: function(){},
      installState: function(){},
      runningState: function(){},
      "InstallState": {
        "DISABLED": "disabled",
        "INSTALLED": "installed",
        "NOT_INSTALLED": "not_installed"
      },
      "RunningState": {
        "CANNOT_RUN": "cannot_run",
        "READY_TO_RUN": "ready_to_run",
        "RUNNING": "running"
      },
    },
    "runtime": {
      id: undefined,
      connect: FuncNativeCode("connect"), // nativeProxy(function(){}, "connect"), // FuncNativeCode("connect"),
      sendMessage: FuncNativeCode("sendMessage"), // nativeProxy(function(){}, "sendMessage"),
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
      },
    }
  };

  var setAllNativeFns = function(obj) {
    for (var i in obj) {
      if (typeof obj[i] === 'function') {
        Object.defineProperty(obj[i], "toString", {enumerable: false, writable: true});
        obj[i].toString = FuncNativeCode;
        // obj[i] = nativeProxy(obj[i]);
      } else if (typeof obj[i] === 'object') {
        setAllNativeFns(obj[i]);
      }
    }
  }

  setAllNativeFns(globalThis.chrome);
  globalThis.chrome.app.getDetails = FuncNativeCode("getDetails");
  globalThis.chrome.app.getIsInstalled = FuncNativeCode("getIsInstalled");
  globalThis.chrome.app.installState = FuncNativeCode("installState");
  globalThis.chrome.app.runningState = FuncNativeCode("runningState");
  // globalThis.chrome.app.getDetails = nativeProxy(function(){}, "getDetails");
  // globalThis.chrome.app.getIsInstalled = nativeProxy(function(){}, "getIsInstalled");
  // globalThis.chrome.app.installState = nativeProxy(function(){}, "installState");
  // globalThis.chrome.app.runningState = nativeProxy(function(){}, "runningState");



  // globalThis.chrome.loadTimes = nativeProxy(globalThis.chrome.loadTimes);
  // globalThis.chrome.csi = nativeProxy(globalThis.chrome.csi);


  // dom.window.setTimeout = function(x) {
  //   x();
  // }
  dom.window.setTimeout.toString = function() {
    return "{ [native code] }";
  }
  dom.window.chrome = globalThis.chrome;

  // import * as Storage from './storage.ts';
  // dom.window.sessionStorage = new Storage.SessionStorage();
  // dom.window.localStorage = new Storage.SessionStorage();

  var pointerListeners = [];
  dom.window._globalObject.HTMLIFrameElement.prototype.getBoundingClientRect = new Proxy(dom.window._globalObject.HTMLIFrameElement.prototype.getBoundingClientRect, {
    apply() {
      logDebug("getBoundingClientRect()");
      return {"x":8,"y":8,"width":310,"height":100,"top":8,"right":318,"bottom":108,"left":8};
    }
  });
  dom.window._globalObject.EventTarget.prototype.addEventListener = new Proxy(dom.window._globalObject.EventTarget.prototype.addEventListener, {
    apply(target, thisArg, argumentsArr) {
      var value =  Reflect.apply(...arguments);
      logDebug(`addEventListener(${argumentsArr[0]}, ...)`)
      if (argumentsArr[0] === 'pointerdown') {
        pointerListeners.push(argumentsArr[1]);
      }
      return value;
    }
  });

  var origCr = dom.window.document.createElement;
  var SUPPORTED_CODECS = {
    'audio/mpeg;': 1,
    'audio/webm; codecs="vorbis"':1,
    'audio/ogg; codecs="vorbis"':1,
    'audio/wav; codecs="1"':1,
    'audio/ogg; codecs="flac"':1,
    'video/mp4; codecs="avc1.42E01E"': 1,
    'video/mp4; codecs="avc1.42E01E, mp4a.40.2"': 1,
    'video/mp4; codecs="avc1.58A01E"': 1,
    'video/mp4; codecs="avc1.4D401E"': 1,
    'video/mp4; codecs="avc1.64001E"': 1,
    'video/webm; codecs="vp8"': 1,
    'video/ogg; codecs="theora"': 1,
  }
  Object.defineProperty(dom.window.document, 'createElement', {
     writable: true,
     value: new Proxy(origCr, {
      apply: function(target, thisArg, argumentsArr) {
        var tag = argumentsArr[0];
        logDebug(`createElement(${JSON.stringify(argumentsArr)})`)
        // console.trace();
        if (argumentsArr[0] === "script") {
          var script = Reflect.apply(...arguments);
          script = new Proxy(script, {
            set: function(obj, property, value) {
              logDebug(`script.${property}=${value}`);
              if (property === 'src') {
                (typeof scriptPromiseResolve !== 'undefined') && scriptPromiseResolve(value);
              } else if (property === 'onerror') {
                return Reflect.set(obj, property, function(){});
              }
              return Reflect.set(...arguments);
            },
          });
          return script;
        } else if (argumentsArr[0] === "audio" || argumentsArr[1] === "video") {
          return {
            canPlayType: function(e) {
              logDebug('canPlayType(' + JSON.stringify(e) + ')')
              if (SUPPORTED_CODECS[e]) {
                return "probably"
              } else {
                return ""
              }
            }
          }
        } else {
          var ele = Reflect.apply(...arguments);
          if (ele instanceof dom.window.HTMLIFrameElement) {
            // debugger;
          }
          // ele = new Proxy(ele, {
          //   get: function(obj, property, value) {
          //     return Reflect.get(...arguments);
          //   },
          // });
          return ele;
          // Trace.addLoggingToNamespace(ele);
          // return Trace.accProxy(ele, argumentsArr[0]);
        }
      }
    })
  });
  return dom;
}