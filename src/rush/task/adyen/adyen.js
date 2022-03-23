import jsdom from 'https://dev.jspm.io/jsdom';
const dom = new jsdom.JSDOM(``, {
  url: "https://www.yeezysupply.com/payment",
  referrer: "https://www.yeezysupply.com/delivery",
  contentType: "text/html",
  includeNodeLocations: false,
});
// window.location = dom.window.location;
// window.location.origin = dom.window.location.origin;
window.document = dom.document;
globalThis.document = dom.document;
globalThis.navigator = dom.window.navigator;

// import * as Storage from './storage.ts';
globalThis.sessionStorage = {};
window.sessionStorage = globalThis.sessionStorage;
globalThis.navigator = window.navigator;

import './datastore.js';
const adyenPublicKey = window.DATA_STORE.app.config.adyenPublicKey;
import Runtime from './runtime.js';
Runtime();
import './vendor.app.js';

import initApp from './app.app.js';
initApp(dom.window);
import './vendors~chk-delivery~chk-payment~chk-payment-callback~frontend-c.js';
import Adyen from './df.js';
Adyen(dom.window, dom.window.document);


var enc = window.__LOADABLE_LOADED_CHUNKS__.map(f => f[1]).filter(f => "./node_modules/adyen-cse-js/js/adyen.encrypt.nodom.js" in f)[0]["./node_modules/adyen-cse-js/js/adyen.encrypt.nodom.js"]
var exp = {};
enc.call(exp, {}, exp, function(m) {
  //console.log(`require(${m})`)

})
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
var card = JSON.parse(Deno.args[0]);
var mtime = Deno.args[1];
card.expirationMonth = card.expirationMonth.toString();
card.expirationYear = card.expirationYear.toString();
var cardEnc = exp.adyen.encrypt.createEncryption(adyenPublicKey, {}).encrypt({
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
