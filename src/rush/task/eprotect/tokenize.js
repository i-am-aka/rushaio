import jsdom from 'https://dev.jspm.io/jsdom';
const dom = new jsdom.JSDOM(``, {
  url: "https://secure.nordstrom.com/checkout",
  contentType: "text/html",
  includeNodeLocations: false,
});
window.document = dom.window.document;
window.domWindow = dom.window;
globalThis.document = dom.window.document;
globalThis.navigator = dom.window.navigator;
globalThis.global = dom.window.document;
import * as Storage from './storage.ts';
globalThis.sessionStorage = new Storage.SessionStorage();
window.sessionStorage = globalThis.sessionStorage;
globalThis.navigator = window.navigator;

// var JQUERY = await import('./jquery-1.12.4.min.js');
// JQUERY.default.call(dom.window, dom.window, dom.window.document);
var EPROTECT = await import('./eprotect.js');
var jQuery = {
    getJSON: function(url, cb) {
        var fn = `jQuery1124043527202544544075_${Date.now()-parseInt(50*Math.random())}`;
        url = url.replace('jsoncallback=?', `jsoncallback=${fn}&_=${Date.now()}`);
        if(url.indexOf('ppstats') !== -1) return;
        fetch(url).then(resp => resp.text()).then(text => {
            var tj = text.substring(fn.length + 1, text.length - 1);
            var json = JSON.parse(tj)
            cb(JSON.parse(tj))
        })
    }
}
EPROTECT.default.call(dom.window, dom.window, dom.window.navigator, jQuery);
const c = (()=>Math.random().toString(36).substring(2, 10) + Math.random().toString(36).substring(2, 10))()
              , l = (()=>Math.random().toString(36).substring(2, 10) + Math.random().toString(36).substring(2, 10))();

debugger;
(new dom.window.eProtect).sendToEprotect({
    paypageId: "dkZyzhtKZ2jpAa9T",
    reportGroup: "*merchant1500",
    orderId: c,
    id: l,
    url: "https://request.eprotect.vantivcnp.com"
}, {
    accountNum: {
        value: Deno.args[0]
    },
    cvv2: {
        value: Deno.args[1]
    }
}, o=>{
    console.log(o.paypageRegistrationId);
    Deno.exit(0);
}, (r) => { console.error(r); Deno.exit(1) }, ()=>{
    console.error('timeout');
    Deno.exit(1);
}, 10e3);
