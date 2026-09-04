// ScriptGo Standard Library: node:events

export type EventKey = string | symbol;

export interface Disposable {
    [Symbol.dispose](): void;
}

class ListenerEntry {
    fn: Function;
    once: boolean;
    raw: Function;

    constructor(fn: Function, once: boolean, raw?: Function) {
        this.fn = fn;
        this.once = once;
        this.raw = raw !== undefined ? raw : fn;
    }
}

class EventBucket {
    name: EventKey;
    listeners: ListenerEntry[];

    constructor(name: EventKey, listeners: ListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export const captureRejectionSymbol: symbol = Symbol("captureRejections");
export const errorMonitor: symbol = Symbol("events.errorMonitor");

export class EventEmitter {
    static defaultMaxListeners: number = 10;
    static captureRejectionSymbol: symbol = captureRejectionSymbol;
    static captureRejections: boolean = false;
    static errorMonitor: symbol = errorMonitor;

    private _events: EventBucket[] = [];
    private _maxListeners: number | undefined = undefined;
    captureRejections: boolean = false;
    [captureRejectionSymbol]?: Function;

    static listenerCount(emitter: EventEmitter, event: EventKey): number {
        return emitter.listenerCount(event);
    }

    static getEventListeners(emitter: AnyEventEmitter, event: EventKey): Function[] {
        return getEventListeners(emitter, event);
    }

    static getMaxListeners(emitter: EventEmitter): number {
        return getMaxListeners(emitter);
    }

    static setMaxListeners(n: number, ...emitters: EventEmitter[]): void {
        setMaxListeners(n, ...emitters);
    }

    static once(emitter: AnyEventEmitter, event: EventKey, options?: { signal?: AbortSignal }): Promise<unknown[]> {
        return once(emitter, event, options);
    }

    static on(emitter: AnyEventEmitter, event: EventKey, options?: { signal?: AbortSignal }): AsyncIterableIterator<unknown[]> {
        return on(emitter, event, options);
    }

    static addAbortListener(signal: AbortSignalLike | EventTarget, resource: Function): Disposable {
        return addAbortListener(signal, resource);
    }

    private _findBucketIndex(event: EventKey): number {
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].name === event) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateBucketIndex(event: EventKey): number {
        let idx = this._findBucketIndex(event);
        if (idx < 0) {
            this._events.push(new EventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    addListener(event: EventKey, listener: Function): this {
        return this.on(event, listener);
    }

    on(event: EventKey, listener: Function): this {
        if (typeof listener !== "function") {
            throw new TypeError("The \"listener\" argument must be of type function.");
        }
        if (this.listenerCount("newListener") > 0) {
            this.emit("newListener", event, listener);
        }
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new ListenerEntry(listener, false, listener));
        return this;
    }

    once(event: EventKey, listener: Function): this {
        if (typeof listener !== "function") {
            throw new TypeError("The \"listener\" argument must be of type function.");
        }
        if (this.listenerCount("newListener") > 0) {
            this.emit("newListener", event, listener);
        }
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new ListenerEntry(listener, true, listener));
        return this;
    }

    prependListener(event: EventKey, listener: Function): this {
        if (typeof listener !== "function") {
            throw new TypeError("The \"listener\" argument must be of type function.");
        }
        if (this.listenerCount("newListener") > 0) {
            this.emit("newListener", event, listener);
        }
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new ListenerEntry(listener, false, listener));
        return this;
    }

    prependOnceListener(event: EventKey, listener: Function): this {
        if (typeof listener !== "function") {
            throw new TypeError("The \"listener\" argument must be of type function.");
        }
        if (this.listenerCount("newListener") > 0) {
            this.emit("newListener", event, listener);
        }
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new ListenerEntry(listener, true, listener));
        return this;
    }

