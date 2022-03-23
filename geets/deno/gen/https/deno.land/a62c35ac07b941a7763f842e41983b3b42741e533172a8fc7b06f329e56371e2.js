// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { getLevelByName, LogLevels } from "./levels.ts";
import { red, yellow, blue, bold } from "../fmt/colors.ts";
import { existsSync, exists } from "../fs/exists.ts";
import { BufWriterSync } from "../io/bufio.ts";
const DEFAULT_FORMATTER = "{levelName} {msg}";
export class BaseHandler {
    constructor(levelName, options = {
    }){
        this.level = getLevelByName(levelName);
        this.levelName = levelName;
        this.formatter = options.formatter || DEFAULT_FORMATTER;
    }
    handle(logRecord) {
        if (this.level > logRecord.level) return;
        const msg = this.format(logRecord);
        return this.log(msg);
    }
    format(logRecord) {
        if (this.formatter instanceof Function) {
            return this.formatter(logRecord);
        }
        return this.formatter.replace(/{(\S+)}/g, (match, p1)=>{
            const value = logRecord[p1];
            // do not interpolate missing values
            if (value == null) {
                return match;
            }
            return String(value);
        });
    }
    log(_msg) {
    }
    async setup() {
    }
    async destroy() {
    }
}
export class ConsoleHandler extends BaseHandler {
    format(logRecord) {
        let msg = super.format(logRecord);
        switch(logRecord.level){
            case LogLevels.INFO:
                msg = blue(msg);
                break;
            case LogLevels.WARNING:
                msg = yellow(msg);
                break;
            case LogLevels.ERROR:
                msg = red(msg);
                break;
            case LogLevels.CRITICAL:
                msg = bold(red(msg));
                break;
            default:
                break;
        }
        return msg;
    }
    log(msg) {
        console.log(msg);
    }
}
export class WriterHandler extends BaseHandler {
    #encoder=new TextEncoder();
}
export class FileHandler extends WriterHandler {
    _encoder = new TextEncoder();
    #unloadCallback=()=>this.destroy()
    ;
    constructor(levelName, options){
        super(levelName, options);
        this._filename = options.filename;
        // default to append mode, write only
        this._mode = options.mode ? options.mode : "a";
        this._openOptions = {
            createNew: this._mode === "x",
            create: this._mode !== "x",
            append: this._mode === "a",
            truncate: this._mode !== "a",
            write: true
        };
    }
    async setup() {
        this._file = await Deno.open(this._filename, this._openOptions);
        this._writer = this._file;
        this._buf = new BufWriterSync(this._file);
        addEventListener("unload", this.#unloadCallback);
    }
    handle(logRecord) {
        super.handle(logRecord);
        // Immediately flush if log level is higher than ERROR
        if (logRecord.level > LogLevels.ERROR) {
            this.flush();
        }
    }
    log(msg) {
        this._buf.writeSync(this._encoder.encode(msg + "\n"));
    }
    flush() {
        if (this._buf?.buffered() > 0) {
            this._buf.flush();
        }
    }
    destroy() {
        this.flush();
        this._file?.close();
        this._file = undefined;
        removeEventListener("unload", this.#unloadCallback);
        return Promise.resolve();
    }
}
export class RotatingFileHandler extends FileHandler {
    #maxBytes;
    #maxBackupCount;
    #currentFileSize=0;
    constructor(levelName, options){
        super(levelName, options);
        this.#maxBytes = options.maxBytes;
        this.#maxBackupCount = options.maxBackupCount;
    }
    async setup() {
        if (this.#maxBytes < 1) {
            this.destroy();
            throw new Error("maxBytes cannot be less than 1");
        }
        if (this.#maxBackupCount < 1) {
            this.destroy();
            throw new Error("maxBackupCount cannot be less than 1");
        }
        await super.setup();
        if (this._mode === "w") {
            // Remove old backups too as it doesn't make sense to start with a clean
            // log file, but old backups
            for(let i = 1; i <= this.#maxBackupCount; i++){
                if (await exists(this._filename + "." + i)) {
                    await Deno.remove(this._filename + "." + i);
                }
            }
        } else if (this._mode === "x") {
            // Throw if any backups also exist
            for(let i = 1; i <= this.#maxBackupCount; i++){
                if (await exists(this._filename + "." + i)) {
                    this.destroy();
                    throw new Deno.errors.AlreadyExists("Backup log file " + this._filename + "." + i + " already exists");
                }
            }
        } else {
            this.#currentFileSize = (await Deno.stat(this._filename)).size;
        }
    }
    log(msg) {
        const msgByteLength = this._encoder.encode(msg).byteLength + 1;
        if (this.#currentFileSize + msgByteLength > this.#maxBytes) {
            this.rotateLogFiles();
            this.#currentFileSize = 0;
        }
        this._buf.writeSync(this._encoder.encode(msg + "\n"));
        this.#currentFileSize += msgByteLength;
    }
    rotateLogFiles() {
        this._buf.flush();
        Deno.close(this._file.rid);
        for(let i = this.#maxBackupCount - 1; i >= 0; i--){
            const source = this._filename + (i === 0 ? "" : "." + i);
            const dest = this._filename + "." + (i + 1);
            if (existsSync(source)) {
                Deno.renameSync(source, dest);
            }
        }
        this._file = Deno.openSync(this._filename, this._openOptions);
        this._writer = this._file;
        this._buf = new BufWriterSync(this._file);
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL2xvZy9oYW5kbGVycy50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cbmltcG9ydCB7IGdldExldmVsQnlOYW1lLCBMZXZlbE5hbWUsIExvZ0xldmVscyB9IGZyb20gXCIuL2xldmVscy50c1wiO1xuaW1wb3J0IHR5cGUgeyBMb2dSZWNvcmQgfSBmcm9tIFwiLi9sb2dnZXIudHNcIjtcbmltcG9ydCB7IHJlZCwgeWVsbG93LCBibHVlLCBib2xkIH0gZnJvbSBcIi4uL2ZtdC9jb2xvcnMudHNcIjtcbmltcG9ydCB7IGV4aXN0c1N5bmMsIGV4aXN0cyB9IGZyb20gXCIuLi9mcy9leGlzdHMudHNcIjtcbmltcG9ydCB7IEJ1ZldyaXRlclN5bmMgfSBmcm9tIFwiLi4vaW8vYnVmaW8udHNcIjtcblxuY29uc3QgREVGQVVMVF9GT1JNQVRURVIgPSBcIntsZXZlbE5hbWV9IHttc2d9XCI7XG50eXBlIEZvcm1hdHRlckZ1bmN0aW9uID0gKGxvZ1JlY29yZDogTG9nUmVjb3JkKSA9PiBzdHJpbmc7XG50eXBlIExvZ01vZGUgPSBcImFcIiB8IFwid1wiIHwgXCJ4XCI7XG5cbmludGVyZmFjZSBIYW5kbGVyT3B0aW9ucyB7XG4gIGZvcm1hdHRlcj86IHN0cmluZyB8IEZvcm1hdHRlckZ1bmN0aW9uO1xufVxuXG5leHBvcnQgY2xhc3MgQmFzZUhhbmRsZXIge1xuICBsZXZlbDogbnVtYmVyO1xuICBsZXZlbE5hbWU6IExldmVsTmFtZTtcbiAgZm9ybWF0dGVyOiBzdHJpbmcgfCBGb3JtYXR0ZXJGdW5jdGlvbjtcblxuICBjb25zdHJ1Y3RvcihsZXZlbE5hbWU6IExldmVsTmFtZSwgb3B0aW9uczogSGFuZGxlck9wdGlvbnMgPSB7fSkge1xuICAgIHRoaXMubGV2ZWwgPSBnZXRMZXZlbEJ5TmFtZShsZXZlbE5hbWUpO1xuICAgIHRoaXMubGV2ZWxOYW1lID0gbGV2ZWxOYW1lO1xuXG4gICAgdGhpcy5mb3JtYXR0ZXIgPSBvcHRpb25zLmZvcm1hdHRlciB8fCBERUZBVUxUX0ZPUk1BVFRFUjtcbiAgfVxuXG4gIGhhbmRsZShsb2dSZWNvcmQ6IExvZ1JlY29yZCk6IHZvaWQge1xuICAgIGlmICh0aGlzLmxldmVsID4gbG9nUmVjb3JkLmxldmVsKSByZXR1cm47XG5cbiAgICBjb25zdCBtc2cgPSB0aGlzLmZvcm1hdChsb2dSZWNvcmQpO1xuICAgIHJldHVybiB0aGlzLmxvZyhtc2cpO1xuICB9XG5cbiAgZm9ybWF0KGxvZ1JlY29yZDogTG9nUmVjb3JkKTogc3RyaW5nIHtcbiAgICBpZiAodGhpcy5mb3JtYXR0ZXIgaW5zdGFuY2VvZiBGdW5jdGlvbikge1xuICAgICAgcmV0dXJuIHRoaXMuZm9ybWF0dGVyKGxvZ1JlY29yZCk7XG4gICAgfVxuXG4gICAgcmV0dXJuIHRoaXMuZm9ybWF0dGVyLnJlcGxhY2UoL3soXFxTKyl9L2csIChtYXRjaCwgcDEpOiBzdHJpbmcgPT4ge1xuICAgICAgY29uc3QgdmFsdWUgPSBsb2dSZWNvcmRbcDEgYXMga2V5b2YgTG9nUmVjb3JkXTtcblxuICAgICAgLy8gZG8gbm90IGludGVycG9sYXRlIG1pc3NpbmcgdmFsdWVzXG4gICAgICBpZiAodmFsdWUgPT0gbnVsbCkge1xuICAgICAgICByZXR1cm4gbWF0Y2g7XG4gICAgICB9XG5cbiAgICAgIHJldHVybiBTdHJpbmcodmFsdWUpO1xuICAgIH0pO1xuICB9XG5cbiAgbG9nKF9tc2c6IHN0cmluZyk6IHZvaWQge31cbiAgYXN5bmMgc2V0dXAoKTogUHJvbWlzZTx2b2lkPiB7fVxuICBhc3luYyBkZXN0cm95KCk6IFByb21pc2U8dm9pZD4ge31cbn1cblxuZXhwb3J0IGNsYXNzIENvbnNvbGVIYW5kbGVyIGV4dGVuZHMgQmFzZUhhbmRsZXIge1xuICBmb3JtYXQobG9nUmVjb3JkOiBMb2dSZWNvcmQpOiBzdHJpbmcge1xuICAgIGxldCBtc2cgPSBzdXBlci5mb3JtYXQobG9nUmVjb3JkKTtcblxuICAgIHN3aXRjaCAobG9nUmVjb3JkLmxldmVsKSB7XG4gICAgICBjYXNlIExvZ0xldmVscy5JTkZPOlxuICAgICAgICBtc2cgPSBibHVlKG1zZyk7XG4gICAgICAgIGJyZWFrO1xuICAgICAgY2FzZSBMb2dMZXZlbHMuV0FSTklORzpcbiAgICAgICAgbXNnID0geWVsbG93KG1zZyk7XG4gICAgICAgIGJyZWFrO1xuICAgICAgY2FzZSBMb2dMZXZlbHMuRVJST1I6XG4gICAgICAgIG1zZyA9IHJlZChtc2cpO1xuICAgICAgICBicmVhaztcbiAgICAgIGNhc2UgTG9nTGV2ZWxzLkNSSVRJQ0FMOlxuICAgICAgICBtc2cgPSBib2xkKHJlZChtc2cpKTtcbiAgICAgICAgYnJlYWs7XG4gICAgICBkZWZhdWx0OlxuICAgICAgICBicmVhaztcbiAgICB9XG5cbiAgICByZXR1cm4gbXNnO1xuICB9XG5cbiAgbG9nKG1zZzogc3RyaW5nKTogdm9pZCB7XG4gICAgY29uc29sZS5sb2cobXNnKTtcbiAgfVxufVxuXG5leHBvcnQgYWJzdHJhY3QgY2xhc3MgV3JpdGVySGFuZGxlciBleHRlbmRzIEJhc2VIYW5kbGVyIHtcbiAgcHJvdGVjdGVkIF93cml0ZXIhOiBEZW5vLldyaXRlcjtcbiAgI2VuY29kZXIgPSBuZXcgVGV4dEVuY29kZXIoKTtcblxuICBhYnN0cmFjdCBsb2cobXNnOiBzdHJpbmcpOiB2b2lkO1xufVxuXG5pbnRlcmZhY2UgRmlsZUhhbmRsZXJPcHRpb25zIGV4dGVuZHMgSGFuZGxlck9wdGlvbnMge1xuICBmaWxlbmFtZTogc3RyaW5nO1xuICBtb2RlPzogTG9nTW9kZTtcbn1cblxuZXhwb3J0IGNsYXNzIEZpbGVIYW5kbGVyIGV4dGVuZHMgV3JpdGVySGFuZGxlciB7XG4gIHByb3RlY3RlZCBfZmlsZTogRGVuby5GaWxlIHwgdW5kZWZpbmVkO1xuICBwcm90ZWN0ZWQgX2J1ZiE6IEJ1ZldyaXRlclN5bmM7XG4gIHByb3RlY3RlZCBfZmlsZW5hbWU6IHN0cmluZztcbiAgcHJvdGVjdGVkIF9tb2RlOiBMb2dNb2RlO1xuICBwcm90ZWN0ZWQgX29wZW5PcHRpb25zOiBEZW5vLk9wZW5PcHRpb25zO1xuICBwcm90ZWN0ZWQgX2VuY29kZXIgPSBuZXcgVGV4dEVuY29kZXIoKTtcbiAgI3VubG9hZENhbGxiYWNrID0gKCk6IFByb21pc2U8dm9pZD4gPT4gdGhpcy5kZXN0cm95KCk7XG5cbiAgY29uc3RydWN0b3IobGV2ZWxOYW1lOiBMZXZlbE5hbWUsIG9wdGlvbnM6IEZpbGVIYW5kbGVyT3B0aW9ucykge1xuICAgIHN1cGVyKGxldmVsTmFtZSwgb3B0aW9ucyk7XG4gICAgdGhpcy5fZmlsZW5hbWUgPSBvcHRpb25zLmZpbGVuYW1lO1xuICAgIC8vIGRlZmF1bHQgdG8gYXBwZW5kIG1vZGUsIHdyaXRlIG9ubHlcbiAgICB0aGlzLl9tb2RlID0gb3B0aW9ucy5tb2RlID8gb3B0aW9ucy5tb2RlIDogXCJhXCI7XG4gICAgdGhpcy5fb3Blbk9wdGlvbnMgPSB7XG4gICAgICBjcmVhdGVOZXc6IHRoaXMuX21vZGUgPT09IFwieFwiLFxuICAgICAgY3JlYXRlOiB0aGlzLl9tb2RlICE9PSBcInhcIixcbiAgICAgIGFwcGVuZDogdGhpcy5fbW9kZSA9PT0gXCJhXCIsXG4gICAgICB0cnVuY2F0ZTogdGhpcy5fbW9kZSAhPT0gXCJhXCIsXG4gICAgICB3cml0ZTogdHJ1ZSxcbiAgICB9O1xuICB9XG5cbiAgYXN5bmMgc2V0dXAoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdGhpcy5fZmlsZSA9IGF3YWl0IERlbm8ub3Blbih0aGlzLl9maWxlbmFtZSwgdGhpcy5fb3Blbk9wdGlvbnMpO1xuICAgIHRoaXMuX3dyaXRlciA9IHRoaXMuX2ZpbGU7XG4gICAgdGhpcy5fYnVmID0gbmV3IEJ1ZldyaXRlclN5bmModGhpcy5fZmlsZSk7XG5cbiAgICBhZGRFdmVudExpc3RlbmVyKFwidW5sb2FkXCIsIHRoaXMuI3VubG9hZENhbGxiYWNrKTtcbiAgfVxuXG4gIGhhbmRsZShsb2dSZWNvcmQ6IExvZ1JlY29yZCk6IHZvaWQge1xuICAgIHN1cGVyLmhhbmRsZShsb2dSZWNvcmQpO1xuXG4gICAgLy8gSW1tZWRpYXRlbHkgZmx1c2ggaWYgbG9nIGxldmVsIGlzIGhpZ2hlciB0aGFuIEVSUk9SXG4gICAgaWYgKGxvZ1JlY29yZC5sZXZlbCA+IExvZ0xldmVscy5FUlJPUikge1xuICAgICAgdGhpcy5mbHVzaCgpO1xuICAgIH1cbiAgfVxuXG4gIGxvZyhtc2c6IHN0cmluZyk6IHZvaWQge1xuICAgIHRoaXMuX2J1Zi53cml0ZVN5bmModGhpcy5fZW5jb2Rlci5lbmNvZGUobXNnICsgXCJcXG5cIikpO1xuICB9XG5cbiAgZmx1c2goKTogdm9pZCB7XG4gICAgaWYgKHRoaXMuX2J1Zj8uYnVmZmVyZWQoKSA+IDApIHtcbiAgICAgIHRoaXMuX2J1Zi5mbHVzaCgpO1xuICAgIH1cbiAgfVxuXG4gIGRlc3Ryb3koKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgdGhpcy5mbHVzaCgpO1xuICAgIHRoaXMuX2ZpbGU/LmNsb3NlKCk7XG4gICAgdGhpcy5fZmlsZSA9IHVuZGVmaW5lZDtcbiAgICByZW1vdmVFdmVudExpc3RlbmVyKFwidW5sb2FkXCIsIHRoaXMuI3VubG9hZENhbGxiYWNrKTtcbiAgICByZXR1cm4gUHJvbWlzZS5yZXNvbHZlKCk7XG4gIH1cbn1cblxuaW50ZXJmYWNlIFJvdGF0aW5nRmlsZUhhbmRsZXJPcHRpb25zIGV4dGVuZHMgRmlsZUhhbmRsZXJPcHRpb25zIHtcbiAgbWF4Qnl0ZXM6IG51bWJlcjtcbiAgbWF4QmFja3VwQ291bnQ6IG51bWJlcjtcbn1cblxuZXhwb3J0IGNsYXNzIFJvdGF0aW5nRmlsZUhhbmRsZXIgZXh0ZW5kcyBGaWxlSGFuZGxlciB7XG4gICNtYXhCeXRlczogbnVtYmVyO1xuICAjbWF4QmFja3VwQ291bnQ6IG51bWJlcjtcbiAgI2N1cnJlbnRGaWxlU2l6ZSA9IDA7XG5cbiAgY29uc3RydWN0b3IobGV2ZWxOYW1lOiBMZXZlbE5hbWUsIG9wdGlvbnM6IFJvdGF0aW5nRmlsZUhhbmRsZXJPcHRpb25zKSB7XG4gICAgc3VwZXIobGV2ZWxOYW1lLCBvcHRpb25zKTtcbiAgICB0aGlzLiNtYXhCeXRlcyA9IG9wdGlvbnMubWF4Qnl0ZXM7XG4gICAgdGhpcy4jbWF4QmFja3VwQ291bnQgPSBvcHRpb25zLm1heEJhY2t1cENvdW50O1xuICB9XG5cbiAgYXN5bmMgc2V0dXAoKTogUHJvbWlzZTx2b2lkPiB7XG4gICAgaWYgKHRoaXMuI21heEJ5dGVzIDwgMSkge1xuICAgICAgdGhpcy5kZXN0cm95KCk7XG4gICAgICB0aHJvdyBuZXcgRXJyb3IoXCJtYXhCeXRlcyBjYW5ub3QgYmUgbGVzcyB0aGFuIDFcIik7XG4gICAgfVxuICAgIGlmICh0aGlzLiNtYXhCYWNrdXBDb3VudCA8IDEpIHtcbiAgICAgIHRoaXMuZGVzdHJveSgpO1xuICAgICAgdGhyb3cgbmV3IEVycm9yKFwibWF4QmFja3VwQ291bnQgY2Fubm90IGJlIGxlc3MgdGhhbiAxXCIpO1xuICAgIH1cbiAgICBhd2FpdCBzdXBlci5zZXR1cCgpO1xuXG4gICAgaWYgKHRoaXMuX21vZGUgPT09IFwid1wiKSB7XG4gICAgICAvLyBSZW1vdmUgb2xkIGJhY2t1cHMgdG9vIGFzIGl0IGRvZXNuJ3QgbWFrZSBzZW5zZSB0byBzdGFydCB3aXRoIGEgY2xlYW5cbiAgICAgIC8vIGxvZyBmaWxlLCBidXQgb2xkIGJhY2t1cHNcbiAgICAgIGZvciAobGV0IGkgPSAxOyBpIDw9IHRoaXMuI21heEJhY2t1cENvdW50OyBpKyspIHtcbiAgICAgICAgaWYgKGF3YWl0IGV4aXN0cyh0aGlzLl9maWxlbmFtZSArIFwiLlwiICsgaSkpIHtcbiAgICAgICAgICBhd2FpdCBEZW5vLnJlbW92ZSh0aGlzLl9maWxlbmFtZSArIFwiLlwiICsgaSk7XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9IGVsc2UgaWYgKHRoaXMuX21vZGUgPT09IFwieFwiKSB7XG4gICAgICAvLyBUaHJvdyBpZiBhbnkgYmFja3VwcyBhbHNvIGV4aXN0XG4gICAgICBmb3IgKGxldCBpID0gMTsgaSA8PSB0aGlzLiNtYXhCYWNrdXBDb3VudDsgaSsrKSB7XG4gICAgICAgIGlmIChhd2FpdCBleGlzdHModGhpcy5fZmlsZW5hbWUgKyBcIi5cIiArIGkpKSB7XG4gICAgICAgICAgdGhpcy5kZXN0cm95KCk7XG4gICAgICAgICAgdGhyb3cgbmV3IERlbm8uZXJyb3JzLkFscmVhZHlFeGlzdHMoXG4gICAgICAgICAgICBcIkJhY2t1cCBsb2cgZmlsZSBcIiArIHRoaXMuX2ZpbGVuYW1lICsgXCIuXCIgKyBpICsgXCIgYWxyZWFkeSBleGlzdHNcIixcbiAgICAgICAgICApO1xuICAgICAgICB9XG4gICAgICB9XG4gICAgfSBlbHNlIHtcbiAgICAgIHRoaXMuI2N1cnJlbnRGaWxlU2l6ZSA9IChhd2FpdCBEZW5vLnN0YXQodGhpcy5fZmlsZW5hbWUpKS5zaXplO1xuICAgIH1cbiAgfVxuXG4gIGxvZyhtc2c6IHN0cmluZyk6IHZvaWQge1xuICAgIGNvbnN0IG1zZ0J5dGVMZW5ndGggPSB0aGlzLl9lbmNvZGVyLmVuY29kZShtc2cpLmJ5dGVMZW5ndGggKyAxO1xuXG4gICAgaWYgKHRoaXMuI2N1cnJlbnRGaWxlU2l6ZSArIG1zZ0J5dGVMZW5ndGggPiB0aGlzLiNtYXhCeXRlcykge1xuICAgICAgdGhpcy5yb3RhdGVMb2dGaWxlcygpO1xuICAgICAgdGhpcy4jY3VycmVudEZpbGVTaXplID0gMDtcbiAgICB9XG5cbiAgICB0aGlzLl9idWYud3JpdGVTeW5jKHRoaXMuX2VuY29kZXIuZW5jb2RlKG1zZyArIFwiXFxuXCIpKTtcbiAgICB0aGlzLiNjdXJyZW50RmlsZVNpemUgKz0gbXNnQnl0ZUxlbmd0aDtcbiAgfVxuXG4gIHJvdGF0ZUxvZ0ZpbGVzKCk6IHZvaWQge1xuICAgIHRoaXMuX2J1Zi5mbHVzaCgpO1xuICAgIERlbm8uY2xvc2UodGhpcy5fZmlsZSEucmlkKTtcblxuICAgIGZvciAobGV0IGkgPSB0aGlzLiNtYXhCYWNrdXBDb3VudCAtIDE7IGkgPj0gMDsgaS0tKSB7XG4gICAgICBjb25zdCBzb3VyY2UgPSB0aGlzLl9maWxlbmFtZSArIChpID09PSAwID8gXCJcIiA6IFwiLlwiICsgaSk7XG4gICAgICBjb25zdCBkZXN0ID0gdGhpcy5fZmlsZW5hbWUgKyBcIi5cIiArIChpICsgMSk7XG5cbiAgICAgIGlmIChleGlzdHNTeW5jKHNvdXJjZSkpIHtcbiAgICAgICAgRGVuby5yZW5hbWVTeW5jKHNvdXJjZSwgZGVzdCk7XG4gICAgICB9XG4gICAgfVxuXG4gICAgdGhpcy5fZmlsZSA9IERlbm8ub3BlblN5bmModGhpcy5fZmlsZW5hbWUsIHRoaXMuX29wZW5PcHRpb25zKTtcbiAgICB0aGlzLl93cml0ZXIgPSB0aGlzLl9maWxlO1xuICAgIHRoaXMuX2J1ZiA9IG5ldyBCdWZXcml0ZXJTeW5jKHRoaXMuX2ZpbGUpO1xuICB9XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtTQUNBLGNBQUEsRUFBQSxTQUFBLFNBQUEsV0FBQTtTQUVBLEdBQUEsRUFBQSxNQUFBLEVBQUEsSUFBQSxFQUFBLElBQUEsU0FBQSxnQkFBQTtTQUNBLFVBQUEsRUFBQSxNQUFBLFNBQUEsZUFBQTtTQUNBLGFBQUEsU0FBQSxjQUFBO01BRUEsaUJBQUEsSUFBQSxpQkFBQTthQVFBLFdBQUE7Z0JBS0EsU0FBQSxFQUFBLE9BQUE7O2FBQ0EsS0FBQSxHQUFBLGNBQUEsQ0FBQSxTQUFBO2FBQ0EsU0FBQSxHQUFBLFNBQUE7YUFFQSxTQUFBLEdBQUEsT0FBQSxDQUFBLFNBQUEsSUFBQSxpQkFBQTs7QUFHQSxVQUFBLENBQUEsU0FBQTtpQkFDQSxLQUFBLEdBQUEsU0FBQSxDQUFBLEtBQUE7Y0FFQSxHQUFBLFFBQUEsTUFBQSxDQUFBLFNBQUE7b0JBQ0EsR0FBQSxDQUFBLEdBQUE7O0FBR0EsVUFBQSxDQUFBLFNBQUE7aUJBQ0EsU0FBQSxZQUFBLFFBQUE7d0JBQ0EsU0FBQSxDQUFBLFNBQUE7O29CQUdBLFNBQUEsQ0FBQSxPQUFBLGNBQUEsS0FBQSxFQUFBLEVBQUE7a0JBQ0EsS0FBQSxHQUFBLFNBQUEsQ0FBQSxFQUFBO0FBRUEsY0FBQSxrQ0FBQTtnQkFDQSxLQUFBLElBQUEsSUFBQTt1QkFDQSxLQUFBOzttQkFHQSxNQUFBLENBQUEsS0FBQTs7O0FBSUEsT0FBQSxDQUFBLElBQUE7O1VBQ0EsS0FBQTs7VUFDQSxPQUFBOzs7YUFHQSxjQUFBLFNBQUEsV0FBQTtBQUNBLFVBQUEsQ0FBQSxTQUFBO1lBQ0EsR0FBQSxHQUFBLEtBQUEsQ0FBQSxNQUFBLENBQUEsU0FBQTtlQUVBLFNBQUEsQ0FBQSxLQUFBO2lCQUNBLFNBQUEsQ0FBQSxJQUFBO0FBQ0EsbUJBQUEsR0FBQSxJQUFBLENBQUEsR0FBQTs7aUJBRUEsU0FBQSxDQUFBLE9BQUE7QUFDQSxtQkFBQSxHQUFBLE1BQUEsQ0FBQSxHQUFBOztpQkFFQSxTQUFBLENBQUEsS0FBQTtBQUNBLG1CQUFBLEdBQUEsR0FBQSxDQUFBLEdBQUE7O2lCQUVBLFNBQUEsQ0FBQSxRQUFBO0FBQ0EsbUJBQUEsR0FBQSxJQUFBLENBQUEsR0FBQSxDQUFBLEdBQUE7Ozs7O2VBTUEsR0FBQTs7QUFHQSxPQUFBLENBQUEsR0FBQTtBQUNBLGVBQUEsQ0FBQSxHQUFBLENBQUEsR0FBQTs7O2FBSUEsYUFBQSxTQUFBLFdBQUE7S0FFQSxPQUFBLEtBQUEsV0FBQTs7YUFVQSxXQUFBLFNBQUEsYUFBQTtBQU1BLFlBQUEsT0FBQSxXQUFBO0tBQ0EsY0FBQSxVQUFBLE9BQUE7O2dCQUVBLFNBQUEsRUFBQSxPQUFBO0FBQ0EsYUFBQSxDQUFBLFNBQUEsRUFBQSxPQUFBO2FBQ0EsU0FBQSxHQUFBLE9BQUEsQ0FBQSxRQUFBO0FBQ0EsVUFBQSxtQ0FBQTthQUNBLEtBQUEsR0FBQSxPQUFBLENBQUEsSUFBQSxHQUFBLE9BQUEsQ0FBQSxJQUFBLElBQUEsQ0FBQTthQUNBLFlBQUE7QUFDQSxxQkFBQSxPQUFBLEtBQUEsTUFBQSxDQUFBO0FBQ0Esa0JBQUEsT0FBQSxLQUFBLE1BQUEsQ0FBQTtBQUNBLGtCQUFBLE9BQUEsS0FBQSxNQUFBLENBQUE7QUFDQSxvQkFBQSxPQUFBLEtBQUEsTUFBQSxDQUFBO0FBQ0EsaUJBQUEsRUFBQSxJQUFBOzs7VUFJQSxLQUFBO2FBQ0EsS0FBQSxTQUFBLElBQUEsQ0FBQSxJQUFBLE1BQUEsU0FBQSxPQUFBLFlBQUE7YUFDQSxPQUFBLFFBQUEsS0FBQTthQUNBLElBQUEsT0FBQSxhQUFBLE1BQUEsS0FBQTtBQUVBLHdCQUFBLEVBQUEsTUFBQSxTQUFBLGNBQUE7O0FBR0EsVUFBQSxDQUFBLFNBQUE7QUFDQSxhQUFBLENBQUEsTUFBQSxDQUFBLFNBQUE7QUFFQSxVQUFBLG9EQUFBO1lBQ0EsU0FBQSxDQUFBLEtBQUEsR0FBQSxTQUFBLENBQUEsS0FBQTtpQkFDQSxLQUFBOzs7QUFJQSxPQUFBLENBQUEsR0FBQTthQUNBLElBQUEsQ0FBQSxTQUFBLE1BQUEsUUFBQSxDQUFBLE1BQUEsQ0FBQSxHQUFBLElBQUEsRUFBQTs7QUFHQSxTQUFBO2lCQUNBLElBQUEsRUFBQSxRQUFBLEtBQUEsQ0FBQTtpQkFDQSxJQUFBLENBQUEsS0FBQTs7O0FBSUEsV0FBQTthQUNBLEtBQUE7YUFDQSxLQUFBLEVBQUEsS0FBQTthQUNBLEtBQUEsR0FBQSxTQUFBO0FBQ0EsMkJBQUEsRUFBQSxNQUFBLFNBQUEsY0FBQTtlQUNBLE9BQUEsQ0FBQSxPQUFBOzs7YUFTQSxtQkFBQSxTQUFBLFdBQUE7S0FDQSxRQUFBO0tBQ0EsY0FBQTtLQUNBLGVBQUEsQ0FBQSxDQUFBO2dCQUVBLFNBQUEsRUFBQSxPQUFBO0FBQ0EsYUFBQSxDQUFBLFNBQUEsRUFBQSxPQUFBO2NBQ0EsUUFBQSxHQUFBLE9BQUEsQ0FBQSxRQUFBO2NBQ0EsY0FBQSxHQUFBLE9BQUEsQ0FBQSxjQUFBOztVQUdBLEtBQUE7a0JBQ0EsUUFBQSxHQUFBLENBQUE7aUJBQ0EsT0FBQTtzQkFDQSxLQUFBLEVBQUEsOEJBQUE7O2tCQUVBLGNBQUEsR0FBQSxDQUFBO2lCQUNBLE9BQUE7c0JBQ0EsS0FBQSxFQUFBLG9DQUFBOztjQUVBLEtBQUEsQ0FBQSxLQUFBO2lCQUVBLEtBQUEsTUFBQSxDQUFBO0FBQ0EsY0FBQSxzRUFBQTtBQUNBLGNBQUEsMEJBQUE7b0JBQ0EsQ0FBQSxHQUFBLENBQUEsRUFBQSxDQUFBLFVBQUEsY0FBQSxFQUFBLENBQUE7MEJBQ0EsTUFBQSxNQUFBLFNBQUEsSUFBQSxDQUFBLElBQUEsQ0FBQTswQkFDQSxJQUFBLENBQUEsTUFBQSxNQUFBLFNBQUEsSUFBQSxDQUFBLElBQUEsQ0FBQTs7O3dCQUdBLEtBQUEsTUFBQSxDQUFBO0FBQ0EsY0FBQSxnQ0FBQTtvQkFDQSxDQUFBLEdBQUEsQ0FBQSxFQUFBLENBQUEsVUFBQSxjQUFBLEVBQUEsQ0FBQTswQkFDQSxNQUFBLE1BQUEsU0FBQSxJQUFBLENBQUEsSUFBQSxDQUFBO3lCQUNBLE9BQUE7OEJBQ0EsSUFBQSxDQUFBLE1BQUEsQ0FBQSxhQUFBLEVBQ0EsZ0JBQUEsU0FBQSxTQUFBLElBQUEsQ0FBQSxJQUFBLENBQUEsSUFBQSxlQUFBOzs7O2tCQUtBLGVBQUEsVUFBQSxJQUFBLENBQUEsSUFBQSxNQUFBLFNBQUEsR0FBQSxJQUFBOzs7QUFJQSxPQUFBLENBQUEsR0FBQTtjQUNBLGFBQUEsUUFBQSxRQUFBLENBQUEsTUFBQSxDQUFBLEdBQUEsRUFBQSxVQUFBLEdBQUEsQ0FBQTtrQkFFQSxlQUFBLEdBQUEsYUFBQSxTQUFBLFFBQUE7aUJBQ0EsY0FBQTtrQkFDQSxlQUFBLEdBQUEsQ0FBQTs7YUFHQSxJQUFBLENBQUEsU0FBQSxNQUFBLFFBQUEsQ0FBQSxNQUFBLENBQUEsR0FBQSxJQUFBLEVBQUE7Y0FDQSxlQUFBLElBQUEsYUFBQTs7QUFHQSxrQkFBQTthQUNBLElBQUEsQ0FBQSxLQUFBO0FBQ0EsWUFBQSxDQUFBLEtBQUEsTUFBQSxLQUFBLENBQUEsR0FBQTtnQkFFQSxDQUFBLFNBQUEsY0FBQSxHQUFBLENBQUEsRUFBQSxDQUFBLElBQUEsQ0FBQSxFQUFBLENBQUE7a0JBQ0EsTUFBQSxRQUFBLFNBQUEsSUFBQSxDQUFBLEtBQUEsQ0FBQSxTQUFBLENBQUEsSUFBQSxDQUFBO2tCQUNBLElBQUEsUUFBQSxTQUFBLElBQUEsQ0FBQSxLQUFBLENBQUEsR0FBQSxDQUFBO2dCQUVBLFVBQUEsQ0FBQSxNQUFBO0FBQ0Esb0JBQUEsQ0FBQSxVQUFBLENBQUEsTUFBQSxFQUFBLElBQUE7OzthQUlBLEtBQUEsR0FBQSxJQUFBLENBQUEsUUFBQSxNQUFBLFNBQUEsT0FBQSxZQUFBO2FBQ0EsT0FBQSxRQUFBLEtBQUE7YUFDQSxJQUFBLE9BQUEsYUFBQSxNQUFBLEtBQUEifQ==