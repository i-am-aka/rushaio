import { dew as _baseGetTagDewDew } from "./_baseGetTag.dew.js";
import { dew as _isArrayDewDew } from "./isArray.dew.js";
import { dew as _isObjectLikeDewDew } from "./isObjectLike.dew.js";
var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  var baseGetTag = _baseGetTagDewDew(),
      isArray = _isArrayDewDew(),
      isObjectLike = _isObjectLikeDewDew();
  /** `Object#toString` result references. */


  var stringTag = '[object String]';
  /**
   * Checks if `value` is classified as a `String` primitive or object.
   *
   * @static
   * @since 0.1.0
   * @memberOf _
   * @category Lang
   * @param {*} value The value to check.
   * @returns {boolean} Returns `true` if `value` is a string, else `false`.
   * @example
   *
   * _.isString('abc');
   * // => true
   *
   * _.isString(1);
   * // => false
   */

  function isString(value) {
    return typeof value == 'string' || !isArray(value) && isObjectLike(value) && baseGetTag(value) == stringTag;
  }

  exports = isString;
  return exports;
}