    removeListener(event: EventKey, listener: Function): this {
        if (typeof listener !== "function") {
            throw new TypeError("The \"listener\" argument must be of type function.");
        }
        const idx = this._findBucketIndex(event);
        if (idx >= 0) {
            const bucket = this._events[idx];
            const nextListeners: ListenerEntry[] = [];
            let removed = false;
            for (let i = 0; i < bucket.listeners.length; i++) {
                const entry = bucket.listeners[i];
                if (!removed && (entry.fn === listener || entry.raw === listener)) {
                    removed = true;
                } else {
                    nextListeners.push(entry);
                }
            }
            bucket.listeners = nextListeners;
            if (removed && this.listenerCount("removeListener") > 0) {
                this.emit("removeListener", event, listener);
            }
        }
        return this;
    }

    off(event: EventKey, listener: Function): this {
        return this.removeListener(event, listener);
    }

    removeAllListeners(event?: EventKey): this {
        if (event !== undefined) {
            const idx = this._findBucketIndex(event);
            if (idx >= 0) {
                const bucket = this._events[idx];
                const snapshot = bucket.listeners.slice();
                bucket.listeners = [];
                if (this.listenerCount("removeListener") > 0) {
                    for (let i = 0; i < snapshot.length; i++) {
                        this.emit("removeListener", event, snapshot[i].raw || snapshot[i].fn);
                    }
                }
            }
        } else {
            const oldEvents = this._events;
            this._events = [];
            if (this.listenerCount("removeListener") > 0) {
                for (let b = 0; b < oldEvents.length; b++) {
                    const bucket = oldEvents[b];
                    for (let i = 0; i < bucket.listeners.length; i++) {
                        this.emit("removeListener", bucket.name as EventKey, bucket.listeners[i].raw || bucket.listeners[i].fn);
                    }
                }
            }
        }
        return this;
    }

    setMaxListeners(n: number): this {
        if (typeof n !== "number" || n < 0 || Number.isNaN(n)) {
            throw new RangeError("The value of \"n\" is out of range. It must be a non-negative number.");
        }
        this._maxListeners = n;
        return this;
    }

    getMaxListeners(): number {
        if (this._maxListeners !== undefined) {
            return this._maxListeners;
        }
        return EventEmitter.defaultMaxListeners;
    }

    listeners(event: EventKey): Function[] {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return [];
        const res: Function[] = [];
        for (let i = 0; i < this._events[idx].listeners.length; i++) {
            res.push(this._events[idx].listeners[i].raw || this._events[idx].listeners[i].fn);
        }
        return res;
    }

    rawListeners(event: EventKey): Function[] {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return [];
        const res: Function[] = [];
        for (let i = 0; i < this._events[idx].listeners.length; i++) {
            res.push(this._events[idx].listeners[i].raw);
        }
        return res;
    }

    emit(event: EventKey, arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): boolean {
        if (event === "error") {
            const errorMonitorIdx = this._findBucketIndex(errorMonitor);
            if (errorMonitorIdx >= 0 && this._events[errorMonitorIdx].listeners.length > 0) {
                const emListeners = this._events[errorMonitorIdx].listeners.slice();
                for (let i = 0; i < emListeners.length; i++) {
                    this._callListener(emListeners[i].fn, arg1, arg2, arg3, arg4);
                }
            }

            const errIdx = this._findBucketIndex("error");
            if (errIdx < 0 || this._events[errIdx].listeners.length === 0) {
                const err = arg1;
                if (err instanceof Error) {
                    throw err;
                }
                if (arg1 !== undefined) {
                    throw new Error("Unhandled error: " + String(err));
                }
                throw new Error("Unhandled error event");
            }
        }

        const idx = this._findBucketIndex(event);
        if (idx < 0) return false;
        const bucket = this._events[idx];
        if (bucket.listeners.length === 0) return false;

        const snapshot = bucket.listeners.slice();
        const next: ListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                next.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = next;

        const capture = this.captureRejections || EventEmitter.captureRejections;
        for (let i = 0; i < snapshot.length; i++) {
            try {
                const result: unknown = this._callListener(snapshot[i].fn, arg1, arg2, arg3, arg4);
                if (capture && result !== null && typeof result === "object" && typeof (result as Record<string, unknown>)["catch"] === "function") {
                    (result as Promise<unknown>).catch((rejectionErr: unknown) => {
                        const handler = this[EventEmitter.captureRejectionSymbol];
                        if (typeof handler === "function") {
                            handler.call(this, rejectionErr, event, arg1);
                        } else {
                            this.emit("error", rejectionErr);
                        }
                    });
                }
            } catch (syncErr) {
                if (capture) {
                    const handler = this[EventEmitter.captureRejectionSymbol];
                    if (typeof handler === "function") {
                        handler.call(this, syncErr, event, arg1);
                        continue;
                    }
                }
                throw syncErr;
            }
        }
        return snapshot.length > 0;
    }

