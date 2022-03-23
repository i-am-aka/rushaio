import { CHAR_DOT, CHAR_BACKWARD_SLASH, CHAR_COLON, CHAR_QUESTION_MARK } from "./_constants.ts";
import { assertPath, isPathSeparator, isWindowsDeviceRoot, normalizeString, _format } from "./_util.ts";
import { assert } from "../_util/assert.ts";
export const sep = "\\";
export const delimiter = ";";
export function resolve(...pathSegments) {
    let resolvedDevice = "";
    let resolvedTail = "";
    let resolvedAbsolute = false;
    for(let i = pathSegments.length - 1; i >= -1; i--){
        let path;
        if (i >= 0) {
            path = pathSegments[i];
        } else if (!resolvedDevice) {
            if (globalThis.Deno == null) {
                throw new TypeError("Resolved a drive-letter-less path without a CWD.");
            }
            path = Deno.cwd();
        } else {
            if (globalThis.Deno == null) {
                throw new TypeError("Resolved a relative path without a CWD.");
            }
            // Windows has the concept of drive-specific current working
            // directories. If we've resolved a drive letter but not yet an
            // absolute path, get cwd for that drive, or the process cwd if
            // the drive cwd is not available. We're sure the device is not
            // a UNC path at this points, because UNC paths are always absolute.
            path = Deno.env.get(`=${resolvedDevice}`) || Deno.cwd();
            // Verify that a cwd was found and that it actually points
            // to our drive. If not, default to the drive's root.
            if (path === undefined || path.slice(0, 3).toLowerCase() !== `${resolvedDevice.toLowerCase()}\\`) {
                path = `${resolvedDevice}\\`;
            }
        }
        assertPath(path);
        const len = path.length;
        // Skip empty entries
        if (len === 0) continue;
        let rootEnd = 0;
        let device = "";
        let isAbsolute = false;
        const code = path.charCodeAt(0);
        // Try to match a root
        if (len > 1) {
            if (isPathSeparator(code)) {
                // Possible UNC root
                // If we started with a separator, we know we at least have an
                // absolute path of some kind (UNC or otherwise)
                isAbsolute = true;
                if (isPathSeparator(path.charCodeAt(1))) {
                    // Matched double path separator at beginning
                    let j = 2;
                    let last = j;
                    // Match 1 or more non-path separators
                    for(; j < len; ++j){
                        if (isPathSeparator(path.charCodeAt(j))) break;
                    }
                    if (j < len && j !== last) {
                        const firstPart = path.slice(last, j);
                        // Matched!
                        last = j;
                        // Match 1 or more path separators
                        for(; j < len; ++j){
                            if (!isPathSeparator(path.charCodeAt(j))) break;
                        }
                        if (j < len && j !== last) {
                            // Matched!
                            last = j;
                            // Match 1 or more non-path separators
                            for(; j < len; ++j){
                                if (isPathSeparator(path.charCodeAt(j))) break;
                            }
                            if (j === len) {
                                // We matched a UNC root only
                                device = `\\\\${firstPart}\\${path.slice(last)}`;
                                rootEnd = j;
                            } else if (j !== last) {
                                // We matched a UNC root with leftovers
                                device = `\\\\${firstPart}\\${path.slice(last, j)}`;
                                rootEnd = j;
                            }
                        }
                    }
                } else {
                    rootEnd = 1;
                }
            } else if (isWindowsDeviceRoot(code)) {
                // Possible device root
                if (path.charCodeAt(1) === CHAR_COLON) {
                    device = path.slice(0, 2);
                    rootEnd = 2;
                    if (len > 2) {
                        if (isPathSeparator(path.charCodeAt(2))) {
                            // Treat separator following drive name as an absolute path
                            // indicator
                            isAbsolute = true;
                            rootEnd = 3;
                        }
                    }
                }
            }
        } else if (isPathSeparator(code)) {
            // `path` contains just a path separator
            rootEnd = 1;
            isAbsolute = true;
        }
        if (device.length > 0 && resolvedDevice.length > 0 && device.toLowerCase() !== resolvedDevice.toLowerCase()) {
            continue;
        }
        if (resolvedDevice.length === 0 && device.length > 0) {
            resolvedDevice = device;
        }
        if (!resolvedAbsolute) {
            resolvedTail = `${path.slice(rootEnd)}\\${resolvedTail}`;
            resolvedAbsolute = isAbsolute;
        }
        if (resolvedAbsolute && resolvedDevice.length > 0) break;
    }
    // At this point the path should be resolved to a full absolute path,
    // but handle relative paths to be safe (might happen when process.cwd()
    // fails)
    // Normalize the tail path
    resolvedTail = normalizeString(resolvedTail, !resolvedAbsolute, "\\", isPathSeparator);
    return resolvedDevice + (resolvedAbsolute ? "\\" : "") + resolvedTail || ".";
}
export function normalize(path) {
    assertPath(path);
    const len = path.length;
    if (len === 0) return ".";
    let rootEnd = 0;
    let device;
    let isAbsolute = false;
    const code = path.charCodeAt(0);
    // Try to match a root
    if (len > 1) {
        if (isPathSeparator(code)) {
            // Possible UNC root
            // If we started with a separator, we know we at least have an absolute
            // path of some kind (UNC or otherwise)
            isAbsolute = true;
            if (isPathSeparator(path.charCodeAt(1))) {
                // Matched double path separator at beginning
                let j = 2;
                let last = j;
                // Match 1 or more non-path separators
                for(; j < len; ++j){
                    if (isPathSeparator(path.charCodeAt(j))) break;
                }
                if (j < len && j !== last) {
                    const firstPart = path.slice(last, j);
                    // Matched!
                    last = j;
                    // Match 1 or more path separators
                    for(; j < len; ++j){
                        if (!isPathSeparator(path.charCodeAt(j))) break;
                    }
                    if (j < len && j !== last) {
                        // Matched!
                        last = j;
                        // Match 1 or more non-path separators
                        for(; j < len; ++j){
                            if (isPathSeparator(path.charCodeAt(j))) break;
                        }
                        if (j === len) {
                            // We matched a UNC root only
                            // Return the normalized version of the UNC root since there
                            // is nothing left to process
                            return `\\\\${firstPart}\\${path.slice(last)}\\`;
                        } else if (j !== last) {
                            // We matched a UNC root with leftovers
                            device = `\\\\${firstPart}\\${path.slice(last, j)}`;
                            rootEnd = j;
                        }
                    }
                }
            } else {
                rootEnd = 1;
            }
        } else if (isWindowsDeviceRoot(code)) {
            // Possible device root
            if (path.charCodeAt(1) === CHAR_COLON) {
                device = path.slice(0, 2);
                rootEnd = 2;
                if (len > 2) {
                    if (isPathSeparator(path.charCodeAt(2))) {
                        // Treat separator following drive name as an absolute path
                        // indicator
                        isAbsolute = true;
                        rootEnd = 3;
                    }
                }
            }
        }
    } else if (isPathSeparator(code)) {
        // `path` contains just a path separator, exit early to avoid unnecessary
        // work
        return "\\";
    }
    let tail;
    if (rootEnd < len) {
        tail = normalizeString(path.slice(rootEnd), !isAbsolute, "\\", isPathSeparator);
    } else {
        tail = "";
    }
    if (tail.length === 0 && !isAbsolute) tail = ".";
    if (tail.length > 0 && isPathSeparator(path.charCodeAt(len - 1))) {
        tail += "\\";
    }
    if (device === undefined) {
        if (isAbsolute) {
            if (tail.length > 0) return `\\${tail}`;
            else return "\\";
        } else if (tail.length > 0) {
            return tail;
        } else {
            return "";
        }
    } else if (isAbsolute) {
        if (tail.length > 0) return `${device}\\${tail}`;
        else return `${device}\\`;
    } else if (tail.length > 0) {
        return device + tail;
    } else {
        return device;
    }
}
export function isAbsolute(path) {
    assertPath(path);
    const len = path.length;
    if (len === 0) return false;
    const code = path.charCodeAt(0);
    if (isPathSeparator(code)) {
        return true;
    } else if (isWindowsDeviceRoot(code)) {
        // Possible device root
        if (len > 2 && path.charCodeAt(1) === CHAR_COLON) {
            if (isPathSeparator(path.charCodeAt(2))) return true;
        }
    }
    return false;
}
export function join(...paths) {
    const pathsCount = paths.length;
    if (pathsCount === 0) return ".";
    let joined;
    let firstPart = null;
    for(let i = 0; i < pathsCount; ++i){
        const path = paths[i];
        assertPath(path);
        if (path.length > 0) {
            if (joined === undefined) joined = firstPart = path;
            else joined += `\\${path}`;
        }
    }
    if (joined === undefined) return ".";
    // Make sure that the joined path doesn't start with two slashes, because
    // normalize() will mistake it for an UNC path then.
    //
    // This step is skipped when it is very clear that the user actually
    // intended to point at an UNC path. This is assumed when the first
    // non-empty string arguments starts with exactly two slashes followed by
    // at least one more non-slash character.
    //
    // Note that for normalize() to treat a path as an UNC path it needs to
    // have at least 2 components, so we don't filter for that here.
    // This means that the user can use join to construct UNC paths from
    // a server name and a share name; for example:
    //   path.join('//server', 'share') -> '\\\\server\\share\\')
    let needsReplace = true;
    let slashCount = 0;
    assert(firstPart != null);
    if (isPathSeparator(firstPart.charCodeAt(0))) {
        ++slashCount;
        const firstLen = firstPart.length;
        if (firstLen > 1) {
            if (isPathSeparator(firstPart.charCodeAt(1))) {
                ++slashCount;
                if (firstLen > 2) {
                    if (isPathSeparator(firstPart.charCodeAt(2))) ++slashCount;
                    else {
                        // We matched a UNC path in the first part
                        needsReplace = false;
                    }
                }
            }
        }
    }
    if (needsReplace) {
        // Find any more consecutive slashes we need to replace
        for(; slashCount < joined.length; ++slashCount){
            if (!isPathSeparator(joined.charCodeAt(slashCount))) break;
        }
        // Replace the slashes if needed
        if (slashCount >= 2) joined = `\\${joined.slice(slashCount)}`;
    }
    return normalize(joined);
}
// It will solve the relative path from `from` to `to`, for instance:
//  from = 'C:\\orandea\\test\\aaa'
//  to = 'C:\\orandea\\impl\\bbb'
// The output of the function should be: '..\\..\\impl\\bbb'
export function relative(from, to) {
    assertPath(from);
    assertPath(to);
    if (from === to) return "";
    const fromOrig = resolve(from);
    const toOrig = resolve(to);
    if (fromOrig === toOrig) return "";
    from = fromOrig.toLowerCase();
    to = toOrig.toLowerCase();
    if (from === to) return "";
    // Trim any leading backslashes
    let fromStart = 0;
    let fromEnd = from.length;
    for(; fromStart < fromEnd; ++fromStart){
        if (from.charCodeAt(fromStart) !== CHAR_BACKWARD_SLASH) break;
    }
    // Trim trailing backslashes (applicable to UNC paths only)
    for(; fromEnd - 1 > fromStart; --fromEnd){
        if (from.charCodeAt(fromEnd - 1) !== CHAR_BACKWARD_SLASH) break;
    }
    const fromLen = fromEnd - fromStart;
    // Trim any leading backslashes
    let toStart = 0;
    let toEnd = to.length;
    for(; toStart < toEnd; ++toStart){
        if (to.charCodeAt(toStart) !== CHAR_BACKWARD_SLASH) break;
    }
    // Trim trailing backslashes (applicable to UNC paths only)
    for(; toEnd - 1 > toStart; --toEnd){
        if (to.charCodeAt(toEnd - 1) !== CHAR_BACKWARD_SLASH) break;
    }
    const toLen = toEnd - toStart;
    // Compare paths to find the longest common path from root
    const length = fromLen < toLen ? fromLen : toLen;
    let lastCommonSep = -1;
    let i = 0;
    for(; i <= length; ++i){
        if (i === length) {
            if (toLen > length) {
                if (to.charCodeAt(toStart + i) === CHAR_BACKWARD_SLASH) {
                    // We get here if `from` is the exact base path for `to`.
                    // For example: from='C:\\foo\\bar'; to='C:\\foo\\bar\\baz'
                    return toOrig.slice(toStart + i + 1);
                } else if (i === 2) {
                    // We get here if `from` is the device root.
                    // For example: from='C:\\'; to='C:\\foo'
                    return toOrig.slice(toStart + i);
                }
            }
            if (fromLen > length) {
                if (from.charCodeAt(fromStart + i) === CHAR_BACKWARD_SLASH) {
                    // We get here if `to` is the exact base path for `from`.
                    // For example: from='C:\\foo\\bar'; to='C:\\foo'
                    lastCommonSep = i;
                } else if (i === 2) {
                    // We get here if `to` is the device root.
                    // For example: from='C:\\foo\\bar'; to='C:\\'
                    lastCommonSep = 3;
                }
            }
            break;
        }
        const fromCode = from.charCodeAt(fromStart + i);
        const toCode = to.charCodeAt(toStart + i);
        if (fromCode !== toCode) break;
        else if (fromCode === CHAR_BACKWARD_SLASH) lastCommonSep = i;
    }
    // We found a mismatch before the first common path separator was seen, so
    // return the original `to`.
    if (i !== length && lastCommonSep === -1) {
        return toOrig;
    }
    let out = "";
    if (lastCommonSep === -1) lastCommonSep = 0;
    // Generate the relative path based on the path difference between `to` and
    // `from`
    for(i = fromStart + lastCommonSep + 1; i <= fromEnd; ++i){
        if (i === fromEnd || from.charCodeAt(i) === CHAR_BACKWARD_SLASH) {
            if (out.length === 0) out += "..";
            else out += "\\..";
        }
    }
    // Lastly, append the rest of the destination (`to`) path that comes after
    // the common path parts
    if (out.length > 0) {
        return out + toOrig.slice(toStart + lastCommonSep, toEnd);
    } else {
        toStart += lastCommonSep;
        if (toOrig.charCodeAt(toStart) === CHAR_BACKWARD_SLASH) ++toStart;
        return toOrig.slice(toStart, toEnd);
    }
}
export function toNamespacedPath(path) {
    // Note: this will *probably* throw somewhere.
    if (typeof path !== "string") return path;
    if (path.length === 0) return "";
    const resolvedPath = resolve(path);
    if (resolvedPath.length >= 3) {
        if (resolvedPath.charCodeAt(0) === CHAR_BACKWARD_SLASH) {
            // Possible UNC root
            if (resolvedPath.charCodeAt(1) === CHAR_BACKWARD_SLASH) {
                const code = resolvedPath.charCodeAt(2);
                if (code !== CHAR_QUESTION_MARK && code !== CHAR_DOT) {
                    // Matched non-long UNC root, convert the path to a long UNC path
                    return `\\\\?\\UNC\\${resolvedPath.slice(2)}`;
                }
            }
        } else if (isWindowsDeviceRoot(resolvedPath.charCodeAt(0))) {
            // Possible device root
            if (resolvedPath.charCodeAt(1) === CHAR_COLON && resolvedPath.charCodeAt(2) === CHAR_BACKWARD_SLASH) {
                // Matched device root, convert the path to a long UNC path
                return `\\\\?\\${resolvedPath}`;
            }
        }
    }
    return path;
}
export function dirname(path) {
    assertPath(path);
    const len = path.length;
    if (len === 0) return ".";
    let rootEnd = -1;
    let end = -1;
    let matchedSlash = true;
    let offset = 0;
    const code = path.charCodeAt(0);
    // Try to match a root
    if (len > 1) {
        if (isPathSeparator(code)) {
            // Possible UNC root
            rootEnd = offset = 1;
            if (isPathSeparator(path.charCodeAt(1))) {
                // Matched double path separator at beginning
                let j = 2;
                let last = j;
                // Match 1 or more non-path separators
                for(; j < len; ++j){
                    if (isPathSeparator(path.charCodeAt(j))) break;
                }
                if (j < len && j !== last) {
                    // Matched!
                    last = j;
                    // Match 1 or more path separators
                    for(; j < len; ++j){
                        if (!isPathSeparator(path.charCodeAt(j))) break;
                    }
                    if (j < len && j !== last) {
                        // Matched!
                        last = j;
                        // Match 1 or more non-path separators
                        for(; j < len; ++j){
                            if (isPathSeparator(path.charCodeAt(j))) break;
                        }
                        if (j === len) {
                            // We matched a UNC root only
                            return path;
                        }
                        if (j !== last) {
                            // We matched a UNC root with leftovers
                            // Offset by 1 to include the separator after the UNC root to
                            // treat it as a "normal root" on top of a (UNC) root
                            rootEnd = offset = j + 1;
                        }
                    }
                }
            }
        } else if (isWindowsDeviceRoot(code)) {
            // Possible device root
            if (path.charCodeAt(1) === CHAR_COLON) {
                rootEnd = offset = 2;
                if (len > 2) {
                    if (isPathSeparator(path.charCodeAt(2))) rootEnd = offset = 3;
                }
            }
        }
    } else if (isPathSeparator(code)) {
        // `path` contains just a path separator, exit early to avoid
        // unnecessary work
        return path;
    }
    for(let i = len - 1; i >= offset; --i){
        if (isPathSeparator(path.charCodeAt(i))) {
            if (!matchedSlash) {
                end = i;
                break;
            }
        } else {
            // We saw the first non-path separator
            matchedSlash = false;
        }
    }
    if (end === -1) {
        if (rootEnd === -1) return ".";
        else end = rootEnd;
    }
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
    // Check for a drive letter prefix so as not to mistake the following
    // path separator as an extra separator at the end of the path that can be
    // disregarded
    if (path.length >= 2) {
        const drive = path.charCodeAt(0);
        if (isWindowsDeviceRoot(drive)) {
            if (path.charCodeAt(1) === CHAR_COLON) start = 2;
        }
    }
    if (ext !== undefined && ext.length > 0 && ext.length <= path.length) {
        if (ext.length === path.length && ext === path) return "";
        let extIdx = ext.length - 1;
        let firstNonSlashEnd = -1;
        for(i = path.length - 1; i >= start; --i){
            const code = path.charCodeAt(i);
            if (isPathSeparator(code)) {
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
        for(i = path.length - 1; i >= start; --i){
            if (isPathSeparator(path.charCodeAt(i))) {
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
    let start = 0;
    let startDot = -1;
    let startPart = 0;
    let end = -1;
    let matchedSlash = true;
    // Track the state of characters (if any) we see before our first dot and
    // after any path separator we find
    let preDotState = 0;
    // Check for a drive letter prefix so as not to mistake the following
    // path separator as an extra separator at the end of the path that can be
    // disregarded
    if (path.length >= 2 && path.charCodeAt(1) === CHAR_COLON && isWindowsDeviceRoot(path.charCodeAt(0))) {
        start = startPart = 2;
    }
    for(let i = path.length - 1; i >= start; --i){
        const code = path.charCodeAt(i);
        if (isPathSeparator(code)) {
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
    return _format("\\", pathObject);
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
    const len = path.length;
    if (len === 0) return ret;
    let rootEnd = 0;
    let code = path.charCodeAt(0);
    // Try to match a root
    if (len > 1) {
        if (isPathSeparator(code)) {
            // Possible UNC root
            rootEnd = 1;
            if (isPathSeparator(path.charCodeAt(1))) {
                // Matched double path separator at beginning
                let j = 2;
                let last = j;
                // Match 1 or more non-path separators
                for(; j < len; ++j){
                    if (isPathSeparator(path.charCodeAt(j))) break;
                }
                if (j < len && j !== last) {
                    // Matched!
                    last = j;
                    // Match 1 or more path separators
                    for(; j < len; ++j){
                        if (!isPathSeparator(path.charCodeAt(j))) break;
                    }
                    if (j < len && j !== last) {
                        // Matched!
                        last = j;
                        // Match 1 or more non-path separators
                        for(; j < len; ++j){
                            if (isPathSeparator(path.charCodeAt(j))) break;
                        }
                        if (j === len) {
                            // We matched a UNC root only
                            rootEnd = j;
                        } else if (j !== last) {
                            // We matched a UNC root with leftovers
                            rootEnd = j + 1;
                        }
                    }
                }
            }
        } else if (isWindowsDeviceRoot(code)) {
            // Possible device root
            if (path.charCodeAt(1) === CHAR_COLON) {
                rootEnd = 2;
                if (len > 2) {
                    if (isPathSeparator(path.charCodeAt(2))) {
                        if (len === 3) {
                            // `path` contains just a drive root, exit early to avoid
                            // unnecessary work
                            ret.root = ret.dir = path;
                            return ret;
                        }
                        rootEnd = 3;
                    }
                } else {
                    // `path` contains just a drive root, exit early to avoid
                    // unnecessary work
                    ret.root = ret.dir = path;
                    return ret;
                }
            }
        }
    } else if (isPathSeparator(code)) {
        // `path` contains just a path separator, exit early to avoid
        // unnecessary work
        ret.root = ret.dir = path;
        return ret;
    }
    if (rootEnd > 0) ret.root = path.slice(0, rootEnd);
    let startDot = -1;
    let startPart = rootEnd;
    let end = -1;
    let matchedSlash = true;
    let i = path.length - 1;
    // Track the state of characters (if any) we see before our first dot and
    // after any path separator we find
    let preDotState = 0;
    // Get non-dir info
    for(; i >= rootEnd; --i){
        code = path.charCodeAt(i);
        if (isPathSeparator(code)) {
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
            ret.base = ret.name = path.slice(startPart, end);
        }
    } else {
        ret.name = path.slice(startPart, startDot);
        ret.base = path.slice(startPart, end);
        ret.ext = path.slice(startDot, end);
    }
    // If the directory is the root, use the entire root as the `dir` including
    // the trailing slash if any (`C:\abc` -> `C:\`). Otherwise, strip out the
    // trailing slash (`C:\abc\def` -> `C:\abc`).
    if (startPart > 0 && startPart !== rootEnd) {
        ret.dir = path.slice(0, startPart - 1);
    } else ret.dir = ret.root;
    return ret;
}
/** Converts a file URL to a path string.
 *
 *      fromFileUrl("file:///home/foo"); // "\\home\\foo"
 *      fromFileUrl("file:///C:/Users/foo"); // "C:\\Users\\foo"
 *      fromFileUrl("file://localhost/home/foo"); // "\\\\localhost\\home\\foo"
 */ export function fromFileUrl(url) {
    url = url instanceof URL ? url : new URL(url);
    if (url.protocol != "file:") {
        throw new TypeError("Must be a file URL.");
    }
    let path = decodeURIComponent(url.pathname.replace(/^\/*([A-Za-z]:)(\/|$)/, "$1/").replace(/\//g, "\\").replace(/%(?![0-9A-Fa-f]{2})/g, "%25"));
    if (url.hostname != "") {
        // Note: The `URL` implementation guarantees that the drive letter and
        // hostname are mutually exclusive. Otherwise it would not have been valid
        // to append the hostname and path like this.
        path = `\\\\${url.hostname}${path}`;
    }
    return path;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL3BhdGgvd2luMzIudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCB0aGUgQnJvd3NlcmlmeSBhdXRob3JzLiBNSVQgTGljZW5zZS5cbi8vIFBvcnRlZCBmcm9tIGh0dHBzOi8vZ2l0aHViLmNvbS9icm93c2VyaWZ5L3BhdGgtYnJvd3NlcmlmeS9cbi8qKiBUaGlzIG1vZHVsZSBpcyBicm93c2VyIGNvbXBhdGlibGUuICovXG5cbmltcG9ydCB0eXBlIHsgRm9ybWF0SW5wdXRQYXRoT2JqZWN0LCBQYXJzZWRQYXRoIH0gZnJvbSBcIi4vX2ludGVyZmFjZS50c1wiO1xuaW1wb3J0IHtcbiAgQ0hBUl9ET1QsXG4gIENIQVJfQkFDS1dBUkRfU0xBU0gsXG4gIENIQVJfQ09MT04sXG4gIENIQVJfUVVFU1RJT05fTUFSSyxcbn0gZnJvbSBcIi4vX2NvbnN0YW50cy50c1wiO1xuXG5pbXBvcnQge1xuICBhc3NlcnRQYXRoLFxuICBpc1BhdGhTZXBhcmF0b3IsXG4gIGlzV2luZG93c0RldmljZVJvb3QsXG4gIG5vcm1hbGl6ZVN0cmluZyxcbiAgX2Zvcm1hdCxcbn0gZnJvbSBcIi4vX3V0aWwudHNcIjtcbmltcG9ydCB7IGFzc2VydCB9IGZyb20gXCIuLi9fdXRpbC9hc3NlcnQudHNcIjtcblxuZXhwb3J0IGNvbnN0IHNlcCA9IFwiXFxcXFwiO1xuZXhwb3J0IGNvbnN0IGRlbGltaXRlciA9IFwiO1wiO1xuXG5leHBvcnQgZnVuY3Rpb24gcmVzb2x2ZSguLi5wYXRoU2VnbWVudHM6IHN0cmluZ1tdKTogc3RyaW5nIHtcbiAgbGV0IHJlc29sdmVkRGV2aWNlID0gXCJcIjtcbiAgbGV0IHJlc29sdmVkVGFpbCA9IFwiXCI7XG4gIGxldCByZXNvbHZlZEFic29sdXRlID0gZmFsc2U7XG5cbiAgZm9yIChsZXQgaSA9IHBhdGhTZWdtZW50cy5sZW5ndGggLSAxOyBpID49IC0xOyBpLS0pIHtcbiAgICBsZXQgcGF0aDogc3RyaW5nO1xuICAgIGlmIChpID49IDApIHtcbiAgICAgIHBhdGggPSBwYXRoU2VnbWVudHNbaV07XG4gICAgfSBlbHNlIGlmICghcmVzb2x2ZWREZXZpY2UpIHtcbiAgICAgIGlmIChnbG9iYWxUaGlzLkRlbm8gPT0gbnVsbCkge1xuICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiUmVzb2x2ZWQgYSBkcml2ZS1sZXR0ZXItbGVzcyBwYXRoIHdpdGhvdXQgYSBDV0QuXCIpO1xuICAgICAgfVxuICAgICAgcGF0aCA9IERlbm8uY3dkKCk7XG4gICAgfSBlbHNlIHtcbiAgICAgIGlmIChnbG9iYWxUaGlzLkRlbm8gPT0gbnVsbCkge1xuICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiUmVzb2x2ZWQgYSByZWxhdGl2ZSBwYXRoIHdpdGhvdXQgYSBDV0QuXCIpO1xuICAgICAgfVxuICAgICAgLy8gV2luZG93cyBoYXMgdGhlIGNvbmNlcHQgb2YgZHJpdmUtc3BlY2lmaWMgY3VycmVudCB3b3JraW5nXG4gICAgICAvLyBkaXJlY3Rvcmllcy4gSWYgd2UndmUgcmVzb2x2ZWQgYSBkcml2ZSBsZXR0ZXIgYnV0IG5vdCB5ZXQgYW5cbiAgICAgIC8vIGFic29sdXRlIHBhdGgsIGdldCBjd2QgZm9yIHRoYXQgZHJpdmUsIG9yIHRoZSBwcm9jZXNzIGN3ZCBpZlxuICAgICAgLy8gdGhlIGRyaXZlIGN3ZCBpcyBub3QgYXZhaWxhYmxlLiBXZSdyZSBzdXJlIHRoZSBkZXZpY2UgaXMgbm90XG4gICAgICAvLyBhIFVOQyBwYXRoIGF0IHRoaXMgcG9pbnRzLCBiZWNhdXNlIFVOQyBwYXRocyBhcmUgYWx3YXlzIGFic29sdXRlLlxuICAgICAgcGF0aCA9IERlbm8uZW52LmdldChgPSR7cmVzb2x2ZWREZXZpY2V9YCkgfHwgRGVuby5jd2QoKTtcblxuICAgICAgLy8gVmVyaWZ5IHRoYXQgYSBjd2Qgd2FzIGZvdW5kIGFuZCB0aGF0IGl0IGFjdHVhbGx5IHBvaW50c1xuICAgICAgLy8gdG8gb3VyIGRyaXZlLiBJZiBub3QsIGRlZmF1bHQgdG8gdGhlIGRyaXZlJ3Mgcm9vdC5cbiAgICAgIGlmIChcbiAgICAgICAgcGF0aCA9PT0gdW5kZWZpbmVkIHx8XG4gICAgICAgIHBhdGguc2xpY2UoMCwgMykudG9Mb3dlckNhc2UoKSAhPT0gYCR7cmVzb2x2ZWREZXZpY2UudG9Mb3dlckNhc2UoKX1cXFxcYFxuICAgICAgKSB7XG4gICAgICAgIHBhdGggPSBgJHtyZXNvbHZlZERldmljZX1cXFxcYDtcbiAgICAgIH1cbiAgICB9XG5cbiAgICBhc3NlcnRQYXRoKHBhdGgpO1xuXG4gICAgY29uc3QgbGVuID0gcGF0aC5sZW5ndGg7XG5cbiAgICAvLyBTa2lwIGVtcHR5IGVudHJpZXNcbiAgICBpZiAobGVuID09PSAwKSBjb250aW51ZTtcblxuICAgIGxldCByb290RW5kID0gMDtcbiAgICBsZXQgZGV2aWNlID0gXCJcIjtcbiAgICBsZXQgaXNBYnNvbHV0ZSA9IGZhbHNlO1xuICAgIGNvbnN0IGNvZGUgPSBwYXRoLmNoYXJDb2RlQXQoMCk7XG5cbiAgICAvLyBUcnkgdG8gbWF0Y2ggYSByb290XG4gICAgaWYgKGxlbiA+IDEpIHtcbiAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAgICAgLy8gUG9zc2libGUgVU5DIHJvb3RcblxuICAgICAgICAvLyBJZiB3ZSBzdGFydGVkIHdpdGggYSBzZXBhcmF0b3IsIHdlIGtub3cgd2UgYXQgbGVhc3QgaGF2ZSBhblxuICAgICAgICAvLyBhYnNvbHV0ZSBwYXRoIG9mIHNvbWUga2luZCAoVU5DIG9yIG90aGVyd2lzZSlcbiAgICAgICAgaXNBYnNvbHV0ZSA9IHRydWU7XG5cbiAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoMSkpKSB7XG4gICAgICAgICAgLy8gTWF0Y2hlZCBkb3VibGUgcGF0aCBzZXBhcmF0b3IgYXQgYmVnaW5uaW5nXG4gICAgICAgICAgbGV0IGogPSAyO1xuICAgICAgICAgIGxldCBsYXN0ID0gajtcbiAgICAgICAgICAvLyBNYXRjaCAxIG9yIG1vcmUgbm9uLXBhdGggc2VwYXJhdG9yc1xuICAgICAgICAgIGZvciAoOyBqIDwgbGVuOyArK2opIHtcbiAgICAgICAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KGopKSkgYnJlYWs7XG4gICAgICAgICAgfVxuICAgICAgICAgIGlmIChqIDwgbGVuICYmIGogIT09IGxhc3QpIHtcbiAgICAgICAgICAgIGNvbnN0IGZpcnN0UGFydCA9IHBhdGguc2xpY2UobGFzdCwgaik7XG4gICAgICAgICAgICAvLyBNYXRjaGVkIVxuICAgICAgICAgICAgbGFzdCA9IGo7XG4gICAgICAgICAgICAvLyBNYXRjaCAxIG9yIG1vcmUgcGF0aCBzZXBhcmF0b3JzXG4gICAgICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgICAgIGlmICghaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdChqKSkpIGJyZWFrO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgaWYgKGogPCBsZW4gJiYgaiAhPT0gbGFzdCkge1xuICAgICAgICAgICAgICAvLyBNYXRjaGVkIVxuICAgICAgICAgICAgICBsYXN0ID0gajtcbiAgICAgICAgICAgICAgLy8gTWF0Y2ggMSBvciBtb3JlIG5vbi1wYXRoIHNlcGFyYXRvcnNcbiAgICAgICAgICAgICAgZm9yICg7IGogPCBsZW47ICsraikge1xuICAgICAgICAgICAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KGopKSkgYnJlYWs7XG4gICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgaWYgKGogPT09IGxlbikge1xuICAgICAgICAgICAgICAgIC8vIFdlIG1hdGNoZWQgYSBVTkMgcm9vdCBvbmx5XG4gICAgICAgICAgICAgICAgZGV2aWNlID0gYFxcXFxcXFxcJHtmaXJzdFBhcnR9XFxcXCR7cGF0aC5zbGljZShsYXN0KX1gO1xuICAgICAgICAgICAgICAgIHJvb3RFbmQgPSBqO1xuICAgICAgICAgICAgICB9IGVsc2UgaWYgKGogIT09IGxhc3QpIHtcbiAgICAgICAgICAgICAgICAvLyBXZSBtYXRjaGVkIGEgVU5DIHJvb3Qgd2l0aCBsZWZ0b3ZlcnNcblxuICAgICAgICAgICAgICAgIGRldmljZSA9IGBcXFxcXFxcXCR7Zmlyc3RQYXJ0fVxcXFwke3BhdGguc2xpY2UobGFzdCwgail9YDtcbiAgICAgICAgICAgICAgICByb290RW5kID0gajtcbiAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgfVxuICAgICAgICAgIH1cbiAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICByb290RW5kID0gMTtcbiAgICAgICAgfVxuICAgICAgfSBlbHNlIGlmIChpc1dpbmRvd3NEZXZpY2VSb290KGNvZGUpKSB7XG4gICAgICAgIC8vIFBvc3NpYmxlIGRldmljZSByb290XG5cbiAgICAgICAgaWYgKHBhdGguY2hhckNvZGVBdCgxKSA9PT0gQ0hBUl9DT0xPTikge1xuICAgICAgICAgIGRldmljZSA9IHBhdGguc2xpY2UoMCwgMik7XG4gICAgICAgICAgcm9vdEVuZCA9IDI7XG4gICAgICAgICAgaWYgKGxlbiA+IDIpIHtcbiAgICAgICAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KDIpKSkge1xuICAgICAgICAgICAgICAvLyBUcmVhdCBzZXBhcmF0b3IgZm9sbG93aW5nIGRyaXZlIG5hbWUgYXMgYW4gYWJzb2x1dGUgcGF0aFxuICAgICAgICAgICAgICAvLyBpbmRpY2F0b3JcbiAgICAgICAgICAgICAgaXNBYnNvbHV0ZSA9IHRydWU7XG4gICAgICAgICAgICAgIHJvb3RFbmQgPSAzO1xuICAgICAgICAgICAgfVxuICAgICAgICAgIH1cbiAgICAgICAgfVxuICAgICAgfVxuICAgIH0gZWxzZSBpZiAoaXNQYXRoU2VwYXJhdG9yKGNvZGUpKSB7XG4gICAgICAvLyBgcGF0aGAgY29udGFpbnMganVzdCBhIHBhdGggc2VwYXJhdG9yXG4gICAgICByb290RW5kID0gMTtcbiAgICAgIGlzQWJzb2x1dGUgPSB0cnVlO1xuICAgIH1cblxuICAgIGlmIChcbiAgICAgIGRldmljZS5sZW5ndGggPiAwICYmXG4gICAgICByZXNvbHZlZERldmljZS5sZW5ndGggPiAwICYmXG4gICAgICBkZXZpY2UudG9Mb3dlckNhc2UoKSAhPT0gcmVzb2x2ZWREZXZpY2UudG9Mb3dlckNhc2UoKVxuICAgICkge1xuICAgICAgLy8gVGhpcyBwYXRoIHBvaW50cyB0byBhbm90aGVyIGRldmljZSBzbyBpdCBpcyBub3QgYXBwbGljYWJsZVxuICAgICAgY29udGludWU7XG4gICAgfVxuXG4gICAgaWYgKHJlc29sdmVkRGV2aWNlLmxlbmd0aCA9PT0gMCAmJiBkZXZpY2UubGVuZ3RoID4gMCkge1xuICAgICAgcmVzb2x2ZWREZXZpY2UgPSBkZXZpY2U7XG4gICAgfVxuICAgIGlmICghcmVzb2x2ZWRBYnNvbHV0ZSkge1xuICAgICAgcmVzb2x2ZWRUYWlsID0gYCR7cGF0aC5zbGljZShyb290RW5kKX1cXFxcJHtyZXNvbHZlZFRhaWx9YDtcbiAgICAgIHJlc29sdmVkQWJzb2x1dGUgPSBpc0Fic29sdXRlO1xuICAgIH1cblxuICAgIGlmIChyZXNvbHZlZEFic29sdXRlICYmIHJlc29sdmVkRGV2aWNlLmxlbmd0aCA+IDApIGJyZWFrO1xuICB9XG5cbiAgLy8gQXQgdGhpcyBwb2ludCB0aGUgcGF0aCBzaG91bGQgYmUgcmVzb2x2ZWQgdG8gYSBmdWxsIGFic29sdXRlIHBhdGgsXG4gIC8vIGJ1dCBoYW5kbGUgcmVsYXRpdmUgcGF0aHMgdG8gYmUgc2FmZSAobWlnaHQgaGFwcGVuIHdoZW4gcHJvY2Vzcy5jd2QoKVxuICAvLyBmYWlscylcblxuICAvLyBOb3JtYWxpemUgdGhlIHRhaWwgcGF0aFxuICByZXNvbHZlZFRhaWwgPSBub3JtYWxpemVTdHJpbmcoXG4gICAgcmVzb2x2ZWRUYWlsLFxuICAgICFyZXNvbHZlZEFic29sdXRlLFxuICAgIFwiXFxcXFwiLFxuICAgIGlzUGF0aFNlcGFyYXRvcixcbiAgKTtcblxuICByZXR1cm4gcmVzb2x2ZWREZXZpY2UgKyAocmVzb2x2ZWRBYnNvbHV0ZSA/IFwiXFxcXFwiIDogXCJcIikgKyByZXNvbHZlZFRhaWwgfHwgXCIuXCI7XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBub3JtYWxpemUocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgYXNzZXJ0UGF0aChwYXRoKTtcbiAgY29uc3QgbGVuID0gcGF0aC5sZW5ndGg7XG4gIGlmIChsZW4gPT09IDApIHJldHVybiBcIi5cIjtcbiAgbGV0IHJvb3RFbmQgPSAwO1xuICBsZXQgZGV2aWNlOiBzdHJpbmcgfCB1bmRlZmluZWQ7XG4gIGxldCBpc0Fic29sdXRlID0gZmFsc2U7XG4gIGNvbnN0IGNvZGUgPSBwYXRoLmNoYXJDb2RlQXQoMCk7XG5cbiAgLy8gVHJ5IHRvIG1hdGNoIGEgcm9vdFxuICBpZiAobGVuID4gMSkge1xuICAgIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAgIC8vIFBvc3NpYmxlIFVOQyByb290XG5cbiAgICAgIC8vIElmIHdlIHN0YXJ0ZWQgd2l0aCBhIHNlcGFyYXRvciwgd2Uga25vdyB3ZSBhdCBsZWFzdCBoYXZlIGFuIGFic29sdXRlXG4gICAgICAvLyBwYXRoIG9mIHNvbWUga2luZCAoVU5DIG9yIG90aGVyd2lzZSlcbiAgICAgIGlzQWJzb2x1dGUgPSB0cnVlO1xuXG4gICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdCgxKSkpIHtcbiAgICAgICAgLy8gTWF0Y2hlZCBkb3VibGUgcGF0aCBzZXBhcmF0b3IgYXQgYmVnaW5uaW5nXG4gICAgICAgIGxldCBqID0gMjtcbiAgICAgICAgbGV0IGxhc3QgPSBqO1xuICAgICAgICAvLyBNYXRjaCAxIG9yIG1vcmUgbm9uLXBhdGggc2VwYXJhdG9yc1xuICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgfVxuICAgICAgICBpZiAoaiA8IGxlbiAmJiBqICE9PSBsYXN0KSB7XG4gICAgICAgICAgY29uc3QgZmlyc3RQYXJ0ID0gcGF0aC5zbGljZShsYXN0LCBqKTtcbiAgICAgICAgICAvLyBNYXRjaGVkIVxuICAgICAgICAgIGxhc3QgPSBqO1xuICAgICAgICAgIC8vIE1hdGNoIDEgb3IgbW9yZSBwYXRoIHNlcGFyYXRvcnNcbiAgICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgICBpZiAoIWlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgICB9XG4gICAgICAgICAgaWYgKGogPCBsZW4gJiYgaiAhPT0gbGFzdCkge1xuICAgICAgICAgICAgLy8gTWF0Y2hlZCFcbiAgICAgICAgICAgIGxhc3QgPSBqO1xuICAgICAgICAgICAgLy8gTWF0Y2ggMSBvciBtb3JlIG5vbi1wYXRoIHNlcGFyYXRvcnNcbiAgICAgICAgICAgIGZvciAoOyBqIDwgbGVuOyArK2opIHtcbiAgICAgICAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIGlmIChqID09PSBsZW4pIHtcbiAgICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCBhIFVOQyByb290IG9ubHlcbiAgICAgICAgICAgICAgLy8gUmV0dXJuIHRoZSBub3JtYWxpemVkIHZlcnNpb24gb2YgdGhlIFVOQyByb290IHNpbmNlIHRoZXJlXG4gICAgICAgICAgICAgIC8vIGlzIG5vdGhpbmcgbGVmdCB0byBwcm9jZXNzXG5cbiAgICAgICAgICAgICAgcmV0dXJuIGBcXFxcXFxcXCR7Zmlyc3RQYXJ0fVxcXFwke3BhdGguc2xpY2UobGFzdCl9XFxcXGA7XG4gICAgICAgICAgICB9IGVsc2UgaWYgKGogIT09IGxhc3QpIHtcbiAgICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCBhIFVOQyByb290IHdpdGggbGVmdG92ZXJzXG5cbiAgICAgICAgICAgICAgZGV2aWNlID0gYFxcXFxcXFxcJHtmaXJzdFBhcnR9XFxcXCR7cGF0aC5zbGljZShsYXN0LCBqKX1gO1xuICAgICAgICAgICAgICByb290RW5kID0gajtcbiAgICAgICAgICAgIH1cbiAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIHJvb3RFbmQgPSAxO1xuICAgICAgfVxuICAgIH0gZWxzZSBpZiAoaXNXaW5kb3dzRGV2aWNlUm9vdChjb2RlKSkge1xuICAgICAgLy8gUG9zc2libGUgZGV2aWNlIHJvb3RcblxuICAgICAgaWYgKHBhdGguY2hhckNvZGVBdCgxKSA9PT0gQ0hBUl9DT0xPTikge1xuICAgICAgICBkZXZpY2UgPSBwYXRoLnNsaWNlKDAsIDIpO1xuICAgICAgICByb290RW5kID0gMjtcbiAgICAgICAgaWYgKGxlbiA+IDIpIHtcbiAgICAgICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdCgyKSkpIHtcbiAgICAgICAgICAgIC8vIFRyZWF0IHNlcGFyYXRvciBmb2xsb3dpbmcgZHJpdmUgbmFtZSBhcyBhbiBhYnNvbHV0ZSBwYXRoXG4gICAgICAgICAgICAvLyBpbmRpY2F0b3JcbiAgICAgICAgICAgIGlzQWJzb2x1dGUgPSB0cnVlO1xuICAgICAgICAgICAgcm9vdEVuZCA9IDM7XG4gICAgICAgICAgfVxuICAgICAgICB9XG4gICAgICB9XG4gICAgfVxuICB9IGVsc2UgaWYgKGlzUGF0aFNlcGFyYXRvcihjb2RlKSkge1xuICAgIC8vIGBwYXRoYCBjb250YWlucyBqdXN0IGEgcGF0aCBzZXBhcmF0b3IsIGV4aXQgZWFybHkgdG8gYXZvaWQgdW5uZWNlc3NhcnlcbiAgICAvLyB3b3JrXG4gICAgcmV0dXJuIFwiXFxcXFwiO1xuICB9XG5cbiAgbGV0IHRhaWw6IHN0cmluZztcbiAgaWYgKHJvb3RFbmQgPCBsZW4pIHtcbiAgICB0YWlsID0gbm9ybWFsaXplU3RyaW5nKFxuICAgICAgcGF0aC5zbGljZShyb290RW5kKSxcbiAgICAgICFpc0Fic29sdXRlLFxuICAgICAgXCJcXFxcXCIsXG4gICAgICBpc1BhdGhTZXBhcmF0b3IsXG4gICAgKTtcbiAgfSBlbHNlIHtcbiAgICB0YWlsID0gXCJcIjtcbiAgfVxuICBpZiAodGFpbC5sZW5ndGggPT09IDAgJiYgIWlzQWJzb2x1dGUpIHRhaWwgPSBcIi5cIjtcbiAgaWYgKHRhaWwubGVuZ3RoID4gMCAmJiBpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KGxlbiAtIDEpKSkge1xuICAgIHRhaWwgKz0gXCJcXFxcXCI7XG4gIH1cbiAgaWYgKGRldmljZSA9PT0gdW5kZWZpbmVkKSB7XG4gICAgaWYgKGlzQWJzb2x1dGUpIHtcbiAgICAgIGlmICh0YWlsLmxlbmd0aCA+IDApIHJldHVybiBgXFxcXCR7dGFpbH1gO1xuICAgICAgZWxzZSByZXR1cm4gXCJcXFxcXCI7XG4gICAgfSBlbHNlIGlmICh0YWlsLmxlbmd0aCA+IDApIHtcbiAgICAgIHJldHVybiB0YWlsO1xuICAgIH0gZWxzZSB7XG4gICAgICByZXR1cm4gXCJcIjtcbiAgICB9XG4gIH0gZWxzZSBpZiAoaXNBYnNvbHV0ZSkge1xuICAgIGlmICh0YWlsLmxlbmd0aCA+IDApIHJldHVybiBgJHtkZXZpY2V9XFxcXCR7dGFpbH1gO1xuICAgIGVsc2UgcmV0dXJuIGAke2RldmljZX1cXFxcYDtcbiAgfSBlbHNlIGlmICh0YWlsLmxlbmd0aCA+IDApIHtcbiAgICByZXR1cm4gZGV2aWNlICsgdGFpbDtcbiAgfSBlbHNlIHtcbiAgICByZXR1cm4gZGV2aWNlO1xuICB9XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBpc0Fic29sdXRlKHBhdGg6IHN0cmluZyk6IGJvb2xlYW4ge1xuICBhc3NlcnRQYXRoKHBhdGgpO1xuICBjb25zdCBsZW4gPSBwYXRoLmxlbmd0aDtcbiAgaWYgKGxlbiA9PT0gMCkgcmV0dXJuIGZhbHNlO1xuXG4gIGNvbnN0IGNvZGUgPSBwYXRoLmNoYXJDb2RlQXQoMCk7XG4gIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICByZXR1cm4gdHJ1ZTtcbiAgfSBlbHNlIGlmIChpc1dpbmRvd3NEZXZpY2VSb290KGNvZGUpKSB7XG4gICAgLy8gUG9zc2libGUgZGV2aWNlIHJvb3RcblxuICAgIGlmIChsZW4gPiAyICYmIHBhdGguY2hhckNvZGVBdCgxKSA9PT0gQ0hBUl9DT0xPTikge1xuICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoMikpKSByZXR1cm4gdHJ1ZTtcbiAgICB9XG4gIH1cbiAgcmV0dXJuIGZhbHNlO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gam9pbiguLi5wYXRoczogc3RyaW5nW10pOiBzdHJpbmcge1xuICBjb25zdCBwYXRoc0NvdW50ID0gcGF0aHMubGVuZ3RoO1xuICBpZiAocGF0aHNDb3VudCA9PT0gMCkgcmV0dXJuIFwiLlwiO1xuXG4gIGxldCBqb2luZWQ6IHN0cmluZyB8IHVuZGVmaW5lZDtcbiAgbGV0IGZpcnN0UGFydDogc3RyaW5nIHwgbnVsbCA9IG51bGw7XG4gIGZvciAobGV0IGkgPSAwOyBpIDwgcGF0aHNDb3VudDsgKytpKSB7XG4gICAgY29uc3QgcGF0aCA9IHBhdGhzW2ldO1xuICAgIGFzc2VydFBhdGgocGF0aCk7XG4gICAgaWYgKHBhdGgubGVuZ3RoID4gMCkge1xuICAgICAgaWYgKGpvaW5lZCA9PT0gdW5kZWZpbmVkKSBqb2luZWQgPSBmaXJzdFBhcnQgPSBwYXRoO1xuICAgICAgZWxzZSBqb2luZWQgKz0gYFxcXFwke3BhdGh9YDtcbiAgICB9XG4gIH1cblxuICBpZiAoam9pbmVkID09PSB1bmRlZmluZWQpIHJldHVybiBcIi5cIjtcblxuICAvLyBNYWtlIHN1cmUgdGhhdCB0aGUgam9pbmVkIHBhdGggZG9lc24ndCBzdGFydCB3aXRoIHR3byBzbGFzaGVzLCBiZWNhdXNlXG4gIC8vIG5vcm1hbGl6ZSgpIHdpbGwgbWlzdGFrZSBpdCBmb3IgYW4gVU5DIHBhdGggdGhlbi5cbiAgLy9cbiAgLy8gVGhpcyBzdGVwIGlzIHNraXBwZWQgd2hlbiBpdCBpcyB2ZXJ5IGNsZWFyIHRoYXQgdGhlIHVzZXIgYWN0dWFsbHlcbiAgLy8gaW50ZW5kZWQgdG8gcG9pbnQgYXQgYW4gVU5DIHBhdGguIFRoaXMgaXMgYXNzdW1lZCB3aGVuIHRoZSBmaXJzdFxuICAvLyBub24tZW1wdHkgc3RyaW5nIGFyZ3VtZW50cyBzdGFydHMgd2l0aCBleGFjdGx5IHR3byBzbGFzaGVzIGZvbGxvd2VkIGJ5XG4gIC8vIGF0IGxlYXN0IG9uZSBtb3JlIG5vbi1zbGFzaCBjaGFyYWN0ZXIuXG4gIC8vXG4gIC8vIE5vdGUgdGhhdCBmb3Igbm9ybWFsaXplKCkgdG8gdHJlYXQgYSBwYXRoIGFzIGFuIFVOQyBwYXRoIGl0IG5lZWRzIHRvXG4gIC8vIGhhdmUgYXQgbGVhc3QgMiBjb21wb25lbnRzLCBzbyB3ZSBkb24ndCBmaWx0ZXIgZm9yIHRoYXQgaGVyZS5cbiAgLy8gVGhpcyBtZWFucyB0aGF0IHRoZSB1c2VyIGNhbiB1c2Ugam9pbiB0byBjb25zdHJ1Y3QgVU5DIHBhdGhzIGZyb21cbiAgLy8gYSBzZXJ2ZXIgbmFtZSBhbmQgYSBzaGFyZSBuYW1lOyBmb3IgZXhhbXBsZTpcbiAgLy8gICBwYXRoLmpvaW4oJy8vc2VydmVyJywgJ3NoYXJlJykgLT4gJ1xcXFxcXFxcc2VydmVyXFxcXHNoYXJlXFxcXCcpXG4gIGxldCBuZWVkc1JlcGxhY2UgPSB0cnVlO1xuICBsZXQgc2xhc2hDb3VudCA9IDA7XG4gIGFzc2VydChmaXJzdFBhcnQgIT0gbnVsbCk7XG4gIGlmIChpc1BhdGhTZXBhcmF0b3IoZmlyc3RQYXJ0LmNoYXJDb2RlQXQoMCkpKSB7XG4gICAgKytzbGFzaENvdW50O1xuICAgIGNvbnN0IGZpcnN0TGVuID0gZmlyc3RQYXJ0Lmxlbmd0aDtcbiAgICBpZiAoZmlyc3RMZW4gPiAxKSB7XG4gICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKGZpcnN0UGFydC5jaGFyQ29kZUF0KDEpKSkge1xuICAgICAgICArK3NsYXNoQ291bnQ7XG4gICAgICAgIGlmIChmaXJzdExlbiA+IDIpIHtcbiAgICAgICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKGZpcnN0UGFydC5jaGFyQ29kZUF0KDIpKSkgKytzbGFzaENvdW50O1xuICAgICAgICAgIGVsc2Uge1xuICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCBhIFVOQyBwYXRoIGluIHRoZSBmaXJzdCBwYXJ0XG4gICAgICAgICAgICBuZWVkc1JlcGxhY2UgPSBmYWxzZTtcbiAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH1cbiAgaWYgKG5lZWRzUmVwbGFjZSkge1xuICAgIC8vIEZpbmQgYW55IG1vcmUgY29uc2VjdXRpdmUgc2xhc2hlcyB3ZSBuZWVkIHRvIHJlcGxhY2VcbiAgICBmb3IgKDsgc2xhc2hDb3VudCA8IGpvaW5lZC5sZW5ndGg7ICsrc2xhc2hDb3VudCkge1xuICAgICAgaWYgKCFpc1BhdGhTZXBhcmF0b3Ioam9pbmVkLmNoYXJDb2RlQXQoc2xhc2hDb3VudCkpKSBicmVhaztcbiAgICB9XG5cbiAgICAvLyBSZXBsYWNlIHRoZSBzbGFzaGVzIGlmIG5lZWRlZFxuICAgIGlmIChzbGFzaENvdW50ID49IDIpIGpvaW5lZCA9IGBcXFxcJHtqb2luZWQuc2xpY2Uoc2xhc2hDb3VudCl9YDtcbiAgfVxuXG4gIHJldHVybiBub3JtYWxpemUoam9pbmVkKTtcbn1cblxuLy8gSXQgd2lsbCBzb2x2ZSB0aGUgcmVsYXRpdmUgcGF0aCBmcm9tIGBmcm9tYCB0byBgdG9gLCBmb3IgaW5zdGFuY2U6XG4vLyAgZnJvbSA9ICdDOlxcXFxvcmFuZGVhXFxcXHRlc3RcXFxcYWFhJ1xuLy8gIHRvID0gJ0M6XFxcXG9yYW5kZWFcXFxcaW1wbFxcXFxiYmInXG4vLyBUaGUgb3V0cHV0IG9mIHRoZSBmdW5jdGlvbiBzaG91bGQgYmU6ICcuLlxcXFwuLlxcXFxpbXBsXFxcXGJiYidcbmV4cG9ydCBmdW5jdGlvbiByZWxhdGl2ZShmcm9tOiBzdHJpbmcsIHRvOiBzdHJpbmcpOiBzdHJpbmcge1xuICBhc3NlcnRQYXRoKGZyb20pO1xuICBhc3NlcnRQYXRoKHRvKTtcblxuICBpZiAoZnJvbSA9PT0gdG8pIHJldHVybiBcIlwiO1xuXG4gIGNvbnN0IGZyb21PcmlnID0gcmVzb2x2ZShmcm9tKTtcbiAgY29uc3QgdG9PcmlnID0gcmVzb2x2ZSh0byk7XG5cbiAgaWYgKGZyb21PcmlnID09PSB0b09yaWcpIHJldHVybiBcIlwiO1xuXG4gIGZyb20gPSBmcm9tT3JpZy50b0xvd2VyQ2FzZSgpO1xuICB0byA9IHRvT3JpZy50b0xvd2VyQ2FzZSgpO1xuXG4gIGlmIChmcm9tID09PSB0bykgcmV0dXJuIFwiXCI7XG5cbiAgLy8gVHJpbSBhbnkgbGVhZGluZyBiYWNrc2xhc2hlc1xuICBsZXQgZnJvbVN0YXJ0ID0gMDtcbiAgbGV0IGZyb21FbmQgPSBmcm9tLmxlbmd0aDtcbiAgZm9yICg7IGZyb21TdGFydCA8IGZyb21FbmQ7ICsrZnJvbVN0YXJ0KSB7XG4gICAgaWYgKGZyb20uY2hhckNvZGVBdChmcm9tU3RhcnQpICE9PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSBicmVhaztcbiAgfVxuICAvLyBUcmltIHRyYWlsaW5nIGJhY2tzbGFzaGVzIChhcHBsaWNhYmxlIHRvIFVOQyBwYXRocyBvbmx5KVxuICBmb3IgKDsgZnJvbUVuZCAtIDEgPiBmcm9tU3RhcnQ7IC0tZnJvbUVuZCkge1xuICAgIGlmIChmcm9tLmNoYXJDb2RlQXQoZnJvbUVuZCAtIDEpICE9PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSBicmVhaztcbiAgfVxuICBjb25zdCBmcm9tTGVuID0gZnJvbUVuZCAtIGZyb21TdGFydDtcblxuICAvLyBUcmltIGFueSBsZWFkaW5nIGJhY2tzbGFzaGVzXG4gIGxldCB0b1N0YXJ0ID0gMDtcbiAgbGV0IHRvRW5kID0gdG8ubGVuZ3RoO1xuICBmb3IgKDsgdG9TdGFydCA8IHRvRW5kOyArK3RvU3RhcnQpIHtcbiAgICBpZiAodG8uY2hhckNvZGVBdCh0b1N0YXJ0KSAhPT0gQ0hBUl9CQUNLV0FSRF9TTEFTSCkgYnJlYWs7XG4gIH1cbiAgLy8gVHJpbSB0cmFpbGluZyBiYWNrc2xhc2hlcyAoYXBwbGljYWJsZSB0byBVTkMgcGF0aHMgb25seSlcbiAgZm9yICg7IHRvRW5kIC0gMSA+IHRvU3RhcnQ7IC0tdG9FbmQpIHtcbiAgICBpZiAodG8uY2hhckNvZGVBdCh0b0VuZCAtIDEpICE9PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSBicmVhaztcbiAgfVxuICBjb25zdCB0b0xlbiA9IHRvRW5kIC0gdG9TdGFydDtcblxuICAvLyBDb21wYXJlIHBhdGhzIHRvIGZpbmQgdGhlIGxvbmdlc3QgY29tbW9uIHBhdGggZnJvbSByb290XG4gIGNvbnN0IGxlbmd0aCA9IGZyb21MZW4gPCB0b0xlbiA/IGZyb21MZW4gOiB0b0xlbjtcbiAgbGV0IGxhc3RDb21tb25TZXAgPSAtMTtcbiAgbGV0IGkgPSAwO1xuICBmb3IgKDsgaSA8PSBsZW5ndGg7ICsraSkge1xuICAgIGlmIChpID09PSBsZW5ndGgpIHtcbiAgICAgIGlmICh0b0xlbiA+IGxlbmd0aCkge1xuICAgICAgICBpZiAodG8uY2hhckNvZGVBdCh0b1N0YXJ0ICsgaSkgPT09IENIQVJfQkFDS1dBUkRfU0xBU0gpIHtcbiAgICAgICAgICAvLyBXZSBnZXQgaGVyZSBpZiBgZnJvbWAgaXMgdGhlIGV4YWN0IGJhc2UgcGF0aCBmb3IgYHRvYC5cbiAgICAgICAgICAvLyBGb3IgZXhhbXBsZTogZnJvbT0nQzpcXFxcZm9vXFxcXGJhcic7IHRvPSdDOlxcXFxmb29cXFxcYmFyXFxcXGJheidcbiAgICAgICAgICByZXR1cm4gdG9PcmlnLnNsaWNlKHRvU3RhcnQgKyBpICsgMSk7XG4gICAgICAgIH0gZWxzZSBpZiAoaSA9PT0gMikge1xuICAgICAgICAgIC8vIFdlIGdldCBoZXJlIGlmIGBmcm9tYCBpcyB0aGUgZGV2aWNlIHJvb3QuXG4gICAgICAgICAgLy8gRm9yIGV4YW1wbGU6IGZyb209J0M6XFxcXCc7IHRvPSdDOlxcXFxmb28nXG4gICAgICAgICAgcmV0dXJuIHRvT3JpZy5zbGljZSh0b1N0YXJ0ICsgaSk7XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICAgIGlmIChmcm9tTGVuID4gbGVuZ3RoKSB7XG4gICAgICAgIGlmIChmcm9tLmNoYXJDb2RlQXQoZnJvbVN0YXJ0ICsgaSkgPT09IENIQVJfQkFDS1dBUkRfU0xBU0gpIHtcbiAgICAgICAgICAvLyBXZSBnZXQgaGVyZSBpZiBgdG9gIGlzIHRoZSBleGFjdCBiYXNlIHBhdGggZm9yIGBmcm9tYC5cbiAgICAgICAgICAvLyBGb3IgZXhhbXBsZTogZnJvbT0nQzpcXFxcZm9vXFxcXGJhcic7IHRvPSdDOlxcXFxmb28nXG4gICAgICAgICAgbGFzdENvbW1vblNlcCA9IGk7XG4gICAgICAgIH0gZWxzZSBpZiAoaSA9PT0gMikge1xuICAgICAgICAgIC8vIFdlIGdldCBoZXJlIGlmIGB0b2AgaXMgdGhlIGRldmljZSByb290LlxuICAgICAgICAgIC8vIEZvciBleGFtcGxlOiBmcm9tPSdDOlxcXFxmb29cXFxcYmFyJzsgdG89J0M6XFxcXCdcbiAgICAgICAgICBsYXN0Q29tbW9uU2VwID0gMztcbiAgICAgICAgfVxuICAgICAgfVxuICAgICAgYnJlYWs7XG4gICAgfVxuICAgIGNvbnN0IGZyb21Db2RlID0gZnJvbS5jaGFyQ29kZUF0KGZyb21TdGFydCArIGkpO1xuICAgIGNvbnN0IHRvQ29kZSA9IHRvLmNoYXJDb2RlQXQodG9TdGFydCArIGkpO1xuICAgIGlmIChmcm9tQ29kZSAhPT0gdG9Db2RlKSBicmVhaztcbiAgICBlbHNlIGlmIChmcm9tQ29kZSA9PT0gQ0hBUl9CQUNLV0FSRF9TTEFTSCkgbGFzdENvbW1vblNlcCA9IGk7XG4gIH1cblxuICAvLyBXZSBmb3VuZCBhIG1pc21hdGNoIGJlZm9yZSB0aGUgZmlyc3QgY29tbW9uIHBhdGggc2VwYXJhdG9yIHdhcyBzZWVuLCBzb1xuICAvLyByZXR1cm4gdGhlIG9yaWdpbmFsIGB0b2AuXG4gIGlmIChpICE9PSBsZW5ndGggJiYgbGFzdENvbW1vblNlcCA9PT0gLTEpIHtcbiAgICByZXR1cm4gdG9PcmlnO1xuICB9XG5cbiAgbGV0IG91dCA9IFwiXCI7XG4gIGlmIChsYXN0Q29tbW9uU2VwID09PSAtMSkgbGFzdENvbW1vblNlcCA9IDA7XG4gIC8vIEdlbmVyYXRlIHRoZSByZWxhdGl2ZSBwYXRoIGJhc2VkIG9uIHRoZSBwYXRoIGRpZmZlcmVuY2UgYmV0d2VlbiBgdG9gIGFuZFxuICAvLyBgZnJvbWBcbiAgZm9yIChpID0gZnJvbVN0YXJ0ICsgbGFzdENvbW1vblNlcCArIDE7IGkgPD0gZnJvbUVuZDsgKytpKSB7XG4gICAgaWYgKGkgPT09IGZyb21FbmQgfHwgZnJvbS5jaGFyQ29kZUF0KGkpID09PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSB7XG4gICAgICBpZiAob3V0Lmxlbmd0aCA9PT0gMCkgb3V0ICs9IFwiLi5cIjtcbiAgICAgIGVsc2Ugb3V0ICs9IFwiXFxcXC4uXCI7XG4gICAgfVxuICB9XG5cbiAgLy8gTGFzdGx5LCBhcHBlbmQgdGhlIHJlc3Qgb2YgdGhlIGRlc3RpbmF0aW9uIChgdG9gKSBwYXRoIHRoYXQgY29tZXMgYWZ0ZXJcbiAgLy8gdGhlIGNvbW1vbiBwYXRoIHBhcnRzXG4gIGlmIChvdXQubGVuZ3RoID4gMCkge1xuICAgIHJldHVybiBvdXQgKyB0b09yaWcuc2xpY2UodG9TdGFydCArIGxhc3RDb21tb25TZXAsIHRvRW5kKTtcbiAgfSBlbHNlIHtcbiAgICB0b1N0YXJ0ICs9IGxhc3RDb21tb25TZXA7XG4gICAgaWYgKHRvT3JpZy5jaGFyQ29kZUF0KHRvU3RhcnQpID09PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSArK3RvU3RhcnQ7XG4gICAgcmV0dXJuIHRvT3JpZy5zbGljZSh0b1N0YXJ0LCB0b0VuZCk7XG4gIH1cbn1cblxuZXhwb3J0IGZ1bmN0aW9uIHRvTmFtZXNwYWNlZFBhdGgocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgLy8gTm90ZTogdGhpcyB3aWxsICpwcm9iYWJseSogdGhyb3cgc29tZXdoZXJlLlxuICBpZiAodHlwZW9mIHBhdGggIT09IFwic3RyaW5nXCIpIHJldHVybiBwYXRoO1xuICBpZiAocGF0aC5sZW5ndGggPT09IDApIHJldHVybiBcIlwiO1xuXG4gIGNvbnN0IHJlc29sdmVkUGF0aCA9IHJlc29sdmUocGF0aCk7XG5cbiAgaWYgKHJlc29sdmVkUGF0aC5sZW5ndGggPj0gMykge1xuICAgIGlmIChyZXNvbHZlZFBhdGguY2hhckNvZGVBdCgwKSA9PT0gQ0hBUl9CQUNLV0FSRF9TTEFTSCkge1xuICAgICAgLy8gUG9zc2libGUgVU5DIHJvb3RcblxuICAgICAgaWYgKHJlc29sdmVkUGF0aC5jaGFyQ29kZUF0KDEpID09PSBDSEFSX0JBQ0tXQVJEX1NMQVNIKSB7XG4gICAgICAgIGNvbnN0IGNvZGUgPSByZXNvbHZlZFBhdGguY2hhckNvZGVBdCgyKTtcbiAgICAgICAgaWYgKGNvZGUgIT09IENIQVJfUVVFU1RJT05fTUFSSyAmJiBjb2RlICE9PSBDSEFSX0RPVCkge1xuICAgICAgICAgIC8vIE1hdGNoZWQgbm9uLWxvbmcgVU5DIHJvb3QsIGNvbnZlcnQgdGhlIHBhdGggdG8gYSBsb25nIFVOQyBwYXRoXG4gICAgICAgICAgcmV0dXJuIGBcXFxcXFxcXD9cXFxcVU5DXFxcXCR7cmVzb2x2ZWRQYXRoLnNsaWNlKDIpfWA7XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9IGVsc2UgaWYgKGlzV2luZG93c0RldmljZVJvb3QocmVzb2x2ZWRQYXRoLmNoYXJDb2RlQXQoMCkpKSB7XG4gICAgICAvLyBQb3NzaWJsZSBkZXZpY2Ugcm9vdFxuXG4gICAgICBpZiAoXG4gICAgICAgIHJlc29sdmVkUGF0aC5jaGFyQ29kZUF0KDEpID09PSBDSEFSX0NPTE9OICYmXG4gICAgICAgIHJlc29sdmVkUGF0aC5jaGFyQ29kZUF0KDIpID09PSBDSEFSX0JBQ0tXQVJEX1NMQVNIXG4gICAgICApIHtcbiAgICAgICAgLy8gTWF0Y2hlZCBkZXZpY2Ugcm9vdCwgY29udmVydCB0aGUgcGF0aCB0byBhIGxvbmcgVU5DIHBhdGhcbiAgICAgICAgcmV0dXJuIGBcXFxcXFxcXD9cXFxcJHtyZXNvbHZlZFBhdGh9YDtcbiAgICAgIH1cbiAgICB9XG4gIH1cblxuICByZXR1cm4gcGF0aDtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGRpcm5hbWUocGF0aDogc3RyaW5nKTogc3RyaW5nIHtcbiAgYXNzZXJ0UGF0aChwYXRoKTtcbiAgY29uc3QgbGVuID0gcGF0aC5sZW5ndGg7XG4gIGlmIChsZW4gPT09IDApIHJldHVybiBcIi5cIjtcbiAgbGV0IHJvb3RFbmQgPSAtMTtcbiAgbGV0IGVuZCA9IC0xO1xuICBsZXQgbWF0Y2hlZFNsYXNoID0gdHJ1ZTtcbiAgbGV0IG9mZnNldCA9IDA7XG4gIGNvbnN0IGNvZGUgPSBwYXRoLmNoYXJDb2RlQXQoMCk7XG5cbiAgLy8gVHJ5IHRvIG1hdGNoIGEgcm9vdFxuICBpZiAobGVuID4gMSkge1xuICAgIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAgIC8vIFBvc3NpYmxlIFVOQyByb290XG5cbiAgICAgIHJvb3RFbmQgPSBvZmZzZXQgPSAxO1xuXG4gICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdCgxKSkpIHtcbiAgICAgICAgLy8gTWF0Y2hlZCBkb3VibGUgcGF0aCBzZXBhcmF0b3IgYXQgYmVnaW5uaW5nXG4gICAgICAgIGxldCBqID0gMjtcbiAgICAgICAgbGV0IGxhc3QgPSBqO1xuICAgICAgICAvLyBNYXRjaCAxIG9yIG1vcmUgbm9uLXBhdGggc2VwYXJhdG9yc1xuICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgfVxuICAgICAgICBpZiAoaiA8IGxlbiAmJiBqICE9PSBsYXN0KSB7XG4gICAgICAgICAgLy8gTWF0Y2hlZCFcbiAgICAgICAgICBsYXN0ID0gajtcbiAgICAgICAgICAvLyBNYXRjaCAxIG9yIG1vcmUgcGF0aCBzZXBhcmF0b3JzXG4gICAgICAgICAgZm9yICg7IGogPCBsZW47ICsraikge1xuICAgICAgICAgICAgaWYgKCFpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KGopKSkgYnJlYWs7XG4gICAgICAgICAgfVxuICAgICAgICAgIGlmIChqIDwgbGVuICYmIGogIT09IGxhc3QpIHtcbiAgICAgICAgICAgIC8vIE1hdGNoZWQhXG4gICAgICAgICAgICBsYXN0ID0gajtcbiAgICAgICAgICAgIC8vIE1hdGNoIDEgb3IgbW9yZSBub24tcGF0aCBzZXBhcmF0b3JzXG4gICAgICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KGopKSkgYnJlYWs7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBpZiAoaiA9PT0gbGVuKSB7XG4gICAgICAgICAgICAgIC8vIFdlIG1hdGNoZWQgYSBVTkMgcm9vdCBvbmx5XG4gICAgICAgICAgICAgIHJldHVybiBwYXRoO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgaWYgKGogIT09IGxhc3QpIHtcbiAgICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCBhIFVOQyByb290IHdpdGggbGVmdG92ZXJzXG5cbiAgICAgICAgICAgICAgLy8gT2Zmc2V0IGJ5IDEgdG8gaW5jbHVkZSB0aGUgc2VwYXJhdG9yIGFmdGVyIHRoZSBVTkMgcm9vdCB0b1xuICAgICAgICAgICAgICAvLyB0cmVhdCBpdCBhcyBhIFwibm9ybWFsIHJvb3RcIiBvbiB0b3Agb2YgYSAoVU5DKSByb290XG4gICAgICAgICAgICAgIHJvb3RFbmQgPSBvZmZzZXQgPSBqICsgMTtcbiAgICAgICAgICAgIH1cbiAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9IGVsc2UgaWYgKGlzV2luZG93c0RldmljZVJvb3QoY29kZSkpIHtcbiAgICAgIC8vIFBvc3NpYmxlIGRldmljZSByb290XG5cbiAgICAgIGlmIChwYXRoLmNoYXJDb2RlQXQoMSkgPT09IENIQVJfQ09MT04pIHtcbiAgICAgICAgcm9vdEVuZCA9IG9mZnNldCA9IDI7XG4gICAgICAgIGlmIChsZW4gPiAyKSB7XG4gICAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoMikpKSByb290RW5kID0gb2Zmc2V0ID0gMztcbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cbiAgfSBlbHNlIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAvLyBgcGF0aGAgY29udGFpbnMganVzdCBhIHBhdGggc2VwYXJhdG9yLCBleGl0IGVhcmx5IHRvIGF2b2lkXG4gICAgLy8gdW5uZWNlc3Nhcnkgd29ya1xuICAgIHJldHVybiBwYXRoO1xuICB9XG5cbiAgZm9yIChsZXQgaSA9IGxlbiAtIDE7IGkgPj0gb2Zmc2V0OyAtLWkpIHtcbiAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdChpKSkpIHtcbiAgICAgIGlmICghbWF0Y2hlZFNsYXNoKSB7XG4gICAgICAgIGVuZCA9IGk7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuICAgIH0gZWxzZSB7XG4gICAgICAvLyBXZSBzYXcgdGhlIGZpcnN0IG5vbi1wYXRoIHNlcGFyYXRvclxuICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgfVxuICB9XG5cbiAgaWYgKGVuZCA9PT0gLTEpIHtcbiAgICBpZiAocm9vdEVuZCA9PT0gLTEpIHJldHVybiBcIi5cIjtcbiAgICBlbHNlIGVuZCA9IHJvb3RFbmQ7XG4gIH1cbiAgcmV0dXJuIHBhdGguc2xpY2UoMCwgZW5kKTtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGJhc2VuYW1lKHBhdGg6IHN0cmluZywgZXh0ID0gXCJcIik6IHN0cmluZyB7XG4gIGlmIChleHQgIT09IHVuZGVmaW5lZCAmJiB0eXBlb2YgZXh0ICE9PSBcInN0cmluZ1wiKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcignXCJleHRcIiBhcmd1bWVudCBtdXN0IGJlIGEgc3RyaW5nJyk7XG4gIH1cblxuICBhc3NlcnRQYXRoKHBhdGgpO1xuXG4gIGxldCBzdGFydCA9IDA7XG4gIGxldCBlbmQgPSAtMTtcbiAgbGV0IG1hdGNoZWRTbGFzaCA9IHRydWU7XG4gIGxldCBpOiBudW1iZXI7XG5cbiAgLy8gQ2hlY2sgZm9yIGEgZHJpdmUgbGV0dGVyIHByZWZpeCBzbyBhcyBub3QgdG8gbWlzdGFrZSB0aGUgZm9sbG93aW5nXG4gIC8vIHBhdGggc2VwYXJhdG9yIGFzIGFuIGV4dHJhIHNlcGFyYXRvciBhdCB0aGUgZW5kIG9mIHRoZSBwYXRoIHRoYXQgY2FuIGJlXG4gIC8vIGRpc3JlZ2FyZGVkXG4gIGlmIChwYXRoLmxlbmd0aCA+PSAyKSB7XG4gICAgY29uc3QgZHJpdmUgPSBwYXRoLmNoYXJDb2RlQXQoMCk7XG4gICAgaWYgKGlzV2luZG93c0RldmljZVJvb3QoZHJpdmUpKSB7XG4gICAgICBpZiAocGF0aC5jaGFyQ29kZUF0KDEpID09PSBDSEFSX0NPTE9OKSBzdGFydCA9IDI7XG4gICAgfVxuICB9XG5cbiAgaWYgKGV4dCAhPT0gdW5kZWZpbmVkICYmIGV4dC5sZW5ndGggPiAwICYmIGV4dC5sZW5ndGggPD0gcGF0aC5sZW5ndGgpIHtcbiAgICBpZiAoZXh0Lmxlbmd0aCA9PT0gcGF0aC5sZW5ndGggJiYgZXh0ID09PSBwYXRoKSByZXR1cm4gXCJcIjtcbiAgICBsZXQgZXh0SWR4ID0gZXh0Lmxlbmd0aCAtIDE7XG4gICAgbGV0IGZpcnN0Tm9uU2xhc2hFbmQgPSAtMTtcbiAgICBmb3IgKGkgPSBwYXRoLmxlbmd0aCAtIDE7IGkgPj0gc3RhcnQ7IC0taSkge1xuICAgICAgY29uc3QgY29kZSA9IHBhdGguY2hhckNvZGVBdChpKTtcbiAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAgICAgLy8gSWYgd2UgcmVhY2hlZCBhIHBhdGggc2VwYXJhdG9yIHRoYXQgd2FzIG5vdCBwYXJ0IG9mIGEgc2V0IG9mIHBhdGhcbiAgICAgICAgLy8gc2VwYXJhdG9ycyBhdCB0aGUgZW5kIG9mIHRoZSBzdHJpbmcsIHN0b3Agbm93XG4gICAgICAgIGlmICghbWF0Y2hlZFNsYXNoKSB7XG4gICAgICAgICAgc3RhcnQgPSBpICsgMTtcbiAgICAgICAgICBicmVhaztcbiAgICAgICAgfVxuICAgICAgfSBlbHNlIHtcbiAgICAgICAgaWYgKGZpcnN0Tm9uU2xhc2hFbmQgPT09IC0xKSB7XG4gICAgICAgICAgLy8gV2Ugc2F3IHRoZSBmaXJzdCBub24tcGF0aCBzZXBhcmF0b3IsIHJlbWVtYmVyIHRoaXMgaW5kZXggaW4gY2FzZVxuICAgICAgICAgIC8vIHdlIG5lZWQgaXQgaWYgdGhlIGV4dGVuc2lvbiBlbmRzIHVwIG5vdCBtYXRjaGluZ1xuICAgICAgICAgIG1hdGNoZWRTbGFzaCA9IGZhbHNlO1xuICAgICAgICAgIGZpcnN0Tm9uU2xhc2hFbmQgPSBpICsgMTtcbiAgICAgICAgfVxuICAgICAgICBpZiAoZXh0SWR4ID49IDApIHtcbiAgICAgICAgICAvLyBUcnkgdG8gbWF0Y2ggdGhlIGV4cGxpY2l0IGV4dGVuc2lvblxuICAgICAgICAgIGlmIChjb2RlID09PSBleHQuY2hhckNvZGVBdChleHRJZHgpKSB7XG4gICAgICAgICAgICBpZiAoLS1leHRJZHggPT09IC0xKSB7XG4gICAgICAgICAgICAgIC8vIFdlIG1hdGNoZWQgdGhlIGV4dGVuc2lvbiwgc28gbWFyayB0aGlzIGFzIHRoZSBlbmQgb2Ygb3VyIHBhdGhcbiAgICAgICAgICAgICAgLy8gY29tcG9uZW50XG4gICAgICAgICAgICAgIGVuZCA9IGk7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgIC8vIEV4dGVuc2lvbiBkb2VzIG5vdCBtYXRjaCwgc28gb3VyIHJlc3VsdCBpcyB0aGUgZW50aXJlIHBhdGhcbiAgICAgICAgICAgIC8vIGNvbXBvbmVudFxuICAgICAgICAgICAgZXh0SWR4ID0gLTE7XG4gICAgICAgICAgICBlbmQgPSBmaXJzdE5vblNsYXNoRW5kO1xuICAgICAgICAgIH1cbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cblxuICAgIGlmIChzdGFydCA9PT0gZW5kKSBlbmQgPSBmaXJzdE5vblNsYXNoRW5kO1xuICAgIGVsc2UgaWYgKGVuZCA9PT0gLTEpIGVuZCA9IHBhdGgubGVuZ3RoO1xuICAgIHJldHVybiBwYXRoLnNsaWNlKHN0YXJ0LCBlbmQpO1xuICB9IGVsc2Uge1xuICAgIGZvciAoaSA9IHBhdGgubGVuZ3RoIC0gMTsgaSA+PSBzdGFydDsgLS1pKSB7XG4gICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdChpKSkpIHtcbiAgICAgICAgLy8gSWYgd2UgcmVhY2hlZCBhIHBhdGggc2VwYXJhdG9yIHRoYXQgd2FzIG5vdCBwYXJ0IG9mIGEgc2V0IG9mIHBhdGhcbiAgICAgICAgLy8gc2VwYXJhdG9ycyBhdCB0aGUgZW5kIG9mIHRoZSBzdHJpbmcsIHN0b3Agbm93XG4gICAgICAgIGlmICghbWF0Y2hlZFNsYXNoKSB7XG4gICAgICAgICAgc3RhcnQgPSBpICsgMTtcbiAgICAgICAgICBicmVhaztcbiAgICAgICAgfVxuICAgICAgfSBlbHNlIGlmIChlbmQgPT09IC0xKSB7XG4gICAgICAgIC8vIFdlIHNhdyB0aGUgZmlyc3Qgbm9uLXBhdGggc2VwYXJhdG9yLCBtYXJrIHRoaXMgYXMgdGhlIGVuZCBvZiBvdXJcbiAgICAgICAgLy8gcGF0aCBjb21wb25lbnRcbiAgICAgICAgbWF0Y2hlZFNsYXNoID0gZmFsc2U7XG4gICAgICAgIGVuZCA9IGkgKyAxO1xuICAgICAgfVxuICAgIH1cblxuICAgIGlmIChlbmQgPT09IC0xKSByZXR1cm4gXCJcIjtcbiAgICByZXR1cm4gcGF0aC5zbGljZShzdGFydCwgZW5kKTtcbiAgfVxufVxuXG5leHBvcnQgZnVuY3Rpb24gZXh0bmFtZShwYXRoOiBzdHJpbmcpOiBzdHJpbmcge1xuICBhc3NlcnRQYXRoKHBhdGgpO1xuICBsZXQgc3RhcnQgPSAwO1xuICBsZXQgc3RhcnREb3QgPSAtMTtcbiAgbGV0IHN0YXJ0UGFydCA9IDA7XG4gIGxldCBlbmQgPSAtMTtcbiAgbGV0IG1hdGNoZWRTbGFzaCA9IHRydWU7XG4gIC8vIFRyYWNrIHRoZSBzdGF0ZSBvZiBjaGFyYWN0ZXJzIChpZiBhbnkpIHdlIHNlZSBiZWZvcmUgb3VyIGZpcnN0IGRvdCBhbmRcbiAgLy8gYWZ0ZXIgYW55IHBhdGggc2VwYXJhdG9yIHdlIGZpbmRcbiAgbGV0IHByZURvdFN0YXRlID0gMDtcblxuICAvLyBDaGVjayBmb3IgYSBkcml2ZSBsZXR0ZXIgcHJlZml4IHNvIGFzIG5vdCB0byBtaXN0YWtlIHRoZSBmb2xsb3dpbmdcbiAgLy8gcGF0aCBzZXBhcmF0b3IgYXMgYW4gZXh0cmEgc2VwYXJhdG9yIGF0IHRoZSBlbmQgb2YgdGhlIHBhdGggdGhhdCBjYW4gYmVcbiAgLy8gZGlzcmVnYXJkZWRcblxuICBpZiAoXG4gICAgcGF0aC5sZW5ndGggPj0gMiAmJlxuICAgIHBhdGguY2hhckNvZGVBdCgxKSA9PT0gQ0hBUl9DT0xPTiAmJlxuICAgIGlzV2luZG93c0RldmljZVJvb3QocGF0aC5jaGFyQ29kZUF0KDApKVxuICApIHtcbiAgICBzdGFydCA9IHN0YXJ0UGFydCA9IDI7XG4gIH1cblxuICBmb3IgKGxldCBpID0gcGF0aC5sZW5ndGggLSAxOyBpID49IHN0YXJ0OyAtLWkpIHtcbiAgICBjb25zdCBjb2RlID0gcGF0aC5jaGFyQ29kZUF0KGkpO1xuICAgIGlmIChpc1BhdGhTZXBhcmF0b3IoY29kZSkpIHtcbiAgICAgIC8vIElmIHdlIHJlYWNoZWQgYSBwYXRoIHNlcGFyYXRvciB0aGF0IHdhcyBub3QgcGFydCBvZiBhIHNldCBvZiBwYXRoXG4gICAgICAvLyBzZXBhcmF0b3JzIGF0IHRoZSBlbmQgb2YgdGhlIHN0cmluZywgc3RvcCBub3dcbiAgICAgIGlmICghbWF0Y2hlZFNsYXNoKSB7XG4gICAgICAgIHN0YXJ0UGFydCA9IGkgKyAxO1xuICAgICAgICBicmVhaztcbiAgICAgIH1cbiAgICAgIGNvbnRpbnVlO1xuICAgIH1cbiAgICBpZiAoZW5kID09PSAtMSkge1xuICAgICAgLy8gV2Ugc2F3IHRoZSBmaXJzdCBub24tcGF0aCBzZXBhcmF0b3IsIG1hcmsgdGhpcyBhcyB0aGUgZW5kIG9mIG91clxuICAgICAgLy8gZXh0ZW5zaW9uXG4gICAgICBtYXRjaGVkU2xhc2ggPSBmYWxzZTtcbiAgICAgIGVuZCA9IGkgKyAxO1xuICAgIH1cbiAgICBpZiAoY29kZSA9PT0gQ0hBUl9ET1QpIHtcbiAgICAgIC8vIElmIHRoaXMgaXMgb3VyIGZpcnN0IGRvdCwgbWFyayBpdCBhcyB0aGUgc3RhcnQgb2Ygb3VyIGV4dGVuc2lvblxuICAgICAgaWYgKHN0YXJ0RG90ID09PSAtMSkgc3RhcnREb3QgPSBpO1xuICAgICAgZWxzZSBpZiAocHJlRG90U3RhdGUgIT09IDEpIHByZURvdFN0YXRlID0gMTtcbiAgICB9IGVsc2UgaWYgKHN0YXJ0RG90ICE9PSAtMSkge1xuICAgICAgLy8gV2Ugc2F3IGEgbm9uLWRvdCBhbmQgbm9uLXBhdGggc2VwYXJhdG9yIGJlZm9yZSBvdXIgZG90LCBzbyB3ZSBzaG91bGRcbiAgICAgIC8vIGhhdmUgYSBnb29kIGNoYW5jZSBhdCBoYXZpbmcgYSBub24tZW1wdHkgZXh0ZW5zaW9uXG4gICAgICBwcmVEb3RTdGF0ZSA9IC0xO1xuICAgIH1cbiAgfVxuXG4gIGlmIChcbiAgICBzdGFydERvdCA9PT0gLTEgfHxcbiAgICBlbmQgPT09IC0xIHx8XG4gICAgLy8gV2Ugc2F3IGEgbm9uLWRvdCBjaGFyYWN0ZXIgaW1tZWRpYXRlbHkgYmVmb3JlIHRoZSBkb3RcbiAgICBwcmVEb3RTdGF0ZSA9PT0gMCB8fFxuICAgIC8vIFRoZSAocmlnaHQtbW9zdCkgdHJpbW1lZCBwYXRoIGNvbXBvbmVudCBpcyBleGFjdGx5ICcuLidcbiAgICAocHJlRG90U3RhdGUgPT09IDEgJiYgc3RhcnREb3QgPT09IGVuZCAtIDEgJiYgc3RhcnREb3QgPT09IHN0YXJ0UGFydCArIDEpXG4gICkge1xuICAgIHJldHVybiBcIlwiO1xuICB9XG4gIHJldHVybiBwYXRoLnNsaWNlKHN0YXJ0RG90LCBlbmQpO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gZm9ybWF0KHBhdGhPYmplY3Q6IEZvcm1hdElucHV0UGF0aE9iamVjdCk6IHN0cmluZyB7XG4gIC8qIGVzbGludC1kaXNhYmxlIG1heC1sZW4gKi9cbiAgaWYgKHBhdGhPYmplY3QgPT09IG51bGwgfHwgdHlwZW9mIHBhdGhPYmplY3QgIT09IFwib2JqZWN0XCIpIHtcbiAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFxuICAgICAgYFRoZSBcInBhdGhPYmplY3RcIiBhcmd1bWVudCBtdXN0IGJlIG9mIHR5cGUgT2JqZWN0LiBSZWNlaXZlZCB0eXBlICR7dHlwZW9mIHBhdGhPYmplY3R9YCxcbiAgICApO1xuICB9XG4gIHJldHVybiBfZm9ybWF0KFwiXFxcXFwiLCBwYXRoT2JqZWN0KTtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIHBhcnNlKHBhdGg6IHN0cmluZyk6IFBhcnNlZFBhdGgge1xuICBhc3NlcnRQYXRoKHBhdGgpO1xuXG4gIGNvbnN0IHJldDogUGFyc2VkUGF0aCA9IHsgcm9vdDogXCJcIiwgZGlyOiBcIlwiLCBiYXNlOiBcIlwiLCBleHQ6IFwiXCIsIG5hbWU6IFwiXCIgfTtcblxuICBjb25zdCBsZW4gPSBwYXRoLmxlbmd0aDtcbiAgaWYgKGxlbiA9PT0gMCkgcmV0dXJuIHJldDtcblxuICBsZXQgcm9vdEVuZCA9IDA7XG4gIGxldCBjb2RlID0gcGF0aC5jaGFyQ29kZUF0KDApO1xuXG4gIC8vIFRyeSB0byBtYXRjaCBhIHJvb3RcbiAgaWYgKGxlbiA+IDEpIHtcbiAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKGNvZGUpKSB7XG4gICAgICAvLyBQb3NzaWJsZSBVTkMgcm9vdFxuXG4gICAgICByb290RW5kID0gMTtcbiAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KDEpKSkge1xuICAgICAgICAvLyBNYXRjaGVkIGRvdWJsZSBwYXRoIHNlcGFyYXRvciBhdCBiZWdpbm5pbmdcbiAgICAgICAgbGV0IGogPSAyO1xuICAgICAgICBsZXQgbGFzdCA9IGo7XG4gICAgICAgIC8vIE1hdGNoIDEgb3IgbW9yZSBub24tcGF0aCBzZXBhcmF0b3JzXG4gICAgICAgIGZvciAoOyBqIDwgbGVuOyArK2opIHtcbiAgICAgICAgICBpZiAoaXNQYXRoU2VwYXJhdG9yKHBhdGguY2hhckNvZGVBdChqKSkpIGJyZWFrO1xuICAgICAgICB9XG4gICAgICAgIGlmIChqIDwgbGVuICYmIGogIT09IGxhc3QpIHtcbiAgICAgICAgICAvLyBNYXRjaGVkIVxuICAgICAgICAgIGxhc3QgPSBqO1xuICAgICAgICAgIC8vIE1hdGNoIDEgb3IgbW9yZSBwYXRoIHNlcGFyYXRvcnNcbiAgICAgICAgICBmb3IgKDsgaiA8IGxlbjsgKytqKSB7XG4gICAgICAgICAgICBpZiAoIWlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgICB9XG4gICAgICAgICAgaWYgKGogPCBsZW4gJiYgaiAhPT0gbGFzdCkge1xuICAgICAgICAgICAgLy8gTWF0Y2hlZCFcbiAgICAgICAgICAgIGxhc3QgPSBqO1xuICAgICAgICAgICAgLy8gTWF0Y2ggMSBvciBtb3JlIG5vbi1wYXRoIHNlcGFyYXRvcnNcbiAgICAgICAgICAgIGZvciAoOyBqIDwgbGVuOyArK2opIHtcbiAgICAgICAgICAgICAgaWYgKGlzUGF0aFNlcGFyYXRvcihwYXRoLmNoYXJDb2RlQXQoaikpKSBicmVhaztcbiAgICAgICAgICAgIH1cbiAgICAgICAgICAgIGlmIChqID09PSBsZW4pIHtcbiAgICAgICAgICAgICAgLy8gV2UgbWF0Y2hlZCBhIFVOQyByb290IG9ubHlcblxuICAgICAgICAgICAgICByb290RW5kID0gajtcbiAgICAgICAgICAgIH0gZWxzZSBpZiAoaiAhPT0gbGFzdCkge1xuICAgICAgICAgICAgICAvLyBXZSBtYXRjaGVkIGEgVU5DIHJvb3Qgd2l0aCBsZWZ0b3ZlcnNcblxuICAgICAgICAgICAgICByb290RW5kID0gaiArIDE7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgfVxuICAgICAgICB9XG4gICAgICB9XG4gICAgfSBlbHNlIGlmIChpc1dpbmRvd3NEZXZpY2VSb290KGNvZGUpKSB7XG4gICAgICAvLyBQb3NzaWJsZSBkZXZpY2Ugcm9vdFxuXG4gICAgICBpZiAocGF0aC5jaGFyQ29kZUF0KDEpID09PSBDSEFSX0NPTE9OKSB7XG4gICAgICAgIHJvb3RFbmQgPSAyO1xuICAgICAgICBpZiAobGVuID4gMikge1xuICAgICAgICAgIGlmIChpc1BhdGhTZXBhcmF0b3IocGF0aC5jaGFyQ29kZUF0KDIpKSkge1xuICAgICAgICAgICAgaWYgKGxlbiA9PT0gMykge1xuICAgICAgICAgICAgICAvLyBgcGF0aGAgY29udGFpbnMganVzdCBhIGRyaXZlIHJvb3QsIGV4aXQgZWFybHkgdG8gYXZvaWRcbiAgICAgICAgICAgICAgLy8gdW5uZWNlc3Nhcnkgd29ya1xuICAgICAgICAgICAgICByZXQucm9vdCA9IHJldC5kaXIgPSBwYXRoO1xuICAgICAgICAgICAgICByZXR1cm4gcmV0O1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgcm9vdEVuZCA9IDM7XG4gICAgICAgICAgfVxuICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgIC8vIGBwYXRoYCBjb250YWlucyBqdXN0IGEgZHJpdmUgcm9vdCwgZXhpdCBlYXJseSB0byBhdm9pZFxuICAgICAgICAgIC8vIHVubmVjZXNzYXJ5IHdvcmtcbiAgICAgICAgICByZXQucm9vdCA9IHJldC5kaXIgPSBwYXRoO1xuICAgICAgICAgIHJldHVybiByZXQ7XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH0gZWxzZSBpZiAoaXNQYXRoU2VwYXJhdG9yKGNvZGUpKSB7XG4gICAgLy8gYHBhdGhgIGNvbnRhaW5zIGp1c3QgYSBwYXRoIHNlcGFyYXRvciwgZXhpdCBlYXJseSB0byBhdm9pZFxuICAgIC8vIHVubmVjZXNzYXJ5IHdvcmtcbiAgICByZXQucm9vdCA9IHJldC5kaXIgPSBwYXRoO1xuICAgIHJldHVybiByZXQ7XG4gIH1cblxuICBpZiAocm9vdEVuZCA+IDApIHJldC5yb290ID0gcGF0aC5zbGljZSgwLCByb290RW5kKTtcblxuICBsZXQgc3RhcnREb3QgPSAtMTtcbiAgbGV0IHN0YXJ0UGFydCA9IHJvb3RFbmQ7XG4gIGxldCBlbmQgPSAtMTtcbiAgbGV0IG1hdGNoZWRTbGFzaCA9IHRydWU7XG4gIGxldCBpID0gcGF0aC5sZW5ndGggLSAxO1xuXG4gIC8vIFRyYWNrIHRoZSBzdGF0ZSBvZiBjaGFyYWN0ZXJzIChpZiBhbnkpIHdlIHNlZSBiZWZvcmUgb3VyIGZpcnN0IGRvdCBhbmRcbiAgLy8gYWZ0ZXIgYW55IHBhdGggc2VwYXJhdG9yIHdlIGZpbmRcbiAgbGV0IHByZURvdFN0YXRlID0gMDtcblxuICAvLyBHZXQgbm9uLWRpciBpbmZvXG4gIGZvciAoOyBpID49IHJvb3RFbmQ7IC0taSkge1xuICAgIGNvZGUgPSBwYXRoLmNoYXJDb2RlQXQoaSk7XG4gICAgaWYgKGlzUGF0aFNlcGFyYXRvcihjb2RlKSkge1xuICAgICAgLy8gSWYgd2UgcmVhY2hlZCBhIHBhdGggc2VwYXJhdG9yIHRoYXQgd2FzIG5vdCBwYXJ0IG9mIGEgc2V0IG9mIHBhdGhcbiAgICAgIC8vIHNlcGFyYXRvcnMgYXQgdGhlIGVuZCBvZiB0aGUgc3RyaW5nLCBzdG9wIG5vd1xuICAgICAgaWYgKCFtYXRjaGVkU2xhc2gpIHtcbiAgICAgICAgc3RhcnRQYXJ0ID0gaSArIDE7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuICAgICAgY29udGludWU7XG4gICAgfVxuICAgIGlmIChlbmQgPT09IC0xKSB7XG4gICAgICAvLyBXZSBzYXcgdGhlIGZpcnN0IG5vbi1wYXRoIHNlcGFyYXRvciwgbWFyayB0aGlzIGFzIHRoZSBlbmQgb2Ygb3VyXG4gICAgICAvLyBleHRlbnNpb25cbiAgICAgIG1hdGNoZWRTbGFzaCA9IGZhbHNlO1xuICAgICAgZW5kID0gaSArIDE7XG4gICAgfVxuICAgIGlmIChjb2RlID09PSBDSEFSX0RPVCkge1xuICAgICAgLy8gSWYgdGhpcyBpcyBvdXIgZmlyc3QgZG90LCBtYXJrIGl0IGFzIHRoZSBzdGFydCBvZiBvdXIgZXh0ZW5zaW9uXG4gICAgICBpZiAoc3RhcnREb3QgPT09IC0xKSBzdGFydERvdCA9IGk7XG4gICAgICBlbHNlIGlmIChwcmVEb3RTdGF0ZSAhPT0gMSkgcHJlRG90U3RhdGUgPSAxO1xuICAgIH0gZWxzZSBpZiAoc3RhcnREb3QgIT09IC0xKSB7XG4gICAgICAvLyBXZSBzYXcgYSBub24tZG90IGFuZCBub24tcGF0aCBzZXBhcmF0b3IgYmVmb3JlIG91ciBkb3QsIHNvIHdlIHNob3VsZFxuICAgICAgLy8gaGF2ZSBhIGdvb2QgY2hhbmNlIGF0IGhhdmluZyBhIG5vbi1lbXB0eSBleHRlbnNpb25cbiAgICAgIHByZURvdFN0YXRlID0gLTE7XG4gICAgfVxuICB9XG5cbiAgaWYgKFxuICAgIHN0YXJ0RG90ID09PSAtMSB8fFxuICAgIGVuZCA9PT0gLTEgfHxcbiAgICAvLyBXZSBzYXcgYSBub24tZG90IGNoYXJhY3RlciBpbW1lZGlhdGVseSBiZWZvcmUgdGhlIGRvdFxuICAgIHByZURvdFN0YXRlID09PSAwIHx8XG4gICAgLy8gVGhlIChyaWdodC1tb3N0KSB0cmltbWVkIHBhdGggY29tcG9uZW50IGlzIGV4YWN0bHkgJy4uJ1xuICAgIChwcmVEb3RTdGF0ZSA9PT0gMSAmJiBzdGFydERvdCA9PT0gZW5kIC0gMSAmJiBzdGFydERvdCA9PT0gc3RhcnRQYXJ0ICsgMSlcbiAgKSB7XG4gICAgaWYgKGVuZCAhPT0gLTEpIHtcbiAgICAgIHJldC5iYXNlID0gcmV0Lm5hbWUgPSBwYXRoLnNsaWNlKHN0YXJ0UGFydCwgZW5kKTtcbiAgICB9XG4gIH0gZWxzZSB7XG4gICAgcmV0Lm5hbWUgPSBwYXRoLnNsaWNlKHN0YXJ0UGFydCwgc3RhcnREb3QpO1xuICAgIHJldC5iYXNlID0gcGF0aC5zbGljZShzdGFydFBhcnQsIGVuZCk7XG4gICAgcmV0LmV4dCA9IHBhdGguc2xpY2Uoc3RhcnREb3QsIGVuZCk7XG4gIH1cblxuICAvLyBJZiB0aGUgZGlyZWN0b3J5IGlzIHRoZSByb290LCB1c2UgdGhlIGVudGlyZSByb290IGFzIHRoZSBgZGlyYCBpbmNsdWRpbmdcbiAgLy8gdGhlIHRyYWlsaW5nIHNsYXNoIGlmIGFueSAoYEM6XFxhYmNgIC0+IGBDOlxcYCkuIE90aGVyd2lzZSwgc3RyaXAgb3V0IHRoZVxuICAvLyB0cmFpbGluZyBzbGFzaCAoYEM6XFxhYmNcXGRlZmAgLT4gYEM6XFxhYmNgKS5cbiAgaWYgKHN0YXJ0UGFydCA+IDAgJiYgc3RhcnRQYXJ0ICE9PSByb290RW5kKSB7XG4gICAgcmV0LmRpciA9IHBhdGguc2xpY2UoMCwgc3RhcnRQYXJ0IC0gMSk7XG4gIH0gZWxzZSByZXQuZGlyID0gcmV0LnJvb3Q7XG5cbiAgcmV0dXJuIHJldDtcbn1cblxuLyoqIENvbnZlcnRzIGEgZmlsZSBVUkwgdG8gYSBwYXRoIHN0cmluZy5cbiAqXG4gKiAgICAgIGZyb21GaWxlVXJsKFwiZmlsZTovLy9ob21lL2Zvb1wiKTsgLy8gXCJcXFxcaG9tZVxcXFxmb29cIlxuICogICAgICBmcm9tRmlsZVVybChcImZpbGU6Ly8vQzovVXNlcnMvZm9vXCIpOyAvLyBcIkM6XFxcXFVzZXJzXFxcXGZvb1wiXG4gKiAgICAgIGZyb21GaWxlVXJsKFwiZmlsZTovL2xvY2FsaG9zdC9ob21lL2Zvb1wiKTsgLy8gXCJcXFxcXFxcXGxvY2FsaG9zdFxcXFxob21lXFxcXGZvb1wiXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBmcm9tRmlsZVVybCh1cmw6IHN0cmluZyB8IFVSTCk6IHN0cmluZyB7XG4gIHVybCA9IHVybCBpbnN0YW5jZW9mIFVSTCA/IHVybCA6IG5ldyBVUkwodXJsKTtcbiAgaWYgKHVybC5wcm90b2NvbCAhPSBcImZpbGU6XCIpIHtcbiAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiTXVzdCBiZSBhIGZpbGUgVVJMLlwiKTtcbiAgfVxuICBsZXQgcGF0aCA9IGRlY29kZVVSSUNvbXBvbmVudChcbiAgICB1cmwucGF0aG5hbWVcbiAgICAgIC5yZXBsYWNlKC9eXFwvKihbQS1aYS16XTopKFxcL3wkKS8sIFwiJDEvXCIpXG4gICAgICAucmVwbGFjZSgvXFwvL2csIFwiXFxcXFwiKVxuICAgICAgLnJlcGxhY2UoLyUoPyFbMC05QS1GYS1mXXsyfSkvZywgXCIlMjVcIiksXG4gICk7XG4gIGlmICh1cmwuaG9zdG5hbWUgIT0gXCJcIikge1xuICAgIC8vIE5vdGU6IFRoZSBgVVJMYCBpbXBsZW1lbnRhdGlvbiBndWFyYW50ZWVzIHRoYXQgdGhlIGRyaXZlIGxldHRlciBhbmRcbiAgICAvLyBob3N0bmFtZSBhcmUgbXV0dWFsbHkgZXhjbHVzaXZlLiBPdGhlcndpc2UgaXQgd291bGQgbm90IGhhdmUgYmVlbiB2YWxpZFxuICAgIC8vIHRvIGFwcGVuZCB0aGUgaG9zdG5hbWUgYW5kIHBhdGggbGlrZSB0aGlzLlxuICAgIHBhdGggPSBgXFxcXFxcXFwke3VybC5ob3N0bmFtZX0ke3BhdGh9YDtcbiAgfVxuICByZXR1cm4gcGF0aDtcbn1cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiU0FNQSxRQUFBLEVBQ0EsbUJBQUEsRUFDQSxVQUFBLEVBQ0Esa0JBQUEsU0FDQSxlQUFBO1NBR0EsVUFBQSxFQUNBLGVBQUEsRUFDQSxtQkFBQSxFQUNBLGVBQUEsRUFDQSxPQUFBLFNBQ0EsVUFBQTtTQUNBLE1BQUEsU0FBQSxrQkFBQTthQUVBLEdBQUEsSUFBQSxFQUFBO2FBQ0EsU0FBQSxJQUFBLENBQUE7Z0JBRUEsT0FBQSxJQUFBLFlBQUE7UUFDQSxjQUFBO1FBQ0EsWUFBQTtRQUNBLGdCQUFBLEdBQUEsS0FBQTtZQUVBLENBQUEsR0FBQSxZQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsRUFBQSxDQUFBLEtBQUEsQ0FBQSxFQUFBLENBQUE7WUFDQSxJQUFBO1lBQ0EsQ0FBQSxJQUFBLENBQUE7QUFDQSxnQkFBQSxHQUFBLFlBQUEsQ0FBQSxDQUFBO29CQUNBLGNBQUE7Z0JBQ0EsVUFBQSxDQUFBLElBQUEsSUFBQSxJQUFBOzBCQUNBLFNBQUEsRUFBQSxnREFBQTs7QUFFQSxnQkFBQSxHQUFBLElBQUEsQ0FBQSxHQUFBOztnQkFFQSxVQUFBLENBQUEsSUFBQSxJQUFBLElBQUE7MEJBQ0EsU0FBQSxFQUFBLHVDQUFBOztBQUVBLGNBQUEsMERBQUE7QUFDQSxjQUFBLDZEQUFBO0FBQ0EsY0FBQSw2REFBQTtBQUNBLGNBQUEsNkRBQUE7QUFDQSxjQUFBLGtFQUFBO0FBQ0EsZ0JBQUEsR0FBQSxJQUFBLENBQUEsR0FBQSxDQUFBLEdBQUEsRUFBQSxDQUFBLEVBQUEsY0FBQSxPQUFBLElBQUEsQ0FBQSxHQUFBO0FBRUEsY0FBQSx3REFBQTtBQUNBLGNBQUEsbURBQUE7Z0JBRUEsSUFBQSxLQUFBLFNBQUEsSUFDQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxDQUFBLEVBQUEsV0FBQSxVQUFBLGNBQUEsQ0FBQSxXQUFBLEdBQUEsRUFBQTtBQUVBLG9CQUFBLE1BQUEsY0FBQSxDQUFBLEVBQUE7OztBQUlBLGtCQUFBLENBQUEsSUFBQTtjQUVBLEdBQUEsR0FBQSxJQUFBLENBQUEsTUFBQTtBQUVBLFVBQUEsbUJBQUE7WUFDQSxHQUFBLEtBQUEsQ0FBQTtZQUVBLE9BQUEsR0FBQSxDQUFBO1lBQ0EsTUFBQTtZQUNBLFVBQUEsR0FBQSxLQUFBO2NBQ0EsSUFBQSxHQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQTtBQUVBLFVBQUEsb0JBQUE7WUFDQSxHQUFBLEdBQUEsQ0FBQTtnQkFDQSxlQUFBLENBQUEsSUFBQTtBQUNBLGtCQUFBLGtCQUFBO0FBRUEsa0JBQUEsNERBQUE7QUFDQSxrQkFBQSw4Q0FBQTtBQUNBLDBCQUFBLEdBQUEsSUFBQTtvQkFFQSxlQUFBLENBQUEsSUFBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBO0FBQ0Esc0JBQUEsMkNBQUE7d0JBQ0EsQ0FBQSxHQUFBLENBQUE7d0JBQ0EsSUFBQSxHQUFBLENBQUE7QUFDQSxzQkFBQSxvQ0FBQTswQkFDQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUE7NEJBQ0EsZUFBQSxDQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQTs7d0JBRUEsQ0FBQSxHQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsSUFBQTs4QkFDQSxTQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxJQUFBLEVBQUEsQ0FBQTtBQUNBLDBCQUFBLFNBQUE7QUFDQSw0QkFBQSxHQUFBLENBQUE7QUFDQSwwQkFBQSxnQ0FBQTs4QkFDQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUE7aUNBQ0EsZUFBQSxDQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQTs7NEJBRUEsQ0FBQSxHQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsSUFBQTtBQUNBLDhCQUFBLFNBQUE7QUFDQSxnQ0FBQSxHQUFBLENBQUE7QUFDQSw4QkFBQSxvQ0FBQTtrQ0FDQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUE7b0NBQ0EsZUFBQSxDQUFBLElBQUEsQ0FBQSxVQUFBLENBQUEsQ0FBQTs7Z0NBRUEsQ0FBQSxLQUFBLEdBQUE7QUFDQSxrQ0FBQSwyQkFBQTtBQUNBLHNDQUFBLElBQUEsSUFBQSxFQUFBLFNBQUEsQ0FBQSxFQUFBLEVBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxJQUFBO0FBQ0EsdUNBQUEsR0FBQSxDQUFBO3VDQUNBLENBQUEsS0FBQSxJQUFBO0FBQ0Esa0NBQUEscUNBQUE7QUFFQSxzQ0FBQSxJQUFBLElBQUEsRUFBQSxTQUFBLENBQUEsRUFBQSxFQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsSUFBQSxFQUFBLENBQUE7QUFDQSx1Q0FBQSxHQUFBLENBQUE7Ozs7O0FBS0EsMkJBQUEsR0FBQSxDQUFBOzt1QkFFQSxtQkFBQSxDQUFBLElBQUE7QUFDQSxrQkFBQSxxQkFBQTtvQkFFQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUEsTUFBQSxVQUFBO0FBQ0EsMEJBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxDQUFB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QUE7bUJBQ0EsSUFBQTs7OztlQUlBLFVBQUE7WUFDQSxJQUFBLENBQUEsTUFBQSxHQUFBLENBQUEsWUFBQSxNQUFBLENBQUEsRUFBQSxFQUFBLElBQUE7dUJBQ0EsTUFBQSxDQUFBLEVBQUE7ZUFDQSxJQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7ZUFDQSxNQUFBLEdBQUEsSUFBQTs7ZUFFQSxNQUFBOzs7Z0JBSUEsVUFBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtVQUNBLEdBQUEsR0FBQSxJQUFBLENBQUEsTUFBQTtRQUNBLEdBQUEsS0FBQSxDQUFBLFNBQUEsS0FBQTtVQUVBLElBQUEsR0FBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUE7UUFDQSxlQUFBLENBQUEsSUFBQTtlQUNBLElBQUE7ZUFDQSxtQkFBQSxDQUFBLElBQUE7QUFDQSxVQUFBLHFCQUFBO1lBRUEsR0FBQSxHQUFBLENBQUEsSUFBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUEsTUFBQSxVQUFBO2dCQUNBLGVBQUEsQ0FBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUEsV0FBQSxJQUFBOzs7V0FHQSxLQUFBOztnQkFHQSxJQUFBLElBQUEsS0FBQTtVQUNBLFVBQUEsR0FBQSxLQUFBLENBQUEsTUFBQTtRQUNBLFVBQUEsS0FBQSxDQUFBLFVBQUEsQ0FBQTtRQUVBLE1BQUE7UUFDQSxTQUFBLEdBQUEsSUFBQTtZQUNBLENBQUEsR0FBQSxDQUFBLEVBQUEsQ0FBQSxHQUFBLFVBQUEsSUFBQSxDQUFBO2NBQ0EsSUFBQSxHQUFBLEtBQUEsQ0FBQSxDQUFBO0FBQ0Esa0JBQUEsQ0FBQSxJQUFBO1lBQ0EsSUFBQSxDQUFBLE1BQUEsR0FBQSxDQUFBO2dCQUNBLE1BQUEsS0FBQSxTQUFBLEVBQUEsTUFBQSxHQUFBLFNBQUEsR0FBQSxJQUFBO2lCQUNBLE1BQUEsS0FBQSxFQUFBLEVBQUEsSUFBQTs7O1FBSUEsTUFBQSxLQUFBLFNBQUEsVUFBQSxDQUFBO0FBRUEsTUFBQSx1RUFBQTtBQUNBLE1BQUEsa0RBQUE7QUFDQSxNQUFBO0FBQ0EsTUFBQSxrRUFBQTtBQUNBLE1BQUEsaUVBQUE7QUFDQSxNQUFBLHVFQUFBO0FBQ0EsTUFBQSx1Q0FBQTtBQUNBLE1BQUE7QUFDQSxNQUFBLHFFQUFBO0FBQ0EsTUFBQSw4REFBQTtBQUNBLE1BQUEsa0VBQUE7QUFDQSxNQUFBLDZDQUFBO0FBQ0EsTUFBQSwyREFBQTtRQUNBLFlBQUEsR0FBQSxJQUFBO1FBQ0EsVUFBQSxHQUFBLENBQUE7QUFDQSxVQUFBLENBQUEsU0FBQSxJQUFBLElBQUE7UUFDQSxlQUFBLENBQUEsU0FBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBO1VBQ0EsVUFBQTtjQUNBLFFBQUEsR0FBQSxTQUFBLENBQUEsTUFBQTtZQUNBLFFBQUEsR0FBQSxDQUFBO2dCQUNBLGVBQUEsQ0FBQSxTQUFBLENBQUEsVUFBQSxDQUFBLENBQUE7a0JBQ0EsVUFBQTtvQkFDQSxRQUFBLEdBQUEsQ0FBQTt3QkFDQSxlQUFBLENBQUEsU0FBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBLE1BQUEsVUFBQTs7QUFFQSwwQkFBQSx3Q0FBQTtBQUNBLG9DQUFBLEdBQUEsS0FBQTs7Ozs7O1FBTUEsWUFBQTtBQUNBLFVBQUEscURBQUE7Y0FDQSxVQUFBLEdBQUEsTUFBQSxDQUFBLE1BQUEsSUFBQSxVQUFBO2lCQUNBLGVBQUEsQ0FBQSxNQUFBLENBQUEsVUFBQSxDQUFBLFVBQUE7O0FBR0EsVUFBQSw4QkFBQTtZQUNBLFVBQUEsSUFBQSxDQUFBLEVBQUEsTUFBQSxJQUFBLEVBQUEsRUFBQSxNQUFBLENBQUEsS0FBQSxDQUFBLFVBQUE7O1dBR0EsU0FBQSxDQUFBLE1BQUE7O0FBR0EsRUFBQSxtRUFBQTtBQUNBLEVBQUEsaUNBQUE7QUFDQSxFQUFBLCtCQUFBO0FBQ0EsRUFBQSwwREFBQTtnQkFDQSxRQUFBLENBQUEsSUFBQSxFQUFBLEVBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtBQUNBLGNBQUEsQ0FBQSxFQUFBO1FBRUEsSUFBQSxLQUFBLEVBQUE7VUFFQSxRQUFBLEdBQUEsT0FBQSxDQUFBLElBQUE7VUFDQSxNQUFBLEdBQUEsT0FBQSxDQUFBLEVBQUE7UUFFQSxRQUFBLEtBQUEsTUFBQTtBQUVBLFFBQUEsR0FBQSxRQUFBLENBQUEsV0FBQTtBQUNBLE1BQUEsR0FBQSxNQUFBLENBQUEsV0FBQTtRQUVBLElBQUEsS0FBQSxFQUFBO0FBRUEsTUFBQSw2QkFBQTtRQUNBLFNBQUEsR0FBQSxDQUFBO1FBQ0EsT0FBQSxHQUFBLElBQUEsQ0FBQSxNQUFBO1VBQ0EsU0FBQSxHQUFBLE9BQUEsSUFBQSxTQUFBO1lBQ0EsSUFBQSxDQUFBLFVBQUEsQ0FBQSxTQUFBLE1BQUEsbUJBQUE7O0FBRUEsTUFBQSx5REFBQTtVQUNBLE9BQUEsR0FBQSxDQUFBLEdBQUEsU0FBQSxJQUFBLE9BQUE7WUFDQSxJQUFBLENBQUEsVUFBQSxDQUFBLE9BQUEsR0FBQSxDQUFBLE1BQUEsbUJBQUE7O1VBRUEsT0FBQSxHQUFBLE9BQUEsR0FBQSxTQUFBO0FBRUEsTUFBQSw2QkFBQTtRQUNBLE9BQUEsR0FBQSxDQUFBO1FBQ0EsS0FBQSxHQUFBLEVBQUEsQ0FBQSxNQUFBO1VBQ0EsT0FBQSxHQUFBLEtBQUEsSUFBQSxPQUFBO1lBQ0EsRUFBQSxDQUFBLFVBQUEsQ0FBQSxPQUFBLE1BQUEsbUJBQUE7O0FBRUEsTUFBQSx5REFBQTtVQUNBLEtBQUEsR0FBQSxDQUFBLEdBQUEsT0FBQSxJQUFBLEtBQUE7WUFDQSxFQUFBLENBQUEsVUFBQSxDQUFBLEtBQUEsR0FBQSxDQUFBLE1BQUEsbUJBQUE7O1VBRUEsS0FBQSxHQUFBLEtBQUEsR0FBQSxPQUFBO0FBRUEsTUFBQSx3REFBQTtVQUNBLE1BQUEsR0FBQSxPQUFBLEdBQUEsS0FBQSxHQUFBLE9BQUEsR0FBQSxLQUFBO1FBQ0EsYUFBQSxJQUFBLENBQUE7UUFDQSxDQUFBLEdBQUEsQ0FBQTtVQUNBLENBQUEsSUFBQSxNQUFBLElBQUEsQ0FBQTtZQUNBLENBQUEsS0FBQSxNQUFBO2dCQUNBLEtBQUEsR0FBQSxNQUFBO29CQUNBLEVBQUEsQ0FBQSxVQUFBLENBQUEsT0FBQSxHQUFBLENBQUEsTUFBQSxt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QUFBLEdBQUEsR0FBQSxnQkFBQTtpQkFDQSxHQUFBLE1BQUEsQ0FBQSxFQUFBLEdBQUEsR0FBQSxJQUFBLENBQUEsTUFBQTtlQUNBLElBQUEsQ0FBQSxLQUFBLENBQUEsS0FBQSxFQUFBLEdBQUE7O1lBRUEsQ0FBQSxHQUFBLElBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxFQUFBLENBQUEsSUFBQSxLQUFBLElBQUEsQ0FBQTtnQkFDQSxlQUFBLENBQUEsSUFBQSxDQUFBLFVBQUEsQ0FBQSxDQUFBO0FBQ0Esa0JBQUEsa0VBQUE7QUFDQSxrQkFBQSw4Q0FBQTtxQkFDQSxZQUFBO0FBQ0EseUJBQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQTs7O3VCQUdBLEdBQUEsTUFBQSxDQUFBO0FBQ0Esa0JBQUEsaUVBQUE7QUFDQSxrQkFBQSxlQUFBO0FBQ0EsNEJBQUEsR0FBQSxLQUFBO0FBQ0EsbUJBQUEsR0FBQSxDQUFBLEdBQUEsQ0FBQTs7O1lBSUEsR0FBQSxNQUFBLENBQUE7ZUFDQSxJQUFBLENBQUEsS0FBQSxDQUFBLEtBQUEsRUFBQSxHQUFBOzs7Z0JBSUEsT0FBQSxDQUFBLElBQUE7QUFDQSxjQUFBLENBQUEsSUFBQTtRQUNBLEtBQUEsR0FBQSxDQUFBO1FBQ0EsUUFBQSxJQUFBLENBQUE7UUFDQSxTQUFBLEdBQUEsQ0FBQTtRQUNBLEdBQUEsSUFBQSxDQUFBO1FBQ0EsWUFBQSxHQUFBLElBQUE7QUFDQSxNQUFBLHVFQUFBO0FBQ0EsTUFBQSxpQ0FBQTtRQUNBLFdBQUEsR0FBQSxDQUFBO0FBRUEsTUFBQSxtRUFBQTtBQUNBLE1BQUEsd0VBQUE7QUFDQSxNQUFBLFlBQUE7UUFHQSxJQUFBLENBQUEsTUFBQSxJQUFBLENBQUEsSUFDQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUEsTUFBQSxVQUFBLElBQ0EsbUJBQUEsQ0FBQSxJQUFBLENBQUEsVUFBQSxDQUFBLENBQUE7QUFFQSxhQUFBLEdBQUEsU0FBQSxHQUFBLENBQUE7O1lBR0EsQ0FBQSxHQUFBLElBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQSxFQUFBLENB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Q0EsWUFBQTtBQUNBLHlCQUFBLEdBQUEsQ0FBQSxHQUFBLENBQUE7Ozs7O1lBS0EsR0FBQSxNQUFBLENBQUE7QUFDQSxjQUFBLGlFQUFBO0FBQ0EsY0FBQSxVQUFBO0FBQ0Esd0JBQUEsR0FBQSxLQUFBO0FBQ0EsZUFBQSxHQUFBLENBQUEsR0FBQSxDQUFBOztZQUVBLElBQUEsS0FBQSxRQUFBO0FBQ0EsY0FBQSxnRUFBQTtnQkFDQSxRQUFBLE1BQUEsQ0FBQSxFQUFBLFFBQUEsR0FBQSxDQUFBO3FCQUNBLFdBQUEsS0FBQSxDQUFBLEVBQUEsV0FBQSxHQUFBLENBQUE7bUJBQ0EsUUFBQSxNQUFBLENBQUE7QUFDQSxjQUFBLHFFQUFBO0FBQ0EsY0FBQSxtREFBQTtBQUNBLHVCQUFBLElBQUEsQ0FBQTs7O1FBS0EsUUFBQSxNQUFBLENBQUEsSUFDQSxHQUFBLE1BQUEsQ0FBQSxJQUNBLEVBQUEsc0RBQUE7QUFDQSxlQUFBLEtBQUEsQ0FBQSxJQUNBLEVBQUEsd0RBQUE7S0FDQSxXQUFBLEtBQUEsQ0FBQSxJQUFBLFFBQUEsS0FBQSxHQUFBLEdBQUEsQ0FBQSxJQUFBLFFBQUEsS0FBQSxTQUFBLEdBQUEsQ0FBQTtZQUVBLEdBQUEsTUFBQSxDQUFBO0FBQ0EsZUFBQSxDQUFBLElBQUEsR0FBQSxHQUFBLENBQUEsSUFBQSxHQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsU0FBQSxFQUFBLEdBQUE7OztBQUdBLFdBQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxTQUFBLEVBQUEsUUFBQTtBQUNBLFdBQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxTQUFBLEVBQUEsR0FBQTtBQUNBLFdBQUEsQ0FBQSxHQUFBLEdBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxRQUFBLEVBQUEsR0FBQTs7QUFHQSxNQUFBLHlFQUFBO0FBQ0EsTUFBQSx3RUFBQTtBQUNBLE1BQUEsMkNBQUE7UUFDQSxTQUFBLEdBQUEsQ0FBQSxJQUFBLFNBQUEsS0FBQSxPQUFBO0FBQ0EsV0FBQSxDQUFBLEdBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLENBQUEsRUFBQSxTQUFBLEdBQUEsQ0FBQTtXQUNBLEdBQUEsQ0FBQSxHQUFBLEdBQUEsR0FBQSxDQUFBLElBQUE7V0FFQSxHQUFBOztBQUdBLEVBS0EsQUFMQSx1UEFLQSxBQUxBLEVBS0EsaUJBQ0EsV0FBQSxDQUFBLEdBQUE7QUFDQSxPQUFBLEdBQUEsR0FBQSxZQUFBLEdBQUEsR0FBQSxHQUFBLE9BQUEsR0FBQSxDQUFBLEdBQUE7UUFDQSxHQUFBLENBQUEsUUFBQSxLQUFBLEtBQUE7a0JBQ0EsU0FBQSxFQUFBLG1CQUFBOztRQUVBLElBQUEsR0FBQSxrQkFBQSxDQUNBLEdBQUEsQ0FBQSxRQUFBLENBQ0EsT0FBQSwyQkFBQSxHQUFBLEdBQ0EsT0FBQSxTQUFBLEVBQUEsR0FDQSxPQUFBLDBCQUFBLEdBQUE7UUFFQSxHQUFBLENBQUEsUUFBQTtBQUNBLFVBQUEsb0VBQUE7QUFDQSxVQUFBLHdFQUFBO0FBQ0EsVUFBQSwyQ0FBQTtBQUNBLFlBQUEsSUFBQSxJQUFBLEVBQUEsR0FBQSxDQUFBLFFBQUEsR0FBQSxJQUFBOztXQUVBLElBQUEifQ==