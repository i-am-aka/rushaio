var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  /*
   * "A request-path path-matches a given cookie-path if at least one of the
   * following conditions holds:"
   */
  function pathMatch(reqPath, cookiePath) {
    // "o  The cookie-path and the request-path are identical."
    if (cookiePath === reqPath) {
      return true;
    }

    var idx = reqPath.indexOf(cookiePath);

    if (idx === 0) {
      // "o  The cookie-path is a prefix of the request-path, and the last
      // character of the cookie-path is %x2F ("/")."
      if (cookiePath.substr(-1) === "/") {
        return true;
      } // " o  The cookie-path is a prefix of the request-path, and the first
      // character of the request-path that is not included in the cookie- path
      // is a %x2F ("/") character."


      if (reqPath.substr(cookiePath.length, 1) === "/") {
        return true;
      }
    }

    return false;
  }

  exports.pathMatch = pathMatch;
  return exports;
}