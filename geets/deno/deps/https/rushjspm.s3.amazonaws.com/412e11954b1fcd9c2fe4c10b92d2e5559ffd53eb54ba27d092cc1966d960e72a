var exports = {},
    _dewExec = false;

var _global = typeof globalThis !== "undefined" ? globalThis : typeof self !== "undefined" ? self : global;

export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  /** Detect free variable `global` from Node.js. */
  var freeGlobal = typeof _global == 'object' && _global && _global.Object === Object && _global;
  exports = freeGlobal;
  return exports;
}