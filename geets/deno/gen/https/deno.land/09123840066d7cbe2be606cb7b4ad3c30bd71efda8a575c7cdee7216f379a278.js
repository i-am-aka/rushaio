// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import init, { source, create_hash as createHash, update_hash as updateHash, digest_hash as digestHash } from "./wasm.js";
import * as hex from "../../encoding/hex.ts";
import * as base64 from "../../encoding/base64.ts";
await init(source);
const TYPE_ERROR_MSG = "hash: `data` is invalid type";
export class Hash {
    #hash;
    #digested;
    constructor(algorithm){
        this.#hash = createHash(algorithm);
        this.#digested = false;
    }
    /**
   * Update internal state
   * @param data data to update
   */ update(data) {
        let msg;
        if (typeof data === "string") {
            msg = new TextEncoder().encode(data);
        } else if (typeof data === "object") {
            if (data instanceof ArrayBuffer || ArrayBuffer.isView(data)) {
                msg = new Uint8Array(data);
            } else {
                throw new Error(TYPE_ERROR_MSG);
            }
        } else {
            throw new Error(TYPE_ERROR_MSG);
        }
        updateHash(this.#hash, msg);
        return this;
    }
    /** Returns final hash */ digest() {
        if (this.#digested) throw new Error("hash: already digested");
        this.#digested = true;
        return digestHash(this.#hash);
    }
    /**
   * Returns hash as a string of given format
   * @param format format of output string (hex or base64). Default is hex
   */ toString(format = "hex") {
        const finalized = new Uint8Array(this.digest());
        switch(format){
            case "hex":
                return hex.encodeToString(finalized);
            case "base64":
                return base64.encode(finalized);
            default:
                throw new Error("hash: invalid format");
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL2hhc2gvX3dhc20vaGFzaC50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cblxuaW1wb3J0IGluaXQsIHtcbiAgc291cmNlLFxuICBjcmVhdGVfaGFzaCBhcyBjcmVhdGVIYXNoLFxuICB1cGRhdGVfaGFzaCBhcyB1cGRhdGVIYXNoLFxuICBkaWdlc3RfaGFzaCBhcyBkaWdlc3RIYXNoLFxuICBEZW5vSGFzaCxcbn0gZnJvbSBcIi4vd2FzbS5qc1wiO1xuXG5pbXBvcnQgKiBhcyBoZXggZnJvbSBcIi4uLy4uL2VuY29kaW5nL2hleC50c1wiO1xuaW1wb3J0ICogYXMgYmFzZTY0IGZyb20gXCIuLi8uLi9lbmNvZGluZy9iYXNlNjQudHNcIjtcbmltcG9ydCB0eXBlIHsgSGFzaGVyLCBNZXNzYWdlLCBPdXRwdXRGb3JtYXQgfSBmcm9tIFwiLi4vaGFzaGVyLnRzXCI7XG5cbmF3YWl0IGluaXQoc291cmNlKTtcblxuY29uc3QgVFlQRV9FUlJPUl9NU0cgPSBcImhhc2g6IGBkYXRhYCBpcyBpbnZhbGlkIHR5cGVcIjtcblxuZXhwb3J0IGNsYXNzIEhhc2ggaW1wbGVtZW50cyBIYXNoZXIge1xuICAjaGFzaDogRGVub0hhc2g7XG4gICNkaWdlc3RlZDogYm9vbGVhbjtcblxuICBjb25zdHJ1Y3RvcihhbGdvcml0aG06IHN0cmluZykge1xuICAgIHRoaXMuI2hhc2ggPSBjcmVhdGVIYXNoKGFsZ29yaXRobSk7XG4gICAgdGhpcy4jZGlnZXN0ZWQgPSBmYWxzZTtcbiAgfVxuXG4gIC8qKlxuICAgKiBVcGRhdGUgaW50ZXJuYWwgc3RhdGVcbiAgICogQHBhcmFtIGRhdGEgZGF0YSB0byB1cGRhdGVcbiAgICovXG4gIHVwZGF0ZShkYXRhOiBNZXNzYWdlKTogdGhpcyB7XG4gICAgbGV0IG1zZzogVWludDhBcnJheTtcblxuICAgIGlmICh0eXBlb2YgZGF0YSA9PT0gXCJzdHJpbmdcIikge1xuICAgICAgbXNnID0gbmV3IFRleHRFbmNvZGVyKCkuZW5jb2RlKGRhdGEgYXMgc3RyaW5nKTtcbiAgICB9IGVsc2UgaWYgKHR5cGVvZiBkYXRhID09PSBcIm9iamVjdFwiKSB7XG4gICAgICBpZiAoZGF0YSBpbnN0YW5jZW9mIEFycmF5QnVmZmVyIHx8IEFycmF5QnVmZmVyLmlzVmlldyhkYXRhKSkge1xuICAgICAgICBtc2cgPSBuZXcgVWludDhBcnJheShkYXRhKTtcbiAgICAgIH0gZWxzZSB7XG4gICAgICAgIHRocm93IG5ldyBFcnJvcihUWVBFX0VSUk9SX01TRyk7XG4gICAgICB9XG4gICAgfSBlbHNlIHtcbiAgICAgIHRocm93IG5ldyBFcnJvcihUWVBFX0VSUk9SX01TRyk7XG4gICAgfVxuXG4gICAgdXBkYXRlSGFzaCh0aGlzLiNoYXNoLCBtc2cpO1xuXG4gICAgcmV0dXJuIHRoaXM7XG4gIH1cblxuICAvKiogUmV0dXJucyBmaW5hbCBoYXNoICovXG4gIGRpZ2VzdCgpOiBBcnJheUJ1ZmZlciB7XG4gICAgaWYgKHRoaXMuI2RpZ2VzdGVkKSB0aHJvdyBuZXcgRXJyb3IoXCJoYXNoOiBhbHJlYWR5IGRpZ2VzdGVkXCIpO1xuXG4gICAgdGhpcy4jZGlnZXN0ZWQgPSB0cnVlO1xuICAgIHJldHVybiBkaWdlc3RIYXNoKHRoaXMuI2hhc2gpO1xuICB9XG5cbiAgLyoqXG4gICAqIFJldHVybnMgaGFzaCBhcyBhIHN0cmluZyBvZiBnaXZlbiBmb3JtYXRcbiAgICogQHBhcmFtIGZvcm1hdCBmb3JtYXQgb2Ygb3V0cHV0IHN0cmluZyAoaGV4IG9yIGJhc2U2NCkuIERlZmF1bHQgaXMgaGV4XG4gICAqL1xuICB0b1N0cmluZyhmb3JtYXQ6IE91dHB1dEZvcm1hdCA9IFwiaGV4XCIpOiBzdHJpbmcge1xuICAgIGNvbnN0IGZpbmFsaXplZCA9IG5ldyBVaW50OEFycmF5KHRoaXMuZGlnZXN0KCkpO1xuXG4gICAgc3dpdGNoIChmb3JtYXQpIHtcbiAgICAgIGNhc2UgXCJoZXhcIjpcbiAgICAgICAgcmV0dXJuIGhleC5lbmNvZGVUb1N0cmluZyhmaW5hbGl6ZWQpO1xuICAgICAgY2FzZSBcImJhc2U2NFwiOlxuICAgICAgICByZXR1cm4gYmFzZTY0LmVuY29kZShmaW5hbGl6ZWQpO1xuICAgICAgZGVmYXVsdDpcbiAgICAgICAgdGhyb3cgbmV3IEVycm9yKFwiaGFzaDogaW52YWxpZCBmb3JtYXRcIik7XG4gICAgfVxuICB9XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtPQUVBLElBQUEsSUFDQSxNQUFBLEVBQ0EsV0FBQSxJQUFBLFVBQUEsRUFDQSxXQUFBLElBQUEsVUFBQSxFQUNBLFdBQUEsSUFBQSxVQUFBLFNBRUEsU0FBQTtZQUVBLEdBQUEsT0FBQSxxQkFBQTtZQUNBLE1BQUEsT0FBQSx3QkFBQTtNQUdBLElBQUEsQ0FBQSxNQUFBO01BRUEsY0FBQSxJQUFBLDRCQUFBO2FBRUEsSUFBQTtLQUNBLElBQUE7S0FDQSxRQUFBO2dCQUVBLFNBQUE7Y0FDQSxJQUFBLEdBQUEsVUFBQSxDQUFBLFNBQUE7Y0FDQSxRQUFBLEdBQUEsS0FBQTs7QUFHQSxNQUdBLEFBSEEsZ0VBR0EsQUFIQSxFQUdBLENBQ0EsTUFBQSxDQUFBLElBQUE7WUFDQSxHQUFBO21CQUVBLElBQUEsTUFBQSxNQUFBO0FBQ0EsZUFBQSxPQUFBLFdBQUEsR0FBQSxNQUFBLENBQUEsSUFBQTswQkFDQSxJQUFBLE1BQUEsTUFBQTtnQkFDQSxJQUFBLFlBQUEsV0FBQSxJQUFBLFdBQUEsQ0FBQSxNQUFBLENBQUEsSUFBQTtBQUNBLG1CQUFBLE9BQUEsVUFBQSxDQUFBLElBQUE7OzBCQUVBLEtBQUEsQ0FBQSxjQUFBOzs7c0JBR0EsS0FBQSxDQUFBLGNBQUE7O0FBR0Esa0JBQUEsT0FBQSxJQUFBLEVBQUEsR0FBQTs7O0FBS0EsTUFBQSxxQkFBQSxFQUFBLENBQ0EsTUFBQTtrQkFDQSxRQUFBLFlBQUEsS0FBQSxFQUFBLHNCQUFBO2NBRUEsUUFBQSxHQUFBLElBQUE7ZUFDQSxVQUFBLE9BQUEsSUFBQTs7QUFHQSxNQUdBLEFBSEEsOEhBR0EsQUFIQSxFQUdBLENBQ0EsUUFBQSxDQUFBLE1BQUEsSUFBQSxHQUFBO2NBQ0EsU0FBQSxPQUFBLFVBQUEsTUFBQSxNQUFBO2VBRUEsTUFBQTtrQkFDQSxHQUFBO3VCQUNBLEdBQUEsQ0FBQSxjQUFBLENBQUEsU0FBQTtrQkFDQSxNQUFBO3VCQUNBLE1BQUEsQ0FBQSxNQUFBLENBQUEsU0FBQTs7MEJBRUEsS0FBQSxFQUFBLG9CQUFBIn0=