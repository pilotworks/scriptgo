// ScriptGo Standard Library: node:events

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
        if (idx < 0) {
            if (event === "error") {
                if (arg1 !== undefined) {
                    throw arg1;
                }
                throw new Error("Unhandled error event");
            }
            return false;
        }
        if (this._events[idx].listeners.length === 0) {
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

export class DOMException extends Error {
    name: string;
    message: string;
    code: number;

    constructor(message?: string, name?: string) {
        super(message || "");
        this.name = name || "Error";
        this.message = message || "";
        this.code = 0;
    }
}

export class Event {
    type: string;
    target: unknown = null;
    currentTarget: unknown = null;
    bubbles: boolean = false;
    cancelable: boolean = false;
    defaultPrevented: boolean = false;
    timeStamp: number;

    propagationStopped: boolean = false;
    immediatePropagationStopped: boolean = false;

    constructor(type: string, eventInitDict?: { bubbles?: boolean; cancelable?: boolean }) {
        this.type = type;
        this.bubbles = !!(eventInitDict && eventInitDict.bubbles);
        this.cancelable = !!(eventInitDict && eventInitDict.cancelable);
        this.timeStamp = Date.now();
    }

    preventDefault(): void {
        if (this.cancelable) {
            this.defaultPrevented = true;
        }
    }

    stopPropagation(): void {
        this.propagationStopped = true;
    }

    stopImmediatePropagation(): void {
        this.immediatePropagationStopped = true;
        this.propagationStopped = true;
    }
}

export class CustomEvent extends Event {
    detail: unknown;

    constructor(type: string, eventInitDict?: { bubbles?: boolean; cancelable?: boolean; detail?: unknown }) {
        super(type, eventInitDict);
        this.detail = eventInitDict && eventInitDict.detail !== undefined ? eventInitDict.detail : null;
    }
}

class TargetListenerEntry {
    listener: Function;
    once: boolean;

    constructor(listener: Function, once: boolean) {
        this.listener = listener;
        this.once = once;
    }
}

class TargetListenerBucket {
    type: string;
    entries: TargetListenerEntry[] = [];

    constructor(type: string) {
        this.type = type;
        this.entries = [];
    }
}

export class EventTarget {
    protected _targetBuckets: TargetListenerBucket[] = [];

    constructor() {
        this._targetBuckets = [];
    }

    private _findTargetBucketIndex(type: string): number {
        for (let i = 0; i < this._targetBuckets.length; i++) {
            if (this._targetBuckets[i].type === type) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateTargetBucketIndex(type: string): number {
        let idx = this._findTargetBucketIndex(type);
        if (idx < 0) {
            this._targetBuckets.push(new TargetListenerBucket(type));
            idx = this._targetBuckets.length - 1;
        }
        return idx;
    }

    addEventListener(type: string, callback: Function, options?: unknown): void {
        if (!callback) return;
        let once = false;
        if (typeof options === "boolean") {
            once = options as boolean;
        } else if (typeof options === "object" && options !== null) {
            once = !!(options as Record<string, unknown>).once;
        }
        const idx = this._getOrCreateTargetBucketIndex(type);
        const bucket = this._targetBuckets[idx];
        for (let i = 0; i < bucket.entries.length; i++) {
            if (bucket.entries[i].listener === callback) {
                return;
            }
        }
        bucket.entries.push(new TargetListenerEntry(callback, once));
    }

    removeEventListener(type: string, callback: Function, options?: unknown): void {
        if (!callback) return;
        const idx = this._findTargetBucketIndex(type);
        if (idx >= 0) {
            const bucket = this._targetBuckets[idx];
            const next: TargetListenerEntry[] = [];
            for (let i = 0; i < bucket.entries.length; i++) {
                if (bucket.entries[i].listener !== callback) {
                    next.push(bucket.entries[i]);
                }
            }
            bucket.entries = next;
        }
    }

    dispatchEvent(event: Event): boolean {
        event.target = this;
        event.currentTarget = this;
        const idx = this._findTargetBucketIndex(event.type);
        if (idx < 0) return true;
        const bucket = this._targetBuckets[idx];
        if (bucket.entries.length === 0) return true;

        const snapshot: TargetListenerEntry[] = [];
        for (let i = 0; i < bucket.entries.length; i++) {
            snapshot.push(bucket.entries[i]);
        }
        const remaining: TargetListenerEntry[] = [];
        for (let i = 0; i < bucket.entries.length; i++) {
            if (!bucket.entries[i].once) {
                remaining.push(bucket.entries[i]);
            }
        }
        bucket.entries = remaining;

        for (let i = 0; i < snapshot.length; i++) {
            if (event.immediatePropagationStopped) {
                break;
            }
            snapshot[i].listener(event);
        }
        return !event.defaultPrevented;
    }
}

export class AbortSignal extends EventTarget {
    aborted: boolean = false;
    reason: unknown = undefined;
    onabort: Function | null = null;

    constructor() {
        super();
        this._targetBuckets = [];
    }

    static abort(reason?: unknown): AbortSignal {
        const sig = new AbortSignal();
        sig._signalAbort(reason !== undefined ? reason : new DOMException("This operation was aborted", "AbortError"));
        return sig;
    }

    static timeout(delay: number): AbortSignal {
        const sig = new AbortSignal();
        setTimeout(() => {
            sig._signalAbort(new DOMException("The operation timed out", "TimeoutError"));
        }, delay);
        return sig;
    }

    static any(signals: AbortSignal[]): AbortSignal {
        const result = new AbortSignal();
        for (let i = 0; i < signals.length; i++) {
            if (signals[i].aborted) {
                result._signalAbort(signals[i].reason);
                return result;
            }
        }
        const onAnyAbort = (e: Event) => {
            const target = e.target as AbortSignal;
            result._signalAbort(target ? target.reason : new DOMException("This operation was aborted", "AbortError"));
        };
        for (let i = 0; i < signals.length; i++) {
            signals[i].addEventListener("abort", onAnyAbort, true);
        }
        return result;
    }

    throwIfAborted(): void {
        if (this.aborted) {
            throw this.reason;
        }
    }

    _signalAbort(reason: unknown): void {
        if (this.aborted) return;
        this.aborted = true;
        this.reason = reason;
        const ev = new Event("abort");
        if (this.onabort) {
            this.onabort(ev);
        }
        this.dispatchEvent(ev);
    }
}

export class AbortController {
    readonly signal: AbortSignal;

    constructor() {
        this.signal = new AbortSignal();
    }

    abort(reason?: unknown): void {
        this.signal._signalAbort(reason !== undefined ? reason : new DOMException("This operation was aborted", "AbortError"));
    }
}

export function getEventListeners(emitter: unknown, event: string): Function[] {
    if (emitter instanceof EventEmitter) {
        return (emitter as EventEmitter).listeners(event);
    }
    return [];
}

export function listenerCount(emitter: EventEmitter, event: string): number {
    return emitter.listenerCount(event);
}

export function once(emitter: unknown, event: string): Promise<unknown[]> {
    return new Promise((resolve, reject) => {
        const handler = (...args: unknown[]) => {
            resolve(args);
        };
        if (emitter instanceof EventEmitter) {
            (emitter as EventEmitter).once(event, handler);
        } else if (emitter instanceof EventTarget) {
            (emitter as EventTarget).addEventListener(event, handler, { once: true });
        }
    });
}

export function on(emitter: unknown, event: string): unknown {
    return {
        [Symbol.asyncIterator]() {
            return {
                next(): Promise<IteratorResult<unknown>> {
                    return once(emitter, event).then((val) => ({ value: val, done: false }));
                }
            };
        }
    };
}

export const defaultMaxListeners = 10;
export const captureRejectionsSymbol = Symbol("captureRejections");
export const errorMonitor = Symbol("events.errorMonitor");

export default {
    EventEmitter,
    AbortController,
    AbortSignal,
    EventTarget,
    Event,
    CustomEvent,
    getEventListeners,
    listenerCount,
    once,
    on,
    defaultMaxListeners,
    captureRejectionsSymbol,
    errorMonitor,
};

