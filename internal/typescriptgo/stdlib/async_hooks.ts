// Node.js Async Hooks module (node:async_hooks)

export class AsyncLocalStorage<T = unknown> {
    private _store: unknown = undefined;

    constructor() {
        this._store = undefined;
    }

    disable(): void {
        this._store = undefined;
    }

    getStore(): unknown {
        return this._store;
    }

    enterWith(store: unknown): void {
        this._store = store;
    }

    run<R>(store: unknown, callback: () => R): R {
        const prev = this._store;
        this._store = store;
        try {
            return callback();
        } finally {
            this._store = prev;
        }
    }

    exit<R>(callback: () => R): R {
        const prev = this._store;
        this._store = undefined;
        try {
            return callback();
        } finally {
            this._store = prev;
        }
    }
}

export default {
    AsyncLocalStorage,
};
