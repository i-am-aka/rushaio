import { dew as _pubsuffixPslDewDew } from "./pubsuffix-psl.dew.js";
var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  var pubsuffix = _pubsuffixPslDewDew(); // Gives the permutation of all possible domainMatch()es of a given domain. The
  // array is in shortest-to-longest order.  Handy for indexing.


  function permuteDomain(domain) {
    var pubSuf = pubsuffix.getPublicSuffix(domain);

    if (!pubSuf) {
      return null;
    }

    if (pubSuf == domain) {
      return [domain];
    }

    var prefix = domain.slice(0, -(pubSuf.length + 1)); // ".example.com"

    var parts = prefix.split('.').reverse();
    var cur = pubSuf;
    var permutations = [cur];

    while (parts.length) {
      cur = parts.shift() + '.' + cur;
      permutations.push(cur);
    }

    return permutations;
  }

  exports.permuteDomain = permuteDomain;
  return exports;
}