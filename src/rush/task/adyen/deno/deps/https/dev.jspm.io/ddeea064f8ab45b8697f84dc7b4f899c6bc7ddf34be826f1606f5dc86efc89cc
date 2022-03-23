var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  /* eslint-env browser */
  exports = typeof self == 'object' ? self.FormData : window.FormData;
  return exports;
}