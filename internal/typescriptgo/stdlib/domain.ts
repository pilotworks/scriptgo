// ScriptGo Standard Library: node:domain

class DomainListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class DomainEventBucket {
    name: string;
    listeners: DomainListenerEntry[];

    constructor(name: string, listeners: DomainListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export class EventEmitter {
    private _events: DomainEventBucket[] = [];

    private _findBucketIndex(event: string): number {
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].name === event) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateBucketIndex(event: string): number {
        let idx = this._findBucketIndex(event);
        if (idx < 0) {
            this._events.push(new DomainEventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    addListener(event: string, listener: Function): EventEmitter {
        return this.on(event, listener);
    }

    on(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new DomainListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new DomainListenerEntry(listener, true));
        return this;
    }

    removeListener(event: string, listener: Function): EventEmitter {
        const idx = this._findBucketIndex(event);
        if (idx >= 0) {
            const bucket = this._events[idx];
            const next: DomainListenerEntry[] = [];
            for (let i = 0; i < bucket.listeners.length; i++) {
                if (bucket.listeners[i].fn !== listener) {
                    next.push(bucket.listeners[i]);
                }
            }
            bucket.listeners = next;
        }
        return this;
    }

    off(event: string, listener: Function): EventEmitter {
        return this.removeListener(event, listener);
    }

    removeAllListeners(event?: string): EventEmitter {
        if (event !== undefined) {
            const idx = this._findBucketIndex(event);
            if (idx >= 0) {
                this._events[idx].listeners = [];
            }
        } else {
            this._events = [];
        }
        return this;
    }

    emit(event: string, arg1?: unknown, arg2?: unknown, arg3?: unknown): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0) {
            if (event === "error") {
                if (arg1 instanceof Error) {
                    throw arg1;
                }
                throw new Error("Unhandled 'error' event on EventEmitter");
            }
            return false;
        }
        const bucket = this._events[idx];
        const current = bucket.listeners;
        if (current.length === 0) {
            if (event === "error") {
                if (arg1 instanceof Error) {
                    throw arg1;
                }
                throw new Error("Unhandled 'error' event on EventEmitter");
            }
            return false;
        }
        const remaining: DomainListenerEntry[] = [];
        for (let i = 0; i < current.length; i++) {
            const entry = current[i];
            if (!entry.once) {
                remaining.push(entry);
            }
            const fn = entry.fn;
            if (arg1 === undefined) {
                fn();
            } else if (arg2 === undefined) {
                fn(arg1);
            } else if (arg3 === undefined) {
                fn(arg1, arg2);
            } else {
                fn(arg1, arg2, arg3);
            }
        }
        bucket.listeners = remaining;
        return true;
    }
}

let _domainStack: Domain[] = [];
export let active: Domain | null = null;

export class Domain extends EventEmitter {
    members: unknown[] = [];
    _disposed: boolean = false;

    constructor() {
        super();
        this.members = [];
    }

    add(emitter: unknown): void {
        if (emitter !== null && emitter !== undefined) {
            this.members.push(emitter);
        }
    }

    remove(emitter: unknown): void {
        const next: unknown[] = [];
        for (let i = 0; i < this.members.length; i++) {
            if (this.members[i] !== emitter) {
                next.push(this.members[i]);
            }
        }
        this.members = next;
    }

    enter(): void {
        if (this._disposed) return;
        active = this;
        _domainStack.push(this);
    }

    exit(): void {
        if (_domainStack.length > 0) {
            _domainStack.pop();
            if (_domainStack.length > 0) {
                active = _domainStack[_domainStack.length - 1];
            } else {
                active = null;
            }
        } else {
            active = null;
        }
    }

    bind(callback: Function): Function {
        const self = this;
        const bound = (arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): unknown => {
            self.enter();
            try {
                let res: unknown = undefined;
                if (arg1 === undefined) res = callback();
                else if (arg2 === undefined) res = callback(arg1);
                else if (arg3 === undefined) res = callback(arg1, arg2);
                else if (arg4 === undefined) res = callback(arg1, arg2, arg3);
                else res = callback(arg1, arg2, arg3, arg4);
                self.exit();
                return res;
            } catch (err) {
                self.exit();
                self.emit("error", err);
                return undefined;
            }
        };
        return bound;
    }

    intercept(callback: Function): Function {
        const self = this;
        const intercepted = (err?: unknown, arg1?: unknown, arg2?: unknown, arg3?: unknown): unknown => {
            if (err) {
                self.emit("error", err);
                return undefined;
            }
            self.enter();
            try {
                let res: unknown = undefined;
                if (arg1 === undefined) res = callback();
                else if (arg2 === undefined) res = callback(arg1);
                else if (arg3 === undefined) res = callback(arg1, arg2);
                else res = callback(arg1, arg2, arg3);
                self.exit();
                return res;
            } catch (thrown) {
                self.exit();
                self.emit("error", thrown);
                return undefined;
            }
        };
        return intercepted;
    }

    run(fn: Function, arg1?: unknown, arg2?: unknown, arg3?: unknown): unknown {
        this.enter();
        try {
            let res: unknown = undefined;
            if (arg1 === undefined) res = fn();
            else if (arg2 === undefined) res = fn(arg1);
            else if (arg3 === undefined) res = fn(arg1, arg2);
            else res = fn(arg1, arg2, arg3);
            this.exit();
            return res;
        } catch (err) {
            this.exit();
            this.emit("error", err);
            return undefined;
        }
    }

    dispose(): void {
        this.removeAllListeners();
        this.members = [];
        this._disposed = true;
        this.exit();
    }
}

export function create(): Domain {
    return new Domain();
}

export namespace domain {
    export function create(): Domain {
        return new Domain();
    }
    export let active: Domain | null = null;
}
