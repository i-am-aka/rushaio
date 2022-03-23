import { createDom } from './dom.js';
import { logDebug } from './log.js';

// import Trace from './trace.js';

var modCache = {};

async function importWrappedUrl(dom, url, referer, cookie, setTimeout_, setInterval_) {
  var mc = modCache[url];
  if (mc) {
    if ((Date.now() - mc.time) < 300e3) {
      return bindWrapped(mc.mod, dom, setTimeout_, setInterval_);
    }
  }

  logDebug(`importWrappedUrl(${JSON.stringify(arguments)})`)
  var headers = {
    "accept": "*/*",
    "accept-language": "en-US,en;q=0.9",
    "cache-control": "no-cache",
    "pragma": "no-cache",
    // "sec-fetch-dest": "script",
    // "sec-fetch-mode": "no-cors",
    // "sec-fetch-site": "same-origin",
    "User-Agent": dom.window.navigator.userAgent,
  };
  if (cookie && cookie.length > 0) {
    headers['cookie'] = cookie;
  }
  var resp = await fetch(url, {
    "headers": headers,
    "referrer": referer,
    "referrerPolicy": "no-referrer-when-downgrade",
    "redirect": "follow",
    "body": null,
    "method": "GET",
    // "mode": "cors"/
  });

  var js = await resp.text();
  logDebug('JS LEN: ' + js.length.toString())
  // var beau = await import('https://dev.jspm.io/js-beautify');
  // js = beau.default(js);


  // // TODO edit js if needed
  // var file = await Deno.makeTempFile();

  // await Deno.writeFile(file, new TextEncoder().encode(js));
  // var unWrapFn = await import(file);
  var [mod, wrapped] = await importUnwrapped(js, dom, setTimeout_, setInterval_);
  modCache[url] = {mod, time: Date.now()};
  return wrapped;
}

async function importUnwrapped(js, dom, setTimeout_, setInterval_) {
  var file = await Deno.makeTempFile();
  await Deno.writeFile(file, new TextEncoder().encode("export default function(window, document, navigator, WebGLRenderingContext, HTMLDocument, HTMLDivElement, Document, XMLHttpRequest, setTimeout, setInterval, clearTimeout, clearInterval) {" + js + "}" ));
  logDebug(file);
  // Deno.exit(0);
  return await importWrapped(file, dom, setTimeout_, setInterval_);
}
async function importWrapped(file, dom, setTimeout_, setInterval_) {
  var wrapped = await import(file);
  return [wrapped.default, bindWrapped(wrapped.default, dom, setTimeout_, setInterval_)];
}

function bindWrapped(wrapped, dom, setTimeout_, setInterval_) {
  setTimeout_ = setTimeout_ || setTimeout;
  setInterval_ = setInterval_ || setInterval;

  return wrapped.bind(dom.window, dom.window, dom.window.document, dom.window.navigator, dom.window.WebGLRenderingContext, {}, dom.Document, dom.window.HTMLDivElement, dom.window.XMLHttpRequest, dom.window.setTimeout, dom.window.setInterval, dom.window.clearTimeout, dom.window.clearInterval);// )setTimeout_, setInterval_);
}

function allStorage(dom) {

    var archive = {}, // Notice change here
        keys = Object.keys(dom.window.sessionStorage),
        i = keys.length;

    while ( i-- ) {
        archive[ keys[i] ] = dom.window.sessionStorage.getItem( keys[i] );
    }

    return archive;
}

