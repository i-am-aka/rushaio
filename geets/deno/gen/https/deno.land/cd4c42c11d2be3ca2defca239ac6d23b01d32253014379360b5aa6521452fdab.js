// Copyright 2018-2020 the Deno authors. All rights reserved. MIT license.
/**
 * pooledMap transforms values from an (async) iterable into another async
 * iterable. The transforms are done concurrently, with a max concurrency
 * defined by the poolLimit.
 * 
 * @param poolLimit The maximum count of items being processed concurrently. 
 * @param array The input array for mapping.
 * @param iteratorFn The function to call for every item of the array.
 */ export function pooledMap(poolLimit, array, iteratorFn) {
    // Create the async iterable that is returned from this function.
    const res = new TransformStream({
        async transform (p, controller) {
            controller.enqueue(await p);
        }
    });
    // Start processing items from the iterator
    (async ()=>{
        const writer = res.writable.getWriter();
        const executing = [];
        for await (const item of array){
            const p = Promise.resolve().then(()=>iteratorFn(item)
            );
            writer.write(p);
            const e = p.then(()=>executing.splice(executing.indexOf(e), 1)
            );
            executing.push(e);
            if (executing.length >= poolLimit) {
                await Promise.race(executing);
            }
        }
        // Wait until all ongoing events have processed, then close the writer.
        await Promise.all(executing);
        writer.close();
    })();
    return res.readable.getIterator();
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIjxodHRwczovL2Rlbm8ubGFuZC9zdGRAMC43NS4wL2FzeW5jL3Bvb2wudHM+Il0sInNvdXJjZXNDb250ZW50IjpbIi8vIENvcHlyaWdodCAyMDE4LTIwMjAgdGhlIERlbm8gYXV0aG9ycy4gQWxsIHJpZ2h0cyByZXNlcnZlZC4gTUlUIGxpY2Vuc2UuXG5cbi8qKlxuICogcG9vbGVkTWFwIHRyYW5zZm9ybXMgdmFsdWVzIGZyb20gYW4gKGFzeW5jKSBpdGVyYWJsZSBpbnRvIGFub3RoZXIgYXN5bmNcbiAqIGl0ZXJhYmxlLiBUaGUgdHJhbnNmb3JtcyBhcmUgZG9uZSBjb25jdXJyZW50bHksIHdpdGggYSBtYXggY29uY3VycmVuY3lcbiAqIGRlZmluZWQgYnkgdGhlIHBvb2xMaW1pdC5cbiAqIFxuICogQHBhcmFtIHBvb2xMaW1pdCBUaGUgbWF4aW11bSBjb3VudCBvZiBpdGVtcyBiZWluZyBwcm9jZXNzZWQgY29uY3VycmVudGx5LiBcbiAqIEBwYXJhbSBhcnJheSBUaGUgaW5wdXQgYXJyYXkgZm9yIG1hcHBpbmcuXG4gKiBAcGFyYW0gaXRlcmF0b3JGbiBUaGUgZnVuY3Rpb24gdG8gY2FsbCBmb3IgZXZlcnkgaXRlbSBvZiB0aGUgYXJyYXkuXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBwb29sZWRNYXA8VCwgUj4oXG4gIHBvb2xMaW1pdDogbnVtYmVyLFxuICBhcnJheTogSXRlcmFibGU8VD4gfCBBc3luY0l0ZXJhYmxlPFQ+LFxuICBpdGVyYXRvckZuOiAoZGF0YTogVCkgPT4gUHJvbWlzZTxSPixcbik6IEFzeW5jSXRlcmFibGVJdGVyYXRvcjxSPiB7XG4gIC8vIENyZWF0ZSB0aGUgYXN5bmMgaXRlcmFibGUgdGhhdCBpcyByZXR1cm5lZCBmcm9tIHRoaXMgZnVuY3Rpb24uXG4gIGNvbnN0IHJlcyA9IG5ldyBUcmFuc2Zvcm1TdHJlYW08UHJvbWlzZTxSPiwgUj4oe1xuICAgIGFzeW5jIHRyYW5zZm9ybShcbiAgICAgIHA6IFByb21pc2U8Uj4sXG4gICAgICBjb250cm9sbGVyOiBUcmFuc2Zvcm1TdHJlYW1EZWZhdWx0Q29udHJvbGxlcjxSPixcbiAgICApOiBQcm9taXNlPHZvaWQ+IHtcbiAgICAgIGNvbnRyb2xsZXIuZW5xdWV1ZShhd2FpdCBwKTtcbiAgICB9LFxuICB9KTtcbiAgLy8gU3RhcnQgcHJvY2Vzc2luZyBpdGVtcyBmcm9tIHRoZSBpdGVyYXRvclxuICAoYXN5bmMgKCk6IFByb21pc2U8dm9pZD4gPT4ge1xuICAgIGNvbnN0IHdyaXRlciA9IHJlcy53cml0YWJsZS5nZXRXcml0ZXIoKTtcbiAgICBjb25zdCBleGVjdXRpbmc6IEFycmF5PFByb21pc2U8dW5rbm93bj4+ID0gW107XG4gICAgZm9yIGF3YWl0IChjb25zdCBpdGVtIG9mIGFycmF5KSB7XG4gICAgICBjb25zdCBwID0gUHJvbWlzZS5yZXNvbHZlKCkudGhlbigoKSA9PiBpdGVyYXRvckZuKGl0ZW0pKTtcbiAgICAgIHdyaXRlci53cml0ZShwKTtcbiAgICAgIGNvbnN0IGU6IFByb21pc2U8dW5rbm93bj4gPSBwLnRoZW4oKCkgPT5cbiAgICAgICAgZXhlY3V0aW5nLnNwbGljZShleGVjdXRpbmcuaW5kZXhPZihlKSwgMSlcbiAgICAgICk7XG4gICAgICBleGVjdXRpbmcucHVzaChlKTtcbiAgICAgIGlmIChleGVjdXRpbmcubGVuZ3RoID49IHBvb2xMaW1pdCkge1xuICAgICAgICBhd2FpdCBQcm9taXNlLnJhY2UoZXhlY3V0aW5nKTtcbiAgICAgIH1cbiAgICB9XG4gICAgLy8gV2FpdCB1bnRpbCBhbGwgb25nb2luZyBldmVudHMgaGF2ZSBwcm9jZXNzZWQsIHRoZW4gY2xvc2UgdGhlIHdyaXRlci5cbiAgICBhd2FpdCBQcm9taXNlLmFsbChleGVjdXRpbmcpO1xuICAgIHdyaXRlci5jbG9zZSgpO1xuICB9KSgpO1xuICByZXR1cm4gcmVzLnJlYWRhYmxlLmdldEl0ZXJhdG9yKCk7XG59XG4iXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUEsRUFBQSx3RUFBQTtBQUVBLEVBUUEsQUFSQSwyWEFRQSxBQVJBLEVBUUEsaUJBQ0EsU0FBQSxDQUNBLFNBQUEsRUFDQSxLQUFBLEVBQ0EsVUFBQTtBQUVBLE1BQUEsK0RBQUE7VUFDQSxHQUFBLE9BQUEsZUFBQTtjQUNBLFNBQUEsRUFDQSxDQUFBLEVBQ0EsVUFBQTtBQUVBLHNCQUFBLENBQUEsT0FBQSxPQUFBLENBQUE7OztBQUdBLE1BQUEseUNBQUE7O2NBRUEsTUFBQSxHQUFBLEdBQUEsQ0FBQSxRQUFBLENBQUEsU0FBQTtjQUNBLFNBQUE7eUJBQ0EsSUFBQSxJQUFBLEtBQUE7a0JBQ0EsQ0FBQSxHQUFBLE9BQUEsQ0FBQSxPQUFBLEdBQUEsSUFBQSxLQUFBLFVBQUEsQ0FBQSxJQUFBOztBQUNBLGtCQUFBLENBQUEsS0FBQSxDQUFBLENBQUE7a0JBQ0EsQ0FBQSxHQUFBLENBQUEsQ0FBQSxJQUFBLEtBQ0EsU0FBQSxDQUFBLE1BQUEsQ0FBQSxTQUFBLENBQUEsT0FBQSxDQUFBLENBQUEsR0FBQSxDQUFBOztBQUVBLHFCQUFBLENBQUEsSUFBQSxDQUFBLENBQUE7Z0JBQ0EsU0FBQSxDQUFBLE1BQUEsSUFBQSxTQUFBO3NCQUNBLE9BQUEsQ0FBQSxJQUFBLENBQUEsU0FBQTs7O0FBR0EsVUFBQSxxRUFBQTtjQUNBLE9BQUEsQ0FBQSxHQUFBLENBQUEsU0FBQTtBQUNBLGNBQUEsQ0FBQSxLQUFBOztXQUVBLEdBQUEsQ0FBQSxRQUFBLENBQUEsV0FBQSJ9