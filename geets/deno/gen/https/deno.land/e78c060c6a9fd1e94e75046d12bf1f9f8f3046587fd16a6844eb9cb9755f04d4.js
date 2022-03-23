// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { decode, encode } from "../encoding/utf8.ts";
import { hasOwnProperty } from "../_util/has_own_property.ts";
import { BufReader, BufWriter } from "../io/bufio.ts";
import { readLong, readShort, sliceLongToBytes } from "../io/ioutil.ts";
import { Sha1 } from "../hash/sha1.ts";
import { writeResponse } from "../http/_io.ts";
import { TextProtoReader } from "../textproto/mod.ts";
import { deferred } from "../async/deferred.ts";
import { assert } from "../_util/assert.ts";
import { concat } from "../bytes/mod.ts";
export var OpCode;
(function(OpCode) {
    OpCode[OpCode["Continue"] = 0] = "Continue";
    OpCode[OpCode["TextFrame"] = 1] = "TextFrame";
    OpCode[OpCode["BinaryFrame"] = 2] = "BinaryFrame";
    OpCode[OpCode["Close"] = 8] = "Close";
    OpCode[OpCode["Ping"] = 9] = "Ping";
    OpCode[OpCode["Pong"] = 10] = "Pong";
})(OpCode || (OpCode = {
}));
export function isWebSocketCloseEvent(a) {
    return hasOwnProperty(a, "code");
}
export function isWebSocketPingEvent(a) {
    return Array.isArray(a) && a[0] === "ping" && a[1] instanceof Uint8Array;
}
export function isWebSocketPongEvent(a) {
    return Array.isArray(a) && a[0] === "pong" && a[1] instanceof Uint8Array;
}
/** Unmask masked websocket payload */ export function unmask(payload, mask) {
    if (mask) {
        for(let i = 0, len = payload.length; i < len; i++){
            payload[i] ^= mask[i & 3];
        }
    }
}
/** Write websocket frame to given writer */ export async function writeFrame(frame, writer) {
    const payloadLength = frame.payload.byteLength;
    let header;
    const hasMask = frame.mask ? 128 : 0;
    if (frame.mask && frame.mask.byteLength !== 4) {
        throw new Error("invalid mask. mask must be 4 bytes: length=" + frame.mask.byteLength);
    }
    if (payloadLength < 126) {
        header = new Uint8Array([
            128 | frame.opcode,
            hasMask | payloadLength
        ]);
    } else if (payloadLength < 65535) {
        header = new Uint8Array([
            128 | frame.opcode,
            hasMask | 126,
            payloadLength >>> 8,
            payloadLength & 255, 
        ]);
    } else {
        header = new Uint8Array([
            128 | frame.opcode,
            hasMask | 127,
            ...sliceLongToBytes(payloadLength), 
        ]);
    }
    if (frame.mask) {
        header = concat(header, frame.mask);
    }
    unmask(frame.payload, frame.mask);
    header = concat(header, frame.payload);
    const w = BufWriter.create(writer);
    await w.write(header);
    await w.flush();
}
/** Read websocket frame from given BufReader
 * @throws `Deno.errors.UnexpectedEof` When peer closed connection without close frame
 * @throws `Error` Frame is invalid
 */ export async function readFrame(buf) {
    let b = await buf.readByte();
    assert(b !== null);
    let isLastFrame = false;
    switch(b >>> 4){
        case 8:
            isLastFrame = true;
            break;
        case 0:
            isLastFrame = false;
            break;
        default:
            throw new Error("invalid signature");
    }
    const opcode = b & 15;
    // has_mask & payload
    b = await buf.readByte();
    assert(b !== null);
    const hasMask = b >>> 7;
    let payloadLength = b & 127;
    if (payloadLength === 126) {
        const l = await readShort(buf);
        assert(l !== null);
        payloadLength = l;
    } else if (payloadLength === 127) {
        const l = await readLong(buf);
        assert(l !== null);
        payloadLength = Number(l);
    }
    // mask
    let mask;
    if (hasMask) {
        mask = new Uint8Array(4);
        assert(await buf.readFull(mask) !== null);
    }
    // payload
    const payload = new Uint8Array(payloadLength);
    assert(await buf.readFull(payload) !== null);
    return {
        isLastFrame,
        opcode,
        mask,
        payload
    };
}
class WebSocketImpl {
    sendQueue = [];
    constructor({ conn , bufReader , bufWriter , mask  }){
        this.conn = conn;
        this.mask = mask;
        this.bufReader = bufReader || new BufReader(conn);
        this.bufWriter = bufWriter || new BufWriter(conn);
    }
    async *[Symbol.asyncIterator]() {
        let frames = [];
        let payloadsLength = 0;
        while(!this._isClosed){
            let frame;
            try {
                frame = await readFrame(this.bufReader);
            } catch (e) {
                this.ensureSocketClosed();
                break;
            }
            unmask(frame.payload, frame.mask);
            switch(frame.opcode){
                case OpCode.TextFrame:
                case OpCode.BinaryFrame:
                case OpCode.Continue:
                    frames.push(frame);
                    payloadsLength += frame.payload.length;
                    if (frame.isLastFrame) {
                        const concat = new Uint8Array(payloadsLength);
                        let offs = 0;
                        for (const frame of frames){
                            concat.set(frame.payload, offs);
                            offs += frame.payload.length;
                        }
                        if (frames[0].opcode === OpCode.TextFrame) {
                            // text
                            yield decode(concat);
                        } else {
                            // binary
                            yield concat;
                        }
                        frames = [];
                        payloadsLength = 0;
                    }
                    break;
                case OpCode.Close:
                    {
                        // [0x12, 0x34] -> 0x1234
                        const code = frame.payload[0] << 8 | frame.payload[1];
                        const reason = decode(frame.payload.subarray(2, frame.payload.length));
                        await this.close(code, reason);
                        yield {
                            code,
                            reason
                        };
                        return;
                    }
                case OpCode.Ping:
                    await this.enqueue({
                        opcode: OpCode.Pong,
                        payload: frame.payload,
                        isLastFrame: true
                    });
                    yield [
                        "ping",
                        frame.payload
                    ];
                    break;
                case OpCode.Pong:
                    yield [
                        "pong",
                        frame.payload
                    ];
                    break;
                default:
            }
        }
    }
    dequeue() {
        const [entry] = this.sendQueue;
        if (!entry) return;
        if (this._isClosed) return;
        const { d , frame  } = entry;
        writeFrame(frame, this.bufWriter).then(()=>d.resolve()
        ).catch((e)=>d.reject(e)
        ).finally(()=>{
            this.sendQueue.shift();
            this.dequeue();
        });
    }
    enqueue(frame) {
        if (this._isClosed) {
            throw new Deno.errors.ConnectionReset("Socket has already been closed");
        }
        const d = deferred();
        this.sendQueue.push({
            d,
            frame
        });
        if (this.sendQueue.length === 1) {
            this.dequeue();
        }
        return d;
    }
    send(data) {
        const opcode = typeof data === "string" ? OpCode.TextFrame : OpCode.BinaryFrame;
        const payload = typeof data === "string" ? encode(data) : data;
        const isLastFrame = true;
        const frame = {
            isLastFrame,
            opcode,
            payload,
            mask: this.mask
        };
        return this.enqueue(frame);
    }
    ping(data = "") {
        const payload = typeof data === "string" ? encode(data) : data;
        const frame = {
            isLastFrame: true,
            opcode: OpCode.Ping,
            mask: this.mask,
            payload
        };
        return this.enqueue(frame);
    }
    _isClosed = false;
    get isClosed() {
        return this._isClosed;
    }
    async close(code = 1000, reason) {
        try {
            const header = [
                code >>> 8,
                code & 255
            ];
            let payload;
            if (reason) {
                const reasonBytes = encode(reason);
                payload = new Uint8Array(2 + reasonBytes.byteLength);
                payload.set(header);
                payload.set(reasonBytes, 2);
            } else {
                payload = new Uint8Array(header);
            }
            await this.enqueue({
                isLastFrame: true,
                opcode: OpCode.Close,
                mask: this.mask,
                payload
            });
        } catch (e) {
            throw e;
        } finally{
            this.ensureSocketClosed();
        }
    }
    closeForce() {
        this.ensureSocketClosed();
    }
    ensureSocketClosed() {
        if (this.isClosed) return;
        try {
            this.conn.close();
        } catch (e) {
            console.error(e);
        } finally{
            this._isClosed = true;
            const rest = this.sendQueue;
            this.sendQueue = [];
            rest.forEach((e)=>e.d.reject(new Deno.errors.ConnectionReset("Socket has already been closed"))
            );
        }
    }
}
/** Return whether given headers is acceptable for websocket  */ export function acceptable(req) {
    const upgrade = req.headers.get("upgrade");
    if (!upgrade || upgrade.toLowerCase() !== "websocket") {
        return false;
    }
    const secKey = req.headers.get("sec-websocket-key");
    return req.headers.has("sec-websocket-key") && typeof secKey === "string" && secKey.length > 0;
}
const kGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11";
/** Create sec-websocket-accept header value with given nonce */ export function createSecAccept(nonce) {
    const sha1 = new Sha1();
    sha1.update(nonce + kGUID);
    const bytes = sha1.digest();
    return btoa(String.fromCharCode(...bytes));
}
/** Upgrade given TCP connection into websocket connection */ export async function acceptWebSocket(req) {
    const { conn , headers , bufReader , bufWriter  } = req;
    if (acceptable(req)) {
        const sock = new WebSocketImpl({
            conn,
            bufReader,
            bufWriter
        });
        const secKey = headers.get("sec-websocket-key");
        if (typeof secKey !== "string") {
            throw new Error("sec-websocket-key is not provided");
        }
        const secAccept = createSecAccept(secKey);
        await writeResponse(bufWriter, {
            status: 101,
            headers: new Headers({
                Upgrade: "websocket",
                Connection: "Upgrade",
                "Sec-WebSocket-Accept": secAccept
            })
        });
        return sock;
    }
    throw new Error("request is not acceptable");
}
const kSecChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-.~_";
/** Create WebSocket-Sec-Key. Base64 encoded 16 bytes string */ export function createSecKey() {
    let key = "";
    for(let i = 0; i < 16; i++){
        const j = Math.floor(Math.random() * kSecChars.length);
        key += kSecChars[j];
    }
    return btoa(key);
}
export async function handshake(url, headers, bufReader, bufWriter) {
    const { hostname , pathname , search  } = url;
    const key = createSecKey();
    if (!headers.has("host")) {
        headers.set("host", hostname);
    }
    headers.set("upgrade", "websocket");
    headers.set("connection", "upgrade");
    headers.set("sec-websocket-key", key);
    headers.set("sec-websocket-version", "13");
    let headerStr = `GET ${pathname}${search} HTTP/1.1\r\n`;
    for (const [key, value] of headers){
        headerStr += `${key}: ${value}\r\n`;
    }
    headerStr += "\r\n";
    await bufWriter.write(encode(headerStr));
    await bufWriter.flush();
    const tpReader = new TextProtoReader(bufReader);
    const statusLine = await tpReader.readLine();
    if (statusLine === null) {
        throw new Deno.errors.UnexpectedEof();
    }
    const m = statusLine.match(/^(?<version>\S+) (?<statusCode>\S+) /);
    if (!m) {
        throw new Error("ws: invalid status line: " + statusLine);
    }
    assert(m.groups);
    const { version , statusCode  } = m.groups;
    if (version !== "HTTP/1.1" || statusCode !== "101") {
        throw new Error(`ws: server didn't accept handshake: ` + `version=${version}, statusCode=${statusCode}`);
    }
    const responseHeaders = await tpReader.readMIMEHeader();
    if (responseHeaders === null) {
        throw new Deno.errors.UnexpectedEof();
    }
    const expectedSecAccept = createSecAccept(key);
    const secAccept = responseHeaders.get("sec-websocket-accept");
    if (secAccept !== expectedSecAccept) {
        throw new Error(`ws: unexpected sec-websocket-accept header: ` + `expected=${expectedSecAccept}, actual=${secAccept}`);
    }
}
export function createWebSocket(params) {
    return new WebSocketImpl(params);
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL3dzL21vZC50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cbmltcG9ydCB7IGRlY29kZSwgZW5jb2RlIH0gZnJvbSBcIi4uL2VuY29kaW5nL3V0ZjgudHNcIjtcbmltcG9ydCB7IGhhc093blByb3BlcnR5IH0gZnJvbSBcIi4uL191dGlsL2hhc19vd25fcHJvcGVydHkudHNcIjtcbmltcG9ydCB7IEJ1ZlJlYWRlciwgQnVmV3JpdGVyIH0gZnJvbSBcIi4uL2lvL2J1ZmlvLnRzXCI7XG5pbXBvcnQgeyByZWFkTG9uZywgcmVhZFNob3J0LCBzbGljZUxvbmdUb0J5dGVzIH0gZnJvbSBcIi4uL2lvL2lvdXRpbC50c1wiO1xuaW1wb3J0IHsgU2hhMSB9IGZyb20gXCIuLi9oYXNoL3NoYTEudHNcIjtcbmltcG9ydCB7IHdyaXRlUmVzcG9uc2UgfSBmcm9tIFwiLi4vaHR0cC9faW8udHNcIjtcbmltcG9ydCB7IFRleHRQcm90b1JlYWRlciB9IGZyb20gXCIuLi90ZXh0cHJvdG8vbW9kLnRzXCI7XG5pbXBvcnQgeyBEZWZlcnJlZCwgZGVmZXJyZWQgfSBmcm9tIFwiLi4vYXN5bmMvZGVmZXJyZWQudHNcIjtcbmltcG9ydCB7IGFzc2VydCB9IGZyb20gXCIuLi9fdXRpbC9hc3NlcnQudHNcIjtcbmltcG9ydCB7IGNvbmNhdCB9IGZyb20gXCIuLi9ieXRlcy9tb2QudHNcIjtcblxuZXhwb3J0IGVudW0gT3BDb2RlIHtcbiAgQ29udGludWUgPSAweDAsXG4gIFRleHRGcmFtZSA9IDB4MSxcbiAgQmluYXJ5RnJhbWUgPSAweDIsXG4gIENsb3NlID0gMHg4LFxuICBQaW5nID0gMHg5LFxuICBQb25nID0gMHhhLFxufVxuXG5leHBvcnQgdHlwZSBXZWJTb2NrZXRFdmVudCA9XG4gIHwgc3RyaW5nXG4gIHwgVWludDhBcnJheVxuICB8IFdlYlNvY2tldENsb3NlRXZlbnQgLy8gUmVjZWl2ZWQgYWZ0ZXIgY2xvc2luZyBjb25uZWN0aW9uIGZpbmlzaGVkLlxuICB8IFdlYlNvY2tldFBpbmdFdmVudCAvLyBSZWNlaXZlZCBhZnRlciBwb25nIGZyYW1lIHJlc3BvbmRlZC5cbiAgfCBXZWJTb2NrZXRQb25nRXZlbnQ7XG5cbmV4cG9ydCBpbnRlcmZhY2UgV2ViU29ja2V0Q2xvc2VFdmVudCB7XG4gIGNvZGU6IG51bWJlcjtcbiAgcmVhc29uPzogc3RyaW5nO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gaXNXZWJTb2NrZXRDbG9zZUV2ZW50KFxuICBhOiBXZWJTb2NrZXRFdmVudCxcbik6IGEgaXMgV2ViU29ja2V0Q2xvc2VFdmVudCB7XG4gIHJldHVybiBoYXNPd25Qcm9wZXJ0eShhLCBcImNvZGVcIik7XG59XG5cbmV4cG9ydCB0eXBlIFdlYlNvY2tldFBpbmdFdmVudCA9IFtcInBpbmdcIiwgVWludDhBcnJheV07XG5cbmV4cG9ydCBmdW5jdGlvbiBpc1dlYlNvY2tldFBpbmdFdmVudChcbiAgYTogV2ViU29ja2V0RXZlbnQsXG4pOiBhIGlzIFdlYlNvY2tldFBpbmdFdmVudCB7XG4gIHJldHVybiBBcnJheS5pc0FycmF5KGEpICYmIGFbMF0gPT09IFwicGluZ1wiICYmIGFbMV0gaW5zdGFuY2VvZiBVaW50OEFycmF5O1xufVxuXG5leHBvcnQgdHlwZSBXZWJTb2NrZXRQb25nRXZlbnQgPSBbXCJwb25nXCIsIFVpbnQ4QXJyYXldO1xuXG5leHBvcnQgZnVuY3Rpb24gaXNXZWJTb2NrZXRQb25nRXZlbnQoXG4gIGE6IFdlYlNvY2tldEV2ZW50LFxuKTogYSBpcyBXZWJTb2NrZXRQb25nRXZlbnQge1xuICByZXR1cm4gQXJyYXkuaXNBcnJheShhKSAmJiBhWzBdID09PSBcInBvbmdcIiAmJiBhWzFdIGluc3RhbmNlb2YgVWludDhBcnJheTtcbn1cblxuZXhwb3J0IHR5cGUgV2ViU29ja2V0TWVzc2FnZSA9IHN0cmluZyB8IFVpbnQ4QXJyYXk7XG5cbmV4cG9ydCBpbnRlcmZhY2UgV2ViU29ja2V0RnJhbWUge1xuICBpc0xhc3RGcmFtZTogYm9vbGVhbjtcbiAgb3Bjb2RlOiBPcENvZGU7XG4gIG1hc2s/OiBVaW50OEFycmF5O1xuICBwYXlsb2FkOiBVaW50OEFycmF5O1xufVxuXG5leHBvcnQgaW50ZXJmYWNlIFdlYlNvY2tldCBleHRlbmRzIEFzeW5jSXRlcmFibGU8V2ViU29ja2V0RXZlbnQ+IHtcbiAgcmVhZG9ubHkgY29ubjogRGVuby5Db25uO1xuICByZWFkb25seSBpc0Nsb3NlZDogYm9vbGVhbjtcblxuICBbU3ltYm9sLmFzeW5jSXRlcmF0b3JdKCk6IEFzeW5jSXRlcmFibGVJdGVyYXRvcjxXZWJTb2NrZXRFdmVudD47XG5cbiAgLyoqXG4gICAqIEB0aHJvd3MgYERlbm8uZXJyb3JzLkNvbm5lY3Rpb25SZXNldGBcbiAgICovXG4gIHNlbmQoZGF0YTogV2ViU29ja2V0TWVzc2FnZSk6IFByb21pc2U8dm9pZD47XG5cbiAgLyoqXG4gICAqIEBwYXJhbSBkYXRhXG4gICAqIEB0aHJvd3MgYERlbm8uZXJyb3JzLkNvbm5lY3Rpb25SZXNldGBcbiAgICovXG4gIHBpbmcoZGF0YT86IFdlYlNvY2tldE1lc3NhZ2UpOiBQcm9taXNlPHZvaWQ+O1xuXG4gIC8qKiBDbG9zZSBjb25uZWN0aW9uIGFmdGVyIHNlbmRpbmcgY2xvc2UgZnJhbWUgdG8gcGVlci5cbiAgICogVGhpcyBpcyBjYW5vbmljYWwgd2F5IG9mIGRpc2Nvbm5lY3Rpb24gYnV0IGl0IG1heSBoYW5nIGJlY2F1c2Ugb2YgcGVlcidzIHJlc3BvbnNlIGRlbGF5LlxuICAgKiBEZWZhdWx0IGNsb3NlIGNvZGUgaXMgMTAwMCAoTm9ybWFsIENsb3N1cmUpXG4gICAqIEB0aHJvd3MgYERlbm8uZXJyb3JzLkNvbm5lY3Rpb25SZXNldGBcbiAgICovXG4gIGNsb3NlKCk6IFByb21pc2U8dm9pZD47XG4gIGNsb3NlKGNvZGU6IG51bWJlcik6IFByb21pc2U8dm9pZD47XG4gIGNsb3NlKGNvZGU6IG51bWJlciwgcmVhc29uOiBzdHJpbmcpOiBQcm9taXNlPHZvaWQ+O1xuXG4gIC8qKiBDbG9zZSBjb25uZWN0aW9uIGZvcmNlbHkgd2l0aG91dCBzZW5kaW5nIGNsb3NlIGZyYW1lIHRvIHBlZXIuXG4gICAqICBUaGlzIGlzIGJhc2ljYWxseSB1bmRlc2lyYWJsZSB3YXkgb2YgZGlzY29ubmVjdGlvbi4gVXNlIGNhcmVmdWxseS4gKi9cbiAgY2xvc2VGb3JjZSgpOiB2b2lkO1xufVxuXG4vKiogVW5tYXNrIG1hc2tlZCB3ZWJzb2NrZXQgcGF5bG9hZCAqL1xuZXhwb3J0IGZ1bmN0aW9uIHVubWFzayhwYXlsb2FkOiBVaW50OEFycmF5LCBtYXNrPzogVWludDhBcnJheSk6IHZvaWQge1xuICBpZiAobWFzaykge1xuICAgIGZvciAobGV0IGkgPSAwLCBsZW4gPSBwYXlsb2FkLmxlbmd0aDsgaSA8IGxlbjsgaSsrKSB7XG4gICAgICBwYXlsb2FkW2ldIF49IG1hc2tbaSAmIDNdO1xuICAgIH1cbiAgfVxufVxuXG4vKiogV3JpdGUgd2Vic29ja2V0IGZyYW1lIHRvIGdpdmVuIHdyaXRlciAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIHdyaXRlRnJhbWUoXG4gIGZyYW1lOiBXZWJTb2NrZXRGcmFtZSxcbiAgd3JpdGVyOiBEZW5vLldyaXRlcixcbik6IFByb21pc2U8dm9pZD4ge1xuICBjb25zdCBwYXlsb2FkTGVuZ3RoID0gZnJhbWUucGF5bG9hZC5ieXRlTGVuZ3RoO1xuICBsZXQgaGVhZGVyOiBVaW50OEFycmF5O1xuICBjb25zdCBoYXNNYXNrID0gZnJhbWUubWFzayA/IDB4ODAgOiAwO1xuICBpZiAoZnJhbWUubWFzayAmJiBmcmFtZS5tYXNrLmJ5dGVMZW5ndGggIT09IDQpIHtcbiAgICB0aHJvdyBuZXcgRXJyb3IoXG4gICAgICBcImludmFsaWQgbWFzay4gbWFzayBtdXN0IGJlIDQgYnl0ZXM6IGxlbmd0aD1cIiArIGZyYW1lLm1hc2suYnl0ZUxlbmd0aCxcbiAgICApO1xuICB9XG4gIGlmIChwYXlsb2FkTGVuZ3RoIDwgMTI2KSB7XG4gICAgaGVhZGVyID0gbmV3IFVpbnQ4QXJyYXkoWzB4ODAgfCBmcmFtZS5vcGNvZGUsIGhhc01hc2sgfCBwYXlsb2FkTGVuZ3RoXSk7XG4gIH0gZWxzZSBpZiAocGF5bG9hZExlbmd0aCA8IDB4ZmZmZikge1xuICAgIGhlYWRlciA9IG5ldyBVaW50OEFycmF5KFtcbiAgICAgIDB4ODAgfCBmcmFtZS5vcGNvZGUsXG4gICAgICBoYXNNYXNrIHwgMGIwMTExMTExMCxcbiAgICAgIHBheWxvYWRMZW5ndGggPj4+IDgsXG4gICAgICBwYXlsb2FkTGVuZ3RoICYgMHgwMGZmLFxuICAgIF0pO1xuICB9IGVsc2Uge1xuICAgIGhlYWRlciA9IG5ldyBVaW50OEFycmF5KFtcbiAgICAgIDB4ODAgfCBmcmFtZS5vcGNvZGUsXG4gICAgICBoYXNNYXNrIHwgMGIwMTExMTExMSxcbiAgICAgIC4uLnNsaWNlTG9uZ1RvQnl0ZXMocGF5bG9hZExlbmd0aCksXG4gICAgXSk7XG4gIH1cbiAgaWYgKGZyYW1lLm1hc2spIHtcbiAgICBoZWFkZXIgPSBjb25jYXQoaGVhZGVyLCBmcmFtZS5tYXNrKTtcbiAgfVxuICB1bm1hc2soZnJhbWUucGF5bG9hZCwgZnJhbWUubWFzayk7XG4gIGhlYWRlciA9IGNvbmNhdChoZWFkZXIsIGZyYW1lLnBheWxvYWQpO1xuICBjb25zdCB3ID0gQnVmV3JpdGVyLmNyZWF0ZSh3cml0ZXIpO1xuICBhd2FpdCB3LndyaXRlKGhlYWRlcik7XG4gIGF3YWl0IHcuZmx1c2goKTtcbn1cblxuLyoqIFJlYWQgd2Vic29ja2V0IGZyYW1lIGZyb20gZ2l2ZW4gQnVmUmVhZGVyXG4gKiBAdGhyb3dzIGBEZW5vLmVycm9ycy5VbmV4cGVjdGVkRW9mYCBXaGVuIHBlZXIgY2xvc2VkIGNvbm5lY3Rpb24gd2l0aG91dCBjbG9zZSBmcmFtZVxuICogQHRocm93cyBgRXJyb3JgIEZyYW1lIGlzIGludmFsaWRcbiAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIHJlYWRGcmFtZShidWY6IEJ1ZlJlYWRlcik6IFByb21pc2U8V2ViU29ja2V0RnJhbWU+IHtcbiAgbGV0IGIgPSBhd2FpdCBidWYucmVhZEJ5dGUoKTtcbiAgYXNzZXJ0KGIgIT09IG51bGwpO1xuICBsZXQgaXNMYXN0RnJhbWUgPSBmYWxzZTtcbiAgc3dpdGNoIChiID4+PiA0KSB7XG4gICAgY2FzZSAwYjEwMDA6XG4gICAgICBpc0xhc3RGcmFtZSA9IHRydWU7XG4gICAgICBicmVhaztcbiAgICBjYXNlIDBiMDAwMDpcbiAgICAgIGlzTGFzdEZyYW1lID0gZmFsc2U7XG4gICAgICBicmVhaztcbiAgICBkZWZhdWx0OlxuICAgICAgdGhyb3cgbmV3IEVycm9yKFwiaW52YWxpZCBzaWduYXR1cmVcIik7XG4gIH1cbiAgY29uc3Qgb3Bjb2RlID0gYiAmIDB4MGY7XG4gIC8vIGhhc19tYXNrICYgcGF5bG9hZFxuICBiID0gYXdhaXQgYnVmLnJlYWRCeXRlKCk7XG4gIGFzc2VydChiICE9PSBudWxsKTtcbiAgY29uc3QgaGFzTWFzayA9IGIgPj4+IDc7XG4gIGxldCBwYXlsb2FkTGVuZ3RoID0gYiAmIDBiMDExMTExMTE7XG4gIGlmIChwYXlsb2FkTGVuZ3RoID09PSAxMjYpIHtcbiAgICBjb25zdCBsID0gYXdhaXQgcmVhZFNob3J0KGJ1Zik7XG4gICAgYXNzZXJ0KGwgIT09IG51bGwpO1xuICAgIHBheWxvYWRMZW5ndGggPSBsO1xuICB9IGVsc2UgaWYgKHBheWxvYWRMZW5ndGggPT09IDEyNykge1xuICAgIGNvbnN0IGwgPSBhd2FpdCByZWFkTG9uZyhidWYpO1xuICAgIGFzc2VydChsICE9PSBudWxsKTtcbiAgICBwYXlsb2FkTGVuZ3RoID0gTnVtYmVyKGwpO1xuICB9XG4gIC8vIG1hc2tcbiAgbGV0IG1hc2s6IFVpbnQ4QXJyYXkgfCB1bmRlZmluZWQ7XG4gIGlmIChoYXNNYXNrKSB7XG4gICAgbWFzayA9IG5ldyBVaW50OEFycmF5KDQpO1xuICAgIGFzc2VydCgoYXdhaXQgYnVmLnJlYWRGdWxsKG1hc2spKSAhPT0gbnVsbCk7XG4gIH1cbiAgLy8gcGF5bG9hZFxuICBjb25zdCBwYXlsb2FkID0gbmV3IFVpbnQ4QXJyYXkocGF5bG9hZExlbmd0aCk7XG4gIGFzc2VydCgoYXdhaXQgYnVmLnJlYWRGdWxsKHBheWxvYWQpKSAhPT0gbnVsbCk7XG4gIHJldHVybiB7XG4gICAgaXNMYXN0RnJhbWUsXG4gICAgb3Bjb2RlLFxuICAgIG1hc2ssXG4gICAgcGF5bG9hZCxcbiAgfTtcbn1cblxuY2xhc3MgV2ViU29ja2V0SW1wbCBpbXBsZW1lbnRzIFdlYlNvY2tldCB7XG4gIHJlYWRvbmx5IGNvbm46IERlbm8uQ29ubjtcbiAgcHJpdmF0ZSByZWFkb25seSBtYXNrPzogVWludDhBcnJheTtcbiAgcHJpdmF0ZSByZWFkb25seSBidWZSZWFkZXI6IEJ1ZlJlYWRlcjtcbiAgcHJpdmF0ZSByZWFkb25seSBidWZXcml0ZXI6IEJ1ZldyaXRlcjtcbiAgcHJpdmF0ZSBzZW5kUXVldWU6IEFycmF5PHtcbiAgICBmcmFtZTogV2ViU29ja2V0RnJhbWU7XG4gICAgZDogRGVmZXJyZWQ8dm9pZD47XG4gIH0+ID0gW107XG5cbiAgY29uc3RydWN0b3Ioe1xuICAgIGNvbm4sXG4gICAgYnVmUmVhZGVyLFxuICAgIGJ1ZldyaXRlcixcbiAgICBtYXNrLFxuICB9OiB7XG4gICAgY29ubjogRGVuby5Db25uO1xuICAgIGJ1ZlJlYWRlcj86IEJ1ZlJlYWRlcjtcbiAgICBidWZXcml0ZXI/OiBCdWZXcml0ZXI7XG4gICAgbWFzaz86IFVpbnQ4QXJyYXk7XG4gIH0pIHtcbiAgICB0aGlzLmNvbm4gPSBjb25uO1xuICAgIHRoaXMubWFzayA9IG1hc2s7XG4gICAgdGhpcy5idWZSZWFkZXIgPSBidWZSZWFkZXIgfHwgbmV3IEJ1ZlJlYWRlcihjb25uKTtcbiAgICB0aGlzLmJ1ZldyaXRlciA9IGJ1ZldyaXRlciB8fCBuZXcgQnVmV3JpdGVyKGNvbm4pO1xuICB9XG5cbiAgYXN5bmMgKltTeW1ib2wuYXN5bmNJdGVyYXRvcl0oKTogQXN5bmNJdGVyYWJsZUl0ZXJhdG9yPFdlYlNvY2tldEV2ZW50PiB7XG4gICAgbGV0IGZyYW1lczogV2ViU29ja2V0RnJhbWVbXSA9IFtdO1xuICAgIGxldCBwYXlsb2Fkc0xlbmd0aCA9IDA7XG4gICAgd2hpbGUgKCF0aGlzLl9pc0Nsb3NlZCkge1xuICAgICAgbGV0IGZyYW1lOiBXZWJTb2NrZXRGcmFtZTtcbiAgICAgIHRyeSB7XG4gICAgICAgIGZyYW1lID0gYXdhaXQgcmVhZEZyYW1lKHRoaXMuYnVmUmVhZGVyKTtcbiAgICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgICAgdGhpcy5lbnN1cmVTb2NrZXRDbG9zZWQoKTtcbiAgICAgICAgYnJlYWs7XG4gICAgICB9XG4gICAgICB1bm1hc2soZnJhbWUucGF5bG9hZCwgZnJhbWUubWFzayk7XG4gICAgICBzd2l0Y2ggKGZyYW1lLm9wY29kZSkge1xuICAgICAgICBjYXNlIE9wQ29kZS5UZXh0RnJhbWU6XG4gICAgICAgIGNhc2UgT3BDb2RlLkJpbmFyeUZyYW1lOlxuICAgICAgICBjYXNlIE9wQ29kZS5Db250aW51ZTpcbiAgICAgICAgICBmcmFtZXMucHVzaChmcmFtZSk7XG4gICAgICAgICAgcGF5bG9hZHNMZW5ndGggKz0gZnJhbWUucGF5bG9hZC5sZW5ndGg7XG4gICAgICAgICAgaWYgKGZyYW1lLmlzTGFzdEZyYW1lKSB7XG4gICAgICAgICAgICBjb25zdCBjb25jYXQgPSBuZXcgVWludDhBcnJheShwYXlsb2Fkc0xlbmd0aCk7XG4gICAgICAgICAgICBsZXQgb2ZmcyA9IDA7XG4gICAgICAgICAgICBmb3IgKGNvbnN0IGZyYW1lIG9mIGZyYW1lcykge1xuICAgICAgICAgICAgICBjb25jYXQuc2V0KGZyYW1lLnBheWxvYWQsIG9mZnMpO1xuICAgICAgICAgICAgICBvZmZzICs9IGZyYW1lLnBheWxvYWQubGVuZ3RoO1xuICAgICAgICAgICAgfVxuICAgICAgICAgICAgaWYgKGZyYW1lc1swXS5vcGNvZGUgPT09IE9wQ29kZS5UZXh0RnJhbWUpIHtcbiAgICAgICAgICAgICAgLy8gdGV4dFxuICAgICAgICAgICAgICB5aWVsZCBkZWNvZGUoY29uY2F0KTtcbiAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgIC8vIGJpbmFyeVxuICAgICAgICAgICAgICB5aWVsZCBjb25jYXQ7XG4gICAgICAgICAgICB9XG4gICAgICAgICAgICBmcmFtZXMgPSBbXTtcbiAgICAgICAgICAgIHBheWxvYWRzTGVuZ3RoID0gMDtcbiAgICAgICAgICB9XG4gICAgICAgICAgYnJlYWs7XG4gICAgICAgIGNhc2UgT3BDb2RlLkNsb3NlOiB7XG4gICAgICAgICAgLy8gWzB4MTIsIDB4MzRdIC0+IDB4MTIzNFxuICAgICAgICAgIGNvbnN0IGNvZGUgPSAoZnJhbWUucGF5bG9hZFswXSA8PCA4KSB8IGZyYW1lLnBheWxvYWRbMV07XG4gICAgICAgICAgY29uc3QgcmVhc29uID0gZGVjb2RlKFxuICAgICAgICAgICAgZnJhbWUucGF5bG9hZC5zdWJhcnJheSgyLCBmcmFtZS5wYXlsb2FkLmxlbmd0aCksXG4gICAgICAgICAgKTtcbiAgICAgICAgICBhd2FpdCB0aGlzLmNsb3NlKGNvZGUsIHJlYXNvbik7XG4gICAgICAgICAgeWllbGQgeyBjb2RlLCByZWFzb24gfTtcbiAgICAgICAgICByZXR1cm47XG4gICAgICAgIH1cbiAgICAgICAgY2FzZSBPcENvZGUuUGluZzpcbiAgICAgICAgICBhd2FpdCB0aGlzLmVucXVldWUoe1xuICAgICAgICAgICAgb3Bjb2RlOiBPcENvZGUuUG9uZyxcbiAgICAgICAgICAgIHBheWxvYWQ6IGZyYW1lLnBheWxvYWQsXG4gICAgICAgICAgICBpc0xhc3RGcmFtZTogdHJ1ZSxcbiAgICAgICAgICB9KTtcbiAgICAgICAgICB5aWVsZCBbXCJwaW5nXCIsIGZyYW1lLnBheWxvYWRdIGFzIFdlYlNvY2tldFBpbmdFdmVudDtcbiAgICAgICAgICBicmVhaztcbiAgICAgICAgY2FzZSBPcENvZGUuUG9uZzpcbiAgICAgICAgICB5aWVsZCBbXCJwb25nXCIsIGZyYW1lLnBheWxvYWRdIGFzIFdlYlNvY2tldFBvbmdFdmVudDtcbiAgICAgICAgICBicmVhaztcbiAgICAgICAgZGVmYXVsdDpcbiAgICAgIH1cbiAgICB9XG4gIH1cblxuICBwcml2YXRlIGRlcXVldWUoKTogdm9pZCB7XG4gICAgY29uc3QgW2VudHJ5XSA9IHRoaXMuc2VuZFF1ZXVlO1xuICAgIGlmICghZW50cnkpIHJldHVybjtcbiAgICBpZiAodGhpcy5faXNDbG9zZWQpIHJldHVybjtcbiAgICBjb25zdCB7IGQsIGZyYW1lIH0gPSBlbnRyeTtcbiAgICB3cml0ZUZyYW1lKGZyYW1lLCB0aGlzLmJ1ZldyaXRlcilcbiAgICAgIC50aGVuKCgpID0+IGQucmVzb2x2ZSgpKVxuICAgICAgLmNhdGNoKChlKSA9PiBkLnJlamVjdChlKSlcbiAgICAgIC5maW5hbGx5KCgpID0+IHtcbiAgICAgICAgdGhpcy5zZW5kUXVldWUuc2hpZnQoKTtcbiAgICAgICAgdGhpcy5kZXF1ZXVlKCk7XG4gICAgICB9KTtcbiAgfVxuXG4gIHByaXZhdGUgZW5xdWV1ZShmcmFtZTogV2ViU29ja2V0RnJhbWUpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBpZiAodGhpcy5faXNDbG9zZWQpIHtcbiAgICAgIHRocm93IG5ldyBEZW5vLmVycm9ycy5Db25uZWN0aW9uUmVzZXQoXCJTb2NrZXQgaGFzIGFscmVhZHkgYmVlbiBjbG9zZWRcIik7XG4gICAgfVxuICAgIGNvbnN0IGQgPSBkZWZlcnJlZDx2b2lkPigpO1xuICAgIHRoaXMuc2VuZFF1ZXVlLnB1c2goeyBkLCBmcmFtZSB9KTtcbiAgICBpZiAodGhpcy5zZW5kUXVldWUubGVuZ3RoID09PSAxKSB7XG4gICAgICB0aGlzLmRlcXVldWUoKTtcbiAgICB9XG4gICAgcmV0dXJuIGQ7XG4gIH1cblxuICBzZW5kKGRhdGE6IFdlYlNvY2tldE1lc3NhZ2UpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBjb25zdCBvcGNvZGUgPSB0eXBlb2YgZGF0YSA9PT0gXCJzdHJpbmdcIlxuICAgICAgPyBPcENvZGUuVGV4dEZyYW1lXG4gICAgICA6IE9wQ29kZS5CaW5hcnlGcmFtZTtcbiAgICBjb25zdCBwYXlsb2FkID0gdHlwZW9mIGRhdGEgPT09IFwic3RyaW5nXCIgPyBlbmNvZGUoZGF0YSkgOiBkYXRhO1xuICAgIGNvbnN0IGlzTGFzdEZyYW1lID0gdHJ1ZTtcbiAgICBjb25zdCBmcmFtZSA9IHtcbiAgICAgIGlzTGFzdEZyYW1lLFxuICAgICAgb3Bjb2RlLFxuICAgICAgcGF5bG9hZCxcbiAgICAgIG1hc2s6IHRoaXMubWFzayxcbiAgICB9O1xuICAgIHJldHVybiB0aGlzLmVucXVldWUoZnJhbWUpO1xuICB9XG5cbiAgcGluZyhkYXRhOiBXZWJTb2NrZXRNZXNzYWdlID0gXCJcIik6IFByb21pc2U8dm9pZD4ge1xuICAgIGNvbnN0IHBheWxvYWQgPSB0eXBlb2YgZGF0YSA9PT0gXCJzdHJpbmdcIiA/IGVuY29kZShkYXRhKSA6IGRhdGE7XG4gICAgY29uc3QgZnJhbWUgPSB7XG4gICAgICBpc0xhc3RGcmFtZTogdHJ1ZSxcbiAgICAgIG9wY29kZTogT3BDb2RlLlBpbmcsXG4gICAgICBtYXNrOiB0aGlzLm1hc2ssXG4gICAgICBwYXlsb2FkLFxuICAgIH07XG4gICAgcmV0dXJuIHRoaXMuZW5xdWV1ZShmcmFtZSk7XG4gIH1cblxuICBwcml2YXRlIF9pc0Nsb3NlZCA9IGZhbHNlO1xuICBnZXQgaXNDbG9zZWQoKTogYm9vbGVhbiB7XG4gICAgcmV0dXJuIHRoaXMuX2lzQ2xvc2VkO1xuICB9XG5cbiAgYXN5bmMgY2xvc2UoY29kZSA9IDEwMDAsIHJlYXNvbj86IHN0cmluZyk6IFByb21pc2U8dm9pZD4ge1xuICAgIHRyeSB7XG4gICAgICBjb25zdCBoZWFkZXIgPSBbY29kZSA+Pj4gOCwgY29kZSAmIDB4MDBmZl07XG4gICAgICBsZXQgcGF5bG9hZDogVWludDhBcnJheTtcbiAgICAgIGlmIChyZWFzb24pIHtcbiAgICAgICAgY29uc3QgcmVhc29uQnl0ZXMgPSBlbmNvZGUocmVhc29uKTtcbiAgICAgICAgcGF5bG9hZCA9IG5ldyBVaW50OEFycmF5KDIgKyByZWFzb25CeXRlcy5ieXRlTGVuZ3RoKTtcbiAgICAgICAgcGF5bG9hZC5zZXQoaGVhZGVyKTtcbiAgICAgICAgcGF5bG9hZC5zZXQocmVhc29uQnl0ZXMsIDIpO1xuICAgICAgfSBlbHNlIHtcbiAgICAgICAgcGF5bG9hZCA9IG5ldyBVaW50OEFycmF5KGhlYWRlcik7XG4gICAgICB9XG4gICAgICBhd2FpdCB0aGlzLmVucXVldWUoe1xuICAgICAgICBpc0xhc3RGcmFtZTogdHJ1ZSxcbiAgICAgICAgb3Bjb2RlOiBPcENvZGUuQ2xvc2UsXG4gICAgICAgIG1hc2s6IHRoaXMubWFzayxcbiAgICAgICAgcGF5bG9hZCxcbiAgICAgIH0pO1xuICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgIHRocm93IGU7XG4gICAgfSBmaW5hbGx5IHtcbiAgICAgIHRoaXMuZW5zdXJlU29ja2V0Q2xvc2VkKCk7XG4gICAgfVxuICB9XG5cbiAgY2xvc2VGb3JjZSgpOiB2b2lkIHtcbiAgICB0aGlzLmVuc3VyZVNvY2tldENsb3NlZCgpO1xuICB9XG5cbiAgcHJpdmF0ZSBlbnN1cmVTb2NrZXRDbG9zZWQoKTogdm9pZCB7XG4gICAgaWYgKHRoaXMuaXNDbG9zZWQpIHJldHVybjtcbiAgICB0cnkge1xuICAgICAgdGhpcy5jb25uLmNsb3NlKCk7XG4gICAgfSBjYXRjaCAoZSkge1xuICAgICAgY29uc29sZS5lcnJvcihlKTtcbiAgICB9IGZpbmFsbHkge1xuICAgICAgdGhpcy5faXNDbG9zZWQgPSB0cnVlO1xuICAgICAgY29uc3QgcmVzdCA9IHRoaXMuc2VuZFF1ZXVlO1xuICAgICAgdGhpcy5zZW5kUXVldWUgPSBbXTtcbiAgICAgIHJlc3QuZm9yRWFjaCgoZSkgPT5cbiAgICAgICAgZS5kLnJlamVjdChcbiAgICAgICAgICBuZXcgRGVuby5lcnJvcnMuQ29ubmVjdGlvblJlc2V0KFwiU29ja2V0IGhhcyBhbHJlYWR5IGJlZW4gY2xvc2VkXCIpLFxuICAgICAgICApXG4gICAgICApO1xuICAgIH1cbiAgfVxufVxuXG4vKiogUmV0dXJuIHdoZXRoZXIgZ2l2ZW4gaGVhZGVycyBpcyBhY2NlcHRhYmxlIGZvciB3ZWJzb2NrZXQgICovXG5leHBvcnQgZnVuY3Rpb24gYWNjZXB0YWJsZShyZXE6IHsgaGVhZGVyczogSGVhZGVycyB9KTogYm9vbGVhbiB7XG4gIGNvbnN0IHVwZ3JhZGUgPSByZXEuaGVhZGVycy5nZXQoXCJ1cGdyYWRlXCIpO1xuICBpZiAoIXVwZ3JhZGUgfHwgdXBncmFkZS50b0xvd2VyQ2FzZSgpICE9PSBcIndlYnNvY2tldFwiKSB7XG4gICAgcmV0dXJuIGZhbHNlO1xuICB9XG4gIGNvbnN0IHNlY0tleSA9IHJlcS5oZWFkZXJzLmdldChcInNlYy13ZWJzb2NrZXQta2V5XCIpO1xuICByZXR1cm4gKFxuICAgIHJlcS5oZWFkZXJzLmhhcyhcInNlYy13ZWJzb2NrZXQta2V5XCIpICYmXG4gICAgdHlwZW9mIHNlY0tleSA9PT0gXCJzdHJpbmdcIiAmJlxuICAgIHNlY0tleS5sZW5ndGggPiAwXG4gICk7XG59XG5cbmNvbnN0IGtHVUlEID0gXCIyNThFQUZBNS1FOTE0LTQ3REEtOTVDQS1DNUFCMERDODVCMTFcIjtcblxuLyoqIENyZWF0ZSBzZWMtd2Vic29ja2V0LWFjY2VwdCBoZWFkZXIgdmFsdWUgd2l0aCBnaXZlbiBub25jZSAqL1xuZXhwb3J0IGZ1bmN0aW9uIGNyZWF0ZVNlY0FjY2VwdChub25jZTogc3RyaW5nKTogc3RyaW5nIHtcbiAgY29uc3Qgc2hhMSA9IG5ldyBTaGExKCk7XG4gIHNoYTEudXBkYXRlKG5vbmNlICsga0dVSUQpO1xuICBjb25zdCBieXRlcyA9IHNoYTEuZGlnZXN0KCk7XG4gIHJldHVybiBidG9hKFN0cmluZy5mcm9tQ2hhckNvZGUoLi4uYnl0ZXMpKTtcbn1cblxuLyoqIFVwZ3JhZGUgZ2l2ZW4gVENQIGNvbm5lY3Rpb24gaW50byB3ZWJzb2NrZXQgY29ubmVjdGlvbiAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIGFjY2VwdFdlYlNvY2tldChyZXE6IHtcbiAgY29ubjogRGVuby5Db25uO1xuICBidWZXcml0ZXI6IEJ1ZldyaXRlcjtcbiAgYnVmUmVhZGVyOiBCdWZSZWFkZXI7XG4gIGhlYWRlcnM6IEhlYWRlcnM7XG59KTogUHJvbWlzZTxXZWJTb2NrZXQ+IHtcbiAgY29uc3QgeyBjb25uLCBoZWFkZXJzLCBidWZSZWFkZXIsIGJ1ZldyaXRlciB9ID0gcmVxO1xuICBpZiAoYWNjZXB0YWJsZShyZXEpKSB7XG4gICAgY29uc3Qgc29jayA9IG5ldyBXZWJTb2NrZXRJbXBsKHsgY29ubiwgYnVmUmVhZGVyLCBidWZXcml0ZXIgfSk7XG4gICAgY29uc3Qgc2VjS2V5ID0gaGVhZGVycy5nZXQoXCJzZWMtd2Vic29ja2V0LWtleVwiKTtcbiAgICBpZiAodHlwZW9mIHNlY0tleSAhPT0gXCJzdHJpbmdcIikge1xuICAgICAgdGhyb3cgbmV3IEVycm9yKFwic2VjLXdlYnNvY2tldC1rZXkgaXMgbm90IHByb3ZpZGVkXCIpO1xuICAgIH1cbiAgICBjb25zdCBzZWNBY2NlcHQgPSBjcmVhdGVTZWNBY2NlcHQoc2VjS2V5KTtcbiAgICBhd2FpdCB3cml0ZVJlc3BvbnNlKGJ1ZldyaXRlciwge1xuICAgICAgc3RhdHVzOiAxMDEsXG4gICAgICBoZWFkZXJzOiBuZXcgSGVhZGVycyh7XG4gICAgICAgIFVwZ3JhZGU6IFwid2Vic29ja2V0XCIsXG4gICAgICAgIENvbm5lY3Rpb246IFwiVXBncmFkZVwiLFxuICAgICAgICBcIlNlYy1XZWJTb2NrZXQtQWNjZXB0XCI6IHNlY0FjY2VwdCxcbiAgICAgIH0pLFxuICAgIH0pO1xuICAgIHJldHVybiBzb2NrO1xuICB9XG4gIHRocm93IG5ldyBFcnJvcihcInJlcXVlc3QgaXMgbm90IGFjY2VwdGFibGVcIik7XG59XG5cbmNvbnN0IGtTZWNDaGFycyA9IFwiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXpBQkNERUZHSElKS0xNTk9QUVJTVFVWV1hZWi0ufl9cIjtcblxuLyoqIENyZWF0ZSBXZWJTb2NrZXQtU2VjLUtleS4gQmFzZTY0IGVuY29kZWQgMTYgYnl0ZXMgc3RyaW5nICovXG5leHBvcnQgZnVuY3Rpb24gY3JlYXRlU2VjS2V5KCk6IHN0cmluZyB7XG4gIGxldCBrZXkgPSBcIlwiO1xuICBmb3IgKGxldCBpID0gMDsgaSA8IDE2OyBpKyspIHtcbiAgICBjb25zdCBqID0gTWF0aC5mbG9vcihNYXRoLnJhbmRvbSgpICoga1NlY0NoYXJzLmxlbmd0aCk7XG4gICAga2V5ICs9IGtTZWNDaGFyc1tqXTtcbiAgfVxuICByZXR1cm4gYnRvYShrZXkpO1xufVxuXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gaGFuZHNoYWtlKFxuICB1cmw6IFVSTCxcbiAgaGVhZGVyczogSGVhZGVycyxcbiAgYnVmUmVhZGVyOiBCdWZSZWFkZXIsXG4gIGJ1ZldyaXRlcjogQnVmV3JpdGVyLFxuKTogUHJvbWlzZTx2b2lkPiB7XG4gIGNvbnN0IHsgaG9zdG5hbWUsIHBhdGhuYW1lLCBzZWFyY2ggfSA9IHVybDtcbiAgY29uc3Qga2V5ID0gY3JlYXRlU2VjS2V5KCk7XG5cbiAgaWYgKCFoZWFkZXJzLmhhcyhcImhvc3RcIikpIHtcbiAgICBoZWFkZXJzLnNldChcImhvc3RcIiwgaG9zdG5hbWUpO1xuICB9XG4gIGhlYWRlcnMuc2V0KFwidXBncmFkZVwiLCBcIndlYnNvY2tldFwiKTtcbiAgaGVhZGVycy5zZXQoXCJjb25uZWN0aW9uXCIsIFwidXBncmFkZVwiKTtcbiAgaGVhZGVycy5zZXQoXCJzZWMtd2Vic29ja2V0LWtleVwiLCBrZXkpO1xuICBoZWFkZXJzLnNldChcInNlYy13ZWJzb2NrZXQtdmVyc2lvblwiLCBcIjEzXCIpO1xuXG4gIGxldCBoZWFkZXJTdHIgPSBgR0VUICR7cGF0aG5hbWV9JHtzZWFyY2h9IEhUVFAvMS4xXFxyXFxuYDtcbiAgZm9yIChjb25zdCBba2V5LCB2YWx1ZV0gb2YgaGVhZGVycykge1xuICAgIGhlYWRlclN0ciArPSBgJHtrZXl9OiAke3ZhbHVlfVxcclxcbmA7XG4gIH1cbiAgaGVhZGVyU3RyICs9IFwiXFxyXFxuXCI7XG5cbiAgYXdhaXQgYnVmV3JpdGVyLndyaXRlKGVuY29kZShoZWFkZXJTdHIpKTtcbiAgYXdhaXQgYnVmV3JpdGVyLmZsdXNoKCk7XG5cbiAgY29uc3QgdHBSZWFkZXIgPSBuZXcgVGV4dFByb3RvUmVhZGVyKGJ1ZlJlYWRlcik7XG4gIGNvbnN0IHN0YXR1c0xpbmUgPSBhd2FpdCB0cFJlYWRlci5yZWFkTGluZSgpO1xuICBpZiAoc3RhdHVzTGluZSA9PT0gbnVsbCkge1xuICAgIHRocm93IG5ldyBEZW5vLmVycm9ycy5VbmV4cGVjdGVkRW9mKCk7XG4gIH1cbiAgY29uc3QgbSA9IHN0YXR1c0xpbmUubWF0Y2goL14oPzx2ZXJzaW9uPlxcUyspICg/PHN0YXR1c0NvZGU+XFxTKykgLyk7XG4gIGlmICghbSkge1xuICAgIHRocm93IG5ldyBFcnJvcihcIndzOiBpbnZhbGlkIHN0YXR1cyBsaW5lOiBcIiArIHN0YXR1c0xpbmUpO1xuICB9XG5cbiAgYXNzZXJ0KG0uZ3JvdXBzKTtcbiAgY29uc3QgeyB2ZXJzaW9uLCBzdGF0dXNDb2RlIH0gPSBtLmdyb3VwcztcbiAgaWYgKHZlcnNpb24gIT09IFwiSFRUUC8xLjFcIiB8fCBzdGF0dXNDb2RlICE9PSBcIjEwMVwiKSB7XG4gICAgdGhyb3cgbmV3IEVycm9yKFxuICAgICAgYHdzOiBzZXJ2ZXIgZGlkbid0IGFjY2VwdCBoYW5kc2hha2U6IGAgK1xuICAgICAgICBgdmVyc2lvbj0ke3ZlcnNpb259LCBzdGF0dXNDb2RlPSR7c3RhdHVzQ29kZX1gLFxuICAgICk7XG4gIH1cblxuICBjb25zdCByZXNwb25zZUhlYWRlcnMgPSBhd2FpdCB0cFJlYWRlci5yZWFkTUlNRUhlYWRlcigpO1xuICBpZiAocmVzcG9uc2VIZWFkZXJzID09PSBudWxsKSB7XG4gICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLlVuZXhwZWN0ZWRFb2YoKTtcbiAgfVxuXG4gIGNvbnN0IGV4cGVjdGVkU2VjQWNjZXB0ID0gY3JlYXRlU2VjQWNjZXB0KGtleSk7XG4gIGNvbnN0IHNlY0FjY2VwdCA9IHJlc3BvbnNlSGVhZGVycy5nZXQoXCJzZWMtd2Vic29ja2V0LWFjY2VwdFwiKTtcbiAgaWYgKHNlY0FjY2VwdCAhPT0gZXhwZWN0ZWRTZWNBY2NlcHQpIHtcbiAgICB0aHJvdyBuZXcgRXJyb3IoXG4gICAgICBgd3M6IHVuZXhwZWN0ZWQgc2VjLXdlYnNvY2tldC1hY2NlcHQgaGVhZGVyOiBgICtcbiAgICAgICAgYGV4cGVjdGVkPSR7ZXhwZWN0ZWRTZWNBY2NlcHR9LCBhY3R1YWw9JHtzZWNBY2NlcHR9YCxcbiAgICApO1xuICB9XG59XG5cbmV4cG9ydCBmdW5jdGlvbiBjcmVhdGVXZWJTb2NrZXQocGFyYW1zOiB7XG4gIGNvbm46IERlbm8uQ29ubjtcbiAgYnVmV3JpdGVyPzogQnVmV3JpdGVyO1xuICBidWZSZWFkZXI/OiBCdWZSZWFkZXI7XG4gIG1hc2s/OiBVaW50OEFycmF5O1xufSk6IFdlYlNvY2tldCB7XG4gIHJldHVybiBuZXcgV2ViU29ja2V0SW1wbChwYXJhbXMpO1xufVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLEVBQUEsd0VBQUE7U0FDQSxNQUFBLEVBQUEsTUFBQSxTQUFBLG1CQUFBO1NBQ0EsY0FBQSxTQUFBLDRCQUFBO1NBQ0EsU0FBQSxFQUFBLFNBQUEsU0FBQSxjQUFBO1NBQ0EsUUFBQSxFQUFBLFNBQUEsRUFBQSxnQkFBQSxTQUFBLGVBQUE7U0FDQSxJQUFBLFNBQUEsZUFBQTtTQUNBLGFBQUEsU0FBQSxjQUFBO1NBQ0EsZUFBQSxTQUFBLG1CQUFBO1NBQ0EsUUFBQSxTQUFBLG9CQUFBO1NBQ0EsTUFBQSxTQUFBLGtCQUFBO1NBQ0EsTUFBQSxTQUFBLGVBQUE7V0FFQSxNQUFBO1VBQUEsTUFBQTtBQUFBLFVBQUEsQ0FBQSxNQUFBLEVBQ0EsUUFBQSxLQUFBLENBQUEsS0FBQSxRQUFBO0FBREEsVUFBQSxDQUFBLE1BQUEsRUFFQSxTQUFBLEtBQUEs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OztVQUVBLE1BQUEsR0FBQSxDQUFBLEdBQUEsRUFBQTtBQUNBLE1BQUEsbUJBQUE7QUFDQSxLQUFBLFNBQUEsR0FBQSxDQUFBLFFBQUE7QUFDQSxVQUFBLENBQUEsQ0FBQSxLQUFBLElBQUE7VUFDQSxPQUFBLEdBQUEsQ0FBQSxLQUFBLENBQUE7UUFDQSxhQUFBLEdBQUEsQ0FBQSxHQUFBLEdBQUE7UUFDQSxhQUFBLEtBQUEsR0FBQTtjQUNBLENBQUEsU0FBQSxTQUFBLENBQUEsR0FBQTtBQUNBLGNBQUEsQ0FBQSxDQUFBLEtBQUEsSUFBQTtBQUNBLHFCQUFBLEdBQUEsQ0FBQTtlQUNBLGFBQUEsS0FBQSxHQUFBO2NBQ0EsQ0FBQSxTQUFBLFFBQUEsQ0FBQSxHQUFBO0FBQ0EsY0FBQSxDQUFBLENBQUEsS0FBQSxJQUFBO0FBQ0EscUJBQUEsR0FBQSxNQUFBLENBQUEsQ0FBQTs7QUFFQSxNQUFBLEtBQUE7UUFDQSxJQUFBO1FBQ0EsT0FBQTtBQUNBLFlBQUEsT0FBQSxVQUFBLENBQUEsQ0FBQTtBQUNBLGNBQUEsT0FBQSxHQUFBLENBQUEsUUFBQSxDQUFBLElBQUEsTUFBQSxJQUFBOztBQUVBLE1BQUEsUUFBQTtVQUNBLE9BQUEsT0FBQSxVQUFBLENBQUEsYUFBQTtBQUNBLFVBQUEsT0FBQSxHQUFBLENBQUEsUUFBQSxDQUFBLE9BQUEsTUFBQSxJQUFBOztBQUVBLG1CQUFBO0FBQ0EsY0FBQTtBQUNBLFlBQUE7QUFDQSxlQUFBOzs7TUFJQSxhQUFBO0FBS0EsYUFBQTtrQkFNQSxJQUFBLEdBQ0EsU0FBQSxHQUNBLFNBQUEsR0FDQSxJQUFBO2FBT0EsSUFBQSxHQUFBLElBQUE7YUFDQSxJQUFBLEdBQUEsSUFBQTthQUNBLFNBQUEsR0FBQSxTQUFBLFFBQUEsU0FBQSxDQUFBLElBQUE7YUFDQSxTQUFBLEdBQUEsU0FBQSxRQUFBLFNBQUEsQ0FBQSxJQUFBOztZQUdBLE1BQUEsQ0FBQSxhQUFBO1lBQ0EsTUFBQTtZQUNBLGNBQUEsR0FBQSxDQUFBO29CQUNBLFNBQUE7Z0JBQ0EsS0FBQTs7QUFFQSxxQkFB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QUVBLEtBQUEsRUFBQSx5QkFBQTs7TUFHQSxTQUFBLElBQUEsd0RBQUE7QUFFQSxFQUFBLDJEQUFBLEVBQUEsaUJBQ0EsWUFBQTtRQUNBLEdBQUE7WUFDQSxDQUFBLEdBQUEsQ0FBQSxFQUFBLENBQUEsR0FBQSxFQUFBLEVBQUEsQ0FBQTtjQUNBLENBQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLElBQUEsQ0FBQSxNQUFBLEtBQUEsU0FBQSxDQUFBLE1BQUE7QUFDQSxXQUFBLElBQUEsU0FBQSxDQUFBLENBQUE7O1dBRUEsSUFBQSxDQUFBLEdBQUE7O3NCQUdBLFNBQUEsQ0FDQSxHQUFBLEVBQ0EsT0FBQSxFQUNBLFNBQUEsRUFDQSxTQUFBO1lBRUEsUUFBQSxHQUFBLFFBQUEsR0FBQSxNQUFBLE1BQUEsR0FBQTtVQUNBLEdBQUEsR0FBQSxZQUFBO1NBRUEsT0FBQSxDQUFBLEdBQUEsRUFBQSxJQUFBO0FBQ0EsZUFBQSxDQUFBLEdBQUEsRUFBQSxJQUFBLEdBQUEsUUFBQTs7QUFFQSxXQUFBLENBQUEsR0FBQSxFQUFBLE9BQUEsSUFBQSxTQUFBO0FBQ0EsV0FBQSxDQUFBLEdBQUEsRUFBQSxVQUFBLElBQUEsT0FBQTtBQUNBLFdBQUEsQ0FBQSxHQUFBLEVBQUEsaUJBQUEsR0FBQSxHQUFBO0FBQ0EsV0FBQSxDQUFBLEdBQUEsRUFBQSxxQkFBQSxJQUFBLEVBQUE7UUFFQSxTQUFBLElBQUEsSUFBQSxFQUFBLFFBQUEsR0FBQSxNQUFBLENBQUEsYUFBQTtnQkFDQSxHQUFBLEVBQUEsS0FBQSxLQUFBLE9BQUE7QUFDQSxpQkFBQSxPQUFBLEdBQUEsQ0FBQSxFQUFBLEVBQUEsS0FBQSxDQUFBLElBQUE7O0FBRUEsYUFBQSxLQUFBLElBQUE7VUFFQSxTQUFBLENBQUEsS0FBQSxDQUFBLE1BQUEsQ0FBQSxTQUFBO1VBQ0EsU0FBQSxDQUFBLEtBQUE7VUFFQSxRQUFBLE9BQUEsZUFBQSxDQUFBLFNBQUE7VUFDQSxVQUFBLFNBQUEsUUFBQSxDQUFBLFFBQUE7UUFDQSxVQUFBLEtBQUEsSUFBQTtrQkFDQSxJQUFBLENBQUEsTUFBQSxDQUFBLGFBQUE7O1VBRUEsQ0FBQSxHQUFBLFVBQUEsQ0FBQSxLQUFBO1NBQ0EsQ0FBQTtrQkFDQSxLQUFBLEVBQUEseUJBQUEsSUFBQSxVQUFBOztBQUdBLFVBQUEsQ0FBQSxDQUFBLENBQUEsTUFBQTtZQUNBLE9BQUEsR0FBQSxVQUFBLE1BQUEsQ0FBQSxDQUFBLE1BQUE7UUFDQSxPQUFBLE1BQUEsUUFBQSxLQUFBLFVBQUEsTUFBQSxHQUFBO2tCQUNBLEtBQUEsRUFDQSxvQ0FBQSxLQUNBLFFBQUEsRUFBQSxPQUFBLENBQUEsYUFBQSxFQUFBLFVBQUE7O1VBSUEsZUFBQSxTQUFBLFFBQUEsQ0FBQSxjQUFBO1FBQ0EsZUFBQSxLQUFBLElBQUE7a0JBQ0EsSUFBQSxDQUFBLE1BQUEsQ0FBQSxhQUFBOztVQUdBLGlCQUFBLEdBQUEsZUFBQSxDQUFBLEdBQUE7VUFDQSxTQUFBLEdBQUEsZUFBQSxDQUFBLEdBQUEsRUFBQSxvQkFBQTtRQUNBLFNBQUEsS0FBQSxpQkFBQTtrQkFDQSxLQUFBLEVBQ0EsNENBQUEsS0FDQSxTQUFBLEVBQUEsaUJBQUEsQ0FBQSxTQUFBLEVBQUEsU0FBQTs7O2dCQUtBLGVBQUEsQ0FBQSxNQUtBO2VBQ0EsYUFBQSxDQUFBLE1BQUEifQ==