import { interceptXhr, interceptXhrProxy } from './xhr.js';
export async function genPx(pxArgs) {
  // TODO  worker

  var sock, dom;
  var timeouts = [], intervals = [];
  var reqs = [];
  try {
    var reqCb = (ru) => {reqs.push(ru)};
    dom = createDom(pxArgs.Url, pxArgs.UserAgent);

    var xhrProx = pxArgs.XhrProxySock || Deno.env.get("XHR_PROXY");
    if (xhrProx != "" && xhrProx != null) {
      logDebug('intercept')
      sock = await Deno.connect({ path: xhrProx, transport: "unix" });
      await interceptXhrProxy(pxArgs.Url, dom, sock, reqCb);
    } else {
      interceptXhr(dom, pxArgs.Url);
    }

    // var setTimeout_ = function() {
    //   var t = setTimeout(...arguments);
    //   timeouts.push(t);
    //   return t;
    // };
    // var setInterval_ = function() {
    //   var t = setInterval(...arguments);
    //   intervals.push(t);
    //   return t;
    // };


    setPxFields(dom, pxArgs);

    var isCapSolve = pxArgs.CapJsSrc && pxArgs.CapJsSrc.length > 0;

    if (isCapSolve) {
      delete dom.window._pxVid;
      delete dom.window._pxUuid;
       // pxArgs.CapJsSrc = 'http://localhost:8001/captcha.deobf.js';
       var pxcap = await importWrappedUrl(dom, pxArgs.CapJsSrc);
       setTimeout(pxcap, 0);
       await delay(1000); // todo remove , add clear signal for ready
       // console.log(dom.window.document.innerHTML);
    }

    var px = await importWrappedUrl(dom, pxArgs.JsSrc); // ), undefined, undefined, setTimeout_, setInterval_);
    // dom.window.setTimeout = setTimeout_;
    // dom.window.setTimeout = setTimeout_;

    var initialPxCkieName, initialPxCkie;
    var px3 = getCookie(dom, "_px3");
    var px2 = getCookie(dom, "_px2");
    if (px3) {
      initialPxCkieName = "_px3";
      initialPxCkie = px3;
    } else if (px2) {
      initialPxCkieName = "_px2";
      initialPxCkie = px2;
    } else {
      initialPxCkieName = "_px3";
    }
    var cookiesToRemove = [];
    if (initialPxCkie) {
      var cookiesToRemove = [[initialPxCkieName, initialPxCkie]];
    }
    var pxsid = getCookie(dom, "pxsid");
    if (pxsid) {
      dom.window.sessionStorage.setItem("pxsid", pxsid);
    }
    dom.window.sessionStorage.setItem("PXu6b0qd2S_px_c_p_PXu6b0qd2S", "0");

    logDebug('INITIAL', initialPxCkieName, initialPxCkie);
    if (DEBUG) {
      // setupTrace();
    }
    // TODO add tracing to px fns so we can know order of calls
    px();

    if (isCapSolve) {
      logDebug('PXACTION', dom.window._pxAction);
      logDebug('PX_VALS', Object.entries(dom.window).filter(function([key, value]) { return key.indexOf("px") !== -1 }).map(s => s.toString()).join("\n"));

      t0 = Date.now();

      while((Date.now() - t0) < 30e3 && (reqs.length === 0 || reqs[reqs.length - 1].indexOf('/b/g') === -1)) {
        // console.log(reqs);
           // console.log(Object.entries(dom.window).filter(([key,value]) => key && key.indexOf && key.indexOf('px') !== -1));
           await delay(250);
      }

        // console.log(dom.window.sessionStorage.getItem("pxsid"));
        logDebug(`handleCaptcha(${pxArgs.RecapToken})`);

        var initialPx3 = getCookie(dom, initialPxCkieName);
        cookiesToRemove.push([initialPxCkieName, initialPx3]);
        logDebug(`initialPx3=${initialPx3}`);
        dom.window.handleCaptcha(pxArgs.RecapToken);
        t0 = Date.now();
        while((Date.now() - t0) < 20e3 && getCookie(dom, initialPxCkieName) === initialPx3) {
          logDebug(dom.window.document.cookie);
          logDebug(getCookie(dom, initialPxCkieName));
          // logDebug(dom.window.document.body.innerHTML);
          await delay(250);
        }
        logDebug(getCookie(dom, initialPxCkieName));
        logDebug('SOLVED 8)', dom.window.document.cookie);
      } else {
        var t0 = Date.now();
        while((Date.now() - t0) < 30e3 && reqs.length < 2) {
          // else if (getCookie(dom, "_px2") || getCookie(dom, "_px3")) {
          //   break;
          // }
          // logDebug(dom.window.document.cookie);
          await delay(250);
        }
        // initialPxCkie = getCookie(dom, initialPxCkieName);
        // var t0 = Date.now();
        // while((Date.now() - t0) < 30e3) {
        //   if (initialPxCkie) {
        //     if (getCookie(dom, initialPxCkieName) !== initialPxCkie) {
        //       break;
        //     }
        //   } else if (getCookie(dom, "_px2") || getCookie(dom, "_px3")) {
        //     break;
        //   }
        //   logDebug(dom.window.document.cookie);
        //   await delay(250);
        // }
    }

    // dom.window.document.cookie = `pxsid=${dom.window.sessionStorage.getItem("pxsid")}`;
    var ckie = dom.window.document.cookie;
    cookiesToRemove.forEach(function([name, value]) {
      ckie = ckie.replace(`${name}=${value}; `, '');
    });
    return ckie;
  } finally {
    // for(var i = 0; i < intervals.length; i++) {
    //   try {
    //     clearInterval(intervals[i]);
    //   }catch(e) {};
    // }
    // for(var i = 0; i < timeouts.length; i++) {
    //   try {
    //     clearTimeout(timeouts[i]);
    //   } catch(e) {};
    // }
    try {
      if (dom) {
        setTimeout(() => dom.window.close(), 0);
      }
    } catch(e) {
      console.log('WTF')
      console.error(e);
    }
    try {
      if (sock) {
        sock.closed = true;
        await sock.close()
      }
    } catch(e) {}
  }
  // await delay(10000);
  // var initialCkie = dom.window.document.cookie;
  // logDebug('INITIAL', initialCkie)
  // while(dom.window.document.cookie === initialCkie) {
  //   await delay(500);
  // }
  // while (1) {
  //   console.log(dom.window.document.cookie)
  //   await delay(500);
  // }
}
// var jsUrl = Deno.args[0] || "https://www.hibbett.com/on/demandware.store/Sites-Hibbett-US-Site/default/Product-Show/IZ/ajdckzhd/captcha?a=c&u=88071440-a530-11ea-835a-c32c2a7d54cf&v=&m=0";
// var jsReferer = Deno.args[1] || "https://www.hibbett.com/implus-all-sport-team-sock--2-pack/6591Z.html?dwvar_6591Z_color=0001";
// var jsCkie = Deno.args[2] || "__cfduid=dc0f48271b210111f06800ce1f68dc76d1591143053; dwac_269962172965e441be86869251=qgfCo0WXImDusYTAdrZRPd41GSn-G9ThURU%3D|dw-only|||USD|false|America%2FChicago|true; cqcid=bcKNVmHB2C6KaJ6BEVyFkIBYiD; sid=qgfCo0WXImDusYTAdrZRPd41GSn-G9ThURU; dwanonymous_9bc4215d5ba3521ff1e6f9e5299a7444=bcKNVmHB2C6KaJ6BEVyFkIBYiD; __cq_dnt=0; dw_dnt=0; dwsid=fBDlTukaEqdnZtQE_DFELQgiPsRSluQm2MJ7N6KdOGkgNv7JgpDf5JoI2GcoAp5DsodkPjv5jFitfSUzVvTUSQ==; ku1-sid=2ZAKBZmFeHN6JQE6saUvY; ku1-vid=04f16dd1-9775-0885-92f9-57a818e1cba5; dw=1; dw_cookies_accepted=1; BVImplpixel_pie=18471_3_0; __olapicU=1591143062613; tfc-l=%7B%22a%22%3A%7B%22v%22%3A%22cd731956-cf8d-4399-9043-582c52f2ff31%22%2C%22e%22%3A1591229462%7D%2C%22u%22%3A%7B%22v%22%3A%22V6%7Cunk_20a10d2e-8b26-4190-ad11-0a4f5a0e332e!1654042261%7Ctjsc997r0r4bntnem5elaqfp3d!1654042261%22%2C%22e%22%3A1654042262%7D%2C%22s%22%3A%7B%22v%22%3A%22session.params%3D%257C1654042261%22%2C%22e%22%3A1654042262%7D%2C%22k%22%3A%7B%22v%22%3A%22tjsc997r0r4bntnem5elaqfp3d%22%2C%22e%22%3A1654042261%7D%7D; tfc-s=%7B%22v%22%3A%22tfc-fitrec-product%3D1%22%7D; liveagent_oref=; rdf-uuid=0367a651-4616-4c04-a43c-70774d4eb30f_1591083294341; rdf-count=1; liveagent_sid=96f1c8e6-8ac1-4ecf-882c-9e12003dcef7; liveagent_vc=2; liveagent_ptid=96f1c8e6-8ac1-4ecf-882c-9e12003dcef7";
// var px = await importWrappedUrl(jsUrl, jsReferer, jsCkie);
const decoder = new TextDecoder('utf-8');
// var px = await importWrapped('./captcha.js');
// dom.window.document.cookie = jsCkie;

