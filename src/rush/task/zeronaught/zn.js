import jsdom from 'https://dev.jspm.io/jsdom';
const dom = new jsdom.JSDOM(``, {
  url: "https://shop.nordstrom.com/s/nike-air-jordan-5-retro-glow-in-the-dark-mid-top-sneaker-women/5368413",
  // referrer: "https://www.yeezysupply.com/delivery",
  contentType: "text/html",
  includeNodeLocations: false,
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
globalThis.navigator = window.navigator;
import Trace from './trace.js';

var { XMLHttpRequest} = dom.window;
var origSetReqHeader = XMLHttpRequest.prototype.setRequestHeader;
var headercnt = 0;
XMLHttpRequest.prototype.setRequestHeader = function() {
  console.log(JSON.stringify([arguments[0], arguments[1]]));
  headercnt++;
  if (headercnt === 6)  {
    Deno.exit(0);
  }
  return origSetReqHeader.apply(this, arguments);
}
// Object.defineProperty(namespaceObject, name, { writable: true, value: functionLogger.getLoggableFunction(potentialFunction, name) });
async function afterReadyCb() {
    // console.log('afterReady')
    // Trace.addLoggingToNamespace(window.document);
    // Trace.addLoggingToNamespace(window.domWindow);
    var xhr = new XMLHttpRequest();
    xhr.open("get", window.domWindow.location.href, true);
    xhr.send("");
}
window.domWindow.addEventListener('afterReady', afterReadyCb, false)

var ES = await import(Deno.args[0]);
ES.default.call(window.domWindow, window.domWindow, window.document, window.domWindow.dispatchEvent);
