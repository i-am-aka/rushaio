export class SessionStorage {
    constructor() {
        this._storage = {};
    }
    get length() {
        return Object.keys(this._storage).length;
    }
    key(index) {
        return Object.keys(this._storage)[index];
    }
    getItem(key) {
        return this._storage[key];
    }
    setItem(key, value) {
        this._storage[key] = value;
    }
    removeItem(key) {
        delete this._storage[key];
    }
    clear() {
        this._storage = {};
    }
}
//# sourceMappingURL=file:///Users/andyh2/drive/ysfc/src/rush/task_single/eprotect/deno/gen/file/Users/andyh2/drive/ysfc/src/rush/task_single/eprotect/storage.ts.js.map