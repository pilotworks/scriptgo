// Node.js Async Hooks module (node:async_hooks)

export class AsyncHook {
    private _enabled: boolean = false;

    constructor() {
        this._enabled = false;
    }

    enable(): this {
        this._enabled = true;
        return this;
    }

    disable(): this {
        this._enabled = false;
        return this;
    }

    executionAsyncResource(): unknown {
        return null;
    }

    executionAsyncId(): number {
        return 1;
    }

    triggerAsyncId(): number {
        return 0;
    }

    return(): void {}
}

export function createHook(callbacks?: unknown): AsyncHook {
    return new AsyncHook();
}

export function executionAsyncId(): number {
    return 1;
}

export function triggerAsyncId(): number {
    return 0;
}

export function executionAsyncResource(): unknown {
    return null;
}

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

export class AsyncResource {
    type: string = "";
    private _asyncId: number = 1;
    private _triggerAsyncId: number = 0;

    constructor(type: string, triggerAsyncId: number = 0) {
        this.type = type;
        this._asyncId = 1;
        this._triggerAsyncId = triggerAsyncId;
    }

    static bind<Func extends Function>(fn: Func, type?: string, thisArg?: unknown): Func {
        return fn;
    }

    bind<Func extends Function>(fn: Func, thisArg?: unknown): Func {
        return fn;
    }

    runInAsyncScope<R, A = unknown>(fn: (arg: A) => R, thisArg?: unknown, arg?: A): R {
        return fn(arg as A);
    }

    emitDestroy(): this {
        return this;
    }

    asyncId(): number {
        return this._asyncId;
    }

    triggerAsyncId(): number {
        return this._triggerAsyncId;
    }
}

export default {
    createHook,
    AsyncHook,
    executionAsyncId,
    triggerAsyncId,
    executionAsyncResource,
    AsyncLocalStorage,
    AsyncResource,
};
