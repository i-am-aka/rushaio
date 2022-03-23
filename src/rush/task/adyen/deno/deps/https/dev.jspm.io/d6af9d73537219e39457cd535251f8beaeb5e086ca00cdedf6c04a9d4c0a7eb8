import { dew as _stringifyDewDew } from "./stringify.dew.js";
import { dew as _parseDewDew } from "./parse.dew.js";
import { dew as _formatsDewDew } from "./formats.dew.js";
var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  var stringify = _stringifyDewDew();

  var parse = _parseDewDew();

  var formats = _formatsDewDew();

  exports = {
    formats: formats,
    parse: parse,
    stringify: stringify
  };
  return exports;
}