    private _callListener(fn: Function, a1?: unknown, a2?: unknown, a3?: unknown, a4?: unknown): unknown {
        if (a1 === undefined) {
            return fn();
        } else if (a2 === undefined) {
            return fn(a1);
        } else if (a3 === undefined) {
            return fn(a1, a2);
        } else if (a4 === undefined) {
            return fn(a1, a2, a3);
        } else {
            return fn(a1, a2, a3, a4);
        }
    }

    listenerCount(event: EventKey): number {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return 0;
        return this._events[idx].listeners.length;
    }

    eventNames(): EventKey[] {
        const res: EventKey[] = [];
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].listeners.length > 0) {
                res.push(this._events[i].name);
            }
        }
        return res;
    }
}

export class EventEmitterAsyncResource extends EventEmitter {
    name: string;
    readonly asyncId: number = 1;
    readonly triggerAsyncId: number = 0;
    readonly asyncResource: { type: string };

    constructor(options?: { name?: string; triggerAsyncId?: number }) {
        super();
        this.name = options && options.name ? options.name : "EventEmitterAsyncResource";
        this.asyncResource = { type: this.name };
        if (options && options.triggerAsyncId !== undefined) {
            this.triggerAsyncId = options.triggerAsyncId;
        }
    }

    destroyed: boolean = false;

    emitDestroy(): void {
        this.destroyed = true;
    }
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
    target: EventTarget | null = null;
    currentTarget: EventTarget | null = null;
    srcElement: EventTarget | null = null;
    returnValue: boolean = true;
    cancelBubble: boolean = false;

    constructor(type: string, eventInitDict?: { bubbles?: boolean; cancelable?: boolean; composed?: boolean }) {
        this.type = type;
        if (eventInitDict) {
            if (eventInitDict.bubbles !== undefined) this.bubbles = eventInitDict.bubbles;
            if (eventInitDict.cancelable !== undefined) this.cancelable = eventInitDict.cancelable;
            if (eventInitDict.composed !== undefined) this.composed = eventInitDict.composed;
        }
        this.timeStamp = Date.now();
    }

    preventDefault(): void {
        if (this.cancelable) {
            this.defaultPrevented = true;
            this.returnValue = false;
        }
    }

    stopPropagation(): void {
        this.cancelBubble = true;
    }

    stopImmediatePropagation(): void {
        this.cancelBubble = true;
    }

    composedPath(): EventTarget[] {
        if (this.target) {
            return [this.target];
        }
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

    constructor(type: string, eventInitDict?: { detail?: unknown; bubbles?: boolean; cancelable?: boolean; composed?: boolean }) {
        super(type, eventInitDict);
        if (eventInitDict && eventInitDict.detail !== undefined) {
            this.detail = eventInitDict.detail;
        }
    }
}

class EventTargetListener {
    type: string;
    callback: Function;
    once: boolean;
    signal: AbortSignal | null;

    constructor(type: string, callback: Function, once: boolean, signal: AbortSignal | null) {
        this.type = type;
        this.callback = callback;
        this.once = once;
        this.signal = signal;
    }
}

export interface AddEventListenerOptions {
    once?: boolean;
    signal?: AbortSignal;
}

export class EventTarget {
    _listeners: EventTargetListener[] = [];

