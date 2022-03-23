// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { encode } from "../encoding/utf8.ts";
import { BufReader, BufWriter } from "../io/bufio.ts";
import { assert } from "../_util/assert.ts";
import { deferred, MuxAsyncIterator } from "../async/mod.ts";
import { bodyReader, chunkedBodyReader, emptyReader, readRequest, writeResponse } from "./_io.ts";
export class ServerRequest {
    done = deferred();
    _contentLength = undefined;
    /**
   * Value of Content-Length header.
   * If null, then content length is invalid or not given (e.g. chunked encoding).
   */ get contentLength() {
        // undefined means not cached.
        // null means invalid or not provided.
        if (this._contentLength === undefined) {
            const cl = this.headers.get("content-length");
            if (cl) {
                this._contentLength = parseInt(cl);
                // Convert NaN to null (as NaN harder to test)
                if (Number.isNaN(this._contentLength)) {
                    this._contentLength = null;
                }
            } else {
                this._contentLength = null;
            }
        }
        return this._contentLength;
    }
    _body = null;
    /**
   * Body of the request.  The easiest way to consume the body is:
   *
   *     const buf: Uint8Array = await Deno.readAll(req.body);
   */ get body() {
        if (!this._body) {
            if (this.contentLength != null) {
                this._body = bodyReader(this.contentLength, this.r);
            } else {
                const transferEncoding = this.headers.get("transfer-encoding");
                if (transferEncoding != null) {
                    const parts = transferEncoding.split(",").map((e)=>e.trim().toLowerCase()
                    );
                    assert(parts.includes("chunked"), 'transfer-encoding must include "chunked" if content-length is not set');
                    this._body = chunkedBodyReader(this.headers, this.r);
                } else {
                    // Neither content-length nor transfer-encoding: chunked
                    this._body = emptyReader();
                }
            }
        }
        return this._body;
    }
    async respond(r) {
        let err;
        try {
            // Write our response!
            await writeResponse(this.w, r);
        } catch (e) {
            try {
                // Eagerly close on error.
                this.conn.close();
            } catch  {
            }
            err = e;
        }
        // Signal that this request has been processed and the next pipelined
        // request on the same connection can be accepted.
        this.done.resolve(err);
        if (err) {
            // Error during responding, rethrow.
            throw err;
        }
    }
    finalized = false;
    async finalize() {
        if (this.finalized) return;
        // Consume unread body
        const body = this.body;
        const buf = new Uint8Array(1024);
        while(await body.read(buf) !== null){
        }
        this.finalized = true;
    }
}
export class Server {
    closing = false;
    connections = [];
    constructor(listener){
        this.listener = listener;
    }
    close() {
        this.closing = true;
        this.listener.close();
        for (const conn of this.connections){
            try {
                conn.close();
            } catch (e) {
                // Connection might have been already closed
                if (!(e instanceof Deno.errors.BadResource)) {
                    throw e;
                }
            }
        }
    }
    // Yields all HTTP requests on a single TCP connection.
    async *iterateHttpRequests(conn) {
        const reader = new BufReader(conn);
        const writer = new BufWriter(conn);
        while(!this.closing){
            let request;
            try {
                request = await readRequest(conn, reader);
            } catch (error) {
                if (error instanceof Deno.errors.InvalidData || error instanceof Deno.errors.UnexpectedEof) {
                    // An error was thrown while parsing request headers.
                    await writeResponse(writer, {
                        status: 400,
                        body: encode(`${error.message}\r\n\r\n`)
                    });
                }
                break;
            }
            if (request === null) {
                break;
            }
            request.w = writer;
            yield request;
            // Wait for the request to be processed before we accept a new request on
            // this connection.
            const responseError = await request.done;
            if (responseError) {
                // Something bad happened during response.
                // (likely other side closed during pipelined req)
                // req.done implies this connection already closed, so we can just return.
                this.untrackConnection(request.conn);
                return;
            }
            // Consume unread body and trailers if receiver didn't consume those data
            await request.finalize();
        }
        this.untrackConnection(conn);
        try {
            conn.close();
        } catch (e) {
        }
    }
    trackConnection(conn) {
        this.connections.push(conn);
    }
    untrackConnection(conn) {
        const index = this.connections.indexOf(conn);
        if (index !== -1) {
            this.connections.splice(index, 1);
        }
    }
    // Accepts a new TCP connection and yields all HTTP requests that arrive on
    // it. When a connection is accepted, it also creates a new iterator of the
    // same kind and adds it to the request multiplexer so that another TCP
    // connection can be accepted.
    async *acceptConnAndIterateHttpRequests(mux) {
        if (this.closing) return;
        // Wait for a new connection.
        let conn;
        try {
            conn = await this.listener.accept();
        } catch (error) {
            if (error instanceof Deno.errors.BadResource || error instanceof Deno.errors.InvalidData || error instanceof Deno.errors.UnexpectedEof) {
                return mux.add(this.acceptConnAndIterateHttpRequests(mux));
            }
            throw error;
        }
        this.trackConnection(conn);
        // Try to accept another connection and add it to the multiplexer.
        mux.add(this.acceptConnAndIterateHttpRequests(mux));
        // Yield the requests that arrive on the just-accepted connection.
        yield* this.iterateHttpRequests(conn);
    }
    [Symbol.asyncIterator]() {
        const mux = new MuxAsyncIterator();
        mux.add(this.acceptConnAndIterateHttpRequests(mux));
        return mux.iterate();
    }
}
/**
 * Parse addr from string
 *
 *     const addr = "::1:8000";
 *     parseAddrFromString(addr);
 *
 * @param addr Address string
 */ export function _parseAddrFromStr(addr) {
    let url;
    try {
        const host = addr.startsWith(":") ? `0.0.0.0${addr}` : addr;
        url = new URL(`http://${host}`);
    } catch  {
        throw new TypeError("Invalid address.");
    }
    if (url.username || url.password || url.pathname != "/" || url.search || url.hash) {
        throw new TypeError("Invalid address.");
    }
    return {
        hostname: url.hostname,
        port: url.port === "" ? 80 : Number(url.port)
    };
}
/**
 * Create a HTTP server
 *
 *     import { serve } from "https://deno.land/std/http/server.ts";
 *     const body = "Hello World\n";
 *     const server = serve({ port: 8000 });
 *     for await (const req of server) {
 *       req.respond({ body });
 *     }
 */ export function serve(addr) {
    if (typeof addr === "string") {
        addr = _parseAddrFromStr(addr);
    }
    const listener = Deno.listen(addr);
    return new Server(listener);
}
/**
 * Start an HTTP server with given options and request handler
 *
 *     const body = "Hello World\n";
 *     const options = { port: 8000 };
 *     listenAndServe(options, (req) => {
 *       req.respond({ body });
 *     });
 *
 * @param options Server configuration
 * @param handler Request handler
 */ export async function listenAndServe(addr, handler) {
    const server = serve(addr);
    for await (const request of server){
        handler(request);
    }
}
/**
 * Create an HTTPS server with given options
 *
 *     const body = "Hello HTTPS";
 *     const options = {
 *       hostname: "localhost",
 *       port: 443,
 *       certFile: "./path/to/localhost.crt",
 *       keyFile: "./path/to/localhost.key",
 *     };
 *     for await (const req of serveTLS(options)) {
 *       req.respond({ body });
 *     }
 *
 * @param options Server configuration
 * @return Async iterable server instance for incoming requests
 */ export function serveTLS(options) {
    const tlsOptions = {
        ...options,
        transport: "tcp"
    };
    const listener = Deno.listenTls(tlsOptions);
    return new Server(listener);
}
/**
 * Start an HTTPS server with given options and request handler
 *
 *     const body = "Hello HTTPS";
 *     const options = {
 *       hostname: "localhost",
 *       port: 443,
 *       certFile: "./path/to/localhost.crt",
 *       keyFile: "./path/to/localhost.key",
 *     };
 *     listenAndServeTLS(options, (req) => {
 *       req.respond({ body });
 *     });
 *
 * @param options Server configuration
 * @param handler Request handler
 */ export async function listenAndServeTLS(options, handler) {
    const server = serveTLS(options);
    for await (const request of server){
        handler(request);
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL2h0dHAvc2VydmVyLnRzPiJdLCJzb3VyY2VzQ29udGVudCI6WyIvLyBDb3B5cmlnaHQgMjAxOC0yMDIwIHRoZSBEZW5vIGF1dGhvcnMuIEFsbCByaWdodHMgcmVzZXJ2ZWQuIE1JVCBsaWNlbnNlLlxuaW1wb3J0IHsgZW5jb2RlIH0gZnJvbSBcIi4uL2VuY29kaW5nL3V0ZjgudHNcIjtcbmltcG9ydCB7IEJ1ZlJlYWRlciwgQnVmV3JpdGVyIH0gZnJvbSBcIi4uL2lvL2J1ZmlvLnRzXCI7XG5pbXBvcnQgeyBhc3NlcnQgfSBmcm9tIFwiLi4vX3V0aWwvYXNzZXJ0LnRzXCI7XG5pbXBvcnQgeyBEZWZlcnJlZCwgZGVmZXJyZWQsIE11eEFzeW5jSXRlcmF0b3IgfSBmcm9tIFwiLi4vYXN5bmMvbW9kLnRzXCI7XG5pbXBvcnQge1xuICBib2R5UmVhZGVyLFxuICBjaHVua2VkQm9keVJlYWRlcixcbiAgZW1wdHlSZWFkZXIsXG4gIHJlYWRSZXF1ZXN0LFxuICB3cml0ZVJlc3BvbnNlLFxufSBmcm9tIFwiLi9faW8udHNcIjtcblxuZXhwb3J0IGNsYXNzIFNlcnZlclJlcXVlc3Qge1xuICB1cmwhOiBzdHJpbmc7XG4gIG1ldGhvZCE6IHN0cmluZztcbiAgcHJvdG8hOiBzdHJpbmc7XG4gIHByb3RvTWlub3IhOiBudW1iZXI7XG4gIHByb3RvTWFqb3IhOiBudW1iZXI7XG4gIGhlYWRlcnMhOiBIZWFkZXJzO1xuICBjb25uITogRGVuby5Db25uO1xuICByITogQnVmUmVhZGVyO1xuICB3ITogQnVmV3JpdGVyO1xuICBkb25lOiBEZWZlcnJlZDxFcnJvciB8IHVuZGVmaW5lZD4gPSBkZWZlcnJlZCgpO1xuXG4gIHByaXZhdGUgX2NvbnRlbnRMZW5ndGg6IG51bWJlciB8IHVuZGVmaW5lZCB8IG51bGwgPSB1bmRlZmluZWQ7XG4gIC8qKlxuICAgKiBWYWx1ZSBvZiBDb250ZW50LUxlbmd0aCBoZWFkZXIuXG4gICAqIElmIG51bGwsIHRoZW4gY29udGVudCBsZW5ndGggaXMgaW52YWxpZCBvciBub3QgZ2l2ZW4gKGUuZy4gY2h1bmtlZCBlbmNvZGluZykuXG4gICAqL1xuICBnZXQgY29udGVudExlbmd0aCgpOiBudW1iZXIgfCBudWxsIHtcbiAgICAvLyB1bmRlZmluZWQgbWVhbnMgbm90IGNhY2hlZC5cbiAgICAvLyBudWxsIG1lYW5zIGludmFsaWQgb3Igbm90IHByb3ZpZGVkLlxuICAgIGlmICh0aGlzLl9jb250ZW50TGVuZ3RoID09PSB1bmRlZmluZWQpIHtcbiAgICAgIGNvbnN0IGNsID0gdGhpcy5oZWFkZXJzLmdldChcImNvbnRlbnQtbGVuZ3RoXCIpO1xuICAgICAgaWYgKGNsKSB7XG4gICAgICAgIHRoaXMuX2NvbnRlbnRMZW5ndGggPSBwYXJzZUludChjbCk7XG4gICAgICAgIC8vIENvbnZlcnQgTmFOIHRvIG51bGwgKGFzIE5hTiBoYXJkZXIgdG8gdGVzdClcbiAgICAgICAgaWYgKE51bWJlci5pc05hTih0aGlzLl9jb250ZW50TGVuZ3RoKSkge1xuICAgICAgICAgIHRoaXMuX2NvbnRlbnRMZW5ndGggPSBudWxsO1xuICAgICAgICB9XG4gICAgICB9IGVsc2Uge1xuICAgICAgICB0aGlzLl9jb250ZW50TGVuZ3RoID0gbnVsbDtcbiAgICAgIH1cbiAgICB9XG4gICAgcmV0dXJuIHRoaXMuX2NvbnRlbnRMZW5ndGg7XG4gIH1cblxuICBwcml2YXRlIF9ib2R5OiBEZW5vLlJlYWRlciB8IG51bGwgPSBudWxsO1xuXG4gIC8qKlxuICAgKiBCb2R5IG9mIHRoZSByZXF1ZXN0LiAgVGhlIGVhc2llc3Qgd2F5IHRvIGNvbnN1bWUgdGhlIGJvZHkgaXM6XG4gICAqXG4gICAqICAgICBjb25zdCBidWY6IFVpbnQ4QXJyYXkgPSBhd2FpdCBEZW5vLnJlYWRBbGwocmVxLmJvZHkpO1xuICAgKi9cbiAgZ2V0IGJvZHkoKTogRGVuby5SZWFkZXIge1xuICAgIGlmICghdGhpcy5fYm9keSkge1xuICAgICAgaWYgKHRoaXMuY29udGVudExlbmd0aCAhPSBudWxsKSB7XG4gICAgICAgIHRoaXMuX2JvZHkgPSBib2R5UmVhZGVyKHRoaXMuY29udGVudExlbmd0aCwgdGhpcy5yKTtcbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIGNvbnN0IHRyYW5zZmVyRW5jb2RpbmcgPSB0aGlzLmhlYWRlcnMuZ2V0KFwidHJhbnNmZXItZW5jb2RpbmdcIik7XG4gICAgICAgIGlmICh0cmFuc2ZlckVuY29kaW5nICE9IG51bGwpIHtcbiAgICAgICAgICBjb25zdCBwYXJ0cyA9IHRyYW5zZmVyRW5jb2RpbmdcbiAgICAgICAgICAgIC5zcGxpdChcIixcIilcbiAgICAgICAgICAgIC5tYXAoKGUpOiBzdHJpbmcgPT4gZS50cmltKCkudG9Mb3dlckNhc2UoKSk7XG4gICAgICAgICAgYXNzZXJ0KFxuICAgICAgICAgICAgcGFydHMuaW5jbHVkZXMoXCJjaHVua2VkXCIpLFxuICAgICAgICAgICAgJ3RyYW5zZmVyLWVuY29kaW5nIG11c3QgaW5jbHVkZSBcImNodW5rZWRcIiBpZiBjb250ZW50LWxlbmd0aCBpcyBub3Qgc2V0JyxcbiAgICAgICAgICApO1xuICAgICAgICAgIHRoaXMuX2JvZHkgPSBjaHVua2VkQm9keVJlYWRlcih0aGlzLmhlYWRlcnMsIHRoaXMucik7XG4gICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgLy8gTmVpdGhlciBjb250ZW50LWxlbmd0aCBub3IgdHJhbnNmZXItZW5jb2Rpbmc6IGNodW5rZWRcbiAgICAgICAgICB0aGlzLl9ib2R5ID0gZW1wdHlSZWFkZXIoKTtcbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cbiAgICByZXR1cm4gdGhpcy5fYm9keTtcbiAgfVxuXG4gIGFzeW5jIHJlc3BvbmQocjogUmVzcG9uc2UpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBsZXQgZXJyOiBFcnJvciB8IHVuZGVmaW5lZDtcbiAgICB0cnkge1xuICAgICAgLy8gV3JpdGUgb3VyIHJlc3BvbnNlIVxuICAgICAgYXdhaXQgd3JpdGVSZXNwb25zZSh0aGlzLncsIHIpO1xuICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgIHRyeSB7XG4gICAgICAgIC8vIEVhZ2VybHkgY2xvc2Ugb24gZXJyb3IuXG4gICAgICAgIHRoaXMuY29ubi5jbG9zZSgpO1xuICAgICAgfSBjYXRjaCB7XG4gICAgICAgIC8vIFBhc3NcbiAgICAgIH1cbiAgICAgIGVyciA9IGU7XG4gICAgfVxuICAgIC8vIFNpZ25hbCB0aGF0IHRoaXMgcmVxdWVzdCBoYXMgYmVlbiBwcm9jZXNzZWQgYW5kIHRoZSBuZXh0IHBpcGVsaW5lZFxuICAgIC8vIHJlcXVlc3Qgb24gdGhlIHNhbWUgY29ubmVjdGlvbiBjYW4gYmUgYWNjZXB0ZWQuXG4gICAgdGhpcy5kb25lLnJlc29sdmUoZXJyKTtcbiAgICBpZiAoZXJyKSB7XG4gICAgICAvLyBFcnJvciBkdXJpbmcgcmVzcG9uZGluZywgcmV0aHJvdy5cbiAgICAgIHRocm93IGVycjtcbiAgICB9XG4gIH1cblxuICBwcml2YXRlIGZpbmFsaXplZCA9IGZhbHNlO1xuICBhc3luYyBmaW5hbGl6ZSgpOiBQcm9taXNlPHZvaWQ+IHtcbiAgICBpZiAodGhpcy5maW5hbGl6ZWQpIHJldHVybjtcbiAgICAvLyBDb25zdW1lIHVucmVhZCBib2R5XG4gICAgY29uc3QgYm9keSA9IHRoaXMuYm9keTtcbiAgICBjb25zdCBidWYgPSBuZXcgVWludDhBcnJheSgxMDI0KTtcbiAgICB3aGlsZSAoKGF3YWl0IGJvZHkucmVhZChidWYpKSAhPT0gbnVsbCkge1xuICAgICAgLy8gUGFzc1xuICAgIH1cbiAgICB0aGlzLmZpbmFsaXplZCA9IHRydWU7XG4gIH1cbn1cblxuZXhwb3J0IGNsYXNzIFNlcnZlciBpbXBsZW1lbnRzIEFzeW5jSXRlcmFibGU8U2VydmVyUmVxdWVzdD4ge1xuICBwcml2YXRlIGNsb3NpbmcgPSBmYWxzZTtcbiAgcHJpdmF0ZSBjb25uZWN0aW9uczogRGVuby5Db25uW10gPSBbXTtcblxuICBjb25zdHJ1Y3RvcihwdWJsaWMgbGlzdGVuZXI6IERlbm8uTGlzdGVuZXIpIHt9XG5cbiAgY2xvc2UoKTogdm9pZCB7XG4gICAgdGhpcy5jbG9zaW5nID0gdHJ1ZTtcbiAgICB0aGlzLmxpc3RlbmVyLmNsb3NlKCk7XG4gICAgZm9yIChjb25zdCBjb25uIG9mIHRoaXMuY29ubmVjdGlvbnMpIHtcbiAgICAgIHRyeSB7XG4gICAgICAgIGNvbm4uY2xvc2UoKTtcbiAgICAgIH0gY2F0Y2ggKGUpIHtcbiAgICAgICAgLy8gQ29ubmVjdGlvbiBtaWdodCBoYXZlIGJlZW4gYWxyZWFkeSBjbG9zZWRcbiAgICAgICAgaWYgKCEoZSBpbnN0YW5jZW9mIERlbm8uZXJyb3JzLkJhZFJlc291cmNlKSkge1xuICAgICAgICAgIHRocm93IGU7XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH1cblxuICAvLyBZaWVsZHMgYWxsIEhUVFAgcmVxdWVzdHMgb24gYSBzaW5nbGUgVENQIGNvbm5lY3Rpb24uXG4gIHByaXZhdGUgYXN5bmMgKml0ZXJhdGVIdHRwUmVxdWVzdHMoXG4gICAgY29ubjogRGVuby5Db25uLFxuICApOiBBc3luY0l0ZXJhYmxlSXRlcmF0b3I8U2VydmVyUmVxdWVzdD4ge1xuICAgIGNvbnN0IHJlYWRlciA9IG5ldyBCdWZSZWFkZXIoY29ubik7XG4gICAgY29uc3Qgd3JpdGVyID0gbmV3IEJ1ZldyaXRlcihjb25uKTtcblxuICAgIHdoaWxlICghdGhpcy5jbG9zaW5nKSB7XG4gICAgICBsZXQgcmVxdWVzdDogU2VydmVyUmVxdWVzdCB8IG51bGw7XG4gICAgICB0cnkge1xuICAgICAgICByZXF1ZXN0ID0gYXdhaXQgcmVhZFJlcXVlc3QoY29ubiwgcmVhZGVyKTtcbiAgICAgIH0gY2F0Y2ggKGVycm9yKSB7XG4gICAgICAgIGlmIChcbiAgICAgICAgICBlcnJvciBpbnN0YW5jZW9mIERlbm8uZXJyb3JzLkludmFsaWREYXRhIHx8XG4gICAgICAgICAgZXJyb3IgaW5zdGFuY2VvZiBEZW5vLmVycm9ycy5VbmV4cGVjdGVkRW9mXG4gICAgICAgICkge1xuICAgICAgICAgIC8vIEFuIGVycm9yIHdhcyB0aHJvd24gd2hpbGUgcGFyc2luZyByZXF1ZXN0IGhlYWRlcnMuXG4gICAgICAgICAgYXdhaXQgd3JpdGVSZXNwb25zZSh3cml0ZXIsIHtcbiAgICAgICAgICAgIHN0YXR1czogNDAwLFxuICAgICAgICAgICAgYm9keTogZW5jb2RlKGAke2Vycm9yLm1lc3NhZ2V9XFxyXFxuXFxyXFxuYCksXG4gICAgICAgICAgfSk7XG4gICAgICAgIH1cbiAgICAgICAgYnJlYWs7XG4gICAgICB9XG4gICAgICBpZiAocmVxdWVzdCA9PT0gbnVsbCkge1xuICAgICAgICBicmVhaztcbiAgICAgIH1cblxuICAgICAgcmVxdWVzdC53ID0gd3JpdGVyO1xuICAgICAgeWllbGQgcmVxdWVzdDtcblxuICAgICAgLy8gV2FpdCBmb3IgdGhlIHJlcXVlc3QgdG8gYmUgcHJvY2Vzc2VkIGJlZm9yZSB3ZSBhY2NlcHQgYSBuZXcgcmVxdWVzdCBvblxuICAgICAgLy8gdGhpcyBjb25uZWN0aW9uLlxuICAgICAgY29uc3QgcmVzcG9uc2VFcnJvciA9IGF3YWl0IHJlcXVlc3QuZG9uZTtcbiAgICAgIGlmIChyZXNwb25zZUVycm9yKSB7XG4gICAgICAgIC8vIFNvbWV0aGluZyBiYWQgaGFwcGVuZWQgZHVyaW5nIHJlc3BvbnNlLlxuICAgICAgICAvLyAobGlrZWx5IG90aGVyIHNpZGUgY2xvc2VkIGR1cmluZyBwaXBlbGluZWQgcmVxKVxuICAgICAgICAvLyByZXEuZG9uZSBpbXBsaWVzIHRoaXMgY29ubmVjdGlvbiBhbHJlYWR5IGNsb3NlZCwgc28gd2UgY2FuIGp1c3QgcmV0dXJuLlxuICAgICAgICB0aGlzLnVudHJhY2tDb25uZWN0aW9uKHJlcXVlc3QuY29ubik7XG4gICAgICAgIHJldHVybjtcbiAgICAgIH1cbiAgICAgIC8vIENvbnN1bWUgdW5yZWFkIGJvZHkgYW5kIHRyYWlsZXJzIGlmIHJlY2VpdmVyIGRpZG4ndCBjb25zdW1lIHRob3NlIGRhdGFcbiAgICAgIGF3YWl0IHJlcXVlc3QuZmluYWxpemUoKTtcbiAgICB9XG5cbiAgICB0aGlzLnVudHJhY2tDb25uZWN0aW9uKGNvbm4pO1xuICAgIHRyeSB7XG4gICAgICBjb25uLmNsb3NlKCk7XG4gICAgfSBjYXRjaCAoZSkge1xuICAgICAgLy8gbWlnaHQgaGF2ZSBiZWVuIGFscmVhZHkgY2xvc2VkXG4gICAgfVxuICB9XG5cbiAgcHJpdmF0ZSB0cmFja0Nvbm5lY3Rpb24oY29ubjogRGVuby5Db25uKTogdm9pZCB7XG4gICAgdGhpcy5jb25uZWN0aW9ucy5wdXNoKGNvbm4pO1xuICB9XG5cbiAgcHJpdmF0ZSB1bnRyYWNrQ29ubmVjdGlvbihjb25uOiBEZW5vLkNvbm4pOiB2b2lkIHtcbiAgICBjb25zdCBpbmRleCA9IHRoaXMuY29ubmVjdGlvbnMuaW5kZXhPZihjb25uKTtcbiAgICBpZiAoaW5kZXggIT09IC0xKSB7XG4gICAgICB0aGlzLmNvbm5lY3Rpb25zLnNwbGljZShpbmRleCwgMSk7XG4gICAgfVxuICB9XG5cbiAgLy8gQWNjZXB0cyBhIG5ldyBUQ1AgY29ubmVjdGlvbiBhbmQgeWllbGRzIGFsbCBIVFRQIHJlcXVlc3RzIHRoYXQgYXJyaXZlIG9uXG4gIC8vIGl0LiBXaGVuIGEgY29ubmVjdGlvbiBpcyBhY2NlcHRlZCwgaXQgYWxzbyBjcmVhdGVzIGEgbmV3IGl0ZXJhdG9yIG9mIHRoZVxuICAvLyBzYW1lIGtpbmQgYW5kIGFkZHMgaXQgdG8gdGhlIHJlcXVlc3QgbXVsdGlwbGV4ZXIgc28gdGhhdCBhbm90aGVyIFRDUFxuICAvLyBjb25uZWN0aW9uIGNhbiBiZSBhY2NlcHRlZC5cbiAgcHJpdmF0ZSBhc3luYyAqYWNjZXB0Q29ubkFuZEl0ZXJhdGVIdHRwUmVxdWVzdHMoXG4gICAgbXV4OiBNdXhBc3luY0l0ZXJhdG9yPFNlcnZlclJlcXVlc3Q+LFxuICApOiBBc3luY0l0ZXJhYmxlSXRlcmF0b3I8U2VydmVyUmVxdWVzdD4ge1xuICAgIGlmICh0aGlzLmNsb3NpbmcpIHJldHVybjtcbiAgICAvLyBXYWl0IGZvciBhIG5ldyBjb25uZWN0aW9uLlxuICAgIGxldCBjb25uOiBEZW5vLkNvbm47XG4gICAgdHJ5IHtcbiAgICAgIGNvbm4gPSBhd2FpdCB0aGlzLmxpc3RlbmVyLmFjY2VwdCgpO1xuICAgIH0gY2F0Y2ggKGVycm9yKSB7XG4gICAgICBpZiAoXG4gICAgICAgIGVycm9yIGluc3RhbmNlb2YgRGVuby5lcnJvcnMuQmFkUmVzb3VyY2UgfHxcbiAgICAgICAgZXJyb3IgaW5zdGFuY2VvZiBEZW5vLmVycm9ycy5JbnZhbGlkRGF0YSB8fFxuICAgICAgICBlcnJvciBpbnN0YW5jZW9mIERlbm8uZXJyb3JzLlVuZXhwZWN0ZWRFb2ZcbiAgICAgICkge1xuICAgICAgICByZXR1cm4gbXV4LmFkZCh0aGlzLmFjY2VwdENvbm5BbmRJdGVyYXRlSHR0cFJlcXVlc3RzKG11eCkpO1xuICAgICAgfVxuICAgICAgdGhyb3cgZXJyb3I7XG4gICAgfVxuICAgIHRoaXMudHJhY2tDb25uZWN0aW9uKGNvbm4pO1xuICAgIC8vIFRyeSB0byBhY2NlcHQgYW5vdGhlciBjb25uZWN0aW9uIGFuZCBhZGQgaXQgdG8gdGhlIG11bHRpcGxleGVyLlxuICAgIG11eC5hZGQodGhpcy5hY2NlcHRDb25uQW5kSXRlcmF0ZUh0dHBSZXF1ZXN0cyhtdXgpKTtcbiAgICAvLyBZaWVsZCB0aGUgcmVxdWVzdHMgdGhhdCBhcnJpdmUgb24gdGhlIGp1c3QtYWNjZXB0ZWQgY29ubmVjdGlvbi5cbiAgICB5aWVsZCogdGhpcy5pdGVyYXRlSHR0cFJlcXVlc3RzKGNvbm4pO1xuICB9XG5cbiAgW1N5bWJvbC5hc3luY0l0ZXJhdG9yXSgpOiBBc3luY0l0ZXJhYmxlSXRlcmF0b3I8U2VydmVyUmVxdWVzdD4ge1xuICAgIGNvbnN0IG11eDogTXV4QXN5bmNJdGVyYXRvcjxTZXJ2ZXJSZXF1ZXN0PiA9IG5ldyBNdXhBc3luY0l0ZXJhdG9yKCk7XG4gICAgbXV4LmFkZCh0aGlzLmFjY2VwdENvbm5BbmRJdGVyYXRlSHR0cFJlcXVlc3RzKG11eCkpO1xuICAgIHJldHVybiBtdXguaXRlcmF0ZSgpO1xuICB9XG59XG5cbi8qKiBPcHRpb25zIGZvciBjcmVhdGluZyBhbiBIVFRQIHNlcnZlci4gKi9cbmV4cG9ydCB0eXBlIEhUVFBPcHRpb25zID0gT21pdDxEZW5vLkxpc3Rlbk9wdGlvbnMsIFwidHJhbnNwb3J0XCI+O1xuXG4vKipcbiAqIFBhcnNlIGFkZHIgZnJvbSBzdHJpbmdcbiAqXG4gKiAgICAgY29uc3QgYWRkciA9IFwiOjoxOjgwMDBcIjtcbiAqICAgICBwYXJzZUFkZHJGcm9tU3RyaW5nKGFkZHIpO1xuICpcbiAqIEBwYXJhbSBhZGRyIEFkZHJlc3Mgc3RyaW5nXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBfcGFyc2VBZGRyRnJvbVN0cihhZGRyOiBzdHJpbmcpOiBIVFRQT3B0aW9ucyB7XG4gIGxldCB1cmw6IFVSTDtcbiAgdHJ5IHtcbiAgICBjb25zdCBob3N0ID0gYWRkci5zdGFydHNXaXRoKFwiOlwiKSA/IGAwLjAuMC4wJHthZGRyfWAgOiBhZGRyO1xuICAgIHVybCA9IG5ldyBVUkwoYGh0dHA6Ly8ke2hvc3R9YCk7XG4gIH0gY2F0Y2gge1xuICAgIHRocm93IG5ldyBUeXBlRXJyb3IoXCJJbnZhbGlkIGFkZHJlc3MuXCIpO1xuICB9XG4gIGlmIChcbiAgICB1cmwudXNlcm5hbWUgfHxcbiAgICB1cmwucGFzc3dvcmQgfHxcbiAgICB1cmwucGF0aG5hbWUgIT0gXCIvXCIgfHxcbiAgICB1cmwuc2VhcmNoIHx8XG4gICAgdXJsLmhhc2hcbiAgKSB7XG4gICAgdGhyb3cgbmV3IFR5cGVFcnJvcihcIkludmFsaWQgYWRkcmVzcy5cIik7XG4gIH1cblxuICByZXR1cm4ge1xuICAgIGhvc3RuYW1lOiB1cmwuaG9zdG5hbWUsXG4gICAgcG9ydDogdXJsLnBvcnQgPT09IFwiXCIgPyA4MCA6IE51bWJlcih1cmwucG9ydCksXG4gIH07XG59XG5cbi8qKlxuICogQ3JlYXRlIGEgSFRUUCBzZXJ2ZXJcbiAqXG4gKiAgICAgaW1wb3J0IHsgc2VydmUgfSBmcm9tIFwiaHR0cHM6Ly9kZW5vLmxhbmQvc3RkL2h0dHAvc2VydmVyLnRzXCI7XG4gKiAgICAgY29uc3QgYm9keSA9IFwiSGVsbG8gV29ybGRcXG5cIjtcbiAqICAgICBjb25zdCBzZXJ2ZXIgPSBzZXJ2ZSh7IHBvcnQ6IDgwMDAgfSk7XG4gKiAgICAgZm9yIGF3YWl0IChjb25zdCByZXEgb2Ygc2VydmVyKSB7XG4gKiAgICAgICByZXEucmVzcG9uZCh7IGJvZHkgfSk7XG4gKiAgICAgfVxuICovXG5leHBvcnQgZnVuY3Rpb24gc2VydmUoYWRkcjogc3RyaW5nIHwgSFRUUE9wdGlvbnMpOiBTZXJ2ZXIge1xuICBpZiAodHlwZW9mIGFkZHIgPT09IFwic3RyaW5nXCIpIHtcbiAgICBhZGRyID0gX3BhcnNlQWRkckZyb21TdHIoYWRkcik7XG4gIH1cblxuICBjb25zdCBsaXN0ZW5lciA9IERlbm8ubGlzdGVuKGFkZHIpO1xuICByZXR1cm4gbmV3IFNlcnZlcihsaXN0ZW5lcik7XG59XG5cbi8qKlxuICogU3RhcnQgYW4gSFRUUCBzZXJ2ZXIgd2l0aCBnaXZlbiBvcHRpb25zIGFuZCByZXF1ZXN0IGhhbmRsZXJcbiAqXG4gKiAgICAgY29uc3QgYm9keSA9IFwiSGVsbG8gV29ybGRcXG5cIjtcbiAqICAgICBjb25zdCBvcHRpb25zID0geyBwb3J0OiA4MDAwIH07XG4gKiAgICAgbGlzdGVuQW5kU2VydmUob3B0aW9ucywgKHJlcSkgPT4ge1xuICogICAgICAgcmVxLnJlc3BvbmQoeyBib2R5IH0pO1xuICogICAgIH0pO1xuICpcbiAqIEBwYXJhbSBvcHRpb25zIFNlcnZlciBjb25maWd1cmF0aW9uXG4gKiBAcGFyYW0gaGFuZGxlciBSZXF1ZXN0IGhhbmRsZXJcbiAqL1xuZXhwb3J0IGFzeW5jIGZ1bmN0aW9uIGxpc3RlbkFuZFNlcnZlKFxuICBhZGRyOiBzdHJpbmcgfCBIVFRQT3B0aW9ucyxcbiAgaGFuZGxlcjogKHJlcTogU2VydmVyUmVxdWVzdCkgPT4gdm9pZCxcbik6IFByb21pc2U8dm9pZD4ge1xuICBjb25zdCBzZXJ2ZXIgPSBzZXJ2ZShhZGRyKTtcblxuICBmb3IgYXdhaXQgKGNvbnN0IHJlcXVlc3Qgb2Ygc2VydmVyKSB7XG4gICAgaGFuZGxlcihyZXF1ZXN0KTtcbiAgfVxufVxuXG4vKiogT3B0aW9ucyBmb3IgY3JlYXRpbmcgYW4gSFRUUFMgc2VydmVyLiAqL1xuZXhwb3J0IHR5cGUgSFRUUFNPcHRpb25zID0gT21pdDxEZW5vLkxpc3RlblRsc09wdGlvbnMsIFwidHJhbnNwb3J0XCI+O1xuXG4vKipcbiAqIENyZWF0ZSBhbiBIVFRQUyBzZXJ2ZXIgd2l0aCBnaXZlbiBvcHRpb25zXG4gKlxuICogICAgIGNvbnN0IGJvZHkgPSBcIkhlbGxvIEhUVFBTXCI7XG4gKiAgICAgY29uc3Qgb3B0aW9ucyA9IHtcbiAqICAgICAgIGhvc3RuYW1lOiBcImxvY2FsaG9zdFwiLFxuICogICAgICAgcG9ydDogNDQzLFxuICogICAgICAgY2VydEZpbGU6IFwiLi9wYXRoL3RvL2xvY2FsaG9zdC5jcnRcIixcbiAqICAgICAgIGtleUZpbGU6IFwiLi9wYXRoL3RvL2xvY2FsaG9zdC5rZXlcIixcbiAqICAgICB9O1xuICogICAgIGZvciBhd2FpdCAoY29uc3QgcmVxIG9mIHNlcnZlVExTKG9wdGlvbnMpKSB7XG4gKiAgICAgICByZXEucmVzcG9uZCh7IGJvZHkgfSk7XG4gKiAgICAgfVxuICpcbiAqIEBwYXJhbSBvcHRpb25zIFNlcnZlciBjb25maWd1cmF0aW9uXG4gKiBAcmV0dXJuIEFzeW5jIGl0ZXJhYmxlIHNlcnZlciBpbnN0YW5jZSBmb3IgaW5jb21pbmcgcmVxdWVzdHNcbiAqL1xuZXhwb3J0IGZ1bmN0aW9uIHNlcnZlVExTKG9wdGlvbnM6IEhUVFBTT3B0aW9ucyk6IFNlcnZlciB7XG4gIGNvbnN0IHRsc09wdGlvbnM6IERlbm8uTGlzdGVuVGxzT3B0aW9ucyA9IHtcbiAgICAuLi5vcHRpb25zLFxuICAgIHRyYW5zcG9ydDogXCJ0Y3BcIixcbiAgfTtcbiAgY29uc3QgbGlzdGVuZXIgPSBEZW5vLmxpc3RlblRscyh0bHNPcHRpb25zKTtcbiAgcmV0dXJuIG5ldyBTZXJ2ZXIobGlzdGVuZXIpO1xufVxuXG4vKipcbiAqIFN0YXJ0IGFuIEhUVFBTIHNlcnZlciB3aXRoIGdpdmVuIG9wdGlvbnMgYW5kIHJlcXVlc3QgaGFuZGxlclxuICpcbiAqICAgICBjb25zdCBib2R5ID0gXCJIZWxsbyBIVFRQU1wiO1xuICogICAgIGNvbnN0IG9wdGlvbnMgPSB7XG4gKiAgICAgICBob3N0bmFtZTogXCJsb2NhbGhvc3RcIixcbiAqICAgICAgIHBvcnQ6IDQ0MyxcbiAqICAgICAgIGNlcnRGaWxlOiBcIi4vcGF0aC90by9sb2NhbGhvc3QuY3J0XCIsXG4gKiAgICAgICBrZXlGaWxlOiBcIi4vcGF0aC90by9sb2NhbGhvc3Qua2V5XCIsXG4gKiAgICAgfTtcbiAqICAgICBsaXN0ZW5BbmRTZXJ2ZVRMUyhvcHRpb25zLCAocmVxKSA9PiB7XG4gKiAgICAgICByZXEucmVzcG9uZCh7IGJvZHkgfSk7XG4gKiAgICAgfSk7XG4gKlxuICogQHBhcmFtIG9wdGlvbnMgU2VydmVyIGNvbmZpZ3VyYXRpb25cbiAqIEBwYXJhbSBoYW5kbGVyIFJlcXVlc3QgaGFuZGxlclxuICovXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gbGlzdGVuQW5kU2VydmVUTFMoXG4gIG9wdGlvbnM6IEhUVFBTT3B0aW9ucyxcbiAgaGFuZGxlcjogKHJlcTogU2VydmVyUmVxdWVzdCkgPT4gdm9pZCxcbik6IFByb21pc2U8dm9pZD4ge1xuICBjb25zdCBzZXJ2ZXIgPSBzZXJ2ZVRMUyhvcHRpb25zKTtcblxuICBmb3IgYXdhaXQgKGNvbnN0IHJlcXVlc3Qgb2Ygc2VydmVyKSB7XG4gICAgaGFuZGxlcihyZXF1ZXN0KTtcbiAgfVxufVxuXG4vKipcbiAqIEludGVyZmFjZSBvZiBIVFRQIHNlcnZlciByZXNwb25zZS5cbiAqIElmIGJvZHkgaXMgYSBSZWFkZXIsIHJlc3BvbnNlIHdvdWxkIGJlIGNodW5rZWQuXG4gKiBJZiBib2R5IGlzIGEgc3RyaW5nLCBpdCB3b3VsZCBiZSBVVEYtOCBlbmNvZGVkIGJ5IGRlZmF1bHQuXG4gKi9cbmV4cG9ydCBpbnRlcmZhY2UgUmVzcG9uc2Uge1xuICBzdGF0dXM/OiBudW1iZXI7XG4gIGhlYWRlcnM/OiBIZWFkZXJzO1xuICBib2R5PzogVWludDhBcnJheSB8IERlbm8uUmVhZGVyIHwgc3RyaW5nO1xuICB0cmFpbGVycz86ICgpID0+IFByb21pc2U8SGVhZGVycz4gfCBIZWFkZXJzO1xufVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLEVBQUEsd0VBQUE7U0FDQSxNQUFBLFNBQUEsbUJBQUE7U0FDQSxTQUFBLEVBQUEsU0FBQSxTQUFBLGNBQUE7U0FDQSxNQUFBLFNBQUEsa0JBQUE7U0FDQSxRQUFBLEVBQUEsZ0JBQUEsU0FBQSxlQUFBO1NBRUEsVUFBQSxFQUNBLGlCQUFBLEVBQ0EsV0FBQSxFQUNBLFdBQUEsRUFDQSxhQUFBLFNBQ0EsUUFBQTthQUVBLGFBQUE7QUFVQSxRQUFBLEdBQUEsUUFBQTtBQUVBLGtCQUFBLEdBQUEsU0FBQTtBQUNBLE1BR0EsQUFIQSw2SEFHQSxBQUhBLEVBR0EsS0FDQSxhQUFBO0FBQ0EsVUFBQSw0QkFBQTtBQUNBLFVBQUEsb0NBQUE7aUJBQ0EsY0FBQSxLQUFBLFNBQUE7a0JBQ0EsRUFBQSxRQUFBLE9BQUEsQ0FBQSxHQUFBLEVBQUEsY0FBQTtnQkFDQSxFQUFBO3FCQUNBLGNBQUEsR0FBQSxRQUFBLENBQUEsRUFBQTtBQUNBLGtCQUFBLDRDQUFBO29CQUNBLE1BQUEsQ0FBQSxLQUFBLE1BQUEsY0FBQTt5QkFDQSxjQUFBLEdBQUEsSUFBQTs7O3FCQUdBLGNBQUEsR0FBQSxJQUFBOzs7b0JBR0EsY0FBQTs7QUFHQSxTQUFBLEdBQUEsSUFBQTtBQUVBLE1BSUEsQUFKQSw0SUFJQSxBQUpBLEVBSUEsS0FDQSxJQUFBO2tCQUNBLEtBQUE7cUJBQ0EsYUFBQSxJQUFBLElBQUE7cUJBQ0EsS0FBQSxHQUFBLFVBQUEsTUFBQSxhQUFBLE9BQUEsQ0FBQTs7c0JBRUEsZ0JB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