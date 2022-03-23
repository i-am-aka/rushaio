// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { LogLevels, getLevelByName, getLevelName } from "./levels.ts";
export class LogRecord {
    #args;
    #datetime;
    constructor(options){
        this.msg = options.msg;
        this.#args = [
            ...options.args
        ];
        this.level = options.level;
        this.loggerName = options.loggerName;
        this.#datetime = new Date();
        this.levelName = getLevelName(options.level);
    }
    get args() {
        return [
            ...this.#args
        ];
    }
    get datetime() {
        return new Date(this.#datetime.getTime());
    }
}
export class Logger {
    #level;
    #handlers;
    #loggerName;
    constructor(loggerName, levelName, options = {
    }){
        this.#loggerName = loggerName;
        this.#level = getLevelByName(levelName);
        this.#handlers = options.handlers || [];
    }
    get level() {
        return this.#level;
    }
    set level(level) {
        this.#level = level;
    }
    get levelName() {
        return getLevelName(this.#level);
    }
    set levelName(levelName) {
        this.#level = getLevelByName(levelName);
    }
    get loggerName() {
        return this.#loggerName;
    }
    set handlers(hndls) {
        this.#handlers = hndls;
    }
    get handlers() {
        return this.#handlers;
    }
    /** If the level of the logger is greater than the level to log, then nothing
   * is logged, otherwise a log record is passed to each log handler.  `msg` data
   * passed in is returned.  If a function is passed in, it is only evaluated
   * if the msg will be logged and the return value will be the result of the
   * function, not the function itself, unless the function isn't called, in which
   * case undefined is returned.  All types are coerced to strings for logging.
   */ _log(level, msg, ...args) {
        if (this.level > level) {
            return msg instanceof Function ? undefined : msg;
        }
        let fnResult;
        let logMessage;
        if (msg instanceof Function) {
            fnResult = msg();
            logMessage = this.asString(fnResult);
        } else {
            logMessage = this.asString(msg);
        }
        const record = new LogRecord({
            msg: logMessage,
            args: args,
            level: level,
            loggerName: this.loggerName
        });
        this.#handlers.forEach((handler)=>{
            handler.handle(record);
        });
        return msg instanceof Function ? fnResult : msg;
    }
    asString(data) {
        if (typeof data === "string") {
            return data;
        } else if (data === null || typeof data === "number" || typeof data === "bigint" || typeof data === "boolean" || typeof data === "undefined" || typeof data === "symbol") {
            return String(data);
        } else if (typeof data === "object") {
            return JSON.stringify(data);
        }
        return "undefined";
    }
    debug(msg, ...args) {
        return this._log(LogLevels.DEBUG, msg, ...args);
    }
    info(msg, ...args) {
        return this._log(LogLevels.INFO, msg, ...args);
    }
    warning(msg, ...args) {
        return this._log(LogLevels.WARNING, msg, ...args);
    }
    error(msg, ...args) {
        return this._log(LogLevels.ERROR, msg, ...args);
    }
    critical(msg, ...args) {
        return this._log(LogLevels.CRITICAL, msg, ...args);
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL2xvZy9sb2dnZXIudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCAyMDE4LTIwMjAgdGhlIERlbm8gYXV0aG9ycy4gQWxsIHJpZ2h0cyByZXNlcnZlZC4gTUlUIGxpY2Vuc2UuXG5pbXBvcnQgeyBMb2dMZXZlbHMsIGdldExldmVsQnlOYW1lLCBnZXRMZXZlbE5hbWUgfSBmcm9tIFwiLi9sZXZlbHMudHNcIjtcbmltcG9ydCB0eXBlIHsgTGV2ZWxOYW1lIH0gZnJvbSBcIi4vbGV2ZWxzLnRzXCI7XG5pbXBvcnQgdHlwZSB7IEJhc2VIYW5kbGVyIH0gZnJvbSBcIi4vaGFuZGxlcnMudHNcIjtcblxuLy8gZXNsaW50LWRpc2FibGUtbmV4dC1saW5lIEB0eXBlc2NyaXB0LWVzbGludC9uby1leHBsaWNpdC1hbnlcbmV4cG9ydCB0eXBlIEdlbmVyaWNGdW5jdGlvbiA9ICguLi5hcmdzOiBhbnlbXSkgPT4gYW55O1xuXG5leHBvcnQgaW50ZXJmYWNlIExvZ1JlY29yZE9wdGlvbnMge1xuICBtc2c6IHN0cmluZztcbiAgYXJnczogdW5rbm93bltdO1xuICBsZXZlbDogbnVtYmVyO1xuICBsb2dnZXJOYW1lOiBzdHJpbmc7XG59XG5cbmV4cG9ydCBjbGFzcyBMb2dSZWNvcmQge1xuICByZWFkb25seSBtc2c6IHN0cmluZztcbiAgI2FyZ3M6IHVua25vd25bXTtcbiAgI2RhdGV0aW1lOiBEYXRlO1xuICByZWFkb25seSBsZXZlbDogbnVtYmVyO1xuICByZWFkb25seSBsZXZlbE5hbWU6IHN0cmluZztcbiAgcmVhZG9ubHkgbG9nZ2VyTmFtZTogc3RyaW5nO1xuXG4gIGNvbnN0cnVjdG9yKG9wdGlvbnM6IExvZ1JlY29yZE9wdGlvbnMpIHtcbiAgICB0aGlzLm1zZyA9IG9wdGlvbnMubXNnO1xuICAgIHRoaXMuI2FyZ3MgPSBbLi4ub3B0aW9ucy5hcmdzXTtcbiAgICB0aGlzLmxldmVsID0gb3B0aW9ucy5sZXZlbDtcbiAgICB0aGlzLmxvZ2dlck5hbWUgPSBvcHRpb25zLmxvZ2dlck5hbWU7XG4gICAgdGhpcy4jZGF0ZXRpbWUgPSBuZXcgRGF0ZSgpO1xuICAgIHRoaXMubGV2ZWxOYW1lID0gZ2V0TGV2ZWxOYW1lKG9wdGlvbnMubGV2ZWwpO1xuICB9XG4gIGdldCBhcmdzKCk6IHVua25vd25bXSB7XG4gICAgcmV0dXJuIFsuLi50aGlzLiNhcmdzXTtcbiAgfVxuICBnZXQgZGF0ZXRpbWUoKTogRGF0ZSB7XG4gICAgcmV0dXJuIG5ldyBEYXRlKHRoaXMuI2RhdGV0aW1lLmdldFRpbWUoKSk7XG4gIH1cbn1cblxuZXhwb3J0IGludGVyZmFjZSBMb2dnZXJPcHRpb25zIHtcbiAgaGFuZGxlcnM/OiBCYXNlSGFuZGxlcltdO1xufVxuXG5leHBvcnQgY2xhc3MgTG9nZ2VyIHtcbiAgI2xldmVsOiBMb2dMZXZlbHM7XG4gICNoYW5kbGVyczogQmFzZUhhbmRsZXJbXTtcbiAgcmVhZG9ubHkgI2xvZ2dlck5hbWU6IHN0cmluZztcblxuICBjb25zdHJ1Y3RvcihcbiAgICBsb2dnZXJOYW1lOiBzdHJpbmcsXG4gICAgbGV2ZWxOYW1lOiBMZXZlbE5hbWUsXG4gICAgb3B0aW9uczogTG9nZ2VyT3B0aW9ucyA9IHt9LFxuICApIHtcbiAgICB0aGlzLiNsb2dnZXJOYW1lID0gbG9nZ2VyTmFtZTtcbiAgICB0aGlzLiNsZXZlbCA9IGdldExldmVsQnlOYW1lKGxldmVsTmFtZSk7XG4gICAgdGhpcy4jaGFuZGxlcnMgPSBvcHRpb25zLmhhbmRsZXJzIHx8IFtdO1xuICB9XG5cbiAgZ2V0IGxldmVsKCk6IExvZ0xldmVscyB7XG4gICAgcmV0dXJuIHRoaXMuI2xldmVsO1xuICB9XG4gIHNldCBsZXZlbChsZXZlbDogTG9nTGV2ZWxzKSB7XG4gICAgdGhpcy4jbGV2ZWwgPSBsZXZlbDtcbiAgfVxuXG4gIGdldCBsZXZlbE5hbWUoKTogTGV2ZWxOYW1lIHtcbiAgICByZXR1cm4gZ2V0TGV2ZWxOYW1lKHRoaXMuI2xldmVsKTtcbiAgfVxuICBzZXQgbGV2ZWxOYW1lKGxldmVsTmFtZTogTGV2ZWxOYW1lKSB7XG4gICAgdGhpcy4jbGV2ZWwgPSBnZXRMZXZlbEJ5TmFtZShsZXZlbE5hbWUpO1xuICB9XG5cbiAgZ2V0IGxvZ2dlck5hbWUoKTogc3RyaW5nIHtcbiAgICByZXR1cm4gdGhpcy4jbG9nZ2VyTmFtZTtcbiAgfVxuXG4gIHNldCBoYW5kbGVycyhobmRsczogQmFzZUhhbmRsZXJbXSkge1xuICAgIHRoaXMuI2hhbmRsZXJzID0gaG5kbHM7XG4gIH1cbiAgZ2V0IGhhbmRsZXJzKCk6IEJhc2VIYW5kbGVyW10ge1xuICAgIHJldHVybiB0aGlzLiNoYW5kbGVycztcbiAgfVxuXG4gIC8qKiBJZiB0aGUgbGV2ZWwgb2YgdGhlIGxvZ2dlciBpcyBncmVhdGVyIHRoYW4gdGhlIGxldmVsIHRvIGxvZywgdGhlbiBub3RoaW5nXG4gICAqIGlzIGxvZ2dlZCwgb3RoZXJ3aXNlIGEgbG9nIHJlY29yZCBpcyBwYXNzZWQgdG8gZWFjaCBsb2cgaGFuZGxlci4gIGBtc2dgIGRhdGFcbiAgICogcGFzc2VkIGluIGlzIHJldHVybmVkLiAgSWYgYSBmdW5jdGlvbiBpcyBwYXNzZWQgaW4sIGl0IGlzIG9ubHkgZXZhbHVhdGVkXG4gICAqIGlmIHRoZSBtc2cgd2lsbCBiZSBsb2dnZWQgYW5kIHRoZSByZXR1cm4gdmFsdWUgd2lsbCBiZSB0aGUgcmVzdWx0IG9mIHRoZVxuICAgKiBmdW5jdGlvbiwgbm90IHRoZSBmdW5jdGlvbiBpdHNlbGYsIHVubGVzcyB0aGUgZnVuY3Rpb24gaXNuJ3QgY2FsbGVkLCBpbiB3aGljaFxuICAgKiBjYXNlIHVuZGVmaW5lZCBpcyByZXR1cm5lZC4gIEFsbCB0eXBlcyBhcmUgY29lcmNlZCB0byBzdHJpbmdzIGZvciBsb2dnaW5nLlxuICAgKi9cbiAgcHJpdmF0ZSBfbG9nPFQ+KFxuICAgIGxldmVsOiBudW1iZXIsXG4gICAgbXNnOiAoVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCkgfCAoKCkgPT4gVCksXG4gICAgLi4uYXJnczogdW5rbm93bltdXG4gICk6IFQgfCB1bmRlZmluZWQge1xuICAgIGlmICh0aGlzLmxldmVsID4gbGV2ZWwpIHtcbiAgICAgIHJldHVybiBtc2cgaW5zdGFuY2VvZiBGdW5jdGlvbiA/IHVuZGVmaW5lZCA6IG1zZztcbiAgICB9XG5cbiAgICBsZXQgZm5SZXN1bHQ6IFQgfCB1bmRlZmluZWQ7XG4gICAgbGV0IGxvZ01lc3NhZ2U6IHN0cmluZztcbiAgICBpZiAobXNnIGluc3RhbmNlb2YgRnVuY3Rpb24pIHtcbiAgICAgIGZuUmVzdWx0ID0gbXNnKCk7XG4gICAgICBsb2dNZXNzYWdlID0gdGhpcy5hc1N0cmluZyhmblJlc3VsdCk7XG4gICAgfSBlbHNlIHtcbiAgICAgIGxvZ01lc3NhZ2UgPSB0aGlzLmFzU3RyaW5nKG1zZyk7XG4gICAgfVxuICAgIGNvbnN0IHJlY29yZDogTG9nUmVjb3JkID0gbmV3IExvZ1JlY29yZCh7XG4gICAgICBtc2c6IGxvZ01lc3NhZ2UsXG4gICAgICBhcmdzOiBhcmdzLFxuICAgICAgbGV2ZWw6IGxldmVsLFxuICAgICAgbG9nZ2VyTmFtZTogdGhpcy5sb2dnZXJOYW1lLFxuICAgIH0pO1xuXG4gICAgdGhpcy4jaGFuZGxlcnMuZm9yRWFjaCgoaGFuZGxlcik6IHZvaWQgPT4ge1xuICAgICAgaGFuZGxlci5oYW5kbGUocmVjb3JkKTtcbiAgICB9KTtcblxuICAgIHJldHVybiBtc2cgaW5zdGFuY2VvZiBGdW5jdGlvbiA/IGZuUmVzdWx0IDogbXNnO1xuICB9XG5cbiAgYXNTdHJpbmcoZGF0YTogdW5rbm93bik6IHN0cmluZyB7XG4gICAgaWYgKHR5cGVvZiBkYXRhID09PSBcInN0cmluZ1wiKSB7XG4gICAgICByZXR1cm4gZGF0YTtcbiAgICB9IGVsc2UgaWYgKFxuICAgICAgZGF0YSA9PT0gbnVsbCB8fFxuICAgICAgdHlwZW9mIGRhdGEgPT09IFwibnVtYmVyXCIgfHxcbiAgICAgIHR5cGVvZiBkYXRhID09PSBcImJpZ2ludFwiIHx8XG4gICAgICB0eXBlb2YgZGF0YSA9PT0gXCJib29sZWFuXCIgfHxcbiAgICAgIHR5cGVvZiBkYXRhID09PSBcInVuZGVmaW5lZFwiIHx8XG4gICAgICB0eXBlb2YgZGF0YSA9PT0gXCJzeW1ib2xcIlxuICAgICkge1xuICAgICAgcmV0dXJuIFN0cmluZyhkYXRhKTtcbiAgICB9IGVsc2UgaWYgKHR5cGVvZiBkYXRhID09PSBcIm9iamVjdFwiKSB7XG4gICAgICByZXR1cm4gSlNPTi5zdHJpbmdpZnkoZGF0YSk7XG4gICAgfVxuICAgIHJldHVybiBcInVuZGVmaW5lZFwiO1xuICB9XG5cbiAgZGVidWc8VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuICBkZWJ1ZzxUPihtc2c6IFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQsIC4uLmFyZ3M6IHVua25vd25bXSk6IFQ7XG4gIGRlYnVnPFQ+KFxuICAgIG1zZzogKFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQpIHwgKCgpID0+IFQpLFxuICAgIC4uLmFyZ3M6IHVua25vd25bXVxuICApOiBUIHwgdW5kZWZpbmVkIHtcbiAgICByZXR1cm4gdGhpcy5fbG9nKExvZ0xldmVscy5ERUJVRywgbXNnLCAuLi5hcmdzKTtcbiAgfVxuXG4gIGluZm88VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuICBpbmZvPFQ+KG1zZzogVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCwgLi4uYXJnczogdW5rbm93bltdKTogVDtcbiAgaW5mbzxUPihcbiAgICBtc2c6IChUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBUKSB8ICgoKSA9PiBUKSxcbiAgICAuLi5hcmdzOiB1bmtub3duW11cbiAgKTogVCB8IHVuZGVmaW5lZCB7XG4gICAgcmV0dXJuIHRoaXMuX2xvZyhMb2dMZXZlbHMuSU5GTywgbXNnLCAuLi5hcmdzKTtcbiAgfVxuXG4gIHdhcm5pbmc8VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuICB3YXJuaW5nPFQ+KG1zZzogVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCwgLi4uYXJnczogdW5rbm93bltdKTogVDtcbiAgd2FybmluZzxUPihcbiAgICBtc2c6IChUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBUKSB8ICgoKSA9PiBUKSxcbiAgICAuLi5hcmdzOiB1bmtub3duW11cbiAgKTogVCB8IHVuZGVmaW5lZCB7XG4gICAgcmV0dXJuIHRoaXMuX2xvZyhMb2dMZXZlbHMuV0FSTklORywgbXNnLCAuLi5hcmdzKTtcbiAgfVxuXG4gIGVycm9yPFQ+KG1zZzogKCkgPT4gVCwgLi4uYXJnczogdW5rbm93bltdKTogVCB8IHVuZGVmaW5lZDtcbiAgZXJyb3I8VD4obXNnOiBUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUO1xuICBlcnJvcjxUPihcbiAgICBtc2c6IChUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBUKSB8ICgoKSA9PiBUKSxcbiAgICAuLi5hcmdzOiB1bmtub3duW11cbiAgKTogVCB8IHVuZGVmaW5lZCB7XG4gICAgcmV0dXJuIHRoaXMuX2xvZyhMb2dMZXZlbHMuRVJST1IsIG1zZywgLi4uYXJncyk7XG4gIH1cblxuICBjcml0aWNhbDxUPihtc2c6ICgpID0+IFQsIC4uLmFyZ3M6IHVua25vd25bXSk6IFQgfCB1bmRlZmluZWQ7XG4gIGNyaXRpY2FsPFQ+KFxuICAgIG1zZzogVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCxcbiAgICAuLi5hcmdzOiB1bmtub3duW11cbiAgKTogVDtcbiAgY3JpdGljYWw8VD4oXG4gICAgbXNnOiAoVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCkgfCAoKCkgPT4gVCksXG4gICAgLi4uYXJnczogdW5rbm93bltdXG4gICk6IFQgfCB1bmRlZmluZWQge1xuICAgIHJldHVybiB0aGlzLl9sb2coTG9nTGV2ZWxzLkNSSVRJQ0FMLCBtc2csIC4uLmFyZ3MpO1xuICB9XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtTQUNBLFNBQUEsRUFBQSxjQUFBLEVBQUEsWUFBQSxTQUFBLFdBQUE7YUFjQSxTQUFBO0tBRUEsSUFBQTtLQUNBLFFBQUE7Z0JBS0EsT0FBQTthQUNBLEdBQUEsR0FBQSxPQUFBLENBQUEsR0FBQTtjQUNBLElBQUE7ZUFBQSxPQUFBLENBQUEsSUFBQTs7YUFDQSxLQUFBLEdBQUEsT0FBQSxDQUFBLEtBQUE7YUFDQSxVQUFBLEdBQUEsT0FBQSxDQUFBLFVBQUE7Y0FDQSxRQUFBLE9BQUEsSUFBQTthQUNBLFNBQUEsR0FBQSxZQUFBLENBQUEsT0FBQSxDQUFBLEtBQUE7O1FBRUEsSUFBQTs7cUJBQ0EsSUFBQTs7O1FBRUEsUUFBQTttQkFDQSxJQUFBLE9BQUEsUUFBQSxDQUFBLE9BQUE7OzthQVFBLE1BQUE7S0FDQSxLQUFBO0tBQ0EsUUFBQTtLQUNBLFVBQUE7Z0JBR0EsVUFBQSxFQUNBLFNBQUEsRUFDQSxPQUFBOztjQUVBLFVBQUEsR0FBQSxVQUFBO2NBQ0EsS0FBQSxHQUFBLGNBQUEsQ0FBQSxTQUFBO2NBQ0EsUUFBQSxHQUFBLE9BQUEsQ0FBQSxRQUFBOztRQUdBLEtBQUE7cUJBQ0EsS0FBQTs7UUFFQSxLQUFBLENBQUEsS0FBQTtjQUNBLEtBQUEsR0FBQSxLQUFBOztRQUdBLFNBQUE7ZUFDQSxZQUFBLE9BQUEsS0FBQTs7UUFFQSxTQUFBLENBQUEsU0FBQTtjQUNBLEtBQUEsR0FBQSxjQUFBLENBQUEsU0FBQTs7UUFHQSxVQUFBO3FCQUNBLFVBQUE7O1FBR0EsUUFBQSxDQUFBLEtBQUE7Y0FDQSxRQUFBLEdBQUEsS0FBQTs7UUFFQSxRQUFBO3FCQUNBLFFBQUE7O0FBR0EsTUFNQSxBQU5BLGdlQU1BLEFBTkEsRUFNQSxDQUNBLElBQUEsQ0FDQSxLQUFBLEVBQ0EsR0FBQSxLQUNBLElBQUE7aUJBRUEsS0FBQSxHQUFBLEtBQUE7bUJBQ0EsR0FBQSxZQUFBLFFBQUEsR0FBQSxTQUFBLEdBQUEsR0FBQTs7WUFHQSxRQUFBO1lBQ0EsVUFBQTtZQUNBLEdBQUEsWUFBQSxRQUFBO0FBQ0Esb0JBQUEsR0FBQSxHQUFBO0FBQ0Esc0JBQUEsUUFBQSxRQUFBLENBQUEsUUFBQTs7QUFFQSxzQkFBQSxRQUFBLFFBQUEsQ0FBQSxHQUFBOztjQUVBLE1BQUEsT0FBQSxTQUFBO0FBQ0EsZUFBQSxFQUFBLFVBQUE7QUFDQSxnQkFBQSxFQUFBLElBQUE7QUFDQSxpQkFBQSxFQUFBLEtBQUE7QUFDQSxzQkFBQSxPQUFBLFVBQUE7O2NBR0EsUUFBQSxDQUFBLE9BQUEsRUFBQSxPQUFBO0FBQ0EsbUJBQUEsQ0FBQSxNQUFBLENBQUEsTUFBQTs7ZUFHQSxHQUFBLFlBQUEsUUFBQSxHQUFBLFFBQUEsR0FBQSxHQUFBOztBQUdBLFlBQUEsQ0FBQSxJQUFBO21CQUNBLElBQUEsTUFBQSxNQUFBO21CQUNBLElBQUE7bUJBRUEsSUFBQSxLQUFBLElBQUEsV0FDQSxJQUFBLE1BQUEsTUFBQSxZQUNBLElBQUEsTUFBQSxNQUFBLFlBQ0EsSUFBQSxNQUFBLE9BQUEsWUFDQSxJQUFBLE1BQUEsU0FBQSxZQUNBLElBQUEsTUFBQSxNQUFBO21CQUVBLE1BQUEsQ0FBQSxJQUFBOzBCQUNBLElBQUEsTUFBQSxNQUFBO21CQUNBLElBQUEsQ0FBQSxTQUFBLENBQUEsSUFBQTs7Z0JBRUEsU0FBQTs7QUFLQSxTQUFBLENBQ0EsR0FBQSxLQUNBLElBQUE7b0JBRUEsSUFBQSxDQUFBLFNBQUEsQ0FBQSxLQUFBLEVBQUEsR0FBQSxLQUFBLElBQUE7O0FBS0EsUUFBQSxDQUNBLEdBQUEsS0FDQSxJQUFBO29CQUVBLElBQUEsQ0FBQSxTQUFBLENBQUEsSUFBQSxFQUFBLEdBQUEsS0FBQSxJQUFBOztBQUtBLFdBQUEsQ0FDQSxHQUFBLEtBQ0EsSUFBQTtvQkFFQSxJQUFBLENBQUEsU0FBQSxDQUFBLE9BQUEsRUFBQSxHQUFBLEtBQUEsSUFBQTs7QUFLQSxTQUFBLENBQ0EsR0FBQSxLQUNBLElBQUE7b0JBRUEsSUFBQSxDQUFBLFNBQUEsQ0FBQSxLQUFBLEVBQUEsR0FBQSxLQUFBLElBQUE7O0FBUUEsWUFBQSxDQUNBLEdBQUEsS0FDQSxJQUFBO29CQUVBLElBQUEsQ0FBQSxTQUFBLENBQUEsUUFBQSxFQUFBLEdBQUEsS0FBQSxJQUFBIn0=