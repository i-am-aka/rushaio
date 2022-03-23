var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  function isASCIIDigit(c) {
    return c >= 0x30 && c <= 0x39;
  }

  function isASCIIAlpha(c) {
    return c >= 0x41 && c <= 0x5A || c >= 0x61 && c <= 0x7A;
  }

  function isASCIIAlphanumeric(c) {
    return isASCIIAlpha(c) || isASCIIDigit(c);
  }

  function isASCIIHex(c) {
    return isASCIIDigit(c) || c >= 0x41 && c <= 0x46 || c >= 0x61 && c <= 0x66;
  }

  exports = {
    isASCIIDigit,
    isASCIIAlpha,
    isASCIIAlphanumeric,
    isASCIIHex
  };
  return exports;
}