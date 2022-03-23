// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
export class DenoStdInternalError extends Error {
    constructor(message){
        super(message);
        this.name = "DenoStdInternalError";
    }
}
/** Make an assertion, if not `true`, then throw. */ export function assert(expr, msg = "") {
    if (!expr) {
        throw new DenoStdInternalError(msg);
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL191dGlsL2Fzc2VydC50cz4iXSwic291cmNlc0NvbnRlbnQiOlsiLy8gQ29weXJpZ2h0IDIwMTgtMjAyMCB0aGUgRGVubyBhdXRob3JzLiBBbGwgcmlnaHRzIHJlc2VydmVkLiBNSVQgbGljZW5zZS5cblxuZXhwb3J0IGNsYXNzIERlbm9TdGRJbnRlcm5hbEVycm9yIGV4dGVuZHMgRXJyb3Ige1xuICBjb25zdHJ1Y3RvcihtZXNzYWdlOiBzdHJpbmcpIHtcbiAgICBzdXBlcihtZXNzYWdlKTtcbiAgICB0aGlzLm5hbWUgPSBcIkRlbm9TdGRJbnRlcm5hbEVycm9yXCI7XG4gIH1cbn1cblxuLyoqIE1ha2UgYW4gYXNzZXJ0aW9uLCBpZiBub3QgYHRydWVgLCB0aGVuIHRocm93LiAqL1xuZXhwb3J0IGZ1bmN0aW9uIGFzc2VydChleHByOiB1bmtub3duLCBtc2cgPSBcIlwiKTogYXNzZXJ0cyBleHByIHtcbiAgaWYgKCFleHByKSB7XG4gICAgdGhyb3cgbmV3IERlbm9TdGRJbnRlcm5hbEVycm9yKG1zZyk7XG4gIH1cbn1cbiJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQSxFQUFBLHdFQUFBO2FBRUEsb0JBQUEsU0FBQSxLQUFBO2dCQUNBLE9BQUE7QUFDQSxhQUFBLENBQUEsT0FBQTthQUNBLElBQUEsSUFBQSxvQkFBQTs7O0FBSUEsRUFBQSxnREFBQSxFQUFBLGlCQUNBLE1BQUEsQ0FBQSxJQUFBLEVBQUEsR0FBQTtTQUNBLElBQUE7a0JBQ0Esb0JBQUEsQ0FBQSxHQUFBIn0=