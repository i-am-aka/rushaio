var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;
  var replace = String.prototype.replace;
  var percentTwenties = /%20/g;
  exports = {
    'default': 'RFC3986',
    formatters: {
      RFC1738: function (value) {
        return replace.call(value, percentTwenties, '+');
      },
      RFC3986: function (value) {
        return value;
      }
    },
    RFC1738: 'RFC1738',
    RFC3986: 'RFC3986'
  };
  return exports;
}