    addEventListener(type: string, callback: Function, options?: AddEventListenerOptions | boolean): void {
        if (typeof callback !== "function") {
            return;
        }
        let once = false;
        let signal: AbortSignal | null = null;
        if (options && typeof options === "object") {
            const opts = options as AddEventListenerOptions;
            if (opts.once === true) {
                once = true;
            }
            if (opts.signal) {
                signal = opts.signal;
                if (signal.aborted) {
                    return;
                }
                const abortHandler = () => {
                    this.removeEventListener(type, callback);
                };
                signal.addEventListener("abort", abortHandler, { once: true });
            }
        }
        this._listeners.push(new EventTargetListener(type, callback, once, signal));
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
        event.srcElement = this;
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
                try {
                    snapshot[i].callback(event);
                } catch (e) {
                    // DOM EventTarget dispatch continues
                }
            }
        }
        return !event.defaultPrevented;
    }
}

export class NodeEventTarget extends EventTarget {
    addListener(type: string, listener: Function): this {
        this.addEventListener(type, listener);
        return this;
    }

    emit(type: string, arg?: unknown): boolean {
        return this.dispatchEvent(new Event(type));
    }

    eventNames(): string[] {
        const names: string[] = [];
        for (let i = 0; i < this._listeners.length; i++) {
            const t = this._listeners[i].type;
            if (names.indexOf(t) === -1) {
                names.push(t);
            }
        }
        return names;
    }

    listenerCount(type: string): number {
        let count = 0;
        for (let i = 0; i < this._listeners.length; i++) {
            if (this._listeners[i].type === type) {
                count++;
            }
        }
        return count;
    }

    private _maxListeners: number = 10;

    setMaxListeners(n: number): this {
        this._maxListeners = n;
        return this;
    }

    getMaxListeners(): number {
        return this._maxListeners;
    }

    off(type: string, listener: Function): this {
        return this.removeListener(type, listener);
    }

    on(type: string, listener: Function): this {
        return this.addListener(type, listener);
    }

    once(type: string, listener: Function): this {
        this.addEventListener(type, listener, { once: true });
        return this;
    }

    removeAllListeners(type?: string): this {
        if (type !== undefined) {
            const remaining: EventTargetListener[] = [];
            for (let i = 0; i < this._listeners.length; i++) {
                if (this._listeners[i].type !== type) {
                    remaining.push(this._listeners[i]);
                }
            }
            this._listeners = remaining;
        } else {
            this._listeners = [];
        }
        return this;
    }

    removeListener(type: string, listener: Function): this {
        this.removeEventListener(type, listener);
        return this;
    }
}

export class AbortSignal extends EventTarget {
    aborted: boolean = false;
    reason: unknown = undefined;
    onabort: Function | null = null;

    static abort(reason: unknown = undefined): AbortSignal {
        const sig = new AbortSignal();
        sig.aborted = true;
        sig.reason = reason !== undefined ? reason : new Error("The operation was aborted");
        return sig;
    }

    static timeout(delay: number): AbortSignal {
        const sig = new AbortSignal();
        setTimeout(() => {
            if (!sig.aborted) {
                sig.aborted = true;
                sig.reason = new Error("The operation was aborted due to timeout");
                if (sig.onabort) {
                    sig.onabort();
                }
                sig.dispatchEvent(new Event("abort"));
            }
        }, delay);
        return sig;
    }

    static any(signals: Iterable<AbortSignal> | AbortSignal[]): AbortSignal {
        const sig = new AbortSignal();
        const arr = Array.isArray(signals) ? signals : Array.from(signals);
        for (let i = 0; i < arr.length; i++) {
            const s = arr[i];
            if (s.aborted) {
                sig.aborted = true;
                sig.reason = s.reason;
                return sig;
            }
            s.addEventListener("abort", () => {
                if (!sig.aborted) {
                    sig.aborted = true;
                    sig.reason = s.reason;
                    if (sig.onabort) {
                        sig.onabort();
                    }
                    sig.dispatchEvent(new Event("abort"));
                }
            });
        }
        return sig;
    }

    throwIfAborted(): void {
        if (this.aborted) {
            throw this.reason !== undefined ? this.reason : new Error("The operation was aborted");
        }
    }
}

