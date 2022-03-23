// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/* Resolves after the given number of milliseconds. */ export function delay(ms) {
    return new Promise((res)=>setTimeout(()=>{
            res();
        }, ms)
    );
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL2FzeW5jL2RlbGF5LnRzPiJdLCJzb3VyY2VzQ29udGVudCI6WyIvLyBDb3B5cmlnaHQgMjAxOC0yMDIwIHRoZSBEZW5vIGF1dGhvcnMuIEFsbCByaWdodHMgcmVzZXJ2ZWQuIE1JVCBsaWNlbnNlLlxuLyogUmVzb2x2ZXMgYWZ0ZXIgdGhlIGdpdmVuIG51bWJlciBvZiBtaWxsaXNlY29uZHMuICovXG5leHBvcnQgZnVuY3Rpb24gZGVsYXkobXM6IG51bWJlcik6IFByb21pc2U8dm9pZD4ge1xuICByZXR1cm4gbmV3IFByb21pc2UoKHJlcyk6IG51bWJlciA9PlxuICAgIHNldFRpbWVvdXQoKCk6IHZvaWQgPT4ge1xuICAgICAgcmVzKCk7XG4gICAgfSwgbXMpXG4gICk7XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtBQUNBLEVBQUEsa0RBQUEsRUFBQSxpQkFDQSxLQUFBLENBQUEsRUFBQTtlQUNBLE9BQUEsRUFBQSxHQUFBLEdBQ0EsVUFBQTtBQUNBLGVBQUE7V0FDQSxFQUFBIn0=