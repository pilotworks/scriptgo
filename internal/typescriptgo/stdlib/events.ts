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
    static captureRejectionSymbol: symbol = Symbol("captureRejections");
    static captureRejections: boolean = false;
    static errorMonitor: symbol = Symbol("events.errorMonitor");

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
        if (event !== undefined && event.length > 0) {
            const idx = this._findBucketIndex(event);
            if (idx >= 0) {
                this._events[idx].listeners = [];
            }
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

    listeners(event: string): Function[] {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return [];
        const res: Function[] = [];
        for (let i = 0; i < this._events[idx].listeners.length; i++) {
            res.push(this._events[idx].listeners[i].fn);
        }
        return res;
    }

    rawListeners(event: string): Function[] {
        return this.listeners(event);
    }

    emit(event: string, arg1?: unknown, arg2?: unknown, arg3?: unknown): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return false;
        const bucket = this._events[idx];
        const snapshot = bucket.listeners.slice();
        const next: ListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                next.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = next;
        for (let i = 0; i < snapshot.length; i++) {
            snapshot[i].fn(arg1, arg2, arg3);
        }
        return snapshot.length > 0;
    }

    listenerCount(event: string): number {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return 0;
        return this._events[idx].listeners.length;
    }

    eventNames(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].listeners.length > 0) {
                res.push(this._events[i].name);
            }
        }
        return res;
    }
}

export class NodeEventTarget extends EventEmitter {}

export class EventEmitterAsyncResource extends EventEmitter {
    asyncId: number = 1;
    triggerAsyncId: number = 0;
    asyncResource: unknown = null;

    emitDestroy(): void {}
}

export class Event {
    type: string;
    bubbles: boolean = false;
    cancelable: boolean = false;
    composed: boolean = false;
    defaultPrevented: boolean = false;
    isTrusted: boolean = false;
    timeStamp: number = 0;
    eventPhase: number = 0;
    target: unknown = null;
    currentTarget: unknown = null;
    srcElement: unknown = null;
    returnValue: boolean = true;
    cancelBubble: boolean = false;

    constructor(type: string, eventInitDict?: Record<string, boolean>) {
        this.type = type;
        this.bubbles = false;
        this.cancelable = true;
        this.composed = false;
        this.defaultPrevented = false;
        this.isTrusted = false;
        this.returnValue = true;
        this.cancelBubble = false;
        this.eventPhase = 0;
        this.timeStamp = Date.now();
    }

    preventDefault(): void {
        this.defaultPrevented = true;
    }

    stopPropagation(): void {
        this.cancelBubble = true;
    }

    stopImmediatePropagation(): void {
        this.cancelBubble = true;
    }

    composedPath(): unknown[] {
        return [];
    }

    initEvent(type: string, bubbles: boolean = false, cancelable: boolean = false): void {
        this.type = type;
        this.bubbles = bubbles;
        this.cancelable = cancelable;
    }
}

export class CustomEvent extends Event {
    detail: unknown;

    constructor(type: string, eventInitDict?: Record<string, unknown>) {
        super(type);
        if (eventInitDict !== undefined && eventInitDict !== null) {
            this.detail = eventInitDict["detail"];
        }
    }
}

class EventTargetListener {
    type: string;
    callback: Function;
    once: boolean;

    constructor(type: string, callback: Function, once: boolean) {
        this.type = type;
        this.callback = callback;
        this.once = once;
    }
}

export class EventTarget {
    private _listeners: EventTargetListener[] = [];

    addEventListener(type: string, callback: Function, options: unknown = null): void {
        let once = false;
        if (options && typeof options === "object") {
            if ((options as Record<string, boolean>)["once"] === true) {
                once = true;
            }
        }
        this._listeners.push(new EventTargetListener(type, callback, once));
    }

    removeEventListener(type: string, callback: Function): void {
        const next: EventTargetListener[] = [];
        let removed = false;
        for (let i = 0; i < this._listeners.length; i++) {
            if (!removed && this._listeners[i].type === type && this._listeners[i].callback === callback) {
                removed = true;
            } else {
                next.push(this._listeners[i]);
            }
        }
        this._listeners = next;
    }

    dispatchEvent(event: Event): boolean {
        event.target = this;
        event.currentTarget = this;
        const snapshot = this._listeners.slice();
        const next: EventTargetListener[] = [];
        for (let i = 0; i < this._listeners.length; i++) {
            if (!this._listeners[i].once || this._listeners[i].type !== event.type) {
                next.push(this._listeners[i]);
            }
        }
        this._listeners = next;
        for (let i = 0; i < snapshot.length; i++) {
            if (snapshot[i].type === event.type) {
                snapshot[i].callback(event);
            }
        }
        return !event.defaultPrevented;
    }
}

export class AbortSignal extends EventTarget {
    aborted: boolean = false;
    reason: unknown = undefined;
    onabort: Function | null = null;

    static abort(reason: unknown = undefined): AbortSignal {
        const sig = new AbortSignal();
        sig.aborted = true;
        sig.reason = reason;
        return sig;
    }

    static timeout(delay: number): AbortSignal {
        const sig = new AbortSignal();
        return sig;
    }

    static any(signals: AbortSignal[]): AbortSignal {
        const sig = new AbortSignal();
        for (let i = 0; i < signals.length; i++) {
            const s = signals[i];
            if (s.aborted) {
                sig.aborted = true;
                sig.reason = s.reason;
                return sig;
            }
            s.addEventListener("abort", () => {
                if (!sig.aborted) {
                    sig.aborted = true;
                    sig.reason = s.reason;
                    sig.dispatchEvent(new Event("abort"));
                }
            });
        }
        return sig;
    }

    throwIfAborted(): void {
        if (this.aborted) {
            throw this.reason;
        }
    }
}

export class AbortController {
    signal: AbortSignal;

    constructor() {
        this.signal = new AbortSignal();
    }

    abort(reason: unknown = undefined): void {
        this.signal.aborted = true;
        this.signal.reason = reason;
        if (this.signal.onabort) {
            this.signal.onabort();
        }
        this.signal.dispatchEvent(new Event("abort"));
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

export function getMaxListeners(emitter: EventEmitter): number {
    return emitter.getMaxListeners();
}

export function setMaxListeners(n: number, ...emitters: EventEmitter[]): void {
    for (let i = 0; i < emitters.length; i++) {
        emitters[i].setMaxListeners(n);
    }
}

export function once(emitter: unknown, event: string): Promise<unknown> {
    return new Promise((resolve, reject) => {
        const handler = (val: unknown) => {
            resolve(val);
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

export function addAbortListener(signal: unknown, listener: Function): { [Symbol.dispose](): void } {
    return {
        [Symbol.dispose]() {}
    };
}

export const defaultMaxListeners = 10;
export const captureRejections = false;
export const captureRejectionSymbol = Symbol("captureRejections");
export const errorMonitor = Symbol("events.errorMonitor");

export default {
    EventEmitter,
    NodeEventTarget,
    EventEmitterAsyncResource,
    EventTarget,
    Event,
    CustomEvent,
    AbortSignal,
    AbortController,
    getEventListeners,
    getMaxListeners,
    setMaxListeners,
    listenerCount,
    once,
    on,
    addAbortListener,
    defaultMaxListeners,
    captureRejections,
    captureRejectionSymbol,
    errorMonitor,
};