export class AbortController {
    signal: AbortSignal;

    constructor() {
        this.signal = new AbortSignal();
    }

    abort(reason: unknown = undefined): void {
        if (this.signal.aborted) {
            return;
        }
        this.signal.aborted = true;
        this.signal.reason = reason !== undefined ? reason : new Error("The operation was aborted");
        if (this.signal.onabort) {
            this.signal.onabort();
        }
        this.signal.dispatchEvent(new Event("abort"));
    }
}

export type AnyEventEmitter = EventEmitter | EventTarget;

export function getEventListeners(emitter: AnyEventEmitter, event: EventKey): Function[] {
    if (emitter instanceof EventEmitter) {
        return (emitter as EventEmitter).listeners(event);
    }
    if (emitter instanceof EventTarget) {
        const res: Function[] = [];
        const listeners = (emitter as EventTarget)._listeners;
        if (Array.isArray(listeners)) {
            const eventStr = String(event);
            for (let i = 0; i < listeners.length; i++) {
                if (listeners[i].type === eventStr) {
                    res.push(listeners[i].callback);
                }
            }
        }
        return res;
    }
    return [];
}

export function listenerCount(emitter: EventEmitter, event: EventKey): number {
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

export function once(
    emitter: AnyEventEmitter,
    event: EventKey,
    options?: { signal?: AbortSignal }
): Promise<unknown[]> {
    if (options && options.signal && options.signal.aborted) {
        return Promise.reject(options.signal.reason || new Error("The operation was aborted"));
    }

    return new Promise<unknown[]>((resolve, reject) => {
        let signalListener: Function | null = null;
        let eventHandler: Function;
        let errorHandler: Function;

        const cleanup = () => {
            if (emitter instanceof EventEmitter) {
                (emitter as EventEmitter).removeListener(event, eventHandler);
                (emitter as EventEmitter).removeListener("error", errorHandler);
            } else if (emitter instanceof EventTarget) {
                (emitter as EventTarget).removeEventListener(String(event), eventHandler);
            }
            if (options && options.signal && signalListener) {
                options.signal.removeEventListener("abort", signalListener);
            }
        };

        eventHandler = (a1?: unknown, a2?: unknown, a3?: unknown, a4?: unknown) => {
            cleanup();
            const args: unknown[] = [];
            if (a1 !== undefined) args.push(a1);
            if (a2 !== undefined) args.push(a2);
            if (a3 !== undefined) args.push(a3);
            if (a4 !== undefined) args.push(a4);
            resolve(args);
        };

        errorHandler = (err: Error) => {
            cleanup();
            reject(err);
        };

        if (options && options.signal) {
            signalListener = () => {
                cleanup();
                reject(options.signal!.reason || new Error("The operation was aborted"));
            };
            options.signal.addEventListener("abort", signalListener, { once: true });
        }

        if (emitter instanceof EventEmitter) {
            (emitter as EventEmitter).on(event, eventHandler);
            if (event !== "error") {
                (emitter as EventEmitter).once("error", errorHandler);
            }
        } else if (emitter instanceof EventTarget) {
            (emitter as EventTarget).addEventListener(String(event), (ev: Event) => {
                cleanup();
                resolve([ev]);
            }, { once: true });
        } else {
            cleanup();
            reject(new TypeError("The \"emitter\" argument must be an instance of EventEmitter or EventTarget"));
        }
    });
}

export function on(
    emitter: AnyEventEmitter,
    event: EventKey,
    options?: { signal?: AbortSignal }
): AsyncIterableIterator<unknown[]> {
    const queue: unknown[][] = [];
    const waiters: Array<{ resolve: (val: IteratorResult<unknown[]>) => void; reject: (err: Error) => void }> = [];
    let finished = false;
    let error: Error | null = null;

    const pushValue = (args: unknown[]) => {
        if (waiters.length > 0) {
            const waiter = waiters.shift()!;
            waiter.resolve({ value: args, done: false });
        } else {
            queue.push(args);
        }
    };

    const pushError = (err: Error) => {
        finished = true;
        error = err;
        cleanup();
        while (waiters.length > 0) {
            const waiter = waiters.shift()!;
            waiter.reject(err);
        }
    };

    const eventListener = (a1?: unknown, a2?: unknown, a3?: unknown, a4?: unknown) => {
        const args: unknown[] = [];
        if (a1 !== undefined) args.push(a1);
        if (a2 !== undefined) args.push(a2);
        if (a3 !== undefined) args.push(a3);
        if (a4 !== undefined) args.push(a4);
        pushValue(args);
    };

    const errorListener = (err: Error) => {
        pushError(err);
    };

    let signalListener: Function | null = null;
    if (options && options.signal) {
        const sig = options.signal;
        const getReasonError = (): Error => {
            const r = sig.reason;
            return r instanceof Error ? r : new Error(String(r || "The operation was aborted"));
        };
        if (sig.aborted) {
            pushError(getReasonError());
        } else {
            signalListener = () => {
                pushError(getReasonError());
            };
            sig.addEventListener("abort", signalListener, { once: true });
        }
    }

    if (emitter instanceof EventEmitter) {
        (emitter as EventEmitter).on(event, eventListener);
        if (event !== "error") {
            (emitter as EventEmitter).once("error", errorListener);
        }
    } else if (emitter instanceof EventTarget) {
        (emitter as EventTarget).addEventListener(String(event), (ev: Event) => {
            pushValue([ev]);
        });
    }

    const cleanup = () => {
        if (emitter instanceof EventEmitter) {
            (emitter as EventEmitter).removeListener(event, eventListener);
            (emitter as EventEmitter).removeListener("error", errorListener);
        } else if (emitter instanceof EventTarget) {
            (emitter as EventTarget).removeEventListener(String(event), eventListener);
        }
        if (options && options.signal && signalListener) {
            options.signal.removeEventListener("abort", signalListener);
        }
    };

    const iterator: AsyncIterableIterator<unknown[]> = {
        [Symbol.asyncIterator]() {
            return iterator;
        },
        next(): Promise<IteratorResult<unknown[]>> {
            if (error !== null) {
                return Promise.reject(error);
            }
            if (queue.length > 0) {
                const val = queue.shift()!;
                return Promise.resolve({ value: val, done: false });
            }
            if (finished) {
                return Promise.resolve({ value: undefined as unknown as unknown[], done: true });
            }
            return new Promise((resolve, reject) => {
                waiters.push({ resolve, reject });
            });
        },
        return(): Promise<IteratorResult<unknown[]>> {
            finished = true;
            cleanup();
            while (waiters.length > 0) {
                const waiter = waiters.shift()!;
                waiter.resolve({ value: undefined as unknown as unknown[], done: true });
            }
            return Promise.resolve({ value: undefined as unknown as unknown[], done: true });
        },
        throw(err?: Error): Promise<IteratorResult<unknown[]>> {
            finished = true;
            cleanup();
            const errObj: Error = err || new Error("Iterator was closed");
            while (waiters.length > 0) {
                const waiter = waiters.shift()!;
                waiter.reject(errObj);
            }
            return Promise.reject(errObj);
        }
    };

    return iterator;
}

export interface AbortSignalLike {
    aborted?: boolean;
    addEventListener(type: string, listener: unknown, options?: unknown): void;
    removeEventListener(type: string, listener?: unknown, options?: unknown): void;
}

export function addAbortListener(signal: AbortSignalLike | EventTarget, resource: Function): Disposable {
    if (!signal) {
        throw new TypeError("The \"signal\" argument must be an instance of AbortSignal");
    }
    if (typeof resource !== "function") {
        throw new TypeError("The \"resource\" argument must be a function");
    }
    const target = signal as EventTarget;
    target.addEventListener("abort", resource, { once: true });
    return {
        [Symbol.dispose]() {
            target.removeEventListener("abort", resource);
        }
    };
}

export const defaultMaxListeners = 10;
export const captureRejections = false;

export default {
    EventEmitter,
    EventEmitterAsyncResource,
    EventTarget,
    NodeEventTarget,
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
