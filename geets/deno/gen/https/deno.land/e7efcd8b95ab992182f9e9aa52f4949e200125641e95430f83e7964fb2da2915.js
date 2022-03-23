// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
import { Logger } from "./logger.ts";
import { BaseHandler, ConsoleHandler, WriterHandler, FileHandler, RotatingFileHandler } from "./handlers.ts";
import { assert } from "../_util/assert.ts";
export { LogLevels } from "./levels.ts";
export { Logger } from "./logger.ts";
export class LoggerConfig {
}
const DEFAULT_LEVEL = "INFO";
const DEFAULT_CONFIG = {
    handlers: {
        default: new ConsoleHandler(DEFAULT_LEVEL)
    },
    loggers: {
        default: {
            level: DEFAULT_LEVEL,
            handlers: [
                "default"
            ]
        }
    }
};
const state = {
    handlers: new Map(),
    loggers: new Map(),
    config: DEFAULT_CONFIG
};
export const handlers = {
    BaseHandler,
    ConsoleHandler,
    WriterHandler,
    FileHandler,
    RotatingFileHandler
};
export function getLogger(name) {
    if (!name) {
        const d = state.loggers.get("default");
        assert(d != null, `"default" logger must be set for getting logger without name`);
        return d;
    }
    const result = state.loggers.get(name);
    if (!result) {
        const logger = new Logger(name, "NOTSET", {
            handlers: []
        });
        state.loggers.set(name, logger);
        return logger;
    }
    return result;
}
export function debug(msg, ...args) {
    // Assist TS compiler with pass-through generic type
    if (msg instanceof Function) {
        return getLogger("default").debug(msg, ...args);
    }
    return getLogger("default").debug(msg, ...args);
}
export function info(msg, ...args) {
    // Assist TS compiler with pass-through generic type
    if (msg instanceof Function) {
        return getLogger("default").info(msg, ...args);
    }
    return getLogger("default").info(msg, ...args);
}
export function warning(msg, ...args) {
    // Assist TS compiler with pass-through generic type
    if (msg instanceof Function) {
        return getLogger("default").warning(msg, ...args);
    }
    return getLogger("default").warning(msg, ...args);
}
export function error(msg, ...args) {
    // Assist TS compiler with pass-through generic type
    if (msg instanceof Function) {
        return getLogger("default").error(msg, ...args);
    }
    return getLogger("default").error(msg, ...args);
}
export function critical(msg, ...args) {
    // Assist TS compiler with pass-through generic type
    if (msg instanceof Function) {
        return getLogger("default").critical(msg, ...args);
    }
    return getLogger("default").critical(msg, ...args);
}
export async function setup(config) {
    state.config = {
        handlers: {
            ...DEFAULT_CONFIG.handlers,
            ...config.handlers
        },
        loggers: {
            ...DEFAULT_CONFIG.loggers,
            ...config.loggers
        }
    };
    // tear down existing handlers
    state.handlers.forEach((handler)=>{
        handler.destroy();
    });
    state.handlers.clear();
    // setup handlers
    const handlers = state.config.handlers || {
    };
    for(const handlerName in handlers){
        const handler = handlers[handlerName];
        await handler.setup();
        state.handlers.set(handlerName, handler);
    }
    // remove existing loggers
    state.loggers.clear();
    // setup loggers
    const loggers = state.config.loggers || {
    };
    for(const loggerName in loggers){
        const loggerConfig = loggers[loggerName];
        const handlerNames = loggerConfig.handlers || [];
        const handlers = [];
        handlerNames.forEach((handlerName)=>{
            const handler = state.handlers.get(handlerName);
            if (handler) {
                handlers.push(handler);
            }
        });
        const levelName = loggerConfig.level || DEFAULT_LEVEL;
        const logger = new Logger(loggerName, levelName, {
            handlers: handlers
        });
        state.loggers.set(loggerName, logger);
    }
}
await setup(DEFAULT_CONFIG);
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL2xvZy9tb2QudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCAyMDE4LTIwMjAgdGhlIERlbm8gYXV0aG9ycy4gQWxsIHJpZ2h0cyByZXNlcnZlZC4gTUlUIGxpY2Vuc2UuXG5pbXBvcnQgeyBMb2dnZXIgfSBmcm9tIFwiLi9sb2dnZXIudHNcIjtcbmltcG9ydCB0eXBlIHsgR2VuZXJpY0Z1bmN0aW9uIH0gZnJvbSBcIi4vbG9nZ2VyLnRzXCI7XG5pbXBvcnQge1xuICBCYXNlSGFuZGxlcixcbiAgQ29uc29sZUhhbmRsZXIsXG4gIFdyaXRlckhhbmRsZXIsXG4gIEZpbGVIYW5kbGVyLFxuICBSb3RhdGluZ0ZpbGVIYW5kbGVyLFxufSBmcm9tIFwiLi9oYW5kbGVycy50c1wiO1xuaW1wb3J0IHsgYXNzZXJ0IH0gZnJvbSBcIi4uL191dGlsL2Fzc2VydC50c1wiO1xuaW1wb3J0IHR5cGUgeyBMZXZlbE5hbWUgfSBmcm9tIFwiLi9sZXZlbHMudHNcIjtcblxuZXhwb3J0IHsgTG9nTGV2ZWxzIH0gZnJvbSBcIi4vbGV2ZWxzLnRzXCI7XG5leHBvcnQgdHlwZSB7IExldmVsTmFtZSB9IGZyb20gXCIuL2xldmVscy50c1wiO1xuZXhwb3J0IHsgTG9nZ2VyIH0gZnJvbSBcIi4vbG9nZ2VyLnRzXCI7XG5cbmV4cG9ydCBjbGFzcyBMb2dnZXJDb25maWcge1xuICBsZXZlbD86IExldmVsTmFtZTtcbiAgaGFuZGxlcnM/OiBzdHJpbmdbXTtcbn1cblxuZXhwb3J0IGludGVyZmFjZSBMb2dDb25maWcge1xuICBoYW5kbGVycz86IHtcbiAgICBbbmFtZTogc3RyaW5nXTogQmFzZUhhbmRsZXI7XG4gIH07XG4gIGxvZ2dlcnM/OiB7XG4gICAgW25hbWU6IHN0cmluZ106IExvZ2dlckNvbmZpZztcbiAgfTtcbn1cblxuY29uc3QgREVGQVVMVF9MRVZFTCA9IFwiSU5GT1wiO1xuY29uc3QgREVGQVVMVF9DT05GSUc6IExvZ0NvbmZpZyA9IHtcbiAgaGFuZGxlcnM6IHtcbiAgICBkZWZhdWx0OiBuZXcgQ29uc29sZUhhbmRsZXIoREVGQVVMVF9MRVZFTCksXG4gIH0sXG5cbiAgbG9nZ2Vyczoge1xuICAgIGRlZmF1bHQ6IHtcbiAgICAgIGxldmVsOiBERUZBVUxUX0xFVkVMLFxuICAgICAgaGFuZGxlcnM6IFtcImRlZmF1bHRcIl0sXG4gICAgfSxcbiAgfSxcbn07XG5cbmNvbnN0IHN0YXRlID0ge1xuICBoYW5kbGVyczogbmV3IE1hcDxzdHJpbmcsIEJhc2VIYW5kbGVyPigpLFxuICBsb2dnZXJzOiBuZXcgTWFwPHN0cmluZywgTG9nZ2VyPigpLFxuICBjb25maWc6IERFRkFVTFRfQ09ORklHLFxufTtcblxuZXhwb3J0IGNvbnN0IGhhbmRsZXJzID0ge1xuICBCYXNlSGFuZGxlcixcbiAgQ29uc29sZUhhbmRsZXIsXG4gIFdyaXRlckhhbmRsZXIsXG4gIEZpbGVIYW5kbGVyLFxuICBSb3RhdGluZ0ZpbGVIYW5kbGVyLFxufTtcblxuZXhwb3J0IGZ1bmN0aW9uIGdldExvZ2dlcihuYW1lPzogc3RyaW5nKTogTG9nZ2VyIHtcbiAgaWYgKCFuYW1lKSB7XG4gICAgY29uc3QgZCA9IHN0YXRlLmxvZ2dlcnMuZ2V0KFwiZGVmYXVsdFwiKTtcbiAgICBhc3NlcnQoXG4gICAgICBkICE9IG51bGwsXG4gICAgICBgXCJkZWZhdWx0XCIgbG9nZ2VyIG11c3QgYmUgc2V0IGZvciBnZXR0aW5nIGxvZ2dlciB3aXRob3V0IG5hbWVgLFxuICAgICk7XG4gICAgcmV0dXJuIGQ7XG4gIH1cbiAgY29uc3QgcmVzdWx0ID0gc3RhdGUubG9nZ2Vycy5nZXQobmFtZSk7XG4gIGlmICghcmVzdWx0KSB7XG4gICAgY29uc3QgbG9nZ2VyID0gbmV3IExvZ2dlcihuYW1lLCBcIk5PVFNFVFwiLCB7IGhhbmRsZXJzOiBbXSB9KTtcbiAgICBzdGF0ZS5sb2dnZXJzLnNldChuYW1lLCBsb2dnZXIpO1xuICAgIHJldHVybiBsb2dnZXI7XG4gIH1cbiAgcmV0dXJuIHJlc3VsdDtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGRlYnVnPFQ+KG1zZzogKCkgPT4gVCwgLi4uYXJnczogdW5rbm93bltdKTogVCB8IHVuZGVmaW5lZDtcbmV4cG9ydCBmdW5jdGlvbiBkZWJ1ZzxUPihcbiAgbXNnOiBUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBULFxuICAuLi5hcmdzOiB1bmtub3duW11cbik6IFQ7XG5leHBvcnQgZnVuY3Rpb24gZGVidWc8VD4oXG4gIG1zZzogKFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQpIHwgKCgpID0+IFQpLFxuICAuLi5hcmdzOiB1bmtub3duW11cbik6IFQgfCB1bmRlZmluZWQge1xuICAvLyBBc3Npc3QgVFMgY29tcGlsZXIgd2l0aCBwYXNzLXRocm91Z2ggZ2VuZXJpYyB0eXBlXG4gIGlmIChtc2cgaW5zdGFuY2VvZiBGdW5jdGlvbikge1xuICAgIHJldHVybiBnZXRMb2dnZXIoXCJkZWZhdWx0XCIpLmRlYnVnKG1zZywgLi4uYXJncyk7XG4gIH1cbiAgcmV0dXJuIGdldExvZ2dlcihcImRlZmF1bHRcIikuZGVidWcobXNnLCAuLi5hcmdzKTtcbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGluZm88VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuZXhwb3J0IGZ1bmN0aW9uIGluZm88VD4oXG4gIG1zZzogVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCxcbiAgLi4uYXJnczogdW5rbm93bltdXG4pOiBUO1xuZXhwb3J0IGZ1bmN0aW9uIGluZm88VD4oXG4gIG1zZzogKFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQpIHwgKCgpID0+IFQpLFxuICAuLi5hcmdzOiB1bmtub3duW11cbik6IFQgfCB1bmRlZmluZWQge1xuICAvLyBBc3Npc3QgVFMgY29tcGlsZXIgd2l0aCBwYXNzLXRocm91Z2ggZ2VuZXJpYyB0eXBlXG4gIGlmIChtc2cgaW5zdGFuY2VvZiBGdW5jdGlvbikge1xuICAgIHJldHVybiBnZXRMb2dnZXIoXCJkZWZhdWx0XCIpLmluZm8obXNnLCAuLi5hcmdzKTtcbiAgfVxuICByZXR1cm4gZ2V0TG9nZ2VyKFwiZGVmYXVsdFwiKS5pbmZvKG1zZywgLi4uYXJncyk7XG59XG5cbmV4cG9ydCBmdW5jdGlvbiB3YXJuaW5nPFQ+KG1zZzogKCkgPT4gVCwgLi4uYXJnczogdW5rbm93bltdKTogVCB8IHVuZGVmaW5lZDtcbmV4cG9ydCBmdW5jdGlvbiB3YXJuaW5nPFQ+KFxuICBtc2c6IFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQsXG4gIC4uLmFyZ3M6IHVua25vd25bXVxuKTogVDtcbmV4cG9ydCBmdW5jdGlvbiB3YXJuaW5nPFQ+KFxuICBtc2c6IChUIGV4dGVuZHMgR2VuZXJpY0Z1bmN0aW9uID8gbmV2ZXIgOiBUKSB8ICgoKSA9PiBUKSxcbiAgLi4uYXJnczogdW5rbm93bltdXG4pOiBUIHwgdW5kZWZpbmVkIHtcbiAgLy8gQXNzaXN0IFRTIGNvbXBpbGVyIHdpdGggcGFzcy10aHJvdWdoIGdlbmVyaWMgdHlwZVxuICBpZiAobXNnIGluc3RhbmNlb2YgRnVuY3Rpb24pIHtcbiAgICByZXR1cm4gZ2V0TG9nZ2VyKFwiZGVmYXVsdFwiKS53YXJuaW5nKG1zZywgLi4uYXJncyk7XG4gIH1cbiAgcmV0dXJuIGdldExvZ2dlcihcImRlZmF1bHRcIikud2FybmluZyhtc2csIC4uLmFyZ3MpO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gZXJyb3I8VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuZXhwb3J0IGZ1bmN0aW9uIGVycm9yPFQ+KFxuICBtc2c6IFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQsXG4gIC4uLmFyZ3M6IHVua25vd25bXVxuKTogVDtcbmV4cG9ydCBmdW5jdGlvbiBlcnJvcjxUPihcbiAgbXNnOiAoVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCkgfCAoKCkgPT4gVCksXG4gIC4uLmFyZ3M6IHVua25vd25bXVxuKTogVCB8IHVuZGVmaW5lZCB7XG4gIC8vIEFzc2lzdCBUUyBjb21waWxlciB3aXRoIHBhc3MtdGhyb3VnaCBnZW5lcmljIHR5cGVcbiAgaWYgKG1zZyBpbnN0YW5jZW9mIEZ1bmN0aW9uKSB7XG4gICAgcmV0dXJuIGdldExvZ2dlcihcImRlZmF1bHRcIikuZXJyb3IobXNnLCAuLi5hcmdzKTtcbiAgfVxuICByZXR1cm4gZ2V0TG9nZ2VyKFwiZGVmYXVsdFwiKS5lcnJvcihtc2csIC4uLmFyZ3MpO1xufVxuXG5leHBvcnQgZnVuY3Rpb24gY3JpdGljYWw8VD4obXNnOiAoKSA9PiBULCAuLi5hcmdzOiB1bmtub3duW10pOiBUIHwgdW5kZWZpbmVkO1xuZXhwb3J0IGZ1bmN0aW9uIGNyaXRpY2FsPFQ+KFxuICBtc2c6IFQgZXh0ZW5kcyBHZW5lcmljRnVuY3Rpb24gPyBuZXZlciA6IFQsXG4gIC4uLmFyZ3M6IHVua25vd25bXVxuKTogVDtcbmV4cG9ydCBmdW5jdGlvbiBjcml0aWNhbDxUPihcbiAgbXNnOiAoVCBleHRlbmRzIEdlbmVyaWNGdW5jdGlvbiA/IG5ldmVyIDogVCkgfCAoKCkgPT4gVCksXG4gIC4uLmFyZ3M6IHVua25vd25bXVxuKTogVCB8IHVuZGVmaW5lZCB7XG4gIC8vIEFzc2lzdCBUUyBjb21waWxlciB3aXRoIHBhc3MtdGhyb3VnaCBnZW5lcmljIHR5cGVcbiAgaWYgKG1zZyBpbnN0YW5jZW9mIEZ1bmN0aW9uKSB7XG4gICAgcmV0dXJuIGdldExvZ2dlcihcImRlZmF1bHRcIikuY3JpdGljYWwobXNnLCAuLi5hcmdzKTtcbiAgfVxuICByZXR1cm4gZ2V0TG9nZ2VyKFwiZGVmYXVsdFwiKS5jcml0aWNhbChtc2csIC4uLmFyZ3MpO1xufVxuXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gc2V0dXAoY29uZmlnOiBMb2dDb25maWcpOiBQcm9taXNlPHZvaWQ+IHtcbiAgc3RhdGUuY29uZmlnID0ge1xuICAgIGhhbmRsZXJzOiB7IC4uLkRFRkFVTFRfQ09ORklHLmhhbmRsZXJzLCAuLi5jb25maWcuaGFuZGxlcnMgfSxcbiAgICBsb2dnZXJzOiB7IC4uLkRFRkFVTFRfQ09ORklHLmxvZ2dlcnMsIC4uLmNvbmZpZy5sb2dnZXJzIH0sXG4gIH07XG5cbiAgLy8gdGVhciBkb3duIGV4aXN0aW5nIGhhbmRsZXJzXG4gIHN0YXRlLmhhbmRsZXJzLmZvckVhY2goKGhhbmRsZXIpOiB2b2lkID0+IHtcbiAgICBoYW5kbGVyLmRlc3Ryb3koKTtcbiAgfSk7XG4gIHN0YXRlLmhhbmRsZXJzLmNsZWFyKCk7XG5cbiAgLy8gc2V0dXAgaGFuZGxlcnNcbiAgY29uc3QgaGFuZGxlcnMgPSBzdGF0ZS5jb25maWcuaGFuZGxlcnMgfHwge307XG5cbiAgZm9yIChjb25zdCBoYW5kbGVyTmFtZSBpbiBoYW5kbGVycykge1xuICAgIGNvbnN0IGhhbmRsZXIgPSBoYW5kbGVyc1toYW5kbGVyTmFtZV07XG4gICAgYXdhaXQgaGFuZGxlci5zZXR1cCgpO1xuICAgIHN0YXRlLmhhbmRsZXJzLnNldChoYW5kbGVyTmFtZSwgaGFuZGxlcik7XG4gIH1cblxuICAvLyByZW1vdmUgZXhpc3RpbmcgbG9nZ2Vyc1xuICBzdGF0ZS5sb2dnZXJzLmNsZWFyKCk7XG5cbiAgLy8gc2V0dXAgbG9nZ2Vyc1xuICBjb25zdCBsb2dnZXJzID0gc3RhdGUuY29uZmlnLmxvZ2dlcnMgfHwge307XG4gIGZvciAoY29uc3QgbG9nZ2VyTmFtZSBpbiBsb2dnZXJzKSB7XG4gICAgY29uc3QgbG9nZ2VyQ29uZmlnID0gbG9nZ2Vyc1tsb2dnZXJOYW1lXTtcbiAgICBjb25zdCBoYW5kbGVyTmFtZXMgPSBsb2dnZXJDb25maWcuaGFuZGxlcnMgfHwgW107XG4gICAgY29uc3QgaGFuZGxlcnM6IEJhc2VIYW5kbGVyW10gPSBbXTtcblxuICAgIGhhbmRsZXJOYW1lcy5mb3JFYWNoKChoYW5kbGVyTmFtZSk6IHZvaWQgPT4ge1xuICAgICAgY29uc3QgaGFuZGxlciA9IHN0YXRlLmhhbmRsZXJzLmdldChoYW5kbGVyTmFtZSk7XG4gICAgICBpZiAoaGFuZGxlcikge1xuICAgICAgICBoYW5kbGVycy5wdXNoKGhhbmRsZXIpO1xuICAgICAgfVxuICAgIH0pO1xuXG4gICAgY29uc3QgbGV2ZWxOYW1lID0gbG9nZ2VyQ29uZmlnLmxldmVsIHx8IERFRkFVTFRfTEVWRUw7XG4gICAgY29uc3QgbG9nZ2VyID0gbmV3IExvZ2dlcihsb2dnZXJOYW1lLCBsZXZlbE5hbWUsIHsgaGFuZGxlcnM6IGhhbmRsZXJzIH0pO1xuICAgIHN0YXRlLmxvZ2dlcnMuc2V0KGxvZ2dlck5hbWUsIGxvZ2dlcik7XG4gIH1cbn1cblxuYXdhaXQgc2V0dXAoREVGQVVMVF9DT05GSUcpO1xuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLEVBQUEsd0VBQUE7U0FDQSxNQUFBLFNBQUEsV0FBQTtTQUdBLFdBQUEsRUFDQSxjQUFBLEVBQ0EsYUFBQSxFQUNBLFdBQUEsRUFDQSxtQkFBQSxTQUNBLGFBQUE7U0FDQSxNQUFBLFNBQUEsa0JBQUE7U0FHQSxTQUFBLFNBQUEsV0FBQTtTQUVBLE1BQUEsU0FBQSxXQUFBO2FBRUEsWUFBQTs7TUFjQSxhQUFBLElBQUEsSUFBQTtNQUNBLGNBQUE7QUFDQSxZQUFBO0FBQ0EsZUFBQSxNQUFBLGNBQUEsQ0FBQSxhQUFBOztBQUdBLFdBQUE7QUFDQSxlQUFBO0FBQ0EsaUJBQUEsRUFBQSxhQUFBO0FBQ0Esb0JBQUE7aUJBQUEsT0FBQTs7Ozs7TUFLQSxLQUFBO0FBQ0EsWUFBQSxNQUFBLEdBQUE7QUFDQSxXQUFBLE1BQUEsR0FBQTtBQUNBLFVBQUEsRUFBQSxjQUFBOzthQUdBLFFBQUE7QUFDQSxlQUFBO0FBQ0Esa0JBQUE7QUFDQSxpQkFBQTtBQUNBLGVBQUE7QUFDQSx1QkFBQTs7Z0JBR0EsU0FBQSxDQUFBLElBQUE7U0FDQSxJQUFBO2NBQ0EsQ0FBQSxHQUFBLEtBQUEsQ0FBQSxPQUFBLENBQUEsR0FBQSxFQUFBLE9BQUE7QUFDQSxjQUFBLENBQ0EsQ0FBQSxJQUFBLElBQUEsR0FDQSw0REFBQTtlQUVBLENBQUE7O1VBRUEsTUFBQSxHQUFBLEtBQUEsQ0FBQSxPQUFBLENBQUEsR0FBQSxDQUFBLElBQUE7U0FDQSxNQUFBO2NBQ0EsTUFBQSxPQUFBLE1BQUEsQ0FBQSxJQUFBLEdBQUEsTUFBQTtBQUFBLG9CQUFBOztBQUNBLGFBQUEsQ0FBQSxPQUFBLENBQUEsR0FBQSxDQUFBLElBQUEsRUFBQSxNQUFBO2VBQ0EsTUFBQTs7V0FFQSxNQUFBOztnQkFRQSxLQUFBLENBQ0EsR0FBQSxLQUNBLElBQUE7QUFFQSxNQUFBLGtEQUFBO1FBQ0EsR0FBQSxZQUFBLFFBQUE7ZUFDQSxTQUFBLEVBQUEsT0FBQSxHQUFBLEtBQUEsQ0FBQSxHQUFBLEtBQUEsSUFBQTs7V0FFQSxTQUFBLEVBQUEsT0FBQSxHQUFBLEtBQUEsQ0FBQSxHQUFBLEtBQUEsSUFBQTs7Z0JBUUEsSUFBQSxDQUNBLEdBQUEsS0FDQSxJQUFBO0FBRUEsTUFBQSxrREFBQTtRQUNBLEdBQUEsWUFBQSxRQUFBO2VBQ0EsU0FBQSxFQUFBLE9BQUEsR0FBQSxJQUFBLENBQUEsR0FBQSxLQUFBLElBQUE7O1dBRUEsU0FBQSxFQUFBLE9BQUEsR0FBQSxJQUFBLENBQUEsR0FBQSxLQUFBLElBQUE7O2dCQVFBLE9BQUEsQ0FDQSxHQUFBLEtBQ0EsSUFBQTtBQUVBLE1BQUEsa0RBQUE7UUFDQSxHQUFBLFlBQUEsUUFBQTtlQUNBLFNBQUEsRUFBQSxPQUFBLEdBQUEsT0FBQSxDQUFBLEdBQUEsS0FBQSxJQUFBOztXQUVBLFNBQUEsRUFBQSxPQUFBLEdBQUEsT0FBQSxDQUFBLEdBQUEsS0FBQSxJQUFBOztnQkFRQSxLQUFBLENBQ0EsR0FBQSxLQUNBLElBQUE7QUFFQSxNQUFBLGtEQUFBO1FBQ0EsR0FBQSxZQUFBLFFBQUE7ZUFDQSxTQUFBLEVBQUEsT0FBQSxHQUFBLEtBQUEsQ0FBQSxHQUFBLEtBQUEsSUFBQTs7V0FFQSxTQUFBLEVBQUEsT0FBQSxHQUFBLEtBQUEsQ0FBQSxHQUFBLEtBQUEsSUFBQTs7Z0JBUUEsUUFBQSxDQUNBLEdBQUEsS0FDQSxJQUFBO0FBRUEsTUFBQSxrREFBQTtRQUNBLEdBQUEsWUFBQSxRQUFBO2VBQ0EsU0FBQSxFQUFBLE9BQUEsR0FBQSxRQUFBLENBQUEsR0FBQSxLQUFBLElBQUE7O1dBRUEsU0FBQSxFQUFBLE9BQUEsR0FBQSxRQUFBLENBQUEsR0FBQSxLQUFBLElBQUE7O3NCQUdBLEtBQUEsQ0FBQSxNQUFBO0FBQ0EsU0FBQSxDQUFBLE1BQUE7QUFDQSxnQkFBQTtlQUFBLGNBQUEsQ0FBQSxRQUFBO2VBQUEsTUFBQSxDQUFBLFFBQUE7O0FBQ0EsZUFBQTtlQUFBLGNBQUEsQ0FBQSxPQUFBO2VBQUEsTUFBQSxDQUFBLE9BQUE7OztBQUdBLE1BQUEsNEJBQUE7QUFDQSxTQUFBLENBQUEsUUFBQSxDQUFBLE9BQUEsRUFBQSxPQUFBO0FBQ0EsZUFBQSxDQUFBLE9BQUE7O0FBRUEsU0FBQSxDQUFBLFFBQUEsQ0FBQSxLQUFBO0FBRUEsTUFBQSxlQUFBO1VBQ0EsUUFBQSxHQUFBLEtBQUEsQ0FBQSxNQUFBLENBQUEsUUFBQTs7Y0FFQSxXQUFBLElBQUEsUUFBQTtjQUNBLE9BQUEsR0FBQSxRQUFBLENBQUEsV0FBQTtjQUNBLE9BQUEsQ0FBQSxLQUFBO0FBQ0EsYUFBQSxDQUFBLFFBQUEsQ0FBQSxHQUFBLENBQUEsV0FBQSxFQUFBLE9BQUE7O0FBR0EsTUFBQSx3QkFBQTtBQUNBLFNBQUEsQ0FBQSxPQUFBLENBQUEsS0FBQTtBQUVBLE1BQUEsY0FBQTtVQUNBLE9BQUEsR0FBQSxLQUFBLENBQUEsTUFBQSxDQUFBLE9BQUE7O2NBQ0EsVUFBQSxJQUFBLE9BQUE7Y0FDQSxZQUFBLEdBQUEsT0FBQSxDQUFBLFVBQUE7Y0FDQSxZQUFBLEdBQUEsWUFBQSxDQUFBLFFBQUE7Y0FDQSxRQUFBO0FBRUEsb0JBQUEsQ0FBQSxPQUFBLEVBQUEsV0FBQTtrQkFDQSxPQUFBLEdBQUEsS0FBQSxDQUFBLFFBQUEsQ0FBQSxHQUFBLENBQUEsV0FBQTtnQkFDQSxPQUFBO0FBQ0Esd0JBQUEsQ0FBQSxJQUFBLENBQUEsT0FBQTs7O2NBSUEsU0FBQSxHQUFBLFlBQUEsQ0FBQSxLQUFBLElBQUEsYUFBQTtjQUNBLE1BQUEsT0FBQSxNQUFBLENBQUEsVUFBQSxFQUFBLFNBQUE7QUFBQSxvQkFBQSxFQUFBLFFBQUE7O0FBQ0EsYUFBQSxDQUFBLE9BQUEsQ0FBQSxHQUFBLENBQUEsVUFBQSxFQUFBLE1BQUE7OztNQUlBLEtBQUEsQ0FBQSxjQUFBIn0=