function setupTrace() {
  // Trace.addAccessLoggingToNamespace(dom.window.screen, 'window.screen');
  // Trace.addAccessLoggingToNamespace(dom.window.Date, 'window.Date');
  // Trace.addAccessLoggingToNamespace(dom.window.location, 'window.location');
  // Trace.addAccessLoggingToNamespace(dom.window.navigator, 'navigator');
  // Trace.addAccessLoggingToNamespace(dom.window.document, 'document');
  // Trace.addAccessLoggingToNamespace(dom.window, 'window');

  Trace.addLoggingToNamespace(dom.window);
  // Trace.addLoggingToNamespace(dom.window._globalProxy);
  // Trace.addLoggingToNamespace(dom.window.MutationObserver);
  // Trace.addLoggingToNamespace(dom.window.navigator);
  Trace.addLoggingToNamespace(dom.window.document);
  // Trace.addLoggingToNamespace(dom.window.localStorage);
  // Trace.addLoggingToNamespace(dom.window.performance);

  // dom.window.navigator = Trace.accProxy(dom.window.navigator, "navigator");
  // dom.window.document.location = Trace.accProxy(dom.window.document.location, "location");
  // dom.window = Trace.accProxy(dom.window, "window");
  // dom.document = Trace.accProxy(dom.window.document, "document");
}
function setPxFields(dom, fields) {
  if (fields.Cookie) {
    var ckies = fields.Cookie.split("; ");
    for(let i = 0; i < ckies.length; i++) {
      dom.window.document.cookie = ckies[i];
    }
  }
  // dom.window.document.cookie = (fields.Cookie && fields.Cookie.replace(/ /g, "")) || '';
  dom.window._pxAppId = fields.AppId;
  dom.window._pxreCaptchaTheme = 'light';
  dom.window._PXETnJ2Y5H = {
      challenge: {
          view: {
               textFont: "BogleWeb, Helvetica Neue, Helvetica, Arial, sans-serif"
         }
      }
  };
  dom.window._pxJsClientSrc = fields.JsSrc;
  if (dom.window._pxJsClientSrc[0] !== "/") {
    dom.window._pxJsClientSrc = new URL(dom.window._pxJsClientSrc).pathname;
  }
  dom.window._pxFirstPartyEnabled = true;
  if (fields.Vid && fields.Vid.length > 0){
    dom.window._pxVid = fields.Vid || '';
  }
  if (fields.Uuid && fields.Uuid.length > 0) {
    dom.window._pxUuid = fields.Uuid;
  }
  dom.window._pxHostUrl = fields.Host;
}
// setPxFields({
//   appId: 'PXAJDckzHD', jsClientSrc: '/on/demandware.store/Sites-Hibbett-US-Site/default/Product-Show/IZ/ajdckzhd/init.js',
//   vid: '', uuid: '24527400-a533-11ea-b109-51a33c061e2c', hostUrl: 'https://collector-pxajdckzhd.perimeterx.net'
// });

