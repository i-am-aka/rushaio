// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { BufWriter } from "../io/bufio.ts";
import { TextProtoReader } from "../textproto/mod.ts";
import { assert } from "../_util/assert.ts";
import { encoder } from "../encoding/utf8.ts";
import { ServerRequest } from "./server.ts";
import { STATUS_TEXT } from "./http_status.ts";
export function emptyReader() {
    return {
        read (_) {
            return Promise.resolve(null);
        }
    };
}
export function bodyReader(contentLength, r) {
    let totalRead = 0;
    let finished = false;
    async function read(buf) {
        if (finished) return null;
        let result;
        const remaining = contentLength - totalRead;
        if (remaining >= buf.byteLength) {
            result = await r.read(buf);
        } else {
            const readBuf = buf.subarray(0, remaining);
            result = await r.read(readBuf);
        }
        if (result !== null) {
            totalRead += result;
        }
        finished = totalRead === contentLength;
        return result;
    }
    return {
        read
    };
}
export function chunkedBodyReader(h, r) {
    // Based on https://tools.ietf.org/html/rfc2616#section-19.4.6
    const tp = new TextProtoReader(r);
    let finished = false;
    const chunks = [];
    async function read(buf) {
        if (finished) return null;
        const [chunk] = chunks;
        if (chunk) {
            const chunkRemaining = chunk.data.byteLength - chunk.offset;
            const readLength = Math.min(chunkRemaining, buf.byteLength);
            for(let i = 0; i < readLength; i++){
                buf[i] = chunk.data[chunk.offset + i];
            }
            chunk.offset += readLength;
            if (chunk.offset === chunk.data.byteLength) {
                chunks.shift();
                // Consume \r\n;
                if (await tp.readLine() === null) {
                    throw new Deno.errors.UnexpectedEof();
                }
            }
            return readLength;
        }
        const line = await tp.readLine();
        if (line === null) throw new Deno.errors.UnexpectedEof();
        // TODO: handle chunk extension
        const [chunkSizeString] = line.split(";");
        const chunkSize = parseInt(chunkSizeString, 16);
        if (Number.isNaN(chunkSize) || chunkSize < 0) {
            throw new Error("Invalid chunk size");
        }
        if (chunkSize > 0) {
            if (chunkSize > buf.byteLength) {
                let eof = await r.readFull(buf);
                if (eof === null) {
                    throw new Deno.errors.UnexpectedEof();
                }
                const restChunk = new Uint8Array(chunkSize - buf.byteLength);
                eof = await r.readFull(restChunk);
                if (eof === null) {
                    throw new Deno.errors.UnexpectedEof();
                } else {
                    chunks.push({
                        offset: 0,
                        data: restChunk
                    });
                }
                return buf.byteLength;
            } else {
                const bufToFill = buf.subarray(0, chunkSize);
                const eof = await r.readFull(bufToFill);
                if (eof === null) {
                    throw new Deno.errors.UnexpectedEof();
                }
                // Consume \r\n
                if (await tp.readLine() === null) {
                    throw new Deno.errors.UnexpectedEof();
                }
                return chunkSize;
            }
        } else {
            assert(chunkSize === 0);
            // Consume \r\n
            if (await r.readLine() === null) {
                throw new Deno.errors.UnexpectedEof();
            }
            await readTrailers(h, r);
            finished = true;
            return null;
        }
    }
    return {
        read
    };
}
function isProhibidedForTrailer(key) {
    const s = new Set([
        "transfer-encoding",
        "content-length",
        "trailer"
    ]);
    return s.has(key.toLowerCase());
}
/** Read trailer headers from reader and append values to headers. "trailer"
 * field will be deleted. */ export async function readTrailers(headers, r) {
    const trailers = parseTrailer(headers.get("trailer"));
    if (trailers == null) return;
    const trailerNames = [
        ...trailers.keys()
    ];
    const tp = new TextProtoReader(r);
    const result = await tp.readMIMEHeader();
    if (result == null) {
        throw new Deno.errors.InvalidData("Missing trailer header.");
    }
    const undeclared = [
        ...result.keys()
    ].filter((k)=>!trailerNames.includes(k)
    );
    if (undeclared.length > 0) {
        throw new Deno.errors.InvalidData(`Undeclared trailers: ${Deno.inspect(undeclared)}.`);
    }
    for (const [k, v] of result){
        headers.append(k, v);
    }
    const missingTrailers = trailerNames.filter((k)=>!result.has(k)
    );
    if (missingTrailers.length > 0) {
        throw new Deno.errors.InvalidData(`Missing trailers: ${Deno.inspect(missingTrailers)}.`);
    }
    headers.delete("trailer");
}
function parseTrailer(field) {
    if (field == null) {
        return undefined;
    }
    const trailerNames = field.split(",").map((v)=>v.trim().toLowerCase()
    );
    if (trailerNames.length === 0) {
        throw new Deno.errors.InvalidData("Empty trailer header.");
    }
    const prohibited = trailerNames.filter((k)=>isProhibidedForTrailer(k)
    );
    if (prohibited.length > 0) {
        throw new Deno.errors.InvalidData(`Prohibited trailer names: ${Deno.inspect(prohibited)}.`);
    }
    return new Headers(trailerNames.map((key)=>[
            key,
            ""
        ]
    ));
}
export async function writeChunkedBody(w, r) {
    const writer = BufWriter.create(w);
    for await (const chunk of Deno.iter(r)){
        if (chunk.byteLength <= 0) continue;
        const start = encoder.encode(`${chunk.byteLength.toString(16)}\r\n`);
        const end = encoder.encode("\r\n");
        await writer.write(start);
        await writer.write(chunk);
        await writer.write(end);
    }
    const endChunk = encoder.encode("0\r\n\r\n");
    await writer.write(endChunk);
}
/** Write trailer headers to writer. It should mostly should be called after
 * `writeResponse()`. */ export async function writeTrailers(w, headers, trailers) {
    const trailer = headers.get("trailer");
    if (trailer === null) {
        throw new TypeError("Missing trailer header.");
    }
    const transferEncoding = headers.get("transfer-encoding");
    if (transferEncoding === null || !transferEncoding.match(/^chunked/)) {
        throw new TypeError(`Trailers are only allowed for "transfer-encoding: chunked", got "transfer-encoding: ${transferEncoding}".`);
    }
    const writer = BufWriter.create(w);
    const trailerNames = trailer.split(",").map((s)=>s.trim().toLowerCase()
    );
    const prohibitedTrailers = trailerNames.filter((k)=>isProhibidedForTrailer(k)
    );
    if (prohibitedTrailers.length > 0) {
        throw new TypeError(`Prohibited trailer names: ${Deno.inspect(prohibitedTrailers)}.`);
    }
    const undeclared = [
        ...trailers.keys()
    ].filter((k)=>!trailerNames.includes(k)
    );
    if (undeclared.length > 0) {
        throw new TypeError(`Undeclared trailers: ${Deno.inspect(undeclared)}.`);
    }
    for (const [key, value] of trailers){
        await writer.write(encoder.encode(`${key}: ${value}\r\n`));
    }
    await writer.write(encoder.encode("\r\n"));
    await writer.flush();
}
export async function writeResponse(w, r) {
    const protoMajor = 1;
    const protoMinor = 1;
    const statusCode = r.status || 200;
    const statusText = STATUS_TEXT.get(statusCode);
    const writer = BufWriter.create(w);
    if (!statusText) {
        throw new Deno.errors.InvalidData("Bad status code");
    }
    if (!r.body) {
        r.body = new Uint8Array();
    }
    if (typeof r.body === "string") {
        r.body = encoder.encode(r.body);
    }
    let out = `HTTP/${protoMajor}.${protoMinor} ${statusCode} ${statusText}\r\n`;
    const headers = r.headers ?? new Headers();
    if (r.body && !headers.get("content-length")) {
        if (r.body instanceof Uint8Array) {
            out += `content-length: ${r.body.byteLength}\r\n`;
        } else if (!headers.get("transfer-encoding")) {
            out += "transfer-encoding: chunked\r\n";
        }
    }
    for (const [key, value] of headers){
        out += `${key}: ${value}\r\n`;
    }
    out += `\r\n`;
    const header = encoder.encode(out);
    const n = await writer.write(header);
    assert(n === header.byteLength);
    if (r.body instanceof Uint8Array) {
        const n = await writer.write(r.body);
        assert(n === r.body.byteLength);
    } else if (headers.has("content-length")) {
        const contentLength = headers.get("content-length");
        assert(contentLength != null);
        const bodyLength = parseInt(contentLength);
        const n = await Deno.copy(r.body, writer);
        assert(n === bodyLength);
    } else {
        await writeChunkedBody(writer, r.body);
    }
    if (r.trailers) {
        const t = await r.trailers();
        await writeTrailers(writer, headers, t);
    }
    await writer.flush();
}
/**
 * ParseHTTPVersion parses a HTTP version string.
 * "HTTP/1.0" returns (1, 0).
 * Ported from https://github.com/golang/go/blob/f5c43b9/src/net/http/request.go#L766-L792
 */ export function parseHTTPVersion(vers) {
    switch(vers){
        case "HTTP/1.1":
            return [
                1,
                1
            ];
        case "HTTP/1.0":
            return [
                1,
                0
            ];
        default:
            {
                const Big = 1000000; // arbitrary upper bound
                if (!vers.startsWith("HTTP/")) {
                    break;
                }
                const dot = vers.indexOf(".");
                if (dot < 0) {
                    break;
                }
                const majorStr = vers.substring(vers.indexOf("/") + 1, dot);
                const major = Number(majorStr);
                if (!Number.isInteger(major) || major < 0 || major > Big) {
                    break;
                }
                const minorStr = vers.substring(dot + 1);
                const minor = Number(minorStr);
                if (!Number.isInteger(minor) || minor < 0 || minor > Big) {
                    break;
                }
                return [
                    major,
                    minor
                ];
            }
    }
    throw new Error(`malformed HTTP version ${vers}`);
}
export async function readRequest(conn, bufr) {
    const tp = new TextProtoReader(bufr);
    const firstLine = await tp.readLine(); // e.g. GET /index.html HTTP/1.0
    if (firstLine === null) return null;
    const headers = await tp.readMIMEHeader();
    if (headers === null) throw new Deno.errors.UnexpectedEof();
    const req = new ServerRequest();
    req.conn = conn;
    req.r = bufr;
    [req.method, req.url, req.proto] = firstLine.split(" ", 3);
    [req.protoMinor, req.protoMajor] = parseHTTPVersion(req.proto);
    req.headers = headers;
    fixLength(req);
    return req;
}
function fixLength(req) {
    const contentLength = req.headers.get("Content-Length");
    if (contentLength) {
        const arrClen = contentLength.split(",");
        if (arrClen.length > 1) {
            const distinct = [
                ...new Set(arrClen.map((e)=>e.trim()
                ))
            ];
            if (distinct.length > 1) {
                throw Error("cannot contain multiple Content-Length headers");
            } else {
                req.headers.set("Content-Length", distinct[0]);
            }
        }
        const c = req.headers.get("Content-Length");
        if (req.method === "HEAD" && c && c !== "0") {
            throw Error("http: method cannot contain a Content-Length");
        }
        if (c && req.headers.has("transfer-encoding")) {
            // A sender MUST NOT send a Content-Length header field in any message
            // that contains a Transfer-Encoding header field.
            // rfc: https://tools.ietf.org/html/rfc7230#section-3.3.2
            throw new Error("http: Transfer-Encoding and Content-Length cannot be send together");
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL2h0dHAvX2lvLnRzPiJdLCJzb3VyY2VzQ29udGVudCI6WyIvLyBDb3B5cmlnaHQgMjAxOC0yMDIwIHRoZSBEZW5vIGF1dGhvcnMuIEFsbCByaWdodHMgcmVzZXJ2ZWQuIE1JVCBsaWNlbnNlLlxuaW1wb3J0IHsgQnVmUmVhZGVyLCBCdWZXcml0ZXIgfSBmcm9tIFwiLi4vaW8vYnVmaW8udHNcIjtcbmltcG9ydCB7IFRleHRQcm90b1JlYWRlciB9IGZyb20gXCIuLi90ZXh0cHJvdG8vbW9kLnRzXCI7XG5pbXBvcnQgeyBhc3NlcnQgfSBmcm9tIFwiLi4vX3V0aWwvYXNzZXJ0LnRzXCI7XG5pbXBvcnQgeyBlbmNvZGVyIH0gZnJvbSBcIi4uL2VuY29kaW5nL3V0ZjgudHNcIjtcbmltcG9ydCB7IFJlc3BvbnNlLCBTZXJ2ZXJSZXF1ZXN0IH0gZnJvbSBcIi4vc2VydmVyLnRzXCI7XG5pbXBvcnQgeyBTVEFUVVNfVEVYVCB9IGZyb20gXCIuL2h0dHBfc3RhdHVzLnRzXCI7XG5cbmV4cG9ydCBmdW5jdGlvbiBlbXB0eVJlYWRlcigpOiBEZW5vLlJlYWRlciB7XG4gIHJldHVybiB7XG4gICAgcmVhZChfOiBVaW50OEFycmF5KTogUHJvbWlzZTxudW1iZXIgfCBudWxsPiB7XG4gICAgICByZXR1cm4gUHJvbWlzZS5yZXNvbHZlKG51bGwpO1xuICAgIH0sXG4gIH07XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBib2R5UmVhZGVyKGNvbnRlbnRMZW5ndGg6IG51bWJlciwgcjogQnVmUmVhZGVyKTogRGVuby5SZWFkZXIge1xuICBsZXQgdG90YWxSZWFkID0gMDtcbiAgbGV0IGZpbmlzaGVkID0gZmFsc2U7XG4gIGFzeW5jIGZ1bmN0aW9uIHJlYWQoYnVmOiBVaW50OEFycmF5KTogUHJvbWlzZTxudW1iZXIgfCBudWxsPiB7XG4gICAgaWYgKGZpbmlzaGVkKSByZXR1cm4gbnVsbDtcbiAgICBsZXQgcmVzdWx0OiBudW1iZXIgfCBudWxsO1xuICAgIGNvbnN0IHJlbWFpbmluZyA9IGNvbnRlbnRMZW5ndGggLSB0b3RhbFJlYWQ7XG4gICAgaWYgKHJlbWFpbmluZyA+PSBidWYuYnl0ZUxlbmd0aCkge1xuICAgICAgcmVzdWx0ID0gYXdhaXQgci5yZWFkKGJ1Zik7XG4gICAgfSBlbHNlIHtcbiAgICAgIGNvbnN0IHJlYWRCdWYgPSBidWYuc3ViYXJyYXkoMCwgcmVtYWluaW5nKTtcbiAgICAgIHJlc3VsdCA9IGF3YWl0IHIucmVhZChyZWFkQnVmKTtcbiAgICB9XG4gICAgaWYgKHJlc3VsdCAhPT0gbnVsbCkge1xuICAgICAgdG90YWxSZWFkICs9IHJlc3VsdDtcbiAgICB9XG4gICAgZmluaXNoZWQgPSB0b3RhbFJlYWQgPT09IGNvbnRlbnRMZW5ndGg7XG4gICAgcmV0dXJuIHJlc3VsdDtcbiAgfVxuICByZXR1cm4geyByZWFkIH07XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBjaHVua2VkQm9keVJlYWRlcihoOiBIZWFkZXJzLCByOiBCdWZSZWFkZXIpOiBEZW5vLlJlYWRlciB7XG4gIC8vIEJhc2VkIG9uIGh0dHBzOi8vdG9vbHMuaWV0Zi5vcmcvaHRtbC9yZmMyNjE2I3NlY3Rpb24tMTkuNC42XG4gIGNvbnN0IHRwID0gbmV3IFRleHRQcm90b1JlYWRlcihyKTtcbiAgbGV0IGZpbmlzaGVkID0gZmFsc2U7XG4gIGNvbnN0IGNodW5rczogQXJyYXk8e1xuICAgIG9mZnNldDogbnVtYmVyO1xuICAgIGRhdGE6IFVpbnQ4QXJyYXk7XG4gIH0+ID0gW107XG4gIGFzeW5jIGZ1bmN0aW9uIHJlYWQoYnVmOiBVaW50OEFycmF5KTogUHJvbWlzZTxudW1iZXIgfCBudWxsPiB7XG4gICAgaWYgKGZpbmlzaGVkKSByZXR1cm4gbnVsbDtcbiAgICBjb25zdCBbY2h1bmtdID0gY2h1bmtzO1xuICAgIGlmIChjaHVuaykge1xuICAgICAgY29uc3QgY2h1bmtSZW1haW5pbmcgPSBjaHVuay5kYXRhLmJ5dGVMZW5ndGggLSBjaHVuay5vZmZzZXQ7XG4gICAgICBjb25zdCByZWFkTGVuZ3RoID0gTWF0aC5taW4oY2h1bmtSZW1haW5pbmcsIGJ1Zi5ieXRlTGVuZ3RoKTtcbiAgICAgIGZvciAobGV0IGkgPSAwOyBpIDwgcmVhZExlbmd0aDsgaSsrKSB7XG4gICAgICAgIGJ1ZltpXSA9IGNodW5rLmRhdGFbY2h1bmsub2Zmc2V0ICsgaV07XG4gICAgICB9XG4gICAgICBjaHVuay5vZmZzZXQgKz0gcmVhZExlbmd0aDtcbiAgICAgIGlmIChjaHVuay5vZmZzZXQgPT09IGNodW5rLmRhdGEuYnl0ZUxlbmd0aCkge1xuICAgICAgICBjaHVua3Muc2hpZnQoKTtcbiAgICAgICAgLy8gQ29uc3VtZSBcXHJcXG47XG4gICAgICAgIGlmICgoYXdhaXQgdHAucmVhZExpbmUoKSkgPT09IG51bGwpIHtcbiAgICAgICAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuVW5leHBlY3RlZEVvZigpO1xuICAgICAgICB9XG4gICAgICB9XG4gICAgICByZXR1cm4gcmVhZExlbmd0aDtcbiAgICB9XG4gICAgY29uc3QgbGluZSA9IGF3YWl0IHRwLnJlYWRMaW5lKCk7XG4gICAgaWYgKGxpbmUgPT09IG51bGwpIHRocm93IG5ldyBEZW5vLmVycm9ycy5VbmV4cGVjdGVkRW9mKCk7XG4gICAgLy8gVE9ETzogaGFuZGxlIGNodW5rIGV4dGVuc2lvblxuICAgIGNvbnN0IFtjaHVua1NpemVTdHJpbmddID0gbGluZS5zcGxpdChcIjtcIik7XG4gICAgY29uc3QgY2h1bmtTaXplID0gcGFyc2VJbnQoY2h1bmtTaXplU3RyaW5nLCAxNik7XG4gICAgaWYgKE51bWJlci5pc05hTihjaHVua1NpemUpIHx8IGNodW5rU2l6ZSA8IDApIHtcbiAgICAgIHRocm93IG5ldyBFcnJvcihcIkludmFsaWQgY2h1bmsgc2l6ZVwiKTtcbiAgICB9XG4gICAgaWYgKGNodW5rU2l6ZSA+IDApIHtcbiAgICAgIGlmIChjaHVua1NpemUgPiBidWYuYnl0ZUxlbmd0aCkge1xuICAgICAgICBsZXQgZW9mID0gYXdhaXQgci5yZWFkRnVsbChidWYpO1xuICAgICAgICBpZiAoZW9mID09PSBudWxsKSB7XG4gICAgICAgICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLlVuZXhwZWN0ZWRFb2YoKTtcbiAgICAgICAgfVxuICAgICAgICBjb25zdCByZXN0Q2h1bmsgPSBuZXcgVWludDhBcnJheShjaHVua1NpemUgLSBidWYuYnl0ZUxlbmd0aCk7XG4gICAgICAgIGVvZiA9IGF3YWl0IHIucmVhZEZ1bGwocmVzdENodW5rKTtcbiAgICAgICAgaWYgKGVvZiA9PT0gbnVsbCkge1xuICAgICAgICAgIHRocm93IG5ldyBEZW5vLmVycm9ycy5VbmV4cGVjdGVkRW9mKCk7XG4gICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgY2h1bmtzLnB1c2goe1xuICAgICAgICAgICAgb2Zmc2V0OiAwLFxuICAgICAgICAgICAgZGF0YTogcmVzdENodW5rLFxuICAgICAgICAgIH0pO1xuICAgICAgICB9XG4gICAgICAgIHJldHVybiBidWYuYnl0ZUxlbmd0aDtcbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIGNvbnN0IGJ1ZlRvRmlsbCA9IGJ1Zi5zdWJhcnJheSgwLCBjaHVua1NpemUpO1xuICAgICAgICBjb25zdCBlb2YgPSBhd2FpdCByLnJlYWRGdWxsKGJ1ZlRvRmlsbCk7XG4gICAgICAgIGlmIChlb2YgPT09IG51bGwpIHtcbiAgICAgICAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuVW5leHBlY3RlZEVvZigpO1xuICAgICAgICB9XG4gICAgICAgIC8vIENvbnN1bWUgXFxyXFxuXG4gICAgICAgIGlmICgoYXdhaXQgdHAucmVhZExpbmUoKSkgPT09IG51bGwpIHtcbiAgICAgICAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuVW5leHBlY3RlZEVvZigpO1xuICAgICAgICB9XG4gICAgICAgIHJldHVybiBjaHVua1NpemU7XG4gICAgICB9XG4gICAgfSBlbHNlIHtcbiAgICAgIGFzc2VydChjaHVua1NpemUgPT09IDApO1xuICAgICAgLy8gQ29uc3VtZSBcXHJcXG5cbiAgICAgIGlmICgoYXdhaXQgci5yZWFkTGluZSgpKSA9PT0gbnVsbCkge1xuICAgICAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuVW5leHBlY3RlZEVvZigpO1xuICAgICAgfVxuICAgICAgYXdhaXQgcmVhZFRyYWlsZXJzKGgsIHIpO1xuICAgICAgZmluaXNoZWQgPSB0cnVlO1xuICAgICAgcmV0dXJuIG51bGw7XG4gICAgfVxuICB9XG4gIHJldHVybiB7IHJlYWQgfTtcbn1cblxuZnVuY3Rpb24gaXNQcm9oaWJpZGVkRm9yVHJhaWxlcihrZXk6IHN0cmluZyk6IGJvb2xlYW4ge1xuICBjb25zdCBzID0gbmV3IFNldChbXCJ0cmFuc2Zlci1lbmNvZGluZ1wiLCBcImNvbnRlbnQtbGVuZ3RoXCIsIFwidHJhaWxlclwiXSk7XG4gIHJldHVybiBzLmhhcyhrZXkudG9Mb3dlckNhc2UoKSk7XG59XG5cbi8qKiBSZWFkIHRyYWlsZXIgaGVhZGVycyBmcm9tIHJlYWRlciBhbmQgYXBwZW5kIHZhbHVlcyB0byBoZWFkZXJzLiBcInRyYWlsZXJcIlxuICogZmllbGQgd2lsbCBiZSBkZWxldGVkLiAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIHJlYWRUcmFpbGVycyhcbiAgaGVhZGVyczogSGVhZGVycyxcbiAgcjogQnVmUmVhZGVyLFxuKTogUHJvbWlzZTx2b2lkPiB7XG4gIGNvbnN0IHRyYWlsZXJzID0gcGFyc2VUcmFpbGVyKGhlYWRlcnMuZ2V0KFwidHJhaWxlclwiKSk7XG4gIGlmICh0cmFpbGVycyA9PSBudWxsKSByZXR1cm47XG4gIGNvbnN0IHRyYWlsZXJOYW1lcyA9IFsuLi50cmFpbGVycy5rZXlzKCldO1xuICBjb25zdCB0cCA9IG5ldyBUZXh0UHJvdG9SZWFkZXIocik7XG4gIGNvbnN0IHJlc3VsdCA9IGF3YWl0IHRwLnJlYWRNSU1FSGVhZGVyKCk7XG4gIGlmIChyZXN1bHQgPT0gbnVsbCkge1xuICAgIHRocm93IG5ldyBEZW5vLmVycm9ycy5JbnZhbGlkRGF0YShcIk1pc3NpbmcgdHJhaWxlciBoZWFkZXIuXCIpO1xuICB9XG4gIGNvbnN0IHVuZGVjbGFyZWQgPSBbLi4ucmVzdWx0LmtleXMoKV0uZmlsdGVyKFxuICAgIChrKSA9PiAhdHJhaWxlck5hbWVzLmluY2x1ZGVzKGspLFxuICApO1xuICBpZiAodW5kZWNsYXJlZC5sZW5ndGggPiAwKSB7XG4gICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLkludmFsaWREYXRhKFxuICAgICAgYFVuZGVjbGFyZWQgdHJhaWxlcnM6ICR7RGVuby5pbnNwZWN0KHVuZGVjbGFyZWQpfS5gLFxuICAgICk7XG4gIH1cbiAgZm9yIChjb25zdCBbaywgdl0gb2YgcmVzdWx0KSB7XG4gICAgaGVhZGVycy5hcHBlbmQoaywgdik7XG4gIH1cbiAgY29uc3QgbWlzc2luZ1RyYWlsZXJzID0gdHJhaWxlck5hbWVzLmZpbHRlcigoaykgPT4gIXJlc3VsdC5oYXMoaykpO1xuICBpZiAobWlzc2luZ1RyYWlsZXJzLmxlbmd0aCA+IDApIHtcbiAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuSW52YWxpZERhdGEoXG4gICAgICBgTWlzc2luZyB0cmFpbGVyczogJHtEZW5vLmluc3BlY3QobWlzc2luZ1RyYWlsZXJzKX0uYCxcbiAgICApO1xuICB9XG4gIGhlYWRlcnMuZGVsZXRlKFwidHJhaWxlclwiKTtcbn1cblxuZnVuY3Rpb24gcGFyc2VUcmFpbGVyKGZpZWxkOiBzdHJpbmcgfCBudWxsKTogSGVhZGVycyB8IHVuZGVmaW5lZCB7XG4gIGlmIChmaWVsZCA9PSBudWxsKSB7XG4gICAgcmV0dXJuIHVuZGVmaW5lZDtcbiAgfVxuICBjb25zdCB0cmFpbGVyTmFtZXMgPSBmaWVsZC5zcGxpdChcIixcIikubWFwKCh2KSA9PiB2LnRyaW0oKS50b0xvd2VyQ2FzZSgpKTtcbiAgaWYgKHRyYWlsZXJOYW1lcy5sZW5ndGggPT09IDApIHtcbiAgICB0aHJvdyBuZXcgRGVuby5lcnJvcnMuSW52YWxpZERhdGEoXCJFbXB0eSB0cmFpbGVyIGhlYWRlci5cIik7XG4gIH1cbiAgY29uc3QgcHJvaGliaXRlZCA9IHRyYWlsZXJOYW1lcy5maWx0ZXIoKGspID0+IGlzUHJvaGliaWRlZEZvclRyYWlsZXIoaykpO1xuICBpZiAocHJvaGliaXRlZC5sZW5ndGggPiAwKSB7XG4gICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLkludmFsaWREYXRhKFxuICAgICAgYFByb2hpYml0ZWQgdHJhaWxlciBuYW1lczogJHtEZW5vLmluc3BlY3QocHJvaGliaXRlZCl9LmAsXG4gICAgKTtcbiAgfVxuICByZXR1cm4gbmV3IEhlYWRlcnModHJhaWxlck5hbWVzLm1hcCgoa2V5KSA9PiBba2V5LCBcIlwiXSkpO1xufVxuXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gd3JpdGVDaHVua2VkQm9keShcbiAgdzogRGVuby5Xcml0ZXIsXG4gIHI6IERlbm8uUmVhZGVyLFxuKTogUHJvbWlzZTx2b2lkPiB7XG4gIGNvbnN0IHdyaXRlciA9IEJ1ZldyaXRlci5jcmVhdGUodyk7XG4gIGZvciBhd2FpdCAoY29uc3QgY2h1bmsgb2YgRGVuby5pdGVyKHIpKSB7XG4gICAgaWYgKGNodW5rLmJ5dGVMZW5ndGggPD0gMCkgY29udGludWU7XG4gICAgY29uc3Qgc3RhcnQgPSBlbmNvZGVyLmVuY29kZShgJHtjaHVuay5ieXRlTGVuZ3RoLnRvU3RyaW5nKDE2KX1cXHJcXG5gKTtcbiAgICBjb25zdCBlbmQgPSBlbmNvZGVyLmVuY29kZShcIlxcclxcblwiKTtcbiAgICBhd2FpdCB3cml0ZXIud3JpdGUoc3RhcnQpO1xuICAgIGF3YWl0IHdyaXRlci53cml0ZShjaHVuayk7XG4gICAgYXdhaXQgd3JpdGVyLndyaXRlKGVuZCk7XG4gIH1cblxuICBjb25zdCBlbmRDaHVuayA9IGVuY29kZXIuZW5jb2RlKFwiMFxcclxcblxcclxcblwiKTtcbiAgYXdhaXQgd3JpdGVyLndyaXRlKGVuZENodW5rKTtcbn1cblxuLyoqIFdyaXRlIHRyYWlsZXIgaGVhZGVycyB0byB3cml0ZXIuIEl0IHNob3VsZCBtb3N0bHkgc2hvdWxkIGJlIGNhbGxlZCBhZnRlclxuICogYHdyaXRlUmVzcG9uc2UoKWAuICovXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gd3JpdGVUcmFpbGVycyhcbiAgdzogRGVuby5Xcml0ZXIsXG4gIGhlYWRlcnM6IEhlYWRlcnMsXG4gIHRyYWlsZXJzOiBIZWFkZXJzLFxuKTogUHJvbWlzZTx2b2lkPiB7XG4gIGNvbnN0IHRyYWlsZXIgPSBoZWFkZXJzLmdldChcInRyYWlsZXJcIik7XG4gIGlmICh0cmFpbGVyID09PSBudWxsKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcihcIk1pc3NpbmcgdHJhaWxlciBoZWFkZXIuXCIpO1xuICB9XG4gIGNvbnN0IHRyYW5zZmVyRW5jb2RpbmcgPSBoZWFkZXJzLmdldChcInRyYW5zZmVyLWVuY29kaW5nXCIpO1xuICBpZiAodHJhbnNmZXJFbmNvZGluZyA9PT0gbnVsbCB8fCAhdHJhbnNmZXJFbmNvZGluZy5tYXRjaCgvXmNodW5rZWQvKSkge1xuICAgIHRocm93IG5ldyBUeXBlRXJyb3IoXG4gICAgICBgVHJhaWxlcnMgYXJlIG9ubHkgYWxsb3dlZCBmb3IgXCJ0cmFuc2Zlci1lbmNvZGluZzogY2h1bmtlZFwiLCBnb3QgXCJ0cmFuc2Zlci1lbmNvZGluZzogJHt0cmFuc2ZlckVuY29kaW5nfVwiLmAsXG4gICAgKTtcbiAgfVxuICBjb25zdCB3cml0ZXIgPSBCdWZXcml0ZXIuY3JlYXRlKHcpO1xuICBjb25zdCB0cmFpbGVyTmFtZXMgPSB0cmFpbGVyLnNwbGl0KFwiLFwiKS5tYXAoKHMpID0+IHMudHJpbSgpLnRvTG93ZXJDYXNlKCkpO1xuICBjb25zdCBwcm9oaWJpdGVkVHJhaWxlcnMgPSB0cmFpbGVyTmFtZXMuZmlsdGVyKChrKSA9PlxuICAgIGlzUHJvaGliaWRlZEZvclRyYWlsZXIoaylcbiAgKTtcbiAgaWYgKHByb2hpYml0ZWRUcmFpbGVycy5sZW5ndGggPiAwKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcihcbiAgICAgIGBQcm9oaWJpdGVkIHRyYWlsZXIgbmFtZXM6ICR7RGVuby5pbnNwZWN0KHByb2hpYml0ZWRUcmFpbGVycyl9LmAsXG4gICAgKTtcbiAgfVxuICBjb25zdCB1bmRlY2xhcmVkID0gWy4uLnRyYWlsZXJzLmtleXMoKV0uZmlsdGVyKFxuICAgIChrKSA9PiAhdHJhaWxlck5hbWVzLmluY2x1ZGVzKGspLFxuICApO1xuICBpZiAodW5kZWNsYXJlZC5sZW5ndGggPiAwKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcihgVW5kZWNsYXJlZCB0cmFpbGVyczogJHtEZW5vLmluc3BlY3QodW5kZWNsYXJlZCl9LmApO1xuICB9XG4gIGZvciAoY29uc3QgW2tleSwgdmFsdWVdIG9mIHRyYWlsZXJzKSB7XG4gICAgYXdhaXQgd3JpdGVyLndyaXRlKGVuY29kZXIuZW5jb2RlKGAke2tleX06ICR7dmFsdWV9XFxyXFxuYCkpO1xuICB9XG4gIGF3YWl0IHdyaXRlci53cml0ZShlbmNvZGVyLmVuY29kZShcIlxcclxcblwiKSk7XG4gIGF3YWl0IHdyaXRlci5mbHVzaCgpO1xufVxuXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gd3JpdGVSZXNwb25zZShcbiAgdzogRGVuby5Xcml0ZXIsXG4gIHI6IFJlc3BvbnNlLFxuKTogUHJvbWlzZTx2b2lkPiB7XG4gIGNvbnN0IHByb3RvTWFqb3IgPSAxO1xuICBjb25zdCBwcm90b01pbm9yID0gMTtcbiAgY29uc3Qgc3RhdHVzQ29kZSA9IHIuc3RhdHVzIHx8IDIwMDtcbiAgY29uc3Qgc3RhdHVzVGV4dCA9IFNUQVRVU19URVhULmdldChzdGF0dXNDb2RlKTtcbiAgY29uc3Qgd3JpdGVyID0gQnVmV3JpdGVyLmNyZWF0ZSh3KTtcbiAgaWYgKCFzdGF0dXNUZXh0KSB7XG4gICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLkludmFsaWREYXRhKFwiQmFkIHN0YXR1cyBjb2RlXCIpO1xuICB9XG4gIGlmICghci5ib2R5KSB7XG4gICAgci5ib2R5ID0gbmV3IFVpbnQ4QXJyYXkoKTtcbiAgfVxuICBpZiAodHlwZW9mIHIuYm9keSA9PT0gXCJzdHJpbmdcIikge1xuICAgIHIuYm9keSA9IGVuY29kZXIuZW5jb2RlKHIuYm9keSk7XG4gIH1cblxuICBsZXQgb3V0ID0gYEhUVFAvJHtwcm90b01ham9yfS4ke3Byb3RvTWlub3J9ICR7c3RhdHVzQ29kZX0gJHtzdGF0dXNUZXh0fVxcclxcbmA7XG5cbiAgY29uc3QgaGVhZGVycyA9IHIuaGVhZGVycyA/PyBuZXcgSGVhZGVycygpO1xuXG4gIGlmIChyLmJvZHkgJiYgIWhlYWRlcnMuZ2V0KFwiY29udGVudC1sZW5ndGhcIikpIHtcbiAgICBpZiAoci5ib2R5IGluc3RhbmNlb2YgVWludDhBcnJheSkge1xuICAgICAgb3V0ICs9IGBjb250ZW50LWxlbmd0aDogJHtyLmJvZHkuYnl0ZUxlbmd0aH1cXHJcXG5gO1xuICAgIH0gZWxzZSBpZiAoIWhlYWRlcnMuZ2V0KFwidHJhbnNmZXItZW5jb2RpbmdcIikpIHtcbiAgICAgIG91dCArPSBcInRyYW5zZmVyLWVuY29kaW5nOiBjaHVua2VkXFxyXFxuXCI7XG4gICAgfVxuICB9XG5cbiAgZm9yIChjb25zdCBba2V5LCB2YWx1ZV0gb2YgaGVhZGVycykge1xuICAgIG91dCArPSBgJHtrZXl9OiAke3ZhbHVlfVxcclxcbmA7XG4gIH1cblxuICBvdXQgKz0gYFxcclxcbmA7XG5cbiAgY29uc3QgaGVhZGVyID0gZW5jb2Rlci5lbmNvZGUob3V0KTtcbiAgY29uc3QgbiA9IGF3YWl0IHdyaXRlci53cml0ZShoZWFkZXIpO1xuICBhc3NlcnQobiA9PT0gaGVhZGVyLmJ5dGVMZW5ndGgpO1xuXG4gIGlmIChyLmJvZHkgaW5zdGFuY2VvZiBVaW50OEFycmF5KSB7XG4gICAgY29uc3QgbiA9IGF3YWl0IHdyaXRlci53cml0ZShyLmJvZHkpO1xuICAgIGFzc2VydChuID09PSByLmJvZHkuYnl0ZUxlbmd0aCk7XG4gIH0gZWxzZSBpZiAoaGVhZGVycy5oYXMoXCJjb250ZW50LWxlbmd0aFwiKSkge1xuICAgIGNvbnN0IGNvbnRlbnRMZW5ndGggPSBoZWFkZXJzLmdldChcImNvbnRlbnQtbGVuZ3RoXCIpO1xuICAgIGFzc2VydChjb250ZW50TGVuZ3RoICE9IG51bGwpO1xuICAgIGNvbnN0IGJvZHlMZW5ndGggPSBwYXJzZUludChjb250ZW50TGVuZ3RoKTtcbiAgICBjb25zdCBuID0gYXdhaXQgRGVuby5jb3B5KHIuYm9keSwgd3JpdGVyKTtcbiAgICBhc3NlcnQobiA9PT0gYm9keUxlbmd0aCk7XG4gIH0gZWxzZSB7XG4gICAgYXdhaXQgd3JpdGVDaHVua2VkQm9keSh3cml0ZXIsIHIuYm9keSk7XG4gIH1cbiAgaWYgKHIudHJhaWxlcnMpIHtcbiAgICBjb25zdCB0ID0gYXdhaXQgci50cmFpbGVycygpO1xuICAgIGF3YWl0IHdyaXRlVHJhaWxlcnMod3JpdGVyLCBoZWFkZXJzLCB0KTtcbiAgfVxuICBhd2FpdCB3cml0ZXIuZmx1c2goKTtcbn1cblxuLyoqXG4gKiBQYXJzZUhUVFBWZXJzaW9uIHBhcnNlcyBhIEhUVFAgdmVyc2lvbiBzdHJpbmcuXG4gKiBcIkhUVFAvMS4wXCIgcmV0dXJucyAoMSwgMCkuXG4gKiBQb3J0ZWQgZnJvbSBodHRwczovL2dpdGh1Yi5jb20vZ29sYW5nL2dvL2Jsb2IvZjVjNDNiOS9zcmMvbmV0L2h0dHAvcmVxdWVzdC5nbyNMNzY2LUw3OTJcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIHBhcnNlSFRUUFZlcnNpb24odmVyczogc3RyaW5nKTogW251bWJlciwgbnVtYmVyXSB7XG4gIHN3aXRjaCAodmVycykge1xuICAgIGNhc2UgXCJIVFRQLzEuMVwiOlxuICAgICAgcmV0dXJuIFsxLCAxXTtcblxuICAgIGNhc2UgXCJIVFRQLzEuMFwiOlxuICAgICAgcmV0dXJuIFsxLCAwXTtcblxuICAgIGRlZmF1bHQ6IHtcbiAgICAgIGNvbnN0IEJpZyA9IDEwMDAwMDA7IC8vIGFyYml0cmFyeSB1cHBlciBib3VuZFxuXG4gICAgICBpZiAoIXZlcnMuc3RhcnRzV2l0aChcIkhUVFAvXCIpKSB7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuXG4gICAgICBjb25zdCBkb3QgPSB2ZXJzLmluZGV4T2YoXCIuXCIpO1xuICAgICAgaWYgKGRvdCA8IDApIHtcbiAgICAgICAgYnJlYWs7XG4gICAgICB9XG5cbiAgICAgIGNvbnN0IG1ham9yU3RyID0gdmVycy5zdWJzdHJpbmcodmVycy5pbmRleE9mKFwiL1wiKSArIDEsIGRvdCk7XG4gICAgICBjb25zdCBtYWpvciA9IE51bWJlcihtYWpvclN0cik7XG4gICAgICBpZiAoIU51bWJlci5pc0ludGVnZXIobWFqb3IpIHx8IG1ham9yIDwgMCB8fCBtYWpvciA+IEJpZykge1xuICAgICAgICBicmVhaztcbiAgICAgIH1cblxuICAgICAgY29uc3QgbWlub3JTdHIgPSB2ZXJzLnN1YnN0cmluZyhkb3QgKyAxKTtcbiAgICAgIGNvbnN0IG1pbm9yID0gTnVtYmVyKG1pbm9yU3RyKTtcbiAgICAgIGlmICghTnVtYmVyLmlzSW50ZWdlcihtaW5vcikgfHwgbWlub3IgPCAwIHx8IG1pbm9yID4gQmlnKSB7XG4gICAgICAgIGJyZWFrO1xuICAgICAgfVxuXG4gICAgICByZXR1cm4gW21ham9yLCBtaW5vcl07XG4gICAgfVxuICB9XG5cbiAgdGhyb3cgbmV3IEVycm9yKGBtYWxmb3JtZWQgSFRUUCB2ZXJzaW9uICR7dmVyc31gKTtcbn1cblxuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIHJlYWRSZXF1ZXN0KFxuICBjb25uOiBEZW5vLkNvbm4sXG4gIGJ1ZnI6IEJ1ZlJlYWRlcixcbik6IFByb21pc2U8U2VydmVyUmVxdWVzdCB8IG51bGw+IHtcbiAgY29uc3QgdHAgPSBuZXcgVGV4dFByb3RvUmVhZGVyKGJ1ZnIpO1xuICBjb25zdCBmaXJzdExpbmUgPSBhd2FpdCB0cC5yZWFkTGluZSgpOyAvLyBlLmcuIEdFVCAvaW5kZXguaHRtbCBIVFRQLzEuMFxuICBpZiAoZmlyc3RMaW5lID09PSBudWxsKSByZXR1cm4gbnVsbDtcbiAgY29uc3QgaGVhZGVycyA9IGF3YWl0IHRwLnJlYWRNSU1FSGVhZGVyKCk7XG4gIGlmIChoZWFkZXJzID09PSBudWxsKSB0aHJvdyBuZXcgRGVuby5lcnJvcnMuVW5leHBlY3RlZEVvZigpO1xuXG4gIGNvbnN0IHJlcSA9IG5ldyBTZXJ2ZXJSZXF1ZXN0KCk7XG4gIHJlcS5jb25uID0gY29ubjtcbiAgcmVxLnIgPSBidWZyO1xuICBbcmVxLm1ldGhvZCwgcmVxLnVybCwgcmVxLnByb3RvXSA9IGZpcnN0TGluZS5zcGxpdChcIiBcIiwgMyk7XG4gIFtyZXEucHJvdG9NaW5vciwgcmVxLnByb3RvTWFqb3JdID0gcGFyc2VIVFRQVmVyc2lvbihyZXEucHJvdG8pO1xuICByZXEuaGVhZGVycyA9IGhlYWRlcnM7XG4gIGZpeExlbmd0aChyZXEpO1xuICByZXR1cm4gcmVxO1xufVxuXG5mdW5jdGlvbiBmaXhMZW5ndGgocmVxOiBTZXJ2ZXJSZXF1ZXN0KTogdm9pZCB7XG4gIGNvbnN0IGNvbnRlbnRMZW5ndGggPSByZXEuaGVhZGVycy5nZXQoXCJDb250ZW50LUxlbmd0aFwiKTtcbiAgaWYgKGNvbnRlbnRMZW5ndGgpIHtcbiAgICBjb25zdCBhcnJDbGVuID0gY29udGVudExlbmd0aC5zcGxpdChcIixcIik7XG4gICAgaWYgKGFyckNsZW4ubGVuZ3RoID4gMSkge1xuICAgICAgY29uc3QgZGlzdGluY3QgPSBbLi4ubmV3IFNldChhcnJDbGVuLm1hcCgoZSk6IHN0cmluZyA9PiBlLnRyaW0oKSkpXTtcbiAgICAgIGlmIChkaXN0aW5jdC5sZW5ndGggPiAxKSB7XG4gICAgICAgIHRocm93IEVycm9yKFwiY2Fubm90IGNvbnRhaW4gbXVsdGlwbGUgQ29udGVudC1MZW5ndGggaGVhZGVyc1wiKTtcbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIHJlcS5oZWFkZXJzLnNldChcIkNvbnRlbnQtTGVuZ3RoXCIsIGRpc3RpbmN0WzBdKTtcbiAgICAgIH1cbiAgICB9XG4gICAgY29uc3QgYyA9IHJlcS5oZWFkZXJzLmdldChcIkNvbnRlbnQtTGVuZ3RoXCIpO1xuICAgIGlmIChyZXEubWV0aG9kID09PSBcIkhFQURcIiAmJiBjICYmIGMgIT09IFwiMFwiKSB7XG4gICAgICB0aHJvdyBFcnJvcihcImh0dHA6IG1ldGhvZCBjYW5ub3QgY29udGFpbiBhIENvbnRlbnQtTGVuZ3RoXCIpO1xuICAgIH1cbiAgICBpZiAoYyAmJiByZXEuaGVhZGVycy5oYXMoXCJ0cmFuc2Zlci1lbmNvZGluZ1wiKSkge1xuICAgICAgLy8gQSBzZW5kZXIgTVVTVCBOT1Qgc2VuZCBhIENvbnRlbnQtTGVuZ3RoIGhlYWRlciBmaWVsZCBpbiBhbnkgbWVzc2FnZVxuICAgICAgLy8gdGhhdCBjb250YWlucyBhIFRyYW5zZmVyLUVuY29kaW5nIGhlYWRlciBmaWVsZC5cbiAgICAgIC8vIHJmYzogaHR0cHM6Ly90b29scy5pZXRmLm9yZy9odG1sL3JmYzcyMzAjc2VjdGlvbi0zLjMuMlxuICAgICAgdGhyb3cgbmV3IEVycm9yKFxuICAgICAgICBcImh0dHA6IFRyYW5zZmVyLUVuY29kaW5nIGFuZCBDb250ZW50LUxlbmd0aCBjYW5ub3QgYmUgc2VuZCB0b2dldGhlclwiLFxuICAgICAgKTtcbiAgICB9XG4gIH1cbn1cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxFQUFBLHdFQUFBO1NBQ0EsU0FBQSxTQUFBLGNBQUE7U0FDQSxlQUFBLFNBQUEsbUJBQUE7U0FDQSxNQUFBLFNBQUEsa0JBQUE7U0FDQSxPQUFBLFNBQUEsbUJBQUE7U0FDQSxhQUFBLFNBQUEsV0FBQTtTQUNBLFdBQUEsU0FBQSxnQkFBQTtnQkFFQSxXQUFBOztBQUVBLFlBQUEsRUFBQSxDQUFBO21CQUNBLE9BQUEsQ0FBQSxPQUFBLENBQUEsSUFBQTs7OztnQkFLQSxVQUFBLENBQUEsYUFBQSxFQUFBLENBQUE7UUFDQSxTQUFBLEdBQUEsQ0FBQTtRQUNBLFFBQUEsR0FBQSxLQUFBO21CQUNBLElBQUEsQ0FBQSxHQUFBO1lBQ0EsUUFBQSxTQUFBLElBQUE7WUFDQSxNQUFBO2NBQ0EsU0FBQSxHQUFBLGFBQUEsR0FBQSxTQUFBO1lBQ0EsU0FBQSxJQUFBLEdBQUEs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OzBCQUNBLElBQUEsQ0FBQSxNQUFBLENBQUEsYUFBQTs7a0JBRUEsWUFBQSxDQUFBLENBQUEsRUFBQSxDQUFBO0FBQ0Esb0JBQUEsR0FBQSxJQUFBO21CQUNBLElBQUE7Ozs7QUFHQSxZQUFBOzs7U0FHQSxzQkFBQSxDQUFBLEdBQUE7VUFDQSxDQUFBLE9BQUEsR0FBQTtTQUFBLGlCQUFBO1NBQUEsY0FBQTtTQUFBLE9BQUE7O1dBQ0EsQ0FBQSxDQUFBLEdBQUEsQ0FBQSxHQUFBLENBQUEsV0FBQTs7QUFHQSxFQUNBLEFBREEscUdBQ0EsQUFEQSxFQUNBLHVCQUNBLFlBQUEsQ0FDQSxPQUFBLEVBQ0EsQ0FBQTtVQUVBLFFBQUEsR0FBQSxZQUFBLENBQUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxPQUFBO1FBQ0EsUUFBQSxJQUFBLElBQUE7VUFDQSxZQUFBO1dBQUEsUUFBQSxDQUFBLElBQUE7O1VBQ0EsRUFBQSxPQUFBLGVBQUEsQ0FBQSxDQUFBO1VBQ0EsTUFBQSxTQUFBLEVBQUEsQ0FBQSxjQUFBO1FBQ0EsTUFBQSxJQUFBLElBQUE7a0JBQ0EsSUFBQSxDQUFBLE1BQUEsQ0FBQSxXQUFBLEVBQUEsdUJBQUE7O1VBRUEsVUFBQTtXQUFBLE1BQUEsQ0FBQSxJQUFBO01BQUEsTUFBQSxFQUNBLENBQUEsSUFBQSxZQUFBLENBQUEsUUFBQSxDQUFBLENBQUE7O1FBRUEsVUFBQSxDQUFBLE1BQUEsR0FBQSxDQUFBO2tCQUNBLElBQUEsQ0FBQSxNQUFBLENBQUEsV0FBQSxFQUNBLHFCQUFBLEVBQUEsSUFBQSxDQUFBLE9BQUEsQ0FBQSxVQUFBLEVBQUEsQ0FBQTs7Z0JBR0EsQ0FBQSxFQUFBLENBQUEsS0FBQSxNQUFBO0FBQ0EsZUFBQSxDQUFBLE1BQUEsQ0FBQSxDQUFBLEVBQUEsQ0FBQTs7VUFFQSxlQUFBLEdBQUEsWUFBQSxDQUFBLE1BQUEsRUFBQSxDQUFBLElBQUEsTUFBQSxDQUFBLEdBQUEsQ0FBQSxDQUFBOztRQUNBLGVBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQTtrQkFDQSxJQUFBLENBQUEsTUFBQSxDQUFBLFdBQUEsRUFDQSxrQkFBQSxFQUFBLElBQUEsQ0FBQSxPQUFBLENBQUEsZUFBQSxFQUFBLENBQUE7O0FBR0EsV0FBQSxDQUFBLE1BQUEsRUFBQSxPQUFBOztTQUdBLFlBQUEsQ0FBQSxLQUFBO1FBQ0EsS0FBQSxJQUFBLElBQUE7ZUFDQSxTQUFBOztVQUVBLFlBQUEsR0FBQSxLQUFBLENBQUEsS0FBQSxFQUFBLENBQUEsR0FBQSxHQUFBLEVBQUEsQ0FBQSxHQUFBLENBQUEsQ0FBQSxJQUFBLEdBQUEsV0FBQTs7UUFDQSxZQUFBLENBQUEsTUFBQSxLQUFBLENBQUE7a0JBQ0EsSUFBQSxDQUFBLE1BQUEsQ0FBQSxXQUFBLEVBQUEscUJBQUE7O1VBRUEsVUFBQSxHQUFBLFlBQUEsQ0FBQSxNQUFBLEVBQUEsQ0FBQSxHQUFBLHNCQUFBLENBQUEsQ0FBQTs7UUFDQSxVQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7a0JBQ0EsSUFBQSxDQUFBLE1BQUEsQ0FBQSxXQUFBLEVBQ0EsMEJBQUEsRUFBQSxJQUFBLENBQUEsT0FBQSxDQUFBLFVBQUEsRUFBQSxDQUFBOztlQUdBLE9BQUEsQ0FBQSxZQUFBLENBQUEsR0FBQSxFQUFBLEdBQUE7QUFBQSxlQUFBOzs7OztzQkFHQSxnQkFBQSxDQUNBLENBQUEsRUFDQSxDQUFBO1VBRUEsTUFBQSxHQUFBLFNBQUEsQ0FBQSxNQUFBLENBQUEsQ0FBQTtxQkFDQSxLQUFBLElBQUEsSUFBQSxDQUFBLElBQUEsQ0FBQSxDQUFBO1lBQ0EsS0FBQSxDQUFBLFVBQUEsSUFBQSxDQUFBO2NBQ0EsS0FBQSxHQUFBLE9BQUEsQ0FBQSxNQUFBLElBQUEsS0FBQSxDQUFBLFVBQUEsQ0FBQSxRQUFBLENBQUEsRUFBQSxFQUFBLElBQUE7Y0FDQSxHQUFBLEdBQUEsT0FBQSxDQUFBLE1BQUEsRUFBQSxJQUFBO2NBQ0EsTUFBQSxDQUFBLEtBQUEsQ0FBQSxLQUFBO2NBQ0EsTUFBQSxDQUFBLEtBQUEsQ0FBQSxLQUFBO2NBQ0EsTUFBQSxDQUFBLEtBQUEsQ0FBQSxHQUFBOztVQUdBLFFBQUEsR0FBQSxPQUFBLENBQUEsTUFBQSxFQUFBLFNBQUE7VUFDQSxNQUFBLENBQUEsS0FBQSxDQUFBLFFBQUE7O0FBR0EsRUFDQSxBQURBLGlHQUNBLEFBREEsRUFDQSx1QkFDQSxhQUFBLENBQ0EsQ0FBQSxFQUNBLE9BQUEsRUFDQSxRQUFBO1VBRUEsT0FBQSxHQUFBLE9BQUEsQ0FBQSxHQUFBLEVBQUEsT0FBQTtRQUNBLE9BQUEsS0FBQSxJQUFBO2tCQUNBLFNBQUEsRUFBQSx1QkFBQTs7VUFFQSxnQkFBQSxHQUFBLE9BQUEsQ0FBQSxHQUFBLEVBQUEsaUJBQUE7UUFDQSxnQkFBQSxLQUFBLElBQUEsS0FBQSxnQkFBQSxDQUFBLEtBQUE7a0JBQ0EsU0FBQSxFQUNBLG9GQUFBLEVBQUEsZ0JBQUEsQ0FBQSxFQUFBOztVQUdBLE1BQUEsR0FBQSxTQUFBLENBQUEsTUFBQSxDQUFBLENBQUE7VUFDQSxZQUFBLEdBQUEsT0FBQSxDQUFBLEtBQUEsRUFBQSxDQUFBLEdBQUEsR0FBQSxFQUFBLENBQUEsR0FBQSxDQUFBLENBQUEsSUFBQSxHQUFBLFdBQUE7O1VBQ0Esa0JBQUEsR0FBQSxZQUFBLENBQUEsTUFBQSxFQUFBLENBQUEsR0FDQSxzQkFBQSxDQUFBLENBQUE7O1FBRUEsa0JBQUEsQ0FBQSxNQUFBLEdBQUEsQ0FBQTtrQkFDQSxTQUFBLEVBQ0EsMEJBQUEsRUFBQSxJQUFBLENBQUEsT0FBQSxDQUFBLGtCQUFBLEVBQUEsQ0FBQTs7VUFHQSxVQUFBO1dBQUEsUUFBQSxDQUFBLElBQUE7TUFBQSxNQUFBLEVBQ0EsQ0FBQSxJQUFBLFlBQUEsQ0FBQSxRQUFBLENBQUEsQ0FBQTs7UUFFQSxVQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7a0JBQ0EsU0FBQSxFQUFBLHFCQUFBLEVBQUEsSUFBQSxDQUFBLE9BQUEsQ0FBQSxVQUFBLEVBQUEsQ0FBQTs7Z0JBRUEsR0FBQSxFQUFBLEtBQUEsS0FBQSxRQUFBO2NBQ0EsTUFBQSxDQUFBLEtBQUEsQ0FBQSxPQUFBLENBQUEsTUFBQSxJQUFBLEdBQUEsQ0FBQSxFQUFBLEVB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QUEsQ0FBQTs7O3NCQUlBLFFBQUEsR0FBQSxJQUFBLENBQUEsU0FBQSxDQUFBLElBQUEsQ0FBQSxPQUFBLEVBQUEsQ0FBQSxLQUFBLENBQUEsRUFBQSxHQUFBO3NCQUNBLEtBQUEsR0FBQSxNQUFBLENBQUEsUUFBQTtxQkFDQSxNQUFBLENBQUEsU0FBQSxDQUFBLEtBQUEsS0FBQSxLQUFBLEdBQUEsQ0FBQSxJQUFBLEtBQUEsR0FBQSxHQUFBOzs7c0JBSUEsUUFBQSxHQUFBLElBQUEsQ0FBQSxTQUFBLENBQUEsR0FBQSxHQUFBLENBQUE7c0JBQ0EsS0FBQSxHQUFBLE1BQUEsQ0FBQSxRQUFBO3FCQUNBLE1BQUEsQ0FBQSxTQUFBLENBQUEsS0FBQSxLQUFBLEtBQUEsR0FBQSxDQUFBLElBQUEsS0FBQSxHQUFBLEdBQUE7Ozs7QUFJQSx5QkFBQTtBQUFBLHlCQUFBOzs7O2NBSUEsS0FBQSxFQUFBLHVCQUFBLEVBQUEsSUFBQTs7c0JBR0EsV0FBQSxDQUNBLElBQUEsRUFDQSxJQUFBO1VBRUEsRUFBQSxPQUFBLGVBQUEsQ0FBQSxJQUFBO1VBQ0EsU0FBQSxTQUFBLEVBQUEsQ0FBQSxRQUFBLEdBQUEsQ0FBQSxFQUFBLDhCQUFBO1FBQ0EsU0FBQSxLQUFBLElBQUEsU0FBQSxJQUFBO1VBQ0EsT0FBQSxTQUFBLEVBQUEsQ0FBQSxjQUFBO1FBQ0EsT0FBQSxLQUFBLElBQUEsWUFBQSxJQUFBLENBQUEsTUFBQSxDQUFBLGFBQUE7VUFFQSxHQUFBLE9BQUEsYUFBQTtBQUNBLE9BQUEsQ0FBQSxJQUFBLEdBQUEsSUFBQTtBQUNBLE9BQUEsQ0FBQSxDQUFBLEdBQUEsSUFBQTtLQUNBLEdBQUEsQ0FBQSxNQUFBLEVBQUEsR0FBQSxDQUFBLEdBQUEsRUFBQSxHQUFBLENBQUEsS0FBQSxJQUFBLFNBQUEsQ0FBQSxLQUFBLEVBQUEsQ0FBQSxHQUFBLENBQUE7S0FDQSxHQUFBLENBQUEsVUFBQSxFQUFBLEdBQUEsQ0FBQSxVQUFBLElBQUEsZ0JBQUEsQ0FBQSxHQUFBLENBQUEsS0FBQTtBQUNBLE9BQUEsQ0FBQSxPQUFBLEdBQUEsT0FBQTtBQUNBLGFBQUEsQ0FBQSxHQUFBO1dBQ0EsR0FBQTs7U0FHQSxTQUFBLENBQUEsR0FBQTtVQUNBLGFBQUEsR0FBQSxHQUFBLENBQUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxjQUFBO1FBQ0EsYUFBQTtjQUNBLE9BQUEsR0FBQSxhQUFBLENBQUEsS0FBQSxFQUFBLENBQUE7WUFDQSxPQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7a0JBQ0EsUUFBQTt1QkFBQSxHQUFBLENBQUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxDQUFBLEdBQUEsQ0FBQSxDQUFBLElBQUE7OztnQkFDQSxRQUFBLENBQUEsTUFBQSxHQUFBLENBQUE7c0JBQ0EsS0FBQSxFQUFBLDhDQUFBOztBQUVBLG1CQUFBLENBQUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxjQUFBLEdBQUEsUUFBQSxDQUFBLENBQUE7OztjQUdBLENBQUEsR0FBQSxHQUFBLENBQUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxjQUFBO1lBQ0EsR0FBQSxDQUFBLE1BQUEsTUFBQSxJQUFBLEtBQUEsQ0FBQSxJQUFBLENBQUEsTUFBQSxDQUFBO2tCQUNBLEtBQUEsRUFBQSw0Q0FBQTs7WUFFQSxDQUFBLElBQUEsR0FBQSxDQUFBLE9BQUEsQ0FBQSxHQUFBLEVBQUEsaUJBQUE7QUFDQSxjQUFBLG9FQUFBO0FBQ0EsY0FBQSxnREFBQTtBQUNBLGNBQUEsdURBQUE7c0JBQ0EsS0FBQSxFQUNBLGtFQUFBIn0=