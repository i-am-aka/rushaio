import { dew as _SymbolDewDew } from "./_Symbol.dew.js";
import { dew as _getRawTagDewDew } from "./_getRawTag.dew.js";
import { dew as _objectToStringDewDew } from "./_objectToString.dew.js";
var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  var Symbol = _SymbolDewDew(),
      getRawTag = _getRawTagDewDew(),
      objectToString = _objectToStringDewDew();
  /** `Object#toString` result references. */


  var nullTag = '[object Null]',
      undefinedTag = '[object Undefined]';
  /** Built-in value references. */

  var symToStringTag = Symbol ? Symbol.toStringTag : undefined;
  /**
   * The base implementation of `getTag` without fallbacks for buggy environments.
   *
   * @private
   * @param {*} value The value to query.
   * @returns {string} Returns the `toStringTag`.
   */

  function baseGetTag(value) {
    if (value == null) {
      return value === undefined ? undefinedTag : nullTag;
    }

    return symToStringTag && symToStringTag in Object(value) ? getRawTag(value) : objectToString(value);
  }

  exports = baseGetTag;
  return exports;
}