var scriptPromiseResolve;
var scriptPromise = new Promise(((resolve) => {
  scriptPromiseResolve = resolve;
}));

var DEBUG = Deno.env.get("DEBUG") === "1";

function getCookie(dom, name) {
  const value = `; ${dom.window.document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length >= 2) {
    return parts.pop().split(';').shift();
  }
}


// TODO await captcha render

// function makeTouchEvent(type, x, y) {
  // var event = dom.window.document.createEvent('Event');
  // event.initEvent('touch' + type, true, true);
  // event.constructor.name; // Event (not TouchEvent)

  // var point = {x: 10, y: 10 };
  // event.touches = [{
  //     identifier: Date.now() + i,
  //     pageX: x,
  //     pageY: y,
  //     screenX: x,
  //     screenY: y,
  //     clientX: x,
  //     clientY: y
  // }, {  identifier: Date.now() + i,
  //     pageX: point.x,
  //     pageY: point.y,
  //     screenX: point.x,
  //     screenY: point.y,
  //     clientX: point.x,
  //     clientY: point.y}]
  // return event;
// }

function sendTouchEvent(touch, ts, element, eventType) {
  const touchEvent = new dom.window.TouchEvent(eventType, {
    cancelable: true,
    bubbles: true,
    touches: [touch],
    targetTouches: [],
    changedTouches: [touch],
    shiftKey: false,
    isTrusted: true,
    timeStamp: ts,
  });

  element.dispatchEvent(touchEvent);
}

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}

// setTimeout(function() {
//   // debugger;
//   // Array.from(dom.window.document.getElementsByTagName('iframe')[0].contentDocument.getElementsByTagName('iframe')).filter(function(e) { return e.ontouchstart })[0];
//   var capdoc = dom.window.document.getElementsByTagName('iframe')[0].contentDocument;
//   var capframe = Array.from(capdoc.body.getElementsByTagName('iframe')).filter(function(e) { return e.ontouchstart })[0]

//   logDebug(Array.from(capdoc.body.getElementsByTagName('iframe')).map(function(e) { return [e.contentDocument, e.contentWindow]; }))
//   Deno.exit(0)

//   /*
//   touch event
//   {
//     "isTrusted": true,
//     "altKey": false,
//     "metaKey": false,
//     "ctrlKey": false,
//     "shiftKey": false,
//     "detail": 0,
//     "which": 0,
//     "NONE": 0,
//     "CAPTURING_PHASE": 1,
//     "AT_TARGET": 2,
//     "BUBBLING_PHASE": 3,
//     "type": "touchcancel",
//     "eventPhase": 2,
//     "bubbles": true,
//     "cancelable": false,
//     "defaultPrevented": false,
//     "composed": true,
//     "timeStamp": 10961.239999999634,
//     "returnValue": true,
//     "cancelBubble": false,
//     "touches": [
//       {"identifier":0,"screenX":465.2109375,"screenY":544.1484375,"clientX":93.2109375,"clientY":47.6484375,"pageX":93.2109375,"pageY":47.6484375,"radiusX":11.5,"radiusY":11.5,"rotationAngle":0,"force":1}
//     ]
//   }
//   */
//   var peo = {
//     "isTrusted": true,
//     "pointerId": 2,
//     "width": 23,
//     "height": 23,
//     "pressure": 1,
//     "tiltX": 0,
//     "tiltY": 0,
//     "tangentialPressure": 0,
//     "twist": 0,
//     "pointerType": "touch",
//     "isPrimary": true,
//     "screenX": 467.7265625,
//     "screenY": 557.29296875,
//     "clientX": 95.7265625,
//     "clientY": 60.79296875,
//     "ctrlKey": false,
//     "shiftKey": false,
//     "altKey": false,
//     "metaKey": false,
//     "button": 0,
//     "buttons": 1,
//     "pageX": 95.7265625,
//     "pageY": 60.79296875,
//     "x": 95.7265625,
//     "y": 60.79296875,
//     "offsetX": 95.7265625,
//     "offsetY": 60.79296875,
//     "movementX": 0,
//     "movementY": 0,
//     "layerX": 95,
//     "layerY": 60,
//     "detail": 0,
//     "which": 1,
//     "NONE": 0,
//     "CAPTURING_PHASE": 1,
//     "AT_TARGET": 2,
//     "BUBBLING_PHASE": 3,
//     "type": "pointerdown",
//     "eventPhase": 0,
//     "bubbles": true,
//     "cancelable": true,
//     "defaultPrevented": false,
//     "composed": true,
//     "timeStamp": 6103.1599999996615,
//     "returnValue": true,
//     "cancelBubble": false
//   };
//   logDebug(capframe, capframe.contentDocument, Object.keys(capframe))
//   var pe = capframe.contentDocument.createEvent('event');
//   pe.initEvent('pointerdown', true, true);
//   for (var i in peo) {
//     try{
//       pe[i] = peo[i];
//     }catch(e){
//       console.error(e);
//     }
//   }
//   logDebug('pll', pointerListeners.length)
//   pointerListeners.forEach(function(pl) {
//     pl(pe);
//   });
//   capframe.contentDocument.body.dispatchEvent(pe);
//   debugger;
//   // pe.target = dom.window.document.body;
//   // debugger;
//   // pe.target = capframe.contentDocument.body;

//   // pe.target.dispatchEvent(pe);


//   // var pe = new dom.window.Event('pointerdown', peo);

//   // var touch = {"identifier":0,"screenX":465.2109375,"screenY":544.1484375,"clientX":93.2109375,"clientY":47.6484375,"pageX":93.2109375,"pageY":47.6484375,"radiusX":11.5,"radiusY":11.5,"rotationAngle":0,"force":1};
//   // sendTouchEvent(touch, pe.timeStamp + 5000, capframe, 'touchcancel');
//   // debugger;
//   // capframe.ontouchstart({type: 'touchstart', target: capframe});

// }, 5000);

async function main_bak() {
  // setupTrace();
  setTimeout(px, 0);
  var pxJs2 = 'https://www.hibbett.com' + await scriptPromise;
  logDebug({pxJs2})
  var px2 = await importWrappedUrl(pxJs2, jsReferer, jsCkie);
  // const text = decoder.decode(await Deno.readFile('./iz-client.js'));
  // var px2 = await importUnwrapped(text);
  // debugger;
  px2();
}

