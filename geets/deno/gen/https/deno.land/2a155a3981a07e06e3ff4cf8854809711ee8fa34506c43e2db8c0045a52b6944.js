// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/** This module is browser compatible. */ import { SEP } from "./separator.ts";
/** Determines the common path from a set of paths, using an optional separator,
 * which defaults to the OS default separator.
 *
 *       import { common } from "https://deno.land/std/path/mod.ts";
 *       const p = common([
 *         "./deno/std/path/mod.ts",
 *         "./deno/std/fs/mod.ts",
 *       ]);
 *       console.log(p); // "./deno/std/"
 *
 */ export function common(paths, sep = SEP) {
    const [first = "", ...remaining] = paths;
    if (first === "" || remaining.length === 0) {
        return first.substring(0, first.lastIndexOf(sep) + 1);
    }
    const parts = first.split(sep);
    let endOfPrefix = parts.length;
    for (const path of remaining){
        const compare = path.split(sep);
        for(let i = 0; i < endOfPrefix; i++){
            if (compare[i] !== parts[i]) {
                endOfPrefix = i;
            }
        }
        if (endOfPrefix === 0) {
            return "";
        }
    }
    const prefix = parts.slice(0, endOfPrefix).join(sep);
    return prefix.endsWith(sep) ? prefix : `${prefix}${sep}`;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC42OS4wL3BhdGgvY29tbW9uLnRzPiJdLCJzb3VyY2VzQ29udGVudCI6WyIvLyBDb3B5cmlnaHQgMjAxOC0yMDIwIHRoZSBEZW5vIGF1dGhvcnMuIEFsbCByaWdodHMgcmVzZXJ2ZWQuIE1JVCBsaWNlbnNlLlxuLyoqIFRoaXMgbW9kdWxlIGlzIGJyb3dzZXIgY29tcGF0aWJsZS4gKi9cblxuaW1wb3J0IHsgU0VQIH0gZnJvbSBcIi4vc2VwYXJhdG9yLnRzXCI7XG5cbi8qKiBEZXRlcm1pbmVzIHRoZSBjb21tb24gcGF0aCBmcm9tIGEgc2V0IG9mIHBhdGhzLCB1c2luZyBhbiBvcHRpb25hbCBzZXBhcmF0b3IsXG4gKiB3aGljaCBkZWZhdWx0cyB0byB0aGUgT1MgZGVmYXVsdCBzZXBhcmF0b3IuXG4gKlxuICogICAgICAgaW1wb3J0IHsgY29tbW9uIH0gZnJvbSBcImh0dHBzOi8vZGVuby5sYW5kL3N0ZC9wYXRoL21vZC50c1wiO1xuICogICAgICAgY29uc3QgcCA9IGNvbW1vbihbXG4gKiAgICAgICAgIFwiLi9kZW5vL3N0ZC9wYXRoL21vZC50c1wiLFxuICogICAgICAgICBcIi4vZGVuby9zdGQvZnMvbW9kLnRzXCIsXG4gKiAgICAgICBdKTtcbiAqICAgICAgIGNvbnNvbGUubG9nKHApOyAvLyBcIi4vZGVuby9zdGQvXCJcbiAqXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBjb21tb24ocGF0aHM6IHN0cmluZ1tdLCBzZXAgPSBTRVApOiBzdHJpbmcge1xuICBjb25zdCBbZmlyc3QgPSBcIlwiLCAuLi5yZW1haW5pbmddID0gcGF0aHM7XG4gIGlmIChmaXJzdCA9PT0gXCJcIiB8fCByZW1haW5pbmcubGVuZ3RoID09PSAwKSB7XG4gICAgcmV0dXJuIGZpcnN0LnN1YnN0cmluZygwLCBmaXJzdC5sYXN0SW5kZXhPZihzZXApICsgMSk7XG4gIH1cbiAgY29uc3QgcGFydHMgPSBmaXJzdC5zcGxpdChzZXApO1xuXG4gIGxldCBlbmRPZlByZWZpeCA9IHBhcnRzLmxlbmd0aDtcbiAgZm9yIChjb25zdCBwYXRoIG9mIHJlbWFpbmluZykge1xuICAgIGNvbnN0IGNvbXBhcmUgPSBwYXRoLnNwbGl0KHNlcCk7XG4gICAgZm9yIChsZXQgaSA9IDA7IGkgPCBlbmRPZlByZWZpeDsgaSsrKSB7XG4gICAgICBpZiAoY29tcGFyZVtpXSAhPT0gcGFydHNbaV0pIHtcbiAgICAgICAgZW5kT2ZQcmVmaXggPSBpO1xuICAgICAgfVxuICAgIH1cblxuICAgIGlmIChlbmRPZlByZWZpeCA9PT0gMCkge1xuICAgICAgcmV0dXJuIFwiXCI7XG4gICAgfVxuICB9XG4gIGNvbnN0IHByZWZpeCA9IHBhcnRzLnNsaWNlKDAsIGVuZE9mUHJlZml4KS5qb2luKHNlcCk7XG4gIHJldHVybiBwcmVmaXguZW5kc1dpdGgoc2VwKSA/IHByZWZpeCA6IGAke3ByZWZpeH0ke3NlcH1gO1xufVxuIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBLEVBQUEsd0VBQUE7QUFDQSxFQUFBLHFDQUFBLEVBQUEsVUFFQSxHQUFBLFNBQUEsY0FBQTtBQUVBLEVBVUEsQUFWQSxxV0FVQSxBQVZBLEVBVUEsaUJBQ0EsTUFBQSxDQUFBLEtBQUEsRUFBQSxHQUFBLEdBQUEsR0FBQTtXQUNBLEtBQUEsVUFBQSxTQUFBLElBQUEsS0FBQTtRQUNBLEtBQUEsV0FBQSxTQUFBLENBQUEsTUFBQSxLQUFBLENBQUE7ZUFDQSxLQUFBLENBQUEsU0FBQSxDQUFBLENBQUEsRUFBQSxLQUFBLENBQUEsV0FBQSxDQUFBLEdBQUEsSUFBQSxDQUFBOztVQUVBLEtBQUEsR0FBQSxLQUFBLENBQUEsS0FBQSxDQUFBLEdBQUE7UUFFQSxXQUFBLEdBQUEsS0FBQSxDQUFBLE1BQUE7ZUFDQSxJQUFBLElBQUEsU0FBQTtjQUNBLE9BQUEsR0FBQSxJQUFBLENBQUEsS0FBQSxDQUFBLEdBQUE7Z0JBQ0EsQ0FBQSxHQUFBLENBQUEsRUFBQSxDQUFBLEdBQUEsV0FBQSxFQUFBLENBQUE7Z0JBQ0EsT0FBQSxDQUFBLENBQUEsTUFBQSxLQUFBLENBQUEsQ0FBQTtBQUNBLDJCQUFBLEdBQUEsQ0FBQTs7O1lBSUEsV0FBQSxLQUFBLENBQUE7Ozs7VUFJQSxNQUFBLEdBQUEsS0FBQSxDQUFBLEtBQUEsQ0FBQSxDQUFBLEVBQUEsV0FBQSxFQUFBLElBQUEsQ0FBQSxHQUFBO1dBQ0EsTUFBQSxDQUFBLFFBQUEsQ0FBQSxHQUFBLElBQUEsTUFBQSxNQUFBLE1BQUEsR0FBQSxHQUFBIn0=