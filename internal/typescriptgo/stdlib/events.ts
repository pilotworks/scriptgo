class ListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class EventBucket {
    name: string;
    listeners: ListenerEntry[];

    constructor(name: string, listeners: ListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export class EventEmitter {
    static defaultMaxListeners: number = 10;

    private _events: EventBucket[] = [];
    private _maxListeners: number = 10;

    static listenerCount(emitter: EventEmitter, event: string): number {
        return emitter.listenerCount(event);
    }

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
            this._events.push(new EventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    addListener(event: string, listener: Function): EventEmitter {
        return this.on(event, listener);
    }

    on(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new ListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new ListenerEntry(listener, true));
        return this;
    }

    prependListener(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new ListenerEntry(listener, false));
        return this;
    }

    prependOnceListener(event: string, listener: Function): EventEmitter {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new ListenerEntry(listener, true));
        return this;
    }

    removeListener(event: string, listener: Function): EventEmitter {
        return this.off(event, listener);
    }

    off(event: string, listener: Function): EventEmitter {
        const idx = this._findBucketIndex(event);
        if (idx >= 0) {
            const bucket = this._events[idx];
            const nextListeners: ListenerEntry[] = [];
            let removed = false;
            for (let i = 0; i < bucket.listeners.length; i++) {
                const entry = bucket.listeners[i];
                if (!removed && entry.fn === listener) {
                    removed = true;
                } else {
                    nextListeners.push(entry);
                }
            }
            bucket.listeners = nextListeners;
        }
        return this;
    }

    removeAllListeners(event?: string): EventEmitter {
        if (event !== undefined && event !== "") {
            const nextEvents: EventBucket[] = [];
            for (let i = 0; i < this._events.length; i++) {
                if (this._events[i].name !== event) {
                    nextEvents.push(this._events[i]);
                }
            }
            this._events = nextEvents;
        } else {
            this._events = [];
        }
        return this;
    }

    setMaxListeners(n: number): EventEmitter {
        this._maxListeners = n;
        return this;
    }

    getMaxListeners(): number {
        return this._maxListeners;
    }

    listenerCount(event: string): number {
        const idx = this._findBucketIndex(event);
        if (idx < 0) {
            return 0;
        }
        return this._events[idx].listeners.length;
    }

    listeners(event: string): Function[] {
        const res: Function[] = [];
        const idx = this._findBucketIndex(event);
        if (idx >= 0) {
            const bucket = this._events[idx];
            for (let i = 0; i < bucket.listeners.length; i++) {
                res.push(bucket.listeners[i].fn);
            }
        }
        return res;
    }

    rawListeners(event: string): Function[] {
        return this.listeners(event);
    }

    eventNames(): string[] {
        const names: string[] = [];
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].listeners.length > 0) {
                names.push(this._events[i].name);
            }
        }
        return names;
    }

    emit(event: string, arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0 || this._events[idx].listeners.length === 0) {
            if (event === "error") {
                if (arg1 !== undefined) {
                    throw arg1;
                }
                throw new Error("Unhandled error event");
            }
            return false;
        }

        const bucket = this._events[idx];
        const snapshot: ListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            snapshot.push(bucket.listeners[i]);
        }

        const remaining: ListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                remaining.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = remaining;

        for (let i = 0; i < snapshot.length; i++) {
            const fn = snapshot[i].fn;
            if (arg1 === undefined) {
                fn();
            } else if (arg2 === undefined) {
                fn(arg1);
            } else if (arg3 === undefined) {
                fn(arg1, arg2);
            } else if (arg4 === undefined) {
                fn(arg1, arg2, arg3);
            } else {
                fn(arg1, arg2, arg3, arg4);
            }
        }

        return true;
    }
}
