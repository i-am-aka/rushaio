import { CHAR_DOT, CHAR_FORWARD_SLASH } from "./_constants.ts";
import { assertPath, normalizeString, isPosixPathSeparator, _format } from "./_util.ts";
export const sep = "/";
export const delimiter = ":";
// path.resolve([from ...], to)
export function resolve(...pathSegments) {
    let resolvedPath = "";
    let resolvedAbsolute = false;
    for(let i = pathSegments.length - 1; i >= -1 && !resolvedAbsolute; i--){
        let path;
        if (i >= 0) path = pathSegments[i];
        else {
            if (globalThis.Deno == null) {
                throw new TypeError("Resolved a relative path without a CWD.");
            }
            path = Deno.cwd();
        }
        assertPath(path);
        // Skip empty entries
        if (path.length === 0) {
            continue;
        }
        resolvedPath = `${path}/${resolvedPath}`;
        resolvedAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    }
    // At this point the path should be resolved to a full absolute path, but
    // handle relative paths to be safe (might happen when process.cwd() fails)
    // Normalize the path
    resolvedPath = normalizeString(resolvedPath, !resolvedAbsolute, "/", isPosixPathSeparator);
    if (resolvedAbsolute) {
        if (resolvedPath.length > 0) return `/${resolvedPath}`;
        else return "/";
    } else if (resolvedPath.length > 0) return resolvedPath;
    else return ".";
}
export function normalize(path) {
    assertPath(path);
    if (path.length === 0) return ".";
    const isAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    const trailingSeparator = path.charCodeAt(path.length - 1) === CHAR_FORWARD_SLASH;
    // Normalize the path
    path = normalizeString(path, !isAbsolute, "/", isPosixPathSeparator);
    if (path.length === 0 && !isAbsolute) path = ".";
    if (path.length > 0 && trailingSeparator) path += "/";
    if (isAbsolute) return `/${path}`;
    return path;
}
export function isAbsolute(path) {
    assertPath(path);
    return path.length > 0 && path.charCodeAt(0) === CHAR_FORWARD_SLASH;
}
export function join(...paths) {
    if (paths.length === 0) return ".";
    let joined;
    for(let i = 0, len = paths.length; i < len; ++i){
        const path = paths[i];
        assertPath(path);
        if (path.length > 0) {
            if (!joined) joined = path;
            else joined += `/${path}`;
        }
    }
    if (!joined) return ".";
    return normalize(joined);
}
export function relative(from, to) {
    assertPath(from);
    assertPath(to);
    if (from === to) return "";
    from = resolve(from);
    to = resolve(to);
    if (from === to) return "";
    // Trim any leading backslashes
    let fromStart = 1;
    const fromEnd = from.length;
    for(; fromStart < fromEnd; ++fromStart){
        if (from.charCodeAt(fromStart) !== CHAR_FORWARD_SLASH) break;
    }
    const fromLen = fromEnd - fromStart;
    // Trim any leading backslashes
    let toStart = 1;
    const toEnd = to.length;
    for(; toStart < toEnd; ++toStart){
        if (to.charCodeAt(toStart) !== CHAR_FORWARD_SLASH) break;
    }
    const toLen = toEnd - toStart;
    // Compare paths to find the longest common path from root
    const length = fromLen < toLen ? fromLen : toLen;
    let lastCommonSep = -1;
    let i = 0;
    for(; i <= length; ++i){
        if (i === length) {
            if (toLen > length) {
                if (to.charCodeAt(toStart + i) === CHAR_FORWARD_SLASH) {
                    // We get here if `from` is the exact base path for `to`.
                    // For example: from='/foo/bar'; to='/foo/bar/baz'
                    return to.slice(toStart + i + 1);
                } else if (i === 0) {
                    // We get here if `from` is the root
                    // For example: from='/'; to='/foo'
                    return to.slice(toStart + i);
                }
            } else if (fromLen > length) {
                if (from.charCodeAt(fromStart + i) === CHAR_FORWARD_SLASH) {
                    // We get here if `to` is the exact base path for `from`.
                    // For example: from='/foo/bar/baz'; to='/foo/bar'
                    lastCommonSep = i;
                } else if (i === 0) {
                    // We get here if `to` is the root.
                    // For example: from='/foo'; to='/'
                    lastCommonSep = 0;
                }
            }
            break;
        }
        const fromCode = from.charCodeAt(fromStart + i);
        const toCode = to.charCodeAt(toStart + i);
        if (fromCode !== toCode) break;
        else if (fromCode === CHAR_FORWARD_SLASH) lastCommonSep = i;
    }
    let out = "";
    // Generate the relative path based on the path difference between `to`
    // and `from`
    for(i = fromStart + lastCommonSep + 1; i <= fromEnd; ++i){
        if (i === fromEnd || from.charCodeAt(i) === CHAR_FORWARD_SLASH) {
            if (out.length === 0) out += "..";
            else out += "/..";
        }
    }
    // Lastly, append the rest of the destination (`to`) path that comes after
    // the common path parts
    if (out.length > 0) return out + to.slice(toStart + lastCommonSep);
    else {
        toStart += lastCommonSep;
        if (to.charCodeAt(toStart) === CHAR_FORWARD_SLASH) ++toStart;
        return to.slice(toStart);
    }
}
export function toNamespacedPath(path) {
    // Non-op on posix systems
    return path;
}
export function dirname(path) {
    assertPath(path);
    if (path.length === 0) return ".";
    const hasRoot = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    let end = -1;
    let matchedSlash = true;
    for(let i = path.length - 1; i >= 1; --i){
        if (path.charCodeAt(i) === CHAR_FORWARD_SLASH) {
            if (!matchedSlash) {
                end = i;
                break;
            }
        } else {
            // We saw the first non-path separator
            matchedSlash = false;
        }
    }
    if (end === -1) return hasRoot ? "/" : ".";
    if (hasRoot && end === 1) return "//";
    return path.slice(0, end);
}
export function basename(path, ext = "") {
    if (ext !== undefined && typeof ext !== "string") {
        throw new TypeError('"ext" argument must be a string');
    }
    assertPath(path);
    let start = 0;
    let end = -1;
    let matchedSlash = true;
    let i;
    if (ext !== undefined && ext.length > 0 && ext.length <= path.length) {
        if (ext.length === path.length && ext === path) return "";
        let extIdx = ext.length - 1;
        let firstNonSlashEnd = -1;
        for(i = path.length - 1; i >= 0; --i){
            const code = path.charCodeAt(i);
            if (code === CHAR_FORWARD_SLASH) {
                // If we reached a path separator that was not part of a set of path
                // separators at the end of the string, stop now
                if (!matchedSlash) {
                    start = i + 1;
                    break;
                }
            } else {
                if (firstNonSlashEnd === -1) {
                    // We saw the first non-path separator, remember this index in case
                    // we need it if the extension ends up not matching
                    matchedSlash = false;
                    firstNonSlashEnd = i + 1;
                }
                if (extIdx >= 0) {
                    // Try to match the explicit extension
                    if (code === ext.charCodeAt(extIdx)) {
                        if ((--extIdx) === -1) {
                            // We matched the extension, so mark this as the end of our path
                            // component
                            end = i;
                        }
                    } else {
                        // Extension does not match, so our result is the entire path
                        // component
                        extIdx = -1;
                        end = firstNonSlashEnd;
                    }
                }
            }
        }
        if (start === end) end = firstNonSlashEnd;
        else if (end === -1) end = path.length;
        return path.slice(start, end);
    } else {
        for(i = path.length - 1; i >= 0; --i){
            if (path.charCodeAt(i) === CHAR_FORWARD_SLASH) {
                // If we reached a path separator that was not part of a set of path
                // separators at the end of the string, stop now
                if (!matchedSlash) {
                    start = i + 1;
                    break;
                }
            } else if (end === -1) {
                // We saw the first non-path separator, mark this as the end of our
                // path component
                matchedSlash = false;
                end = i + 1;
            }
        }
        if (end === -1) return "";
        return path.slice(start, end);
    }
}
export function extname(path) {
    assertPath(path);
    let startDot = -1;
    let startPart = 0;
    let end = -1;
    let matchedSlash = true;
    // Track the state of characters (if any) we see before our first dot and
    // after any path separator we find
    let preDotState = 0;
    for(let i = path.length - 1; i >= 0; --i){
        const code = path.charCodeAt(i);
        if (code === CHAR_FORWARD_SLASH) {
            // If we reached a path separator that was not part of a set of path
            // separators at the end of the string, stop now
            if (!matchedSlash) {
                startPart = i + 1;
                break;
            }
            continue;
        }
        if (end === -1) {
            // We saw the first non-path separator, mark this as the end of our
            // extension
            matchedSlash = false;
            end = i + 1;
        }
        if (code === CHAR_DOT) {
            // If this is our first dot, mark it as the start of our extension
            if (startDot === -1) startDot = i;
            else if (preDotState !== 1) preDotState = 1;
        } else if (startDot !== -1) {
            // We saw a non-dot and non-path separator before our dot, so we should
            // have a good chance at having a non-empty extension
            preDotState = -1;
        }
    }
    if (startDot === -1 || end === -1 || // We saw a non-dot character immediately before the dot
    preDotState === 0 || // The (right-most) trimmed path component is exactly '..'
    (preDotState === 1 && startDot === end - 1 && startDot === startPart + 1)) {
        return "";
    }
    return path.slice(startDot, end);
}
export function format(pathObject) {
    /* eslint-disable max-len */ if (pathObject === null || typeof pathObject !== "object") {
        throw new TypeError(`The "pathObject" argument must be of type Object. Received type ${typeof pathObject}`);
    }
    return _format("/", pathObject);
}
export function parse(path) {
    assertPath(path);
    const ret = {
        root: "",
        dir: "",
        base: "",
        ext: "",
        name: ""
    };
    if (path.length === 0) return ret;
    const isAbsolute = path.charCodeAt(0) === CHAR_FORWARD_SLASH;
    let start;
    if (isAbsolute) {
        ret.root = "/";
        start = 1;
    } else {
        start = 0;
    }
    let startDot = -1;
    let startPart = 0;
    let end = -1;
    let matchedSlash = true;
    let i = path.length - 1;
    // Track the state of characters (if any) we see before our first dot and
    // after any path separator we find
    let preDotState = 0;
    // Get non-dir info
    for(; i >= start; --i){
        const code = path.charCodeAt(i);
        if (code === CHAR_FORWARD_SLASH) {
            // If we reached a path separator that was not part of a set of path
            // separators at the end of the string, stop now
            if (!matchedSlash) {
                startPart = i + 1;
                break;
            }
            continue;
        }
        if (end === -1) {
            // We saw the first non-path separator, mark this as the end of our
            // extension
            matchedSlash = false;
            end = i + 1;
        }
        if (code === CHAR_DOT) {
            // If this is our first dot, mark it as the start of our extension
            if (startDot === -1) startDot = i;
            else if (preDotState !== 1) preDotState = 1;
        } else if (startDot !== -1) {
            // We saw a non-dot and non-path separator before our dot, so we should
            // have a good chance at having a non-empty extension
            preDotState = -1;
        }
    }
    if (startDot === -1 || end === -1 || // We saw a non-dot character immediately before the dot
    preDotState === 0 || // The (right-most) trimmed path component is exactly '..'
    (preDotState === 1 && startDot === end - 1 && startDot === startPart + 1)) {
        if (end !== -1) {
            if (startPart === 0 && isAbsolute) {
                ret.base = ret.name = path.slice(1, end);
            } else {
                ret.base = ret.name = path.slice(startPart, end);
            }
        }
    } else {
        if (startPart === 0 && isAbsolute) {
            ret.name = path.slice(1, startDot);
            ret.base = path.slice(1, end);
        } else {
            ret.name = path.slice(startPart, startDot);
            ret.base = path.slice(startPart, end);
        }
        ret.ext = path.slice(startDot, end);
    }
    if (startPart > 0) ret.dir = path.slice(0, startPart - 1);
    else if (isAbsolute) ret.dir = "/";
    return ret;
}
/** Converts a file URL to a path string.
 *
 *      fromFileUrl("file:///home/foo"); // "/home/foo"
 */ export function fromFileUrl(url) {
    url = url instanceof URL ? url : new URL(url);
    if (url.protocol != "file:") {
        throw new TypeError("Must be a file URL.");
    }
    return decodeURIComponent(url.pathname.replace(/%(?![0-9A-Fa-f]{2})/g, "%25"));
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL3BhdGgvcG9zaXgudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCB0aGUgQnJvd3NlcmlmeSBhdXRob3JzLiBNSVQgTGljZW5zZS5cbi8vIFBvcnRlZCBmcm9tIGh0dHBzOi8vZ2l0aHViLmNvbS9icm93c2VyaWZ5L3BhdGgtYnJvd3NlcmlmeS9cbi8qKiBUaGlzIG1vZHVsZSBpcyBicm93c2VyIGNvbXBhdGlibGUuICovXG5cbmltcG9ydCB0eXBlIHsgRm9ybWF0SW5wdXRQYXRoT2JqZWN0LCBQYXJzZWRQYXRoIH0gZnJvbSBcIi4vX2ludGVyZmFjZS50c1wiO1xuaW1wb3J0IHsgQ0hBUl9ET1QsIENIQVJfRk9SV0FSRF9TTEFTSCB9IGZyb20gXCIuL19jb25zdGFudHMudHNcIjtcblxuaW1wb3J0IHtcbiAgYXNzZXJ0UGF0aCxcbiAgbm9ybWFsaXplU3RyaW5nLFxuICBpc1Bvc2l4UGF0aFNlcGFyYXRvcixcbiAgX2Zvcm1hdCxcbn0gZnJvbSBcIi4vX3V0aWwudHNcIjtcblxuZXhwb3J0IGNvbnN0IHNlcCA9IFwiL1wiO1xuZXhwb3J0IGNvbnN0IGRlbGltaXRlciA9IFwiOlwiO1xuXG4vLyBwYXRoLnJlc29sdmUoW2Zyb20gLi4uXSwgdG8pXG5leHBvcnQgZnVuY3Rpb24gcmVzb2x2ZSguLi5wYXRoU2VnbWVudHM6IHN0cmluZ1tdKTogc3RyaW5nIHtcbiAgbGV0IHJlc29sdmVkUGF0aCA9IFwiXCI7XG4gIGxldCByZXNvbHZlZEFic29sdXRlID0gZmFsc2U7XG5cbiAgZm9yIChsZXQgaSA9IHBhdGhTZWdtZW50cy5sZW5ndGggLSAxOyBpID49IC0xICYmICFyZXNvbHZlZEFic29sdXRlOyBpLS0pIHtcbiAgICBsZXQgcGF0aDogc3RyaW5nO1xuXG4gICAgaWYgKGkgPj0gMCkgcGF0aCA9IHBhdGhTZWdtZW50c1tpXTtcbiAgICBlbHNlIHtcbiAgICAgIGlmIChnbG9iYWxUaGlzLkRlbm8gPT0gbnVsbCkge1xuICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiUmVzb2x2ZWQgYSByZWxhdGl2ZSBwYXRoIHdpdGhvdXQgYSBDV0QuXCIpO1xuICAgICAgfVxuICAgICAgcGF0aCA9IERlbm8uY3dkKCk7XG4gICAgfVxuXG4gICAgYXNzZXJ0UGF0aChwYXRoKTtcblxuICAgIC8vIFNraXAgZW1wdHkgZW50cmllc1xuICAgIGlmIChwYXRoLmxlbmd0aCA9PT0gMCkge1xuICAgICAgY29udGludWU7XG4gICAgfVxuXG4gICAgcmVzb2x2ZWRQYXRoID0gYCR7cGF0aH0vJHtyZXNvbHZlZFBhdGh9YDtcbiAgICByZXNvbHZlZEFic29sdXRlID0gcGF0aC5jaGFyQ29kZUF0KDApID09PSBDSEFSX0ZPUldBUkRfU0xBU0g7XG4gIH1cblxuICAvLyBBdCB0aGlzIHBvaW50IHRoZSBwYXRoIHNob3VsZCBiZSByZXNvbHZlZCB0byBhIGZ1bGwgYWJzb2x1dGUgcGF0aCwgYnV0XG4gIC8vIGhhbmRsZSByZWxhdGl2ZSBwYXRocyB0byBiZSBzYWZlIChtaWdodCBoYXBwZW4gd2hlbiBwcm9jZXNzLmN3ZCgpIGZhaWxzKVxuXG4gIC8vIE5vcm1hbGl6ZSB0aGUgcGF0aFxuICByZXNvbHZlZFBhdGggPSBub3JtYWxpemVTdHJpbmcoXG4gICAgcmVzb2x2ZWRQYXRoLFxuICAgICFyZXNvbHZlZEFic29sdXRlLFxuICAgIFwiL1wiLFxuICAgIGlzUG9zaXhQYXRoU2VwYXJhdG9yLFxuICApO1xuXG4gIGlmIChyZXNvbHZlZEFic29sdXRlKSB7XG4gICAgaWYgKHJlc29sdmVkUGF0aC5sZW5ndGggPiAwKSByZXR1cm4gYC8ke3Jlc29sdmVkUGF0aH1gO1xuICAgIGVsc2UgcmV0dXJuIFwiL1wiO1xuICB9IGVsc2UgaWYgKHJlc29sdmVkUGF0aC5sZW5ndGggPiAwKSByZXR1cm4gcmVzb2x2ZWRQYXRoO1xuICBlbHNlIHJldHVybiBcIi5cIjtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIG5vcm1hbGl6ZShwYXRoOiBzdHJpbmcpOiBzdHJpbmcge1xuICBhc3NlcnRQYXRoKHBhdGgpO1xuXG4gIGlmIChwYXRoLmxlbmd0aCA9PT0gMCkgcmV0dXJuIFwiLlwiO1xuXG4gIGNvbnN0IGlzQWJzb2x1dGUgPSBwYXRoLmNoYXJDb2RlQXQoMCkgPT09IENIQVJfRk9SV0FSRF9TTEFTSDtcbiAgY29uc3QgdHJhaWxpbmdTZXBhcmF0b3IgPVxuICAgIHBhdGguY2hhckNvZGVBdChwYXRoLmxlbmd0aCAtIDEpID09PSBDSEFSX0ZPUldBUkRfU0xBU0g7XG5cbiAgLy8gTm9ybWFsaXplIHRoZSBwYXRoXG4gIHBhdGggPSBub3JtYWxpemVTdHJpbmcocGF0aCwgIWlzQWJzb2x1dGUsIFwiL1wiLCBpc1Bvc2l4UGF0aFNlcGFyYXRvcik7XG5cbiAgaWYgKHBhdGgubGVuZ3RoID09PSAwICYmICFpc0Fic29sdXRlKSBwYXRoID0gXCIuXCI7XG4gIGlmIChwYXRoLmxlbmd0aCA+IDAgJiYgdHJhaWxpbmdTZXBhcmF0b3IpIHBhdGggKz0gXCIvXCI7XG5cbiAgaWYgKGlzQWJzb2x1dGUpIHJldHVybiBgLyR7cGF0aH1gO1xuICByZXR1cm4gcGF0aDtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGlzQWJzb2x1dGUocGF0aDogc3RyaW5nKTogYm9vbGVhbiB7XG4gIGFzc2VydFBhdGgocGF0aCk7XG4gIHJldHVybiBwYXRoLmxlbmd0aCA+IDAgJiYgcGF0aC5jaGFyQ29kZUF0KDApID09PSBDSEFSX0ZPUldBUkRfU0xBU0g7XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBqb2luKC4uLnBhdGhzOiBzdHJpbmdbXSk6IHN0cmluZyB7XG4gIGlmIChwYXRocy5sZW5ndGggPT09IDApIHJldHVybiBcIi5cIjtcbiAgbGV0IGpvaW5lZDogc3RyaW5nIHwgdW5kZWZpbmVkO1xuICBmb3IgKGxldCBpID0gMCwgbGVuID0gcGF0aHMubGVuZ3RoOyBpIDwgbGVuOyArK2kpIHtcbiAgICBjb25zdCBwYXRoID0gcGF0aHNbaV07XG4gICAgYXNzZXJ0UGF0aChwYXRoKTtcbiAgICBpZiAocGF0aC5sZW5ndGggPiAwKSB7XG4gICAgICBpZiAoIWpvaW5lZCkgam9pbmVkID0gcGF0aDtcbiAgICAgIGVsc2Ugam9pbmVkICs9IGAvJHtwYXRofWA7XG4gICAgfVxuICB9XG4gIGlmICgham9pbmVkKSByZXR1cm4gXCIuXCI7XG4gIHJldHVybiBub3JtYWxpemUoam9pbmVkKTtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIHJlbGF0aXZlKGZyb206IHN0cmluZywgdG86IHN0cmluZyk6IHN0cmluZyB7XG4gIGFzc2VydFBhdGgoZnJvbSk7XG4gIGFzc2VydFBhdGgodG8pO1xuXG4gIGlmIChmcm9tID09PSB0bykgcmV0dXJuIFwiXCI7XG5cbiAgZnJvbSA9IHJlc29sdmUoZnJvbSk7XG4gIHRvID0gcmVzb2x2ZSh0byk7XG5cbiAgaWYgKGZyb20gPT09IHRvKSByZXR1cm4gXCJcIjtcblxuICAvLyBUcmltIGFueSBsZWFkaW5nIGJhY2tzbGFzaGVzXG4gIGxldCBmcm9tU3RhcnQgPSAxO1xuICBjb25zdCBmcm9tRW5kID0gZnJvbS5sZW5ndGg7XG4gIGZvciAoOyBmcm9tU3RhcnQgPCBmcm9tRW5kOyArK2Zyb21TdGFydCkge1xuICAgIGlmIChmcm9tLmNoYXJDb2RlQXQoZnJvbVN0YXJ0KSAhPT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSBicmVhaztcbiAgfVxuICBjb25zdCBmcm9tTGVuID0gZnJvbUVuZCAtIGZyb21TdGFydDtcblxuICAvLyBUcmltIGFueSBsZWFkaW5nIGJhY2tzbGFzaGVzXG4gIGxldCB0b1N0YXJ0ID0gMTtcbiAgY29uc3QgdG9FbmQgPSB0by5sZW5ndGg7XG4gIGZvciAoOyB0b1N0YXJ0IDwgdG9FbmQ7ICsrdG9TdGFydCkge1xuICAgIGlmICh0by5jaGFyQ29kZUF0KHRvU3RhcnQpICE9PSBDSEFSX0ZPUldBUkRfU0xBU0gpIGJyZWFrO1xuICB9XG4gIGNvbnN0IHRvTGVuID0gdG9FbmQgLSB0b1N0YXJ0O1xuXG4gIC8vIENvbXBhcmUgcGF0aHMgdG8gZmluZCB0aGUgbG9uZ2VzdCBjb21tb24gcGF0aCBmcm9tIHJvb3RcbiAgY29uc3QgbGVuZ3RoID0gZnJvbUxlbiA8IHRvTGVuID8gZnJvbUxlbiA6IHRvTGVuO1xuICBsZXQgbGFzdENvbW1vblNlcCA9IC0xO1xuICBsZXQgaSA9IDA7XG4gIGZvciAoOyBpIDw9IGxlbmd0aDsgKytpKSB7XG4gICAgaWYgKGkgPT09IGxlbmd0aCkge1xuICAgICAgaWYgKHRvTGVuID4gbGVuZ3RoKSB7XG4gICAgICAgIGlmICh0by5jaGFyQ29kZUF0KHRvU3RhcnQgKyBpKSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSB7XG4gICAgICAgICAgLy8gV2UgZ2V0IGhlcmUgaWYgYGZyb21gIGlzIHRoZSBleGFjdCBiYXNlIHBhdGggZm9yIGB0b2AuXG4gICAgICAgICAgLy8gRm9yIGV4YW1wbGU6IGZyb209Jy9mb28vYmFyJzsgdG89Jy9mb28vYmFyL2JheidcbiAgICAgICAgICByZXR1cm4gdG8uc2xpY2UodG9TdGFydCArIGkgKyAxKTtcbiAgICAgICAgfSBlbHNlIGlmIChpID09PSAwKSB7XG4gICAgICAgICAgLy8gV2UgZ2V0IGhlcmUgaWYgYGZyb21gIGlzIHRoZSByb290XG4gICAgICAgICAgLy8gRm9yIGV4YW1wbGU6IGZyb209Jy8nOyB0bz0nL2ZvbydcbiAgICAgICAgICByZXR1cm4gdG8uc2xpY2UodG9TdGFydCArIGkpO1xuICAgICAgICB9XG4gICAgICB9IGVsc2UgaWYgKGZyb21MZW4gPiBsZW5ndGgpIHtcbiAgICAgICAgaWYgKGZyb20uY2hhckNvZGVBdChmcm9tU3RhcnQgKyBpKSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSB7XG4gICAgICAgICAgLy8gV2UgZ2V0IGhlcmUgaWYgYHRvYCBpcyB0aGUgZXhhY3QgYmFzZSBwYXRoIGZvciBgZnJvbWAuXG4gICAgICAgICAgLy8gRm9yIGV4YW1wbGU6IGZyb209Jy9mb28vYmFyL2Jheic7IHRvPScvZm9vL2JhcidcbiAgICAgICAgICBsYXN0Q29tbW9uU2VwID0gaTtcbiAgICAgICAgfSBlbHNlIGlmIChpID09PSAwKSB7XG4gICAgICAgICAgLy8gV2UgZ2V0IGhlcmUgaWYgYHRvYCBpcyB0aGUgcm9vdC5cbiAgICAgICAgICAvLyBGb3IgZXhhbXBsZTogZnJvbT0nL2Zvbyc7IHRvPScvJ1xuICAgICAgICAgIGxhc3RDb21tb25TZXAgPSAwO1xuICAgICAgICB9XG4gICAgICB9XG4gICAgICBicmVhaztcbiAgICB9XG4gICAgY29uc3QgZnJvbUNvZGUgPSBmcm9tLmNoYXJDb2RlQXQoZnJvbVN0YXJ0ICsgaSk7XG4gICAgY29uc3QgdG9Db2RlID0gdG8uY2hhckNvZGVBdCh0b1N0YXJ0ICsgaSk7XG4gICAgaWYgKGZyb21Db2RlICE9PSB0b0NvZGUpIGJyZWFrO1xuICAgIGVsc2UgaWYgKGZyb21Db2RlID09PSBDSEFSX0ZPUldBUkRfU0xBU0gpIGxhc3RDb21tb25TZXAgPSBpO1xuICB9XG5cbiAgbGV0IG91dCA9IFwiXCI7XG4gIC8vIEdlbmVyYXRlIHRoZSByZWxhdGl2ZSBwYXRoIGJhc2VkIG9uIHRoZSBwYXRoIGRpZmZlcmVuY2UgYmV0d2VlbiBgdG9gXG4gIC8vIGFuZCBgZnJvbWBcbiAgZm9yIChpID0gZnJvbVN0YXJ0ICsgbGFzdENvbW1vblNlcCArIDE7IGkgPD0gZnJvbUVuZDsgKytpKSB7XG4gICAgaWYgKGkgPT09IGZyb21FbmQgfHwgZnJvbS5jaGFyQ29kZUF0KGkpID09PSBDSEFSX0ZPUldBUkRfU0xBU0gpIHtcbiAgICAgIGlmIChvdXQubGVuZ3RoID09PSAwKSBvdXQgKz0gXCIuLlwiO1xuICAgICAgZWxzZSBvdXQgKz0gXCIvLi5cIjtcbiAgICB9XG4gIH1cblxuICAvLyBMYXN0bHksIGFwcGVuZCB0aGUgcmVzdCBvZiB0aGUgZGVzdGluYXRpb24gKGB0b2ApIHBhdGggdGhhdCBjb21lcyBhZnRlclxuICAvLyB0aGUgY29tbW9uIHBhdGggcGFydHNcbiAgaWYgKG91dC5sZW5ndGggPiAwKSByZXR1cm4gb3V0ICsgdG8uc2xpY2UodG9TdGFydCArIGxhc3RDb21tb25TZXApO1xuICBlbHNlIHtcbiAgICB0b1N0YXJ0ICs9IGxhc3RDb21tb25TZXA7XG4gICAgaWYgKHRvLmNoYXJDb2RlQXQodG9TdGFydCkgPT09IENIQVJfRk9SV0FSRF9TTEFTSCkgKyt0b1N0YXJ0O1xuICAgIHJldHVybiB0by5zbGljZSh0b1N0YXJ0KTtcbiAgfVxufVxuXG5leHBvcnQgZnVuY3Rpb24gdG9OYW1lc3BhY2VkUGF0aChwYXRoOiBzdHJpbmcpOiBzdHJpbmcge1xuICAvLyBOb24tb3Agb24gcG9zaXggc3lzdGVtc1xuICByZXR1cm4gcGF0aDtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGRpcm5hbWUocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgYXNzZXJ0UGF0aChwYXRoKTtcbiAgaWYgKHBhdGgubGVuZ3RoID09PSAwKSByZXR1cm4gXCIuXCI7XG4gIGNvbnN0IGhhc1Jvb3QgPSBwYXRoLmNoYXJDb2RlQXQoMCkgPT09IENIQVJfRk9SV0FSRF9TTEFTSDtcbiAgbGV0IGVuZCA9IC0xO1xuICBsZXQgbWF0Y2hlZFNsYXNoID0gdHJ1ZTtcbiAgZm9yIChsZXQgaSA9IHBhdGgubGVuZ3RoIC0gMTsgaSA+PSAxOyAtLWkpIHtcbiAgICBpZiAocGF0aC5jaGFyQ29kZUF0KGkpID09PSBDSEFSX0ZPUldBUkRfU0xBU0gpIHtcbiAgICAgIGlmICghbWF0Y2hlZFNsYXNoKSB7XG4gICAgICAgIGVuZCA9IGk7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuICAgIH0gZWxzZSB7XG4gICAgICAvLyBXZSBzYXcgdGhlIGZpcnN0IG5vbi1wYXRoIHNlcGFyYXRvclxuICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgfVxuICB9XG5cbiAgaWYgKGVuZCA9PT0gLTEpIHJldHVybiBoYXNSb290ID8gXCIvXCIgOiBcIi5cIjtcbiAgaWYgKGhhc1Jvb3QgJiYgZW5kID09PSAxKSByZXR1cm4gXCIvL1wiO1xuICByZXR1cm4gcGF0aC5zbGljZSgwLCBlbmQpO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gYmFzZW5hbWUocGF0aDogc3RyaW5nLCBleHQgPSBcIlwiKTogc3RyaW5nIHtcbiAgaWYgKGV4dCAhPT0gdW5kZWZpbmVkICYmIHR5cGVvZiBleHQgIT09IFwic3RyaW5nXCIpIHtcbiAgICB0aHJvdyBuZXcgVHlwZUVycm9yKCdcImV4dFwiIGFyZ3VtZW50IG11c3QgYmUgYSBzdHJpbmcnKTtcbiAgfVxuICBhc3NlcnRQYXRoKHBhdGgpO1xuXG4gIGxldCBzdGFydCA9IDA7XG4gIGxldCBlbmQgPSAtMTtcbiAgbGV0IG1hdGNoZWRTbGFzaCA9IHRydWU7XG4gIGxldCBpOiBudW1iZXI7XG5cbiAgaWYgKGV4dCAhPT0gdW5kZWZpbmVkICYmIGV4dC5sZW5ndGggPiAwICYmIGV4dC5sZW5ndGggPD0gcGF0aC5sZW5ndGgpIHtcbiAgICBpZiAoZXh0Lmxlbmd0aCA9PT0gcGF0aC5sZW5ndGggJiYgZXh0ID09PSBwYXRoKSByZXR1cm4gXCJcIjtcbiAgICBsZXQgZXh0SWR4ID0gZXh0Lmxlbmd0aCAtIDE7XG4gICAgbGV0IGZpcnN0Tm9uU2xhc2hFbmQgPSAtMTtcbiAgICBmb3IgKGkgPSBwYXRoLmxlbmd0aCAtIDE7IGkgPj0gMDsgLS1pKSB7XG4gICAgICBjb25zdCBjb2RlID0gcGF0aC5jaGFyQ29kZUF0KGkpO1xuICAgICAgaWYgKGNvZGUgPT09IENIQVJfRk9SV0FSRF9TTEFTSCkge1xuICAgICAgICAvLyBJZiB3ZSByZWFjaGVkIGEgcGF0aCBzZXBhcmF0b3IgdGhhdCB3YXMgbm90IHBhcnQgb2YgYSBzZXQgb2YgcGF0aFxuICAgICAgICAvLyBzZXBhcmF0b3JzIGF0IHRoZSBlbmQgb2YgdGhlIHN0cmluZywgc3RvcCBub3dcbiAgICAgICAgaWYgKCFtYXRjaGVkU2xhc2gpIHtcbiAgICAgICAgICBzdGFydCA9IGkgKyAxO1xuICAgICAgICAgIGJyZWFrO1xuICAgICAgICB9XG4gICAgICB9IGVsc2Uge1xuICAgICAgICBpZiAoZmlyc3ROb25TbGFzaEVuZCA9PT0gLTEpIHtcbiAgICAgICAgICAvLyBXZSBzYXcgdGhlIGZpcnN0IG5vbi1wYXRoIHNlcGFyYXRvciwgcmVtZW1iZXIgdGhpcyBpbmRleCBpbiBjYXNlXG4gICAgICAgICAgLy8gd2UgbmVlZCBpdCBpZiB0aGUgZXh0ZW5zaW9uIGVuZHMgdXAgbm90IG1hdGNoaW5nXG4gICAgICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgICAgICAgZmlyc3ROb25TbGFzaEVuZCA9IGkgKyAxO1xuICAgICAgICB9XG4gICAgICAgIGlmIChleHRJZHggPj0gMCkge1xuICAgICAgICAgIC8vIFRyeSB0byBtYXRjaCB0aGUgZXhwbGljaXQgZXh0ZW5zaW9uXG4gICAgICAgICAgaWYgKGNvZGUgPT09IGV4dC5jaGFyQ29kZUF0KGV4dElkeCkpIHtcbiAgICAgICAgICAgIGlmICgtLWV4dElkeCA9PT0gLTEpIHtcbiAgICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCB0aGUgZXh0ZW5zaW9uLCBzbyBtYXJrIHRoaXMgYXMgdGhlIGVuZCBvZiBvdXIgcGF0aFxuICAgICAgICAgICAgICAvLyBjb21wb25lbnRcbiAgICAgICAgICAgICAgZW5kID0gaTtcbiAgICAgICAgICAgIH1cbiAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgLy8gRXh0ZW5zaW9uIGRvZXMgbm90IG1hdGNoLCBzbyBvdXIgcmVzdWx0IGlzIHRoZSBlbnRpcmUgcGF0aFxuICAgICAgICAgICAgLy8gY29tcG9uZW50XG4gICAgICAgICAgICBleHRJZHggPSAtMTtcbiAgICAgICAgICAgIGVuZCA9IGZpcnN0Tm9uU2xhc2hFbmQ7XG4gICAgICAgICAgfVxuICAgICAgICB9XG4gICAgICB9XG4gICAgfVxuXG4gICAgaWYgKHN0YXJ0ID09PSBlbmQpIGVuZCA9IGZpcnN0Tm9uU2xhc2hFbmQ7XG4gICAgZWxzZSBpZiAoZW5kID09PSAtMSkgZW5kID0gcGF0aC5sZW5ndGg7XG4gICAgcmV0dXJuIHBhdGguc2xpY2Uoc3RhcnQsIGVuZCk7XG4gIH0gZWxzZSB7XG4gICAgZm9yIChpID0gcGF0aC5sZW5ndGggLSAxOyBpID49IDA7IC0taSkge1xuICAgICAgaWYgKHBhdGguY2hhckNvZGVBdChpKSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSB7XG4gICAgICAgIC8vIElmIHdlIHJlYWNoZWQgYSBwYXRoIHNlcGFyYXRvciB0aGF0IHdhcyBub3QgcGFydCBvZiBhIHNldCBvZiBwYXRoXG4gICAgICAgIC8vIHNlcGFyYXRvcnMgYXQgdGhlIGVuZCBvZiB0aGUgc3RyaW5nLCBzdG9wIG5vd1xuICAgICAgICBpZiAoIW1hdGNoZWRTbGFzaCkge1xuICAgICAgICAgIHN0YXJ0ID0gaSArIDE7XG4gICAgICAgICAgYnJlYWs7XG4gICAgICAgIH1cbiAgICAgIH0gZWxzZSBpZiAoZW5kID09PSAtMSkge1xuICAgICAgICAvLyBXZSBzYXcgdGhlIGZpcnN0IG5vbi1wYXRoIHNlcGFyYXRvciwgbWFyayB0aGlzIGFzIHRoZSBlbmQgb2Ygb3VyXG4gICAgICAgIC8vIHBhdGggY29tcG9uZW50XG4gICAgICAgIG1hdGNoZWRTbGFzaCA9IGZhbHNlO1xuICAgICAgICBlbmQgPSBpICsgMTtcbiAgICAgIH1cbiAgICB9XG5cbiAgICBpZiAoZW5kID09PSAtMSkgcmV0dXJuIFwiXCI7XG4gICAgcmV0dXJuIHBhdGguc2xpY2Uoc3RhcnQsIGVuZCk7XG4gIH1cbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGV4dG5hbWUocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgYXNzZXJ0UGF0aChwYXRoKTtcbiAgbGV0IHN0YXJ0RG90ID0gLTE7XG4gIGxldCBzdGFydFBhcnQgPSAwO1xuICBsZXQgZW5kID0gLTE7XG4gIGxldCBtYXRjaGVkU2xhc2ggPSB0cnVlO1xuICAvLyBUcmFjayB0aGUgc3RhdGUgb2YgY2hhcmFjdGVycyAoaWYgYW55KSB3ZSBzZWUgYmVmb3JlIG91ciBmaXJzdCBkb3QgYW5kXG4gIC8vIGFmdGVyIGFueSBwYXRoIHNlcGFyYXRvciB3ZSBmaW5kXG4gIGxldCBwcmVEb3RTdGF0ZSA9IDA7XG4gIGZvciAobGV0IGkgPSBwYXRoLmxlbmd0aCAtIDE7IGkgPj0gMDsgLS1pKSB7XG4gICAgY29uc3QgY29kZSA9IHBhdGguY2hhckNvZGVBdChpKTtcbiAgICBpZiAoY29kZSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSB7XG4gICAgICAvLyBJZiB3ZSByZWFjaGVkIGEgcGF0aCBzZXBhcmF0b3IgdGhhdCB3YXMgbm90IHBhcnQgb2YgYSBzZXQgb2YgcGF0aFxuICAgICAgLy8gc2VwYXJhdG9ycyBhdCB0aGUgZW5kIG9mIHRoZSBzdHJpbmcsIHN0b3Agbm93XG4gICAgICBpZiAoIW1hdGNoZWRTbGFzaCkge1xuICAgICAgICBzdGFydFBhcnQgPSBpICsgMTtcbiAgICAgICAgYnJlYWs7XG4gICAgICB9XG4gICAgICBjb250aW51ZTtcbiAgICB9XG4gICAgaWYgKGVuZCA9PT0gLTEpIHtcbiAgICAgIC8vIFdlIHNhdyB0aGUgZmlyc3Qgbm9uLXBhdGggc2VwYXJhdG9yLCBtYXJrIHRoaXMgYXMgdGhlIGVuZCBvZiBvdXJcbiAgICAgIC8vIGV4dGVuc2lvblxuICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgICBlbmQgPSBpICsgMTtcbiAgICB9XG4gICAgaWYgKGNvZGUgPT09IENIQVJfRE9UKSB7XG4gICAgICAvLyBJZiB0aGlzIGlzIG91ciBmaXJzdCBkb3QsIG1hcmsgaXQgYXMgdGhlIHN0YXJ0IG9mIG91ciBleHRlbnNpb25cbiAgICAgIGlmIChzdGFydERvdCA9PT0gLTEpIHN0YXJ0RG90ID0gaTtcbiAgICAgIGVsc2UgaWYgKHByZURvdFN0YXRlICE9PSAxKSBwcmVEb3RTdGF0ZSA9IDE7XG4gICAgfSBlbHNlIGlmIChzdGFydERvdCAhPT0gLTEpIHtcbiAgICAgIC8vIFdlIHNhdyBhIG5vbi1kb3QgYW5kIG5vbi1wYXRoIHNlcGFyYXRvciBiZWZvcmUgb3VyIGRvdCwgc28gd2Ugc2hvdWxkXG4gICAgICAvLyBoYXZlIGEgZ29vZCBjaGFuY2UgYXQgaGF2aW5nIGEgbm9uLWVtcHR5IGV4dGVuc2lvblxuICAgICAgcHJlRG90U3RhdGUgPSAtMTtcbiAgICB9XG4gIH1cblxuICBpZiAoXG4gICAgc3RhcnREb3QgPT09IC0xIHx8XG4gICAgZW5kID09PSAtMSB8fFxuICAgIC8vIFdlIHNhdyBhIG5vbi1kb3QgY2hhcmFjdGVyIGltbWVkaWF0ZWx5IGJlZm9yZSB0aGUgZG90XG4gICAgcHJlRG90U3RhdGUgPT09IDAgfHxcbiAgICAvLyBUaGUgKHJpZ2h0LW1vc3QpIHRyaW1tZWQgcGF0aCBjb21wb25lbnQgaXMgZXhhY3RseSAnLi4nXG4gICAgKHByZURvdFN0YXRlID09PSAxICYmIHN0YXJ0RG90ID09PSBlbmQgLSAxICYmIHN0YXJ0RG90ID09PSBzdGFydFBhcnQgKyAxKVxuICApIHtcbiAgICByZXR1cm4gXCJcIjtcbiAgfVxuICByZXR1cm4gcGF0aC5zbGljZShzdGFydERvdCwgZW5kKTtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGZvcm1hdChwYXRoT2JqZWN0OiBGb3JtYXRJbnB1dFBhdGhPYmplY3QpOiBzdHJpbmcge1xuICAvKiBlc2xpbnQtZGlzYWJsZSBtYXgtbGVuICovXG4gIGlmIChwYXRoT2JqZWN0ID09PSBudWxsIHx8IHR5cGVvZiBwYXRoT2JqZWN0ICE9PSBcIm9iamVjdFwiKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcihcbiAgICAgIGBUaGUgXCJwYXRoT2JqZWN0XCIgYXJndW1lbnQgbXVzdCBiZSBvZiB0eXBlIE9iamVjdC4gUmVjZWl2ZWQgdHlwZSAke3R5cGVvZiBwYXRoT2JqZWN0fWAsXG4gICAgKTtcbiAgfVxuICByZXR1cm4gX2Zvcm1hdChcIi9cIiwgcGF0aE9iamVjdCk7XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBwYXJzZShwYXRoOiBzdHJpbmcpOiBQYXJzZWRQYXRoIHtcbiAgYXNzZXJ0UGF0aChwYXRoKTtcblxuICBjb25zdCByZXQ6IFBhcnNlZFBhdGggPSB7IHJvb3Q6IFwiXCIsIGRpcjogXCJcIiwgYmFzZTogXCJcIiwgZXh0OiBcIlwiLCBuYW1lOiBcIlwiIH07XG4gIGlmIChwYXRoLmxlbmd0aCA9PT0gMCkgcmV0dXJuIHJldDtcbiAgY29uc3QgaXNBYnNvbHV0ZSA9IHBhdGguY2hhckNvZGVBdCgwKSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIO1xuICBsZXQgc3RhcnQ6IG51bWJlcjtcbiAgaWYgKGlzQWJzb2x1dGUpIHtcbiAgICByZXQucm9vdCA9IFwiL1wiO1xuICAgIHN0YXJ0ID0gMTtcbiAgfSBlbHNlIHtcbiAgICBzdGFydCA9IDA7XG4gIH1cbiAgbGV0IHN0YXJ0RG90ID0gLTE7XG4gIGxldCBzdGFydFBhcnQgPSAwO1xuICBsZXQgZW5kID0gLTE7XG4gIGxldCBtYXRjaGVkU2xhc2ggPSB0cnVlO1xuICBsZXQgaSA9IHBhdGgubGVuZ3RoIC0gMTtcblxuICAvLyBUcmFjayB0aGUgc3RhdGUgb2YgY2hhcmFjdGVycyAoaWYgYW55KSB3ZSBzZWUgYmVmb3JlIG91ciBmaXJzdCBkb3QgYW5kXG4gIC8vIGFmdGVyIGFueSBwYXRoIHNlcGFyYXRvciB3ZSBmaW5kXG4gIGxldCBwcmVEb3RTdGF0ZSA9IDA7XG5cbiAgLy8gR2V0IG5vbi1kaXIgaW5mb1xuICBmb3IgKDsgaSA+PSBzdGFydDsgLS1pKSB7XG4gICAgY29uc3QgY29kZSA9IHBhdGguY2hhckNvZGVBdChpKTtcbiAgICBpZiAoY29kZSA9PT0gQ0hBUl9GT1JXQVJEX1NMQVNIKSB7XG4gICAgICAvLyBJZiB3ZSByZWFjaGVkIGEgcGF0aCBzZXBhcmF0b3IgdGhhdCB3YXMgbm90IHBhcnQgb2YgYSBzZXQgb2YgcGF0aFxuICAgICAgLy8gc2VwYXJhdG9ycyBhdCB0aGUgZW5kIG9mIHRoZSBzdHJpbmcsIHN0b3Agbm93XG4gICAgICBpZiAoIW1hdGNoZWRTbGFzaCkge1xuICAgICAgICBzdGFydFBhcnQgPSBpICsgMTtcbiAgICAgICAgYnJlYWs7XG4gICAgICB9XG4gICAgICBjb250aW51ZTtcbiAgICB9XG4gICAgaWYgKGVuZCA9PT0gLTEpIHtcbiAgICAgIC8vIFdlIHNhdyB0aGUgZmlyc3Qgbm9uLXBhdGggc2VwYXJhdG9yLCBtYXJrIHRoaXMgYXMgdGhlIGVuZCBvZiBvdXJcbiAgICAgIC8vIGV4dGVuc2lvblxuICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgICBlbmQgPSBpICsgMTtcbiAgICB9XG4gICAgaWYgKGNvZGUgPT09IENIQVJfRE9UKSB7XG4gICAgICAvLyBJZiB0aGlzIGlzIG91ciBmaXJzdCBkb3QsIG1hcmsgaXQgYXMgdGhlIHN0YXJ0IG9mIG91ciBleHRlbnNpb25cbiAgICAgIGlmIChzdGFydERvdCA9PT0gLTEpIHN0YXJ0RG90ID0gaTtcbiAgICAgIGVsc2UgaWYgKHByZURvdFN0YXRlICE9PSAxKSBwcmVEb3RTdGF0ZSA9IDE7XG4gICAgfSBlbHNlIGlmIChzdGFydERvdCAhPT0gLTEpIHtcbiAgICAgIC8vIFdlIHNhdyBhIG5vbi1kb3QgYW5kIG5vbi1wYXRoIHNlcGFyYXRvciBiZWZvcmUgb3VyIGRvdCwgc28gd2Ugc2hvdWxkXG4gICAgICAvLyBoYXZlIGEgZ29vZCBjaGFuY2UgYXQgaGF2aW5nIGEgbm9uLWVtcHR5IGV4dGVuc2lvblxuICAgICAgcHJlRG90U3RhdGUgPSAtMTtcbiAgICB9XG4gIH1cblxuICBpZiAoXG4gICAgc3RhcnREb3QgPT09IC0xIHx8XG4gICAgZW5kID09PSAtMSB8fFxuICAgIC8vIFdlIHNhdyBhIG5vbi1kb3QgY2hhcmFjdGVyIGltbWVkaWF0ZWx5IGJlZm9yZSB0aGUgZG90XG4gICAgcHJlRG90U3RhdGUgPT09IDAgfHxcbiAgICAvLyBUaGUgKHJpZ2h0LW1vc3QpIHRyaW1tZWQgcGF0aCBjb21wb25lbnQgaXMgZXhhY3RseSAnLi4nXG4gICAgKHByZURvdFN0YXRlID09PSAxICYmIHN0YXJ0RG90ID09PSBlbmQgLSAxICYmIHN0YXJ0RG90ID09PSBzdGFydFBhcnQgKyAxKVxuICApIHtcbiAgICBpZiAoZW5kICE9PSAtMSkge1xuICAgICAgaWYgKHN0YXJ0UGFydCA9PT0gMCAmJiBpc0Fic29sdXRlKSB7XG4gICAgICAgIHJldC5iYXNlID0gcmV0Lm5hbWUgPSBwYXRoLnNsaWNlKDEsIGVuZCk7XG4gICAgICB9IGVsc2Uge1xuICAgICAgICByZXQuYmFzZSA9IHJldC5uYW1lID0gcGF0aC5zbGljZShzdGFydFBhcnQsIGVuZCk7XG4gICAgICB9XG4gICAgfVxuICB9IGVsc2Uge1xuICAgIGlmIChzdGFydFBhcnQgPT09IDAgJiYgaXNBYnNvbHV0ZSkge1xuICAgICAgcmV0Lm5hbWUgPSBwYXRoLnNsaWNlKDEsIHN0YXJ0RG90KTtcbiAgICAgIHJldC5iYXNlID0gcGF0aC5zbGljZSgxLCBlbmQpO1xuICAgIH0gZWxzZSB7XG4gICAgICByZXQubmFtZSA9IHBhdGguc2xpY2Uoc3RhcnRQYXJ0LCBzdGFydERvdCk7XG4gICAgICByZXQuYmFzZSA9IHBhdGguc2xpY2Uoc3RhcnRQYXJ0LCBlbmQpO1xuICAgIH1cbiAgICByZXQuZXh0ID0gcGF0aC5zbGljZShzdGFydERvdCwgZW5kKTtcbiAgfVxuXG4gIGlmIChzdGFydFBhcnQgPiAwKSByZXQuZGlyID0gcGF0aC5zbGljZSgwLCBzdGFydFBhcnQgLSAxKTtcbiAgZWxzZSBpZiAoaXNBYnNvbHV0ZSkgcmV0LmRpciA9IFwiL1wiO1xuXG4gIHJldHVybiByZXQ7XG59XG5cbi8qKiBDb252ZXJ0cyBhIGZpbGUgVVJMIHRvIGEgcGF0aCBzdHJpbmcuXG4gKlxuICogICAgICBmcm9tRmlsZVVybChcImZpbGU6Ly8vaG9tZS9mb29cIik7IC8vIFwiL2hvbWUvZm9vXCJcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIGZyb21GaWxlVXJsKHVybDogc3RyaW5nIHwgVVJMKTogc3RyaW5nIHtcbiAgdXJsID0gdXJsIGluc3RhbmNlb2YgVVJMID8gdXJsIDogbmV3IFVSTCh1cmwpO1xuICBpZiAodXJsLnByb3RvY29sICE9IFwiZmlsZTpcIikge1xuICAgIHRocm93IG5ldyBUeXBlRXJyb3IoXCJNdXN0IGJlIGEgZmlsZSBVUkwuXCIpO1xuICB9XG4gIHJldHVybiBkZWNvZGVVUklDb21wb25lbnQoXG4gICAgdXJsLnBhdGhuYW1lLnJlcGxhY2UoLyUoPyFbMC05QS1GYS1mXXsyfSkvZywgXCIlMjVcIiksXG4gICk7XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IlNBS0EsUUFBQSxFQUFBLGtCQUFBLFNBQUEsZUFBQTtTQUdBLFVBQUEsRUFDQSxlQUFBLEVBQ0Esb0JBQUEsRUFDQSxPQUFBLFNBQ0EsVUFBQTthQUVBLEdBQUEsSUFBQSxDQUFBO2FBQ0EsU0FBQSxJQUFBLENBQUE7QUFFQSxFQUFBLDZCQUFBO2dCQUNBLE9BQUEsSUFBQSxZQUFBO1FBQ0EsWUFBQTtRQUNBLGdCQUFBLEdBQUEsS0FBQTtZQUVBLENBQUEsR0FBQSxZQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsRUFBQSxDQUFBLEtBQUEsQ0FBQSxLQUFBLGdCQUFBLEVBQUEsQ0FBQTtZQUNBLElBQUE7WUFFQSxDQUFBLElBQUEsQ0FBQSxFQUFBLElBQUEsR0FBQSxZQUFBLENBQUEsQ0FBQTs7Z0JBRUEsVUFBQSxDQUFBLElBQUEsSUFBQSxJQUFBOzBCQUNBLFNBQUEsRUFBQSx1Q0FBQTs7QUFFQSxnQkFBQSxHQUFBLElBQUEsQ0FBQSxHQUFBOztBQUdBLGtCQUFBLENBQUEsSUFBQTtBQUVBLFVBQUEsbUJBQUE7WUFDQSxJQUFBLENBQUEsTUFBQSxLQUFBLENBQUE7OztBQUlBLG9CQUFBLE1BQUEsSUFBQSxDQUFBLENBQUEsRUFBQSxZQUFBO0FBQ0Esd0JBQUEsR0FBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUEsTUFBQSxrQkFBQTs7QUFHQSxNQUFBLHVFQUFBO0FBQ0EsTUFBQSx5RUFBQTtBQUVBLE1BQUEsbUJBQUE7QUFDQSxnQkFBQSxHQUFBLGVBQUEsQ0FDQSxZQUFBLEdBQ0EsZ0JBQUEsR0FDQSxDQUFBLEdBQ0Esb0JBQUE7UUFHQSxnQkFBQTtZQUNBLFlBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxVQUFBLENBQUEsRUFBQSxZQUFBO3FCQUNBLENBQUE7ZUFDQSxZQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsU0FBQSxZQUFBO2lCQUNBLENBQUE7O2dCQUdBLFNBQUEsQ0FBQSxJQUFBO0FBQ0EsY0FBQSxDQUFBLElBQUE7UUFFQSxJQUFBLENBQUEsTUFBQSxLQUFBLENBQUEsVUFBQSxDQUFBO1VBRUEsVUFBQSxHQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQSxNQUFBLGtCQUFBO1VBQ0EsaUJBQUEsR0FDQSxJQUFBLENBQUEsVUFBQSxDQUFBLElBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxNQUFBLGtCQUFBO0FBRUEsTUFBQSxtQkFBQTtBQUNBLFFBQUEsR0FBQSxlQUFBLENBQUEsSUFBQSxHQUFBLFVBQUEsR0FBQSxDQUFBLEdBQUEsb0JBQUE7UUFFQSxJQUFBLENBQUEsTUFBQSxLQUFBLENBQUEsS0FBQSxVQUFBLEVBQUEsSUFBQSxJQUFBLENBQUE7UUFDQSxJQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsSUFBQSxpQkFBQSxFQUFBLElBQUEsS0FBQSxDQUFBO1FBRUEsVUFBQSxVQUFBLENBQUEsRUFBQSxJQUFBO1dBQ0EsSUFBQTs7Z0JBR0EsVUFBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtXQUNBLElBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxJQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQSxNQUFBLGtCQUFBOztnQkFHQSxJQUFBLElBQUEsS0FBQTtRQUNBLEtBQUEsQ0FBQSxNQUFBLEtBQUEsQ0FBQSxVQUFBLENBQUE7UUFDQSxNQUFBO1lBQ0EsQ0FBQSxHQUFBLENBQUEsRUFBQSxHQUFBLEdBQUEsS0FBQSxDQUFBLE1BQUEsRUFBQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUE7Y0FDQSxJQUFBLEdBQUEsS0FBQSxDQUFBLENBQUE7QUFDQSxrQkFBQSxDQUFBLElBQUE7WUFDQSxJQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7aUJBQ0EsTUFBQSxFQUFBLE1BQUEsR0FBQSxJQUFBO2lCQUNBLE1BQUEsS0FBQSxDQUFBLEVBQUEsSUFBQTs7O1NBR0EsTUFBQSxVQUFBLENBQUE7V0FDQSxTQUFBLENBQUEsTUFBQTs7Z0JBR0EsUUFBQSxDQUFBLElBQUEsRUFBQSxFQUFBO0FBQ0EsY0FBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsRUFBQTtRQUVBLElBQUEsS0FBQSxFQUFBO0FBRUEsUUFBQSxHQUFBLE9BQUEsQ0FBQSxJQUFBO0FBQ0EsTUFBQSxHQUFBLE9BQUEsQ0FBQSxFQUFBO1FBRUEsSUFBQSxLQUFBLEVBQUE7QUFFQSxNQUFBLDZCQUFBO1FBQ0EsU0FBQSxHQUFBLENBQUE7VUFDQSxPQUFBLEdBQUEsSUFBQSxDQUFBLE1BQUE7VUFDQSxTQUFBLEdBQUEsT0FBQSxJQUFBLFNBQUE7WUFDQSxJQUFBLENBQUEsVUFBQSxDQUFBLFNBQUEsTUFBQSxrQkFBQTs7VUFFQSxPQUFBLEdBQUEsT0FBQSxHQUFBLFNBQUE7QUFFQSxNQUFBLDZCQUFBO1FBQ0EsT0FBQSxHQUFBLENBQUE7VUFDQSxLQUFBLEdBQUEsRUFBQSxDQUFBLE1BQUE7VUFDQSxPQUFBLEdBQUEsS0FBQSxJQUFBLE9BQUE7WUFDQSxFQUFBLENBQUEsVUFBQSxDQUFBLE9BQUEsTUFBQSxrQkFBQTs7VUFFQSxLQUFBLEdBQUEsS0FBQSxHQUFBLE9BQUE7QUFFQSxNQUFBLHdEQUFBO1VBQ0EsTUFBQSxHQUFBLE9BQUEsR0FBQSxLQUFBLEdBQUEsT0FBQSxHQUFBLEtBQUE7UUFDQSxhQUFBLElBQUEsQ0FBQTtRQUNBLENBQUEsR0FBQSxDQUFBO1VBQ0EsQ0FBQSxJQUFBLE1BQUEsSUFBQSxDQUFBO1lBQ0EsQ0FBQSxLQUFBLE1BQUE7Z0JBQ0EsS0FBQSxHQUFBLE1BQUE7b0JBQ0EsRUFBQSxDQUFBLFVBQUEsQ0FBQSxPQUFBLEdBQUEsQ0FBQSxNQUFBLGtCQUFBO0FBQ0Esc0JBQUEsdURBQUE7QUFDQSxzQkFBQSxnREFBQTsyQkFDQSxFQUFBLENBQUEsS0FBQSxDQUFBLE9BQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQTsyQkFDQSxDQUFBLEtBQUEsQ0FBQTtBQUNBLHNCQUFBLGtDQUFBO0FBQ0Esc0JBQUEsaUNBQUE7MkJBQ0EsRUFBQSxDQUFBLEtBQUEsQ0FBQSxPQUFBLEdBQUEsQ0FBQTs7dUJBRUEsT0FBQSxHQUFBLE1BQUE7b0JBQ0EsSUFBQSxDQUFBLFVBQUEsQ0FBQSxTQUFBLEdBQUEsQ0FBQSxNQUFBLGtCQUFBO0FBQ0Esc0JBQUEsdURBQUE7QUFDQSxzQkFBQSxnREFBQTtBQUNBLGlDQUFBLEdBQUEsQ0FBQTsyQkFDQSxDQUFBLEtBQUEsQ0FBQTtBQUNBLHNCQUFBLGlDQUFBO0FBQ0Esc0JBQUEsaUNBQUE7QUFDQSxpQ0FBQSxHQUFBLENBQUE7Ozs7O2NBS0EsUUFBQSxHQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsU0FBQSxHQUFBLENBQUE7Y0FDQSxNQUFBLEdBQUEsRUFBQSxDQUFBLFVBQUEsQ0FBQSxPQUFBLEdBQUEsQ0FBQTtZQUNBLFFBQUEsS0FBQSxNQUFBO2lCQUNBLFFBQUEsS0FBQSxrQkFBQSxFQUFBLGFBQUEsR0FBQSxDQUFBOztRQUdBLEdBQUE7QUFDQSxNQUFBLHFFQUFBO0FBQ0EsTUFBQSxXQUFBO1FBQ0EsQ0FBQSxHQUFBLFNBQUEsR0FBQSxhQUFBLEdBQUEsQ0FBQSxFQUFBLENBQUEsSUFBQSxPQUFBLElBQUEsQ0FBQTtZQUNBLENBQUEsS0FBQSxPQUFBLElBQUEsSUFBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBLE1BQUEsa0JBQUE7Z0JBQ0EsR0FBQSxDQUFBLE1BQUEsS0FBQSxDQUFBLEVBQUEsR0FBQSxLQUFBLEVBQUE7aUJBQ0EsR0FBQSxLQUFBLEdBQUE7OztBQUlBLE1BQUEsd0VBQUE7QUFDQSxNQUFBLHNCQUFBO1FBQ0EsR0FBQSxDQUFBLE1BQUEsR0FBQSxDQUFBLFNBQUEsR0FBQSxHQUFBLEVBQUEsQ0FBQSxLQUFBLENBQUEsT0FBQSxHQUFBLGFBQUE7O0FBRUEsZUFBQSxJQUFBLGFBQUE7WUFDQSxFQUFBLENBQUEsVUFBQSxDQUFBLE9BQUEsTUFBQSxrQkFBQSxJQUFBLE9BQUE7ZUFDQSxFQUFBLENBQUEsS0FBQSxDQUFBLE9BQUE7OztnQkFJQSxnQkFBQSxDQUFBLElBQUE7QUFDQSxNQUFBLHdCQUFBO1dBQ0EsSUFBQTs7Z0JBR0EsT0FBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtRQUNBLElBQUEsQ0FBQSxNQUFBLEtBQUEsQ0FBQSxVQUFBLENBQUE7VUFDQSxPQUFBLEdBQUEsSUFBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBLE1BQUEsa0JBQUE7UUFDQSxHQUFBLElBQUEsQ0FBQTtRQUNBLFlBQUEsR0FBQSxJQUFBO1lBQ0EsQ0FBQSxHQUFBLElBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxFQUFBLENBQUEsSUFBQSxDQUFBLElBQUEsQ0FBQTtZQUNBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQSxNQUFBLGtCQUFBO2lCQUNBLFlBQUE7QUFDQSxtQkFBQSxHQUFBLENBQUE7Ozs7QUFJQSxjQUFBLG9DQUFBO0FBQ0Esd0JBQUEsR0FBQSxLQUFBOzs7UUFJQSxHQUFBLE1BQUEsQ0FBQSxTQUFBLE9BQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQTtRQUNBLE9BQUEsSUFBQSxHQUFBLEtBQUEsQ0FBQSxVQUFBLEVBQUE7V0FDQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxHQUFBOztnQkFHQSxRQUFBLENBQUEsSUFBQSxFQUFBLEdBQUE7UUFDQSxHQUFBLEtBQUEsU0FBQSxXQUFBLEdBQUEsTUFBQSxNQUFBO2tCQUNBLFNBQUEsRUFBQSwrQkFBQTs7QUFFQSxjQUFBLENBQUEsSUFBQTtRQUVBLEtBQUEsR0FBQSxDQUFBO1FBQ0EsR0FBQSxJQUFBLENBQUE7UUFDQSxZQUFBLEdBQUEsSUFBQTtRQUNBLENBQUE7UUFFQSxHQUFBLEtBQUEsU0FBQSxJQUFBLEdBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxJQUFBLEdBQUEsQ0FBQSxNQUFBLElBQUEsSUFBQSxDQUFBLE1BQUE7WUFDQSxHQUFBLENBQUEsTUFBQSxLQUFBLElBQUEsQ0FBQSxNQUFBLElBQUEsR0FBQSxLQUFBLElBQUE7WUFDQSxNQUFBLEdBQUEsR0FBQSxDQUFBLE1BQUEsR0FBQSxDQUFB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LHNEQUFBO0FBQ0EsZUFBQSxLQUFBLENBQUEsSUFDQSxFQUFBLHdEQUFBO0tBQ0EsV0FBQSxLQUFBLENBQUEsSUFBQSxRQUFBLEtBQUEsR0FBQSxHQUFBLENBQUEsSUFBQSxRQUFBLEtBQUEsU0FBQSxHQUFBLENBQUE7OztXQUlBLElBQUEsQ0FBQSxLQUFBLENBQUEsUUFBQSxFQUFBLEdBQUE7O2dCQUdBLE1BQUEsQ0FBQSxVQUFBO0FBQ0EsTUFBQSx3QkFBQSxFQUFBLEtBQ0EsVUFBQSxLQUFBLElBQUEsV0FBQSxVQUFBLE1BQUEsTUFBQTtrQkFDQSxTQUFBLEVBQ0EsZ0VBQUEsU0FBQSxVQUFBOztXQUdBLE9BQUEsRUFBQSxDQUFBLEdBQUEsVUFBQTs7Z0JBR0EsS0FBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtVQUVBLEdBQUE7QUFBQSxZQUFBO0FBQUEsV0FBQTtBQUFBLFlBQUE7QUFBQSxXQUFBO0FBQUEsWUFBQTs7UUFDQSxJQUFBLENBQUEsTUFBQSxLQUFBLENBQUEsU0FBQSxHQUFBO1VBQ0EsVUFBQSxHQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQSxNQUFBLGtCQUFBO1FBQ0EsS0FBQTtRQUNBLFVBQUE7QUFDQSxXQUFBLENBQUEsSUFBQSxJQUFBLENBQUE7QUFDQSxhQUFBLEdBQUEsQ0FBQTs7QUFFQSxhQUFBLEdBQUEsQ0FBQTs7UUFFQSxRQUFBLElBQUEsQ0FBQTtRQUNBLFNBQUEsR0FBQSxDQUFBO1FBQ0EsR0FBQSxJQUFBLENBQUE7UUFDQSxZQUFBLEdBQUEsSUFBQTtRQUNBLENBQUEsR0FBQSxJQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7QUFFQSxNQUFBLHVFQUFBO0FBQ0EsTUFBQSxpQ0FBQTtRQUNBLFdBQUEsR0FBQSxDQUFBO0FBRUEsTUFBQSxpQkFBQTtVQUNBLENBQUEsSUFBQSxLQUFBLElBQUEsQ0FBQTtjQUNBLElBQUEsR0FBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUE7WUFDQSxJQUFBLEtBQUEsa0JBQUE7QUFDQSxjQUFBLGtFQUFBO0FBQ0EsY0FBQSw4Q0FBQTtpQkFDQSxZQUFBO0FBQ0EseUJBQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQTs7Ozs7WUFLQSxHQUFBLE1BQUEsQ0FBQTtBQUNBLGNBQUEsaUVBQUE7QUFDQSxjQUFBLFVBQUE7QUFDQSx3QkFBQSxHQUFBLEtBQUE7QUFDQSxlQUFBLEdBQUEsQ0FBQSxHQUFBLENBQUE7O1lBRUEsSUFBQSxLQUFBLFFBQUE7QUFDQSxjQUFBLGdFQUFBO2dCQUNBLFFBQUEsTUFBQSxDQUFBLEVBQUEsUUFBQSxHQUFBLENBQUE7cUJBQ0EsV0FBQSxLQUFBLENBQUEsRUFBQSxXQUFBLEdBQUEsQ0FBQTttQkFDQSxRQUFBLE1BQUEsQ0FBQTtBQUNBLGNBQUEscUVBQUE7QUFDQSxjQUFBLG1EQUFBO0FBQ0EsdUJBQUEsSUFBQSxDQUFBOzs7UUFLQSxRQUFBLE1BQUEsQ0FBQSxJQUNBLEdBQUEsTUFBQSxDQUFBLElBQ0EsRUFBQSxzREFBQTtBQUNBLGVBQUEsS0FBQSxDQUFBLElBQ0EsRUFBQSx3REFBQTtLQUNBLFdBQUEsS0FBQSxDQUFBLElBQUEsUUFBQSxLQUFBLEdBQUEsR0FBQSxDQUFBLElBQUEsUUFBQSxLQUFBLFNBQUEsR0FBQSxDQUFBO1lBRUEsR0FBQSxNQUFBLENBQUE7Z0JBQ0EsU0FBQSxLQUFBLENBQUEsSUFBQSxVQUFBO0FBQ0EsbUJBQUEsQ0FBQSxJQUFBLEdBQUEsR0FBQSxDQUFBLElBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxHQUFBOztBQUVBLG1CQUFBLENBQUEsSUFBQSxHQUFBLEdBQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxTQUFBLEVBQUEsR0FBQTs7OztZQUlBLFNBQUEsS0FBQSxDQUFBLElBQUEsVUFBQTtBQUNBLGVBQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxDQUFBLEVBQUEsUUFBQTtBQUNBLGVBQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxDQUFBLEVBQUEsR0FBQTs7QUFFQSxlQUFBLENBQUEsSUFBQSxHQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsU0FBQSxFQUFBLFFBQUE7QUFDQSxlQUFBLENBQUEsSUFBQSxHQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsU0FBQSxFQUFBLEdBQUE7O0FBRUEsV0FBQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLFFBQUEsRUFBQSxHQUFBOztRQUdBLFNBQUEsR0FBQSxDQUFBLEVBQUEsR0FBQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxTQUFBLEdBQUEsQ0FBQTthQUNBLFVBQUEsRUFBQSxHQUFBLENBQUEsR0FBQSxJQUFBLENBQUE7V0FFQSxHQUFBOztBQUdBLEVBR0EsQUFIQSxvR0FHQSxBQUhBLEVBR0EsaUJBQ0EsV0FBQSxDQUFBLEdBQUE7QUFDQSxPQUFBLEdBQUEsR0FBQSxZQUFBLEdBQUEsR0FBQSxHQUFBLE9BQUEsR0FBQSxDQUFBLEdBQUE7UUFDQSxHQUFBLENBQUEsUUFBQSxLQUFBLEtBQUE7a0JBQ0EsU0FBQSxFQUFBLG1CQUFBOztXQUVBLGtCQUFBLENBQ0EsR0FBQSxDQUFBLFFBQUEsQ0FBQSxPQUFBLDBCQUFBLEdBQUEifQ==