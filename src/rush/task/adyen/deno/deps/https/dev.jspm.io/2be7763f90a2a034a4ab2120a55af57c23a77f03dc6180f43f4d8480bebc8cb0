import { dew as _errorsDewDew } from "./errors.dew.js";
import { dew as _typesDewDew } from "./types.dew.js";
import { dew as _readerDewDew } from "./reader.dew.js";
import { dew as _writerDewDew } from "./writer.dew.js";
var exports = {},
    _dewExec = false;
export function dew() {
  if (_dewExec) return exports;
  _dewExec = true;

  // Copyright 2011 Mark Cavage <mcavage@gmail.com> All rights reserved.
  var errors = _errorsDewDew();

  var types = _typesDewDew();

  var Reader = _readerDewDew();

  var Writer = _writerDewDew(); // --- Exports


  exports = {
    Reader: Reader,
    Writer: Writer
  };

  for (var t in types) {
    if (types.hasOwnProperty(t)) exports[t] = types[t];
  }

  for (var e in errors) {
    if (errors.hasOwnProperty(e)) exports[e] = errors[e];
  }

  return exports;
}