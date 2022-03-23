import adyen from './adyen.js';

adyen(window, {}, {});

var enc = window.adyen.createEncryption({
  "enableValidations": true,
  "randomBytes": null,
  "stopDeviceFingerprint": true,
  "numberIgnoreNonNumeric": true,
  "cvcLengthFornumber": {
    "matcher": {}, // todo maybe this matters
    "requiredLength": 4
  }
});
var gentime = (new Date()).toISOString().substring(0,19) + 'Z';

var toenc = [
  { number: Deno.args[0], generationtime: gentime },
  {expiryMonth: Deno.args[1], generationtime: gentime},
  {expiryYear: Deno.args[2], generationtime: gentime},
  {cvc: Deno.args[3], generationtime: gentime},
];
for(var i = 0; i < toenc.length; i++) {
  if (typeof Deno.args[4] !== 'undefined') {
    toenc[i].dfValue = Deno.args[4];
  }
  console.log(enc.encrypt(toenc[i]));
}
