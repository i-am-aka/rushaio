// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/** Find first index of binary pattern from a. If not found, then return -1
 * @param source source array
 * @param pat pattern to find in source array
 */ export function findIndex(source, pat) {
    const s = pat[0];
    for(let i = 0; i < source.length; i++){
        if (source[i] !== s) continue;
        const pin = i;
        let matched = 1;
        let j = i;
        while(matched < pat.length){
            j++;
            if (source[j] !== pat[j - pin]) {
                break;
            }
            matched++;
        }
        if (matched === pat.length) {
            return pin;
        }
    }
    return -1;
}
/** Find last index of binary pattern from a. If not found, then return -1.
 * @param source source array
 * @param pat pattern to find in source array
 */ export function findLastIndex(source, pat) {
    const e = pat[pat.length - 1];
    for(let i = source.length - 1; i >= 0; i--){
        if (source[i] !== e) continue;
        const pin = i;
        let matched = 1;
        let j = i;
        while(matched < pat.length){
            j--;
            if (source[j] !== pat[pat.length - 1 - (pin - j)]) {
                break;
            }
            matched++;
        }
        if (matched === pat.length) {
            return pin - pat.length + 1;
        }
    }
    return -1;
}
/** Check whether binary arrays are equal to each other.
 * @param source first array to check equality
 * @param match second array to check equality
 */ export function equal(source, match) {
    if (source.length !== match.length) return false;
    for(let i = 0; i < match.length; i++){
        if (source[i] !== match[i]) return false;
    }
    return true;
}
/** Check whether binary array starts with prefix.
 * @param source srouce array
 * @param prefix prefix array to check in source
 */ export function hasPrefix(source, prefix) {
    for(let i = 0, max = prefix.length; i < max; i++){
        if (source[i] !== prefix[i]) return false;
    }
    return true;
}
/** Check whether binary array ends with suffix.
 * @param source source array
 * @param suffix suffix array to check in source
 */ export function hasSuffix(source, suffix) {
    for(let srci = source.length - 1, sfxi = suffix.length - 1; sfxi >= 0; srci--, sfxi--){
        if (source[srci] !== suffix[sfxi]) return false;
    }
    return true;
}
/** Repeat bytes. returns a new byte slice consisting of `count` copies of `b`.
 * @param origin The origin bytes
 * @param count The count you want to repeat.
 */ export function repeat(origin, count) {
    if (count === 0) {
        return new Uint8Array();
    }
    if (count < 0) {
        throw new Error("bytes: negative repeat count");
    } else if (origin.length * count / count !== origin.length) {
        throw new Error("bytes: repeat count causes overflow");
    }
    const int = Math.floor(count);
    if (int !== count) {
        throw new Error("bytes: repeat count must be an integer");
    }
    const nb = new Uint8Array(origin.length * count);
    let bp = copyBytes(origin, nb);
    for(; bp < nb.length; bp *= 2){
        copyBytes(nb.slice(0, bp), nb, bp);
    }
    return nb;
}
/** Concatenate two binary arrays and return new one.
 * @param origin origin array to concatenate
 * @param b array to concatenate with origin
 */ export function concat(origin, b) {
    const output = new Uint8Array(origin.length + b.length);
    output.set(origin, 0);
    output.set(b, origin.length);
    return output;
}
/** Check source array contains pattern array.
 * @param source source array
 * @param pat patter array
 */ export function contains(source, pat) {
    return findIndex(source, pat) != -1;
}
/**
 * Copy bytes from one Uint8Array to another.  Bytes from `src` which don't fit
 * into `dst` will not be copied.
 *
 * @param src Source byte array
 * @param dst Destination byte array
 * @param off Offset into `dst` at which to begin writing values from `src`.
 * @return number of bytes copied
 */ export function copyBytes(src, dst, off = 0) {
    off = Math.max(0, Math.min(off, dst.byteLength));
    const dstBytesAvailable = dst.byteLength - off;
    if (src.byteLength > dstBytesAvailable) {
        src = src.subarray(0, dstBytesAvailable);
    }
    dst.set(src, off);
    return src.byteLength;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL2J5dGVzL21vZC50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cblxuLyoqIEZpbmQgZmlyc3QgaW5kZXggb2YgYmluYXJ5IHBhdHRlcm4gZnJvbSBhLiBJZiBub3QgZm91bmQsIHRoZW4gcmV0dXJuIC0xXG4gKiBAcGFyYW0gc291cmNlIHNvdXJjZSBhcnJheVxuICogQHBhcmFtIHBhdCBwYXR0ZXJuIHRvIGZpbmQgaW4gc291cmNlIGFycmF5XG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBmaW5kSW5kZXgoc291cmNlOiBVaW50OEFycmF5LCBwYXQ6IFVpbnQ4QXJyYXkpOiBudW1iZXIge1xuICBjb25zdCBzID0gcGF0WzBdO1xuICBmb3IgKGxldCBpID0gMDsgaSA8IHNvdXJjZS5sZW5ndGg7IGkrKykge1xuICAgIGlmIChzb3VyY2VbaV0gIT09IHMpIGNvbnRpbnVlO1xuICAgIGNvbnN0IHBpbiA9IGk7XG4gICAgbGV0IG1hdGNoZWQgPSAxO1xuICAgIGxldCBqID0gaTtcbiAgICB3aGlsZSAobWF0Y2hlZCA8IHBhdC5sZW5ndGgpIHtcbiAgICAgIGorKztcbiAgICAgIGlmIChzb3VyY2Vbal0gIT09IHBhdFtqIC0gcGluXSkge1xuICAgICAgICBicmVhaztcbiAgICAgIH1cbiAgICAgIG1hdGNoZWQrKztcbiAgICB9XG4gICAgaWYgKG1hdGNoZWQgPT09IHBhdC5sZW5ndGgpIHtcbiAgICAgIHJldHVybiBwaW47XG4gICAgfVxuICB9XG4gIHJldHVybiAtMTtcbn1cblxuLyoqIEZpbmQgbGFzdCBpbmRleCBvZiBiaW5hcnkgcGF0dGVybiBmcm9tIGEuIElmIG5vdCBmb3VuZCwgdGhlbiByZXR1cm4gLTEuXG4gKiBAcGFyYW0gc291cmNlIHNvdXJjZSBhcnJheVxuICogQHBhcmFtIHBhdCBwYXR0ZXJuIHRvIGZpbmQgaW4gc291cmNlIGFycmF5XG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBmaW5kTGFzdEluZGV4KHNvdXJjZTogVWludDhBcnJheSwgcGF0OiBVaW50OEFycmF5KTogbnVtYmVyIHtcbiAgY29uc3QgZSA9IHBhdFtwYXQubGVuZ3RoIC0gMV07XG4gIGZvciAobGV0IGkgPSBzb3VyY2UubGVuZ3RoIC0gMTsgaSA+PSAwOyBpLS0pIHtcbiAgICBpZiAoc291cmNlW2ldICE9PSBlKSBjb250aW51ZTtcbiAgICBjb25zdCBwaW4gPSBpO1xuICAgIGxldCBtYXRjaGVkID0gMTtcbiAgICBsZXQgaiA9IGk7XG4gICAgd2hpbGUgKG1hdGNoZWQgPCBwYXQubGVuZ3RoKSB7XG4gICAgICBqLS07XG4gICAgICBpZiAoc291cmNlW2pdICE9PSBwYXRbcGF0Lmxlbmd0aCAtIDEgLSAocGluIC0gaildKSB7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuICAgICAgbWF0Y2hlZCsrO1xuICAgIH1cbiAgICBpZiAobWF0Y2hlZCA9PT0gcGF0Lmxlbmd0aCkge1xuICAgICAgcmV0dXJuIHBpbiAtIHBhdC5sZW5ndGggKyAxO1xuICAgIH1cbiAgfVxuICByZXR1cm4gLTE7XG59XG5cbi8qKiBDaGVjayB3aGV0aGVyIGJpbmFyeSBhcnJheXMgYXJlIGVxdWFsIHRvIGVhY2ggb3RoZXIuXG4gKiBAcGFyYW0gc291cmNlIGZpcnN0IGFycmF5IHRvIGNoZWNrIGVxdWFsaXR5XG4gKiBAcGFyYW0gbWF0Y2ggc2Vjb25kIGFycmF5IHRvIGNoZWNrIGVxdWFsaXR5XG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBlcXVhbChzb3VyY2U6IFVpbnQ4QXJyYXksIG1hdGNoOiBVaW50OEFycmF5KTogYm9vbGVhbiB7XG4gIGlmIChzb3VyY2UubGVuZ3RoICE9PSBtYXRjaC5sZW5ndGgpIHJldHVybiBmYWxzZTtcbiAgZm9yIChsZXQgaSA9IDA7IGkgPCBtYXRjaC5sZW5ndGg7IGkrKykge1xuICAgIGlmIChzb3VyY2VbaV0gIT09IG1hdGNoW2ldKSByZXR1cm4gZmFsc2U7XG4gIH1cbiAgcmV0dXJuIHRydWU7XG59XG5cbi8qKiBDaGVjayB3aGV0aGVyIGJpbmFyeSBhcnJheSBzdGFydHMgd2l0aCBwcmVmaXguXG4gKiBAcGFyYW0gc291cmNlIHNyb3VjZSBhcnJheVxuICogQHBhcmFtIHByZWZpeCBwcmVmaXggYXJyYXkgdG8gY2hlY2sgaW4gc291cmNlXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBoYXNQcmVmaXgoc291cmNlOiBVaW50OEFycmF5LCBwcmVmaXg6IFVpbnQ4QXJyYXkpOiBib29sZWFuIHtcbiAgZm9yIChsZXQgaSA9IDAsIG1heCA9IHByZWZpeC5sZW5ndGg7IGkgPCBtYXg7IGkrKykge1xuICAgIGlmIChzb3VyY2VbaV0gIT09IHByZWZpeFtpXSkgcmV0dXJuIGZhbHNlO1xuICB9XG4gIHJldHVybiB0cnVlO1xufVxuXG4vKiogQ2hlY2sgd2hldGhlciBiaW5hcnkgYXJyYXkgZW5kcyB3aXRoIHN1ZmZpeC5cbiAqIEBwYXJhbSBzb3VyY2Ugc291cmNlIGFycmF5XG4gKiBAcGFyYW0gc3VmZml4IHN1ZmZpeCBhcnJheSB0byBjaGVjayBpbiBzb3VyY2VcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIGhhc1N1ZmZpeChzb3VyY2U6IFVpbnQ4QXJyYXksIHN1ZmZpeDogVWludDhBcnJheSk6IGJvb2xlYW4ge1xuICBmb3IgKFxuICAgIGxldCBzcmNpID0gc291cmNlLmxlbmd0aCAtIDEsIHNmeGkgPSBzdWZmaXgubGVuZ3RoIC0gMTtcbiAgICBzZnhpID49IDA7XG4gICAgc3JjaS0tLCBzZnhpLS1cbiAgKSB7XG4gICAgaWYgKHNvdXJjZVtzcmNpXSAhPT0gc3VmZml4W3NmeGldKSByZXR1cm4gZmFsc2U7XG4gIH1cbiAgcmV0dXJuIHRydWU7XG59XG5cbi8qKiBSZXBlYXQgYnl0ZXMuIHJldHVybnMgYSBuZXcgYnl0ZSBzbGljZSBjb25zaXN0aW5nIG9mIGBjb3VudGAgY29waWVzIG9mIGBiYC5cbiAqIEBwYXJhbSBvcmlnaW4gVGhlIG9yaWdpbiBieXRlc1xuICogQHBhcmFtIGNvdW50IFRoZSBjb3VudCB5b3Ugd2FudCB0byByZXBlYXQuXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiByZXBlYXQob3JpZ2luOiBVaW50OEFycmF5LCBjb3VudDogbnVtYmVyKTogVWludDhBcnJheSB7XG4gIGlmIChjb3VudCA9PT0gMCkge1xuICAgIHJldHVybiBuZXcgVWludDhBcnJheSgpO1xuICB9XG5cbiAgaWYgKGNvdW50IDwgMCkge1xuICAgIHRocm93IG5ldyBFcnJvcihcImJ5dGVzOiBuZWdhdGl2ZSByZXBlYXQgY291bnRcIik7XG4gIH0gZWxzZSBpZiAoKG9yaWdpbi5sZW5ndGggKiBjb3VudCkgLyBjb3VudCAhPT0gb3JpZ2luLmxlbmd0aCkge1xuICAgIHRocm93IG5ldyBFcnJvcihcImJ5dGVzOiByZXBlYXQgY291bnQgY2F1c2VzIG92ZXJmbG93XCIpO1xuICB9XG5cbiAgY29uc3QgaW50ID0gTWF0aC5mbG9vcihjb3VudCk7XG5cbiAgaWYgKGludCAhPT0gY291bnQpIHtcbiAgICB0aHJvdyBuZXcgRXJyb3IoXCJieXRlczogcmVwZWF0IGNvdW50IG11c3QgYmUgYW4gaW50ZWdlclwiKTtcbiAgfVxuXG4gIGNvbnN0IG5iID0gbmV3IFVpbnQ4QXJyYXkob3JpZ2luLmxlbmd0aCAqIGNvdW50KTtcblxuICBsZXQgYnAgPSBjb3B5Qnl0ZXMob3JpZ2luLCBuYik7XG5cbiAgZm9yICg7IGJwIDwgbmIubGVuZ3RoOyBicCAqPSAyKSB7XG4gICAgY29weUJ5dGVzKG5iLnNsaWNlKDAsIGJwKSwgbmIsIGJwKTtcbiAgfVxuXG4gIHJldHVybiBuYjtcbn1cblxuLyoqIENvbmNhdGVuYXRlIHR3byBiaW5hcnkgYXJyYXlzIGFuZCByZXR1cm4gbmV3IG9uZS5cbiAqIEBwYXJhbSBvcmlnaW4gb3JpZ2luIGFycmF5IHRvIGNvbmNhdGVuYXRlXG4gKiBAcGFyYW0gYiBhcnJheSB0byBjb25jYXRlbmF0ZSB3aXRoIG9yaWdpblxuICovXG5leHBvcnQgZnVuY3Rpb24gY29uY2F0KG9yaWdpbjogVWludDhBcnJheSwgYjogVWludDhBcnJheSk6IFVpbnQ4QXJyYXkge1xuICBjb25zdCBvdXRwdXQgPSBuZXcgVWludDhBcnJheShvcmlnaW4ubGVuZ3RoICsgYi5sZW5ndGgpO1xuICBvdXRwdXQuc2V0KG9yaWdpbiwgMCk7XG4gIG91dHB1dC5zZXQoYiwgb3JpZ2luLmxlbmd0aCk7XG4gIHJldHVybiBvdXRwdXQ7XG59XG5cbi8qKiBDaGVjayBzb3VyY2UgYXJyYXkgY29udGFpbnMgcGF0dGVybiBhcnJheS5cbiAqIEBwYXJhbSBzb3VyY2Ugc291cmNlIGFycmF5XG4gKiBAcGFyYW0gcGF0IHBhdHRlciBhcnJheVxuICovXG5leHBvcnQgZnVuY3Rpb24gY29udGFpbnMoc291cmNlOiBVaW50OEFycmF5LCBwYXQ6IFVpbnQ4QXJyYXkpOiBib29sZWFuIHtcbiAgcmV0dXJuIGZpbmRJbmRleChzb3VyY2UsIHBhdCkgIT0gLTE7XG59XG5cbi8qKlxuICogQ29weSBieXRlcyBmcm9tIG9uZSBVaW50OEFycmF5IHRvIGFub3RoZXIuICBCeXRlcyBmcm9tIGBzcmNgIHdoaWNoIGRvbid0IGZpdFxuICogaW50byBgZHN0YCB3aWxsIG5vdCBiZSBjb3BpZWQuXG4gKlxuICogQHBhcmFtIHNyYyBTb3VyY2UgYnl0ZSBhcnJheVxuICogQHBhcmFtIGRzdCBEZXN0aW5hdGlvbiBieXRlIGFycmF5XG4gKiBAcGFyYW0gb2ZmIE9mZnNldCBpbnRvIGBkc3RgIGF0IHdoaWNoIHRvIGJlZ2luIHdyaXRpbmcgdmFsdWVzIGZyb20gYHNyY2AuXG4gKiBAcmV0dXJuIG51bWJlciBvZiBieXRlcyBjb3BpZWRcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIGNvcHlCeXRlcyhzcmM6IFVpbnQ4QXJyYXksIGRzdDogVWludDhBcnJheSwgb2ZmID0gMCk6IG51bWJlciB7XG4gIG9mZiA9IE1hdGgubWF4KDAsIE1hdGgubWluKG9mZiwgZHN0LmJ5dGVMZW5ndGgpKTtcbiAgY29uc3QgZHN0Qnl0ZXNBdmFpbGFibGUgPSBkc3QuYnl0ZUxlbmd0aCAtIG9mZjtcbiAgaWYgKHNyYy5ieXRlTGVuZ3RoID4gZHN0Qnl0ZXNBdmFpbGFibGUpIHtcbiAgICBzcmMgPSBzcmMuc3ViYXJyYXkoMCwgZHN0Qnl0ZXNBdmFpbGFibGUpO1xuICB9XG4gIGRzdC5zZXQoc3JjLCBvZmYpO1xuICByZXR1cm4gc3JjLmJ5dGVMZW5ndGg7XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtBQUVBLEVBR0EsQUFIQSx1SkFHQSxBQUhBLEVBR0EsaUJBQ0EsU0FBQSxDQUFBLE1BQUEsRUFBQSxHQUFBO1VBQ0EsQ0FBQSxHQUFBLEdBQUEsQ0FBQSxDQUFBO1lBQ0EsQ0FBQSxHQUFBLENBQUEsRUFBQSxDQUFBLEdBQUEsTUFBQSxDQUFBLE1BQUEsRUFBQSxDQUFBO1lBQ0EsTUFBQSxDQUFBLENBQUEsTUFBQSxDQUFBO2NBQ0EsR0FBQSxHQUFBLENBQUE7WUFDQSxPQUFBLEdBQUEsQ0FBQTtZQUNBLENBQUEsR0FBQSxDQUFBO2NBQ0EsT0FBQSxHQUFBLEdBQUEsQ0FBQSxNQUFBO0FBQ0EsYUFBQTtnQkFDQSxNQUFBLENBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxDQUFBLEdBQUEsR0FBQTs7O0FBR0EsbUJBQUE7O1lBRUEsT0FBQSxLQUFBLEdBQUEsQ0FBQSxNQUFBO21CQUNBLEdBQUE7OztZQUdBLENBQUE7O0FBR0EsRUFHQSxBQUhBLHVKQUdBLEFBSEEsRUFHQSxpQkFDQSxhQUFBLENBQUEsTUFBQSxFQUFBLEdBQUE7VUFDQSxDQUFBLEdBQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQTtZQUNBLENBQUEsR0FBQSxNQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsRUFBQSxDQUFBLElBQUEsQ0FBQSxFQUFBLENBQUE7WUFDQSxNQUFBLENBQUEsQ0FBQSxNQUFBLENBQUE7Y0FDQSxHQUFBLEdBQUEsQ0FBQTtZQUNBLE9BQUEsR0FBQSxDQUFBO1lBQ0EsQ0FBQSxHQUFBLENBQUE7Y0FDQSxPQUFBLEdBQUEsR0FBQSxDQUFBLE1BQUE7QUFDQSxhQUFBO2dCQUNBLE1BQUEsQ0FBQSxDQUFBLE1BQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxJQUFBLEdBQUEsR0FBQSxDQUFBOzs7QUFHQSxtQkFBQTs7WUFFQSxPQUFBLEtBQUEsR0FBQSxDQUFBLE1BQUE7bUJBQ0EsR0FBQSxHQUFBLEdBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQTs7O1lBR0EsQ0FBQTs7QUFHQSxFQUdBLEFBSEEsc0pBR0EsQUFIQSxFQUdBLGlCQUNBLEtBQUEsQ0FBQSxNQUFBLEVBQUEsS0FBQTtRQUNBLE1BQUEsQ0FBQSxNQUFBLEtBQUEsS0FBQSxDQUFBLE1BQUEsU0FBQSxLQUFBO1lBQ0EsQ0FBQSxHQUFBLENBQUEsRUFBQSxDQUFBLEdBQUEsS0FBQSxDQUFBLE1BQUEsRUFBQSxDQUFBO1lBQ0EsTUFBQSxDQUFBLENBQUEsTUFBQSxLQUFBLENBQUEsQ0FBQSxVQUFBLEtBQUE7O1dBRUEsSUFBQTs7QUFHQSxFQUdBLEFBSEEsaUlBR0EsQUFIQSxFQUdBLGlCQUNBLFNBQUEsQ0FBQSxNQUFBLEVBQUEsTUFBQTtZQUNBLENBQUEsR0FBQSxDQUFBLEVBQUEsR0FBQSxHQUFBLE1BQUEsQ0FBQSxNQUFBLEVBQUEsQ0FBQSxHQUFBLEdBQUEsRUFBQSxDQUFBO1lBQ0EsTUFBQSxDQUFBLENBQUEsTUFBQSxNQUFBLENBQUEsQ0FBQSxVQUFBLEtBQUE7O1dBRUEsSUFBQTs7QUFHQSxFQUdBLEFBSEEsK0hBR0EsQUFIQSxFQUdBLGlCQUNBLFNBQUEsQ0FBQSxNQUFBLEVBQUEsTUFBQTtZQUVBLElBQUEsR0FBQSxNQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsRUFBQSxJQUFBLEdBQUEsTUFBQSxDQUFBLE1BQUEsR0FBQSxDQUFBLEVBQ0EsSUFBQSxJQUFBLENBQUEsRUFDQSxJQUFBLElBQUEsSUFBQTtZQUVBLE1BQUEsQ0FBQSxJQUFBLE1BQUEsTUFBQSxDQUFBLElBQUEsVUFBQSxLQUFBOztXQUVBLElBQUE7O0FBR0EsRUFHQSxBQUhBLCtKQUdBLEFBSEEsRUFHQSxpQkFDQSxNQUFBLENBQUEsTUFBQSxFQUFBLEtBQUE7UUFDQSxLQUFBLEtBQUEsQ0FBQTttQkFDQSxVQUFBOztRQUdBLEtBQUEsR0FBQSxDQUFBO2tCQUNBLEtBQUEsRUFBQSw0QkFBQTtlQUNBLE1BQUEsQ0FBQSxNQUFBLEdBQUEsS0FBQSxHQUFBLEtBQUEsS0FBQSxNQUFBLENBQUEsTUFBQTtrQkFDQSxLQUFBLEVBQUEsbUNBQUE7O1VBR0EsR0FBQSxHQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsS0FBQTtRQUVBLEdBQUEsS0FBQSxLQUFBO2tCQUNBLEtBQUEsRUFBQSxzQ0FBQTs7VUFHQSxFQUFBLE9BQUEsVUFBQSxDQUFBLE1BQUEsQ0FBQSxNQUFBLEdBQUEsS0FBQTtRQUVBLEVBQUEsR0FBQSxTQUFBLENBQUEsTUFBQSxFQUFBLEVBQUE7VUFFQSxFQUFBLEdBQUEsRUFBQSxDQUFBLE1BQUEsRUFBQSxFQUFBLElBQUEsQ0FBQTtBQUNBLGlCQUFBLENBQUEsRUFBQSxDQUFBLEtBQUEsQ0FBQSxDQUFBLEVBQUEsRUFBQSxHQUFBLEVBQUEsRUFBQSxFQUFBOztXQUdBLEVBQUE7O0FBR0EsRUFHQSxBQUhBLCtJQUdBLEFBSEEsRUFHQSxpQkFDQSxNQUFBLENBQUEsTUFBQSxFQUFBLENBQUE7VUFDQSxNQUFBLE9BQUEsVUFBQSxDQUFBLE1BQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxDQUFBLE1BQUE7QUFDQSxVQUFBLENBQUEsR0FBQSxDQUFBLE1BQUEsRUFBQSxDQUFBO0FBQ0EsVUFBQSxDQUFBLEdBQUEsQ0FBQSxDQUFBLEVBQUEsTUFBQSxDQUFBLE1BQUE7V0FDQSxNQUFBOztBQUdBLEVBR0EsQUFIQSx1R0FHQSxBQUhBLEVBR0EsaUJBQ0EsUUFBQSxDQUFBLE1BQUEsRUFBQSxHQUFBO1dBQ0EsU0FBQSxDQUFBLE1BQUEsRUFBQSxHQUFBLE1BQUEsQ0FBQTs7QUFHQSxFQVFBLEFBUkEsNFNBUUEsQUFSQSxFQVFBLGlCQUNBLFNBQUEsQ0FBQSxHQUFBLEVBQUEsR0FBQSxFQUFBLEdBQUEsR0FBQSxDQUFBO0FBQ0EsT0FBQSxHQUFBLElBQUEsQ0FBQSxHQUFBLENBQUEsQ0FBQSxFQUFBLElBQUEsQ0FBQSxHQUFBLENBQUEsR0FBQSxFQUFBLEdBQUEsQ0FBQSxVQUFBO1VBQ0EsaUJBQUEsR0FBQSxHQUFBLENBQUEsVUFBQSxHQUFBLEdBQUE7UUFDQSxHQUFBLENBQUEsVUFBQSxHQUFBLGlCQUFBO0FBQ0EsV0FBQSxHQUFBLEdBQUEsQ0FBQSxRQUFBLENBQUEsQ0FBQSxFQUFBLGlCQUFBOztBQUVBLE9BQUEsQ0FBQSxHQUFBLENBQUEsR0FBQSxFQUFBLEdBQUE7V0FDQSxHQUFBLENBQUEsVUFBQSJ9