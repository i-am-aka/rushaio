// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/**
 * Determines whether an object has a property with the specified name.
 * Avoid calling prototype builtin `hasOwnProperty` for two reasons:
 *
 * 1. `hasOwnProperty` is defined on the object as something else:
 *
 *      const options = {
 *        ending: 'utf8',
 *        hasOwnProperty: 'foo'
 *      };
 *      options.hasOwnProperty('ending') // throws a TypeError
 *
 * 2. The object doesn't inherit from `Object.prototype`:
 *
 *       const options = Object.create(null);
 *       options.ending = 'utf8';
 *       options.hasOwnProperty('ending'); // throws a TypeError
 *
 * @param obj A Object.
 * @param v A property name.
 * @see https://eslint.org/docs/rules/no-prototype-builtins
 */ export function hasOwnProperty(obj, v) {
    if (obj == null) {
        return false;
    }
    return Object.prototype.hasOwnProperty.call(obj, v);
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL191dGlsL2hhc19vd25fcHJvcGVydHkudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCAyMDE4LTIwMjAgdGhlIERlbm8gYXV0aG9ycy4gQWxsIHJpZ2h0cyByZXNlcnZlZC4gTUlUIGxpY2Vuc2UuXG5cbi8qKlxuICogRGV0ZXJtaW5lcyB3aGV0aGVyIGFuIG9iamVjdCBoYXMgYSBwcm9wZXJ0eSB3aXRoIHRoZSBzcGVjaWZpZWQgbmFtZS5cbiAqIEF2b2lkIGNhbGxpbmcgcHJvdG90eXBlIGJ1aWx0aW4gYGhhc093blByb3BlcnR5YCBmb3IgdHdvIHJlYXNvbnM6XG4gKlxuICogMS4gYGhhc093blByb3BlcnR5YCBpcyBkZWZpbmVkIG9uIHRoZSBvYmplY3QgYXMgc29tZXRoaW5nIGVsc2U6XG4gKlxuICogICAgICBjb25zdCBvcHRpb25zID0ge1xuICogICAgICAgIGVuZGluZzogJ3V0ZjgnLFxuICogICAgICAgIGhhc093blByb3BlcnR5OiAnZm9vJ1xuICogICAgICB9O1xuICogICAgICBvcHRpb25zLmhhc093blByb3BlcnR5KCdlbmRpbmcnKSAvLyB0aHJvd3MgYSBUeXBlRXJyb3JcbiAqXG4gKiAyLiBUaGUgb2JqZWN0IGRvZXNuJ3QgaW5oZXJpdCBmcm9tIGBPYmplY3QucHJvdG90eXBlYDpcbiAqXG4gKiAgICAgICBjb25zdCBvcHRpb25zID0gT2JqZWN0LmNyZWF0ZShudWxsKTtcbiAqICAgICAgIG9wdGlvbnMuZW5kaW5nID0gJ3V0ZjgnO1xuICogICAgICAgb3B0aW9ucy5oYXNPd25Qcm9wZXJ0eSgnZW5kaW5nJyk7IC8vIHRocm93cyBhIFR5cGVFcnJvclxuICpcbiAqIEBwYXJhbSBvYmogQSBPYmplY3QuXG4gKiBAcGFyYW0gdiBBIHByb3BlcnR5IG5hbWUuXG4gKiBAc2VlIGh0dHBzOi8vZXNsaW50Lm9yZy9kb2NzL3J1bGVzL25vLXByb3RvdHlwZS1idWlsdGluc1xuICovXG5leHBvcnQgZnVuY3Rpb24gaGFzT3duUHJvcGVydHk8VD4ob2JqOiBULCB2OiBQcm9wZXJ0eUtleSk6IGJvb2xlYW4ge1xuICBpZiAob2JqID09IG51bGwpIHtcbiAgICByZXR1cm4gZmFsc2U7XG4gIH1cbiAgcmV0dXJuIE9iamVjdC5wcm90b3R5cGUuaGFzT3duUHJvcGVydHkuY2FsbChvYmosIHYpO1xufVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLEVBQUEsd0VBQUE7QUFFQSxFQXFCQSxBQXJCQSw0ckJBcUJBLEFBckJBLEVBcUJBLGlCQUNBLGNBQUEsQ0FBQSxHQUFBLEVBQUEsQ0FBQTtRQUNBLEdBQUEsSUFBQSxJQUFBO2VBQ0EsS0FBQTs7V0FFQSxNQUFBLENBQUEsU0FBQSxDQUFBLGNBQUEsQ0FBQSxJQUFBLENBQUEsR0FBQSxFQUFBLENBQUEifQ==