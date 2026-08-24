declare namespace __scriptgo {
    function streamGetDefaultHighWaterMark(objectMode: boolean): number;
    function streamSetDefaultHighWaterMark(objectMode: boolean, value: number): void;
}

class StreamListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class StreamEventBucket {
    name: string;
    listeners: StreamListenerEntry[];

    constructor(name: string, listeners: StreamListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

class StreamWriteReq {
    chunk: string;
    encoding: string;
    callback: Function;

    constructor(chunk: string, encoding: string, callback: Function) {
        this.chunk = chunk;
        this.encoding = encoding;
        this.callback = callback;
    }
}

class PipeOptions {
    end: boolean = true;
}

class AbortSignalLike {
    aborted: boolean = false;
    addEventListener: Function | null = null;
}

export class WebReadableStream {
    _stream: Readable | null = null;
    constructor(stream?: Readable | null) {
        this._stream = stream !== undefined ? stream : null;
    }
}

export class ReadableStream extends WebReadableStream {}

export class WebWritableStream {
    _stream: Writable | null = null;
    constructor(stream?: Writable | null) {
        this._stream = stream !== undefined ? stream : null;
    }
}

export class WritableStream extends WebWritableStream {}

export class WebDuplexStream {
    readable: WebReadableStream;
    writable: WebWritableStream;
    constructor(readable: WebReadableStream, writable: WebWritableStream) {
        this.readable = readable;
        this.writable = writable;
    }
}

export class Stream {
    static _defaultHighWaterMark: number = 65536;
    static _defaultObjectModeHighWaterMark: number = 16;

    readable: boolean = true;
    writable: boolean = true;
    destroyed: boolean = false;
    errored: Error | null = null;
    closed: boolean = false;

    private _events: StreamEventBucket[] = [];
    private _maxListeners: number = 10;

    constructor() {}

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
            this._events.push(new StreamEventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    addListener(event: string, listener: Function): this {
        return this.on(event, listener);
    }

    on(event: string, listener: Function): this {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new StreamListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): this {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new StreamListenerEntry(listener, true));
        return this;
    }

    prependListener(event: string, listener: Function): this {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new StreamListenerEntry(listener, false));
        return this;
    }

    prependOnceListener(event: string, listener: Function): this {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.unshift(new StreamListenerEntry(listener, true));
        return this;
    }

    removeListener(event: string, listener: Function): this {
        return this.off(event, listener);
    }

    off(event: string, listener: Function): this {
        const idx = this._findBucketIndex(event);
        if (idx < 0) {
            return this;
        }
        const remaining: StreamListenerEntry[] = [];
        let removed = false;
        const bucket = this._events[idx];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            if (!removed && entry.fn === listener) {
                removed = true;
            } else {
                remaining.push(entry);
            }
        }
        bucket.listeners = remaining;
        return this;
    }

    removeAllListeners(event?: string): this {
        if (event !== undefined && event !== "") {
            const idx = this._findBucketIndex(event);
            if (idx >= 0) {
                this._events[idx].listeners = [];
            }
        } else {
            this._events = [];
        }
        return this;
    }

    setMaxListeners(n: number): this {
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
        const snapshot: StreamListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            snapshot.push(bucket.listeners[i]);
        }

        const remaining: StreamListenerEntry[] = [];
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

    destroy(error?: unknown): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error as Error;
            if (this.listenerCount("error") > 0) {
                this.emit("error", error);
            }
        }
        this.emit("close");
        return this;
    }

    write(chunk: unknown, encoding?: unknown, callback?: unknown): boolean {
        if (this instanceof Writable) {
            return (this as Writable).write(chunk, encoding, callback);
        }
        if (this instanceof Duplex) {
            return (this as Duplex).write(chunk, encoding, callback);
        }
        return true;
    }

    end(chunk?: unknown, encoding?: unknown, callback?: unknown): this {
        if (this instanceof Writable) {
            (this as Writable).end(chunk, encoding, callback);
        } else if (this instanceof Duplex) {
            (this as Duplex).end(chunk, encoding, callback);
        }
        return this;
    }

    read(size: number = -1): unknown {
        if (this instanceof Readable) {
            return (this as Readable).read(size);
        }
        return null;
    }

    pause(): this {
        if (this instanceof Readable) {
            (this as Readable).pause();
        }
        return this;
    }

    resume(): this {
        if (this instanceof Readable) {
            (this as Readable).resume();
        }
        return this;
    }

    pipe(destination: unknown, options?: unknown): unknown {
        if (this instanceof Readable) {
            let opt: PipeOptions | undefined = undefined;
            if (typeof options === "object" && options !== null) {
                opt = options as PipeOptions;
            }
            return (this as Readable).pipe(destination, opt);
        }
        return destination;
    }
}

export function getDefaultHighWaterMark(objectMode: boolean = false): number {
    return __scriptgo.streamGetDefaultHighWaterMark(objectMode);
}

export function setDefaultHighWaterMark(objectMode: boolean, value: number): void {
    if (typeof value !== "number" || value < 0) {
        throw new Error("highWaterMark must be a non-negative number");
    }
    __scriptgo.streamSetDefaultHighWaterMark(objectMode, Math.floor(value));
}

export interface ReadableOptions {
    highWaterMark?: number;
    encoding?: string;
    objectMode?: boolean;
    autoDestroy?: boolean;
    read?: Function;
    destroy?: Function;
}

export class Readable extends Stream {
    readable: boolean = true;
    readableEnded: boolean = false;
    readableFlowing: boolean = false;
    readableHighWaterMark: number = 16384;
    readableLength: number = 0;
    readableObjectMode: boolean = false;
    readableEncoding: string = "utf8";
    destroyed: boolean = false;
    errored: Error | null = null;
    closed: boolean = false;

    _buffer: string[] = [];
    _reading: boolean = false;
    _paused: boolean = false;
    _autoDestroy: boolean = true;
    _disturbed: boolean = false;
    _customRead: Function | null = null;
    _customDestroy: Function | null = null;
    _pipeDests: Stream[] = [];

    constructor(options?: ReadableOptions) {
        super();
        if (options !== undefined && options !== null) {
            if (options.objectMode !== undefined) {
                this.readableObjectMode = options.objectMode;
            }
            if (options.highWaterMark !== undefined && options.highWaterMark > 0) {
                this.readableHighWaterMark = options.highWaterMark;
            } else {
                this.readableHighWaterMark = getDefaultHighWaterMark(this.readableObjectMode);
            }
            if (options.encoding !== undefined) {
                this.readableEncoding = options.encoding;
            }
            if (options.autoDestroy !== undefined) {
                this._autoDestroy = options.autoDestroy;
            }
            if (options.read !== undefined) {
                this._customRead = options.read as Function;
            }
            if (options.destroy !== undefined) {
                this._customDestroy = options.destroy as Function;
            }
        } else {
            this.readableHighWaterMark = getDefaultHighWaterMark(false);
        }
    }

    private _calcLength(): number {
        if (this.readableObjectMode) {
            return this._buffer.length;
        }
        let total = 0;
        for (let i = 0; i < this._buffer.length; i++) {
            total += this._buffer[i].length;
        }
        return total;
    }

    _read(size: number): void {
        if (typeof this._customRead === "function") {
            const fn = this._customRead as Function;
            fn(size);
        }
    }

    read(size: number = -1): string | null {
        this._disturbed = true;
        if (size === 0) {
            this._read(0);
            return null;
        }

        if (this._buffer.length === 0) {
            if (!this.readableEnded && !this._reading) {
                this._reading = true;
                this._read(size > 0 ? size : this.readableHighWaterMark);
                this._reading = false;
            }
            if (this._buffer.length === 0) {
                return null;
            }
        }

        const chunk = this._buffer.shift();
        this.readableLength = this._calcLength();
        if (chunk !== undefined) {
            return chunk;
        }
        return null;
    }

    push(chunk: string | null, encoding?: string): boolean {
        this._disturbed = true;
        if (chunk === null) {
            this.readableEnded = true;
            this.emit("end");
            if (this._autoDestroy) {
                this.destroy();
            }
            return false;
        }

        this._buffer.push(chunk);
        this.readableLength = this._calcLength();
        this.emit("readable");

        if (this.readableFlowing && !this._paused) {
            while (this._buffer.length > 0 && !this._paused) {
                const item = this._buffer.shift();
                this.readableLength = this._calcLength();
                if (item !== undefined) {
                    this.emit("data", item);
                }
            }
        }

        return this.readableLength < this.readableHighWaterMark;
    }

    unshift(chunk: string): boolean {
        this._buffer.unshift(chunk);
        this.readableLength = this._calcLength();
        return true;
    }

    on(event: string, listener: Function): this {
        super.on(event, listener);
        if (event === "data") {
            if (!this._paused) {
                this.resume();
            } else {
                this.readableFlowing = true;
            }
        } else if (event === "end" && (this.readableEnded || this.destroyed) && this._buffer.length === 0) {
            listener();
        }
        return this;
    }

    addListener(event: string, listener: Function): this {
        return this.on(event, listener);
    }

    pause(): this {
        this._paused = true;
        this.readableFlowing = false;
        this.emit("pause");
        return this;
    }

    resume(): this {
        this._paused = false;
        this.readableFlowing = true;
        this.emit("resume");
        while (this._buffer.length > 0 && !this._paused) {
            const item = this._buffer.shift();
            this.readableLength = this._calcLength();
            if (item !== undefined) {
                this.emit("data", item);
            }
        }
        if (this._buffer.length === 0 && this.readableEnded) {
            this.emit("end");
        } else if (this._buffer.length === 0 && !this.readableEnded) {
            this._read(this.readableHighWaterMark);
        }
        return this;
    }

    isPaused(): boolean {
        return this._paused;
    }

    setEncoding(encoding: string): this {
        this.readableEncoding = encoding;
        return this;
    }

    pipe(destination: unknown, options?: PipeOptions): unknown {
        let endOption = true;
        if (options !== undefined && options !== null) {
            endOption = options.end;
        }

        if (destination instanceof Writable) {
            const dest = destination as Writable;
            this._pipeDests.push(destination);
            if (endOption) {
                this.on("end", () => {
                    dest.end();
                });
            }
            this.on("data", (chunk: unknown) => {
                dest.write(chunk);
            });
        } else if (destination instanceof Duplex) {
            const dest = destination as Duplex;
            this._pipeDests.push(destination);
            if (endOption) {
                this.on("end", () => {
                    dest.end();
                });
            }
            this.on("data", (chunk: unknown) => {
                dest.write(chunk);
            });
        } else if (destination instanceof Stream) {
            const dest = destination as Stream;
            this._pipeDests.push(destination);
            if (endOption) {
                this.on("end", () => {
                    dest.end();
                });
            }
            this.on("data", (chunk: unknown) => {
                dest.write(chunk);
            });
        }

        if (this._pipeDests.length === 1 && !this._paused) {
            this.resume();
        }
        return destination;
    }

    unpipe(destination?: unknown): this {
        if (destination === undefined || destination === null) {
            this._pipeDests = [];
            this.pause();
            return this;
        }
        const remaining: Stream[] = [];
        for (let i = 0; i < this._pipeDests.length; i++) {
            const oldStream = this._pipeDests[i];
            if (oldStream !== destination) {
                remaining.push(oldStream);
            }
        }
        this._pipeDests = remaining;
        return this;
    }

    wrap(oldStream: unknown): this {
        if (oldStream instanceof Stream) {
            const old = oldStream as Stream;
            old.on("data", (chunk: unknown) => {
                this.push(String(chunk));
            });
            old.on("end", () => {
                this.push(null);
            });
            old.on("error", (err: unknown) => {
                let errObj: Error = new Error(String(err));
                if (err instanceof Error) {
                    errObj = err as Error;
                }
                this.destroy(errObj);
            });
        }
        return this;
    }

    destroy(error?: unknown): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.readable = false;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error as Error;
            if (this.listenerCount("error") > 0) {
                this.emit("error", error);
            }
        }
        if (typeof this._customDestroy === "function") {
            const fn = this._customDestroy as Function;
            fn(error !== undefined ? error : null, (err: Error | null) => {});
        }
        this.emit("close");
        return this;
    }

    static from(iterable: unknown, options?: unknown): Readable {
        let stream: Readable = new Readable();
        if (options !== undefined && options !== null) {
            stream = new Readable(options as ReadableOptions);
        }
        if (iterable !== null && iterable !== undefined) {
            if (Array.isArray(iterable)) {
                const arr = iterable as string[];
                let idx = 0;
                stream._customRead = (size: number) => {
                    while (idx < arr.length && !stream.isPaused()) {
                        stream.push(String(arr[idx]));
                        idx++;
                    }
                    if (idx >= arr.length) {
                        stream.push(null);
                    }
                };
            } else if (typeof iterable === "string") {
                const str = iterable as string;
                let sent = false;
                stream._customRead = (size: number) => {
                    if (!sent) {
                        sent = true;
                        stream.push(str);
                        stream.push(null);
                    }
                };
            }
        }
        return stream;
    }

    static isDisturbed(stream: unknown): boolean {
        if (stream instanceof Readable) {
            return (stream as Readable)._disturbed;
        }
        return false;
    }

    static fromWeb(readableStream: unknown, options?: unknown): Readable {
        if (options !== undefined && options !== null) {
            return new Readable(options as ReadableOptions);
        }
        return new Readable();
    }

    static toWeb(streamReadable: unknown, options?: unknown): unknown {
        return new WebReadableStream(streamReadable as Readable);
    }
}

export interface WritableOptions {
    highWaterMark?: number;
    decodeStrings?: boolean;
    defaultEncoding?: string;
    objectMode?: boolean;
    autoDestroy?: boolean;
    write?: Function;
    writev?: Function;
    final?: Function;
    destroy?: Function;
}

export class Writable extends Stream {
    writable: boolean = true;
    writableEnded: boolean = false;
    writableFinished: boolean = false;
    writableHighWaterMark: number = 16384;
    writableLength: number = 0;
    writableNeedDrain: boolean = false;
    writableObjectMode: boolean = false;
    writableCorked: number = 0;
    destroyed: boolean = false;
    errored: Error | null = null;
    closed: boolean = false;

    _defaultEncoding: string = "utf8";
    _writeBuffer: StreamWriteReq[] = [];
    _autoDestroy: boolean = true;
    _customWrite: Function | null = null;
    _customWritev: Function | null = null;
    _customFinal: Function | null = null;
    _customDestroy: Function | null = null;

    constructor(options?: WritableOptions) {
        super();
        if (options !== undefined && options !== null) {
            if (options.objectMode !== undefined) {
                this.writableObjectMode = options.objectMode;
            }
            if (options.highWaterMark !== undefined && options.highWaterMark > 0) {
                this.writableHighWaterMark = options.highWaterMark;
            } else {
                this.writableHighWaterMark = getDefaultHighWaterMark(this.writableObjectMode);
            }
            if (options.defaultEncoding !== undefined) {
                this._defaultEncoding = options.defaultEncoding;
            }
            if (options.autoDestroy !== undefined) {
                this._autoDestroy = options.autoDestroy;
            }
            if (options.write !== undefined) {
                this._customWrite = options.write as Function;
            }
            if (options.writev !== undefined) {
                this._customWritev = options.writev as Function;
            }
            if (options.final !== undefined) {
                this._customFinal = options.final as Function;
            }
            if (options.destroy !== undefined) {
                this._customDestroy = options.destroy as Function;
            }
        } else {
            this.writableHighWaterMark = getDefaultHighWaterMark(false);
        }
    }

    _write(chunk: string, encoding: string, callback: Function): void {
        if (typeof this._customWrite === "function") {
            const fn = this._customWrite as Function;
            fn(chunk, encoding, callback);
        } else {
            callback(null);
        }
    }

    _writev(chunks: StreamWriteReq[], callback: Function): void {
        if (typeof this._customWritev === "function") {
            const fn = this._customWritev as Function;
            fn(chunks, callback);
        } else {
            callback(null);
        }
    }

    _final(callback: Function): void {
        if (typeof this._customFinal === "function") {
            const fn = this._customFinal as Function;
            fn(callback);
        } else {
            callback(null);
        }
    }

    setDefaultEncoding(encoding: string): this {
        this._defaultEncoding = encoding;
        return this;
    }

    cork(): void {
        this.writableCorked++;
    }

    uncork(): void {
        if (this.writableCorked > 0) {
            this.writableCorked--;
            if (this.writableCorked === 0) {
                while (this._writeBuffer.length > 0) {
                    const item = this._writeBuffer.shift();
                    if (item !== undefined) {
                        this._write(item.chunk, item.encoding, item.callback);
                    }
                }
                this.writableLength = 0;
                if (this.writableNeedDrain) {
                    this.writableNeedDrain = false;
                    this.emit("drain");
                }
            }
        }
    }

    write(chunk: unknown, encodingOrCallback?: unknown, callback?: unknown): boolean {
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof encodingOrCallback === "function") {
            cb = encodingOrCallback as Function;
        } else if (typeof encodingOrCallback === "string") {
            encoding = encodingOrCallback as string;
            if (typeof callback === "function") {
                cb = callback as Function;
            }
        } else if (typeof callback === "function") {
            cb = callback as Function;
        }

        if (this.writableEnded || this.destroyed) {
            const err = new Error("write after end");
            if (cb !== null) {
                cb(err);
            } else {
                this.emit("error", err);
            }
            return false;
        }

        const chunkStr = String(chunk);
        this.writableLength += chunkStr.length;

        const self = this;
        const onDone = (err?: Error | null) => {
            self.writableLength = 0;
            if (err) {
                self.destroy(err);
                return;
            }
            if (self.writableNeedDrain && self.writableLength < self.writableHighWaterMark) {
                self.writableNeedDrain = false;
                self.emit("drain");
            }
        };

        if (this.writableCorked > 0) {
            this._writeBuffer.push(new StreamWriteReq(chunkStr, encoding, onDone));
        } else {
            this._write(chunkStr, encoding, onDone);
        }

        if (cb !== null) {
            cb(null);
        }

        return this.writableLength < this.writableHighWaterMark;
    }

    end(chunkOrCallback?: unknown, encodingOrCallback?: unknown, callback?: unknown): this {
        let chunk: unknown = null;
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof chunkOrCallback === "function") {
            cb = chunkOrCallback as Function;
        } else if (chunkOrCallback !== undefined && chunkOrCallback !== null) {
            chunk = chunkOrCallback;
            if (typeof encodingOrCallback === "function") {
                cb = encodingOrCallback as Function;
            } else if (typeof encodingOrCallback === "string") {
                encoding = encodingOrCallback as string;
                if (typeof callback === "function") {
                    cb = callback as Function;
                }
            }
        }

        if (chunk !== null) {
            this.write(chunk, encoding);
        }

        this.writableEnded = true;

        const self = this;
        this._final((err?: Error | null) => {
            if (err) {
                self.destroy(err);
                return;
            }
            self.writableFinished = true;
            self.emit("finish");
            if (self._autoDestroy) {
                self.destroy();
            }
        });

        if (cb !== null) {
            cb(null);
        }

        return this;
    }

    destroy(error?: unknown): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.writable = false;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error as Error;
            if (this.listenerCount("error") > 0) {
                this.emit("error", error);
            }
        }
        if (typeof this._customDestroy === "function") {
            const fn = this._customDestroy as Function;
            fn(error !== undefined ? error : null, (err: Error | null) => {});
        }
        this.emit("close");
        return this;
    }

    static fromWeb(writableStream: unknown, options?: unknown): Writable {
        if (options !== undefined && options !== null) {
            return new Writable(options as WritableOptions);
        }
        return new Writable();
    }

    static toWeb(streamWritable: unknown, options?: unknown): unknown {
        return new WebWritableStream(streamWritable as Writable);
    }
}

export interface DuplexOptions {
    allowHalfOpen?: boolean;
    readable?: boolean;
    writable?: boolean;
    highWaterMark?: number;
    decodeStrings?: boolean;
    defaultEncoding?: string;
    encoding?: string;
    objectMode?: boolean;
    autoDestroy?: boolean;
    read?: Function;
    write?: Function;
    writev?: Function;
    final?: Function;
    destroy?: Function;
}

export class Duplex extends Readable {
    writable: boolean = true;
    writableEnded: boolean = false;
    writableFinished: boolean = false;
    writableHighWaterMark: number = 16384;
    writableLength: number = 0;
    writableNeedDrain: boolean = false;
    writableObjectMode: boolean = false;
    writableCorked: number = 0;

    _defaultEncoding: string = "utf8";
    _writeBuffer: StreamWriteReq[] = [];
    _allowHalfOpen: boolean = true;
    _customWrite: Function | null = null;
    _customFinal: Function | null = null;

    constructor(options?: DuplexOptions) {
        super(options);
        if (options !== undefined && options !== null) {
            if (options.allowHalfOpen !== undefined) {
                this._allowHalfOpen = options.allowHalfOpen;
            }
            if (options.writable !== undefined) {
                this.writable = options.writable;
            }
            if (options.write !== undefined) {
                this._customWrite = options.write as Function;
            }
            if (options.final !== undefined) {
                this._customFinal = options.final as Function;
            }
            if (options.defaultEncoding !== undefined) {
                this._defaultEncoding = options.defaultEncoding;
            }
            if (options.objectMode !== undefined) {
                this.writableObjectMode = options.objectMode;
            }
        }
    }

    _write(chunk: string, encoding: string, callback: Function): void {
        if (this instanceof Transform) {
            (this as Transform)._write(chunk, encoding, callback);
            return;
        }
        if (typeof this._customWrite === "function") {
            const fn = this._customWrite as Function;
            fn(chunk, encoding, callback);
        } else {
            callback(null);
        }
    }

    _final(callback: Function): void {
        if (this instanceof Transform) {
            (this as Transform)._final(callback);
            return;
        }
        if (typeof this._customFinal === "function") {
            const fn = this._customFinal as Function;
            fn(callback);
        } else {
            callback(null);
        }
    }

    setDefaultEncoding(encoding: string): this {
        this._defaultEncoding = encoding;
        return this;
    }

    cork(): void {
        this.writableCorked++;
    }

    uncork(): void {
        if (this.writableCorked > 0) {
            this.writableCorked--;
            if (this.writableCorked === 0) {
                while (this._writeBuffer.length > 0) {
                    const req = this._writeBuffer.shift();
                    if (req !== undefined) {
                        this._write(req.chunk, req.encoding, req.callback);
                    }
                }
                this.writableLength = 0;
                if (this.writableNeedDrain) {
                    this.writableNeedDrain = false;
                    this.emit("drain");
                }
            }
        }
    }

    write(chunk: unknown, encodingOrCallback?: unknown, callback?: unknown): boolean {
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof encodingOrCallback === "function") {
            cb = encodingOrCallback as Function;
        } else if (typeof encodingOrCallback === "string") {
            encoding = encodingOrCallback as string;
            if (typeof callback === "function") {
                cb = callback as Function;
            }
        } else if (typeof callback === "function") {
            cb = callback as Function;
        }

        if (this.writableEnded || this.destroyed) {
            const err = new Error("write after end");
            if (cb !== null) {
                cb(err);
            } else {
                this.emit("error", err);
            }
            return false;
        }

        const chunkStr = String(chunk);
        this.writableLength += chunkStr.length;

        const self = this;
        const onDone = (err?: Error | null) => {
            self.writableLength = 0;
            if (err) {
                self.destroy(err);
                return;
            }
            if (self.writableNeedDrain && self.writableLength < self.writableHighWaterMark) {
                self.writableNeedDrain = false;
                self.emit("drain");
            }
        };

        if (this.writableCorked > 0) {
            this._writeBuffer.push(new StreamWriteReq(chunkStr, encoding, onDone));
        } else {
            this._write(chunkStr, encoding, onDone);
        }

        if (cb !== null) {
            cb(null);
        }

        return this.writableLength < this.writableHighWaterMark;
    }

    end(chunkOrCallback?: unknown, encodingOrCallback?: unknown, callback?: unknown): this {
        let chunk: unknown = null;
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof chunkOrCallback === "function") {
            cb = chunkOrCallback as Function;
        } else if (chunkOrCallback !== undefined && chunkOrCallback !== null) {
            chunk = chunkOrCallback;
            if (typeof encodingOrCallback === "function") {
                cb = encodingOrCallback as Function;
            } else if (typeof encodingOrCallback === "string") {
                encoding = encodingOrCallback as string;
                if (typeof callback === "function") {
                    cb = callback as Function;
                }
            }
        }

        if (chunk !== null) {
            this.write(chunk, encoding);
        }

        this.writableEnded = true;

        const self = this;
        this._final((err?: Error | null) => {
            if (err) {
                self.destroy(err);
                return;
            }
            self.writableFinished = true;
            self.emit("finish");
            if (!self._allowHalfOpen && self._autoDestroy) {
                self.destroy();
            }
        });

        if (cb !== null) {
            cb(null);
        }

        return this;
    }

    static from(src: unknown, options?: unknown): Duplex {
        if (src instanceof Duplex) {
            return src as Duplex;
        }
        let duplex: Duplex = new Duplex();
        if (options !== undefined && options !== null) {
            duplex = new Duplex(options as DuplexOptions);
        }
        if (src instanceof Stream) {
            const s = src as Stream;
            const res = new PassThrough();
            s.pipe(res);
            return res;
        }
        return duplex;
    }

    static fromWeb(pair: unknown, options?: unknown): Duplex {
        if (options !== undefined && options !== null) {
            return new Duplex(options as DuplexOptions);
        }
        return new Duplex();
    }

    static toWeb(streamDuplex: unknown, options?: unknown): unknown {
        return new WebDuplexStream(
            new WebReadableStream(streamDuplex as Readable),
            new WebWritableStream(streamDuplex as Writable)
        );
    }
}

export interface TransformOptions extends DuplexOptions {
    transform?: Function;
    flush?: Function;
}

export class Transform extends Duplex {
    _customTransform: Function | null = null;
    _customFlush: Function | null = null;

    constructor(options?: TransformOptions) {
        super(options);
        if (options !== undefined && options !== null) {
            if (options.transform !== undefined) {
                this._customTransform = options.transform as Function;
            }
            if (options.flush !== undefined) {
                this._customFlush = options.flush as Function;
            }
        }
    }

    _transform(chunk: string, encoding: string, callback: Function): void {
        if (typeof this._customTransform === "function") {
            const fn = this._customTransform as Function;
            fn(chunk, encoding, callback);
        } else {
            callback(null, chunk);
        }
    }

    _flush(callback: Function): void {
        if (typeof this._customFlush === "function") {
            const fn = this._customFlush as Function;
            fn(callback);
        } else {
            callback(null);
        }
    }

    _write(chunk: string, encoding: string, callback: Function): void {
        const self = this;
        this._transform(chunk, encoding, (err?: Error | null, data?: string) => {
            if (err) {
                callback(err);
                return;
            }
            if (data) {
                self.push(data);
            }
            callback(null);
        });
    }

    _final(callback: Function): void {
        const self = this;
        this._flush((err?: Error | null, data?: string) => {
            if (err) {
                callback(err);
                return;
            }
            if (data) {
                self.push(data);
            }
            self.push(null);
            self.writableFinished = true;
            self.emit("finish");
            callback(null);
        });
    }
}

export class PassThrough extends Transform {
    _transform(chunk: string, encoding: string, callback: Function): void {
        callback(null, chunk);
    }
}

export function isReadable(stream: unknown): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.readable !== false && !s.destroyed;
    }
    return false;
}

export function isWritable(stream: unknown): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.writable !== false && !s.destroyed;
    }
    return false;
}

export function isErrored(stream: unknown): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.errored !== null && s.errored !== undefined;
    }
    return false;
}

export function addAbortSignal(signal: unknown, stream: unknown): unknown {
    if (signal !== null && signal !== undefined && stream instanceof Stream) {
        const sig = signal as AbortSignalLike;
        const str = stream as Stream;
        if (sig.aborted) {
            str.destroy(new Error("The operation was aborted"));
        } else if (typeof sig.addEventListener === "function") {
            const addEv = sig.addEventListener as Function;
            addEv("abort", () => {
                str.destroy(new Error("The operation was aborted"));
            });
        }
    }
    return stream;
}

export function finished(stream: unknown, optionsOrCallback?: unknown, callback?: unknown): unknown {
    let cb: Function | null = null;

    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
    }

    if (cb !== null) {
        const fn = cb as Function;
        if (stream instanceof Stream) {
            const s = stream as Stream;
            if (s.destroyed || s.closed) {
                fn(s.errored);
                return null;
            }
            let done = false;
            const onFinishOrEnd = () => {
                if (!done) {
                    done = true;
                    fn(null);
                }
            };
            const onError = (err: unknown) => {
                if (!done) {
                    done = true;
                    fn(err);
                }
            };
            s.on("finish", onFinishOrEnd);
            s.on("end", onFinishOrEnd);
            s.on("close", onFinishOrEnd);
            s.on("error", onError);
        }
        return null;
    }

    return Promise.resolve(null);
}

function _extractStreams(first?: unknown, second?: unknown, third?: unknown, fourth?: unknown, fifth?: unknown): Stream[] {
    const res: Stream[] = [];
    if (Array.isArray(first)) {
        const arr = first as Stream[];
        for (let j = 0; j < arr.length; j++) {
            res.push(arr[j]);
        }
    } else if (first instanceof Stream) {
        res.push(first as Stream);
    }
    if (second instanceof Stream) {
        res.push(second as Stream);
    }
    if (third instanceof Stream) {
        res.push(third as Stream);
    }
    if (fourth instanceof Stream) {
        res.push(fourth as Stream);
    }
    if (fifth instanceof Stream) {
        res.push(fifth as Stream);
    }
    return res;
}

function _extractCallback(first?: unknown, second?: unknown, third?: unknown, fourth?: unknown, fifth?: unknown): Function | null {
    if (typeof first === "function") return first as Function;
    if (typeof second === "function") return second as Function;
    if (typeof third === "function") return third as Function;
    if (typeof fourth === "function") return fourth as Function;
    if (typeof fifth === "function") return fifth as Function;
    return null;
}

export function pipeline(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown,
    fifth?: unknown
): unknown {
    const streams: Stream[] = _extractStreams(first, second, third, fourth, fifth);
    const cb: Function | null = _extractCallback(first, second, third, fourth, fifth);

    if (cb !== null) {
        const fn = cb as Function;
        if (streams.length > 0) {
            streams[0].pause();
            const last = streams[streams.length - 1];
            let done = false;
            const onFinish = () => {
                if (!done) {
                    done = true;
                    fn(null);
                }
            };
            const onError = (err: unknown) => {
                if (!done) {
                    done = true;
                    fn(err);
                }
            };
            last.on("finish", onFinish);
            last.on("end", onFinish);
            last.on("close", onFinish);
            for (let i = 0; i < streams.length; i++) {
                streams[i].on("error", onError);
            }
        }
        for (let i = 0; i < streams.length - 1; i++) {
            const current = streams[i];
            const next = streams[i + 1];
            current.pipe(next);
        }
        if (streams.length > 0) {
            streams[0].resume();
            return streams[streams.length - 1];
        }
        return null;
    }

    if (streams.length > 0) {
        streams[0].pause();
        for (let i = 0; i < streams.length - 1; i++) {
            const current = streams[i];
            const next = streams[i + 1];
            current.pipe(next);
        }
        streams[0].resume();
        return Promise.resolve(streams[streams.length - 1]);
    }
    return Promise.resolve(null);
}

function _processComposeArg(arg: unknown, list: Stream[]): void {
    if (arg === undefined || arg === null) {
        return;
    }
    if (Array.isArray(arg)) {
        const arr = arg as Stream[];
        for (let j = 0; j < arr.length; j++) {
            list.push(arr[j]);
        }
    } else if (arg instanceof Stream) {
        list.push(arg as Stream);
    }
}

export function compose(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown
): Duplex {
    const list: Stream[] = [];
    _processComposeArg(first, list);
    _processComposeArg(second, list);
    _processComposeArg(third, list);
    _processComposeArg(fourth, list);

    if (list.length === 0) {
        return new PassThrough();
    }
    if (list.length === 1) {
        if (list[0] instanceof Duplex) {
            return list[0] as Duplex;
        }
        return new PassThrough();
    }

    const firstStream = list[0];
    const lastStream = list[list.length - 1];

    for (let i = 0; i < list.length - 1; i++) {
        const c = list[i];
        c.pipe(list[i + 1]);
    }

    const duplex = new Duplex();
    duplex._customRead = (size: number) => {
        lastStream.read(size);
    };
    duplex._customWrite = (chunk: string, encoding: string, callback: Function) => {
        firstStream.write(chunk);
        callback(null);
    };
    duplex._customFinal = (callback: Function) => {
        firstStream.end();
        callback(null);
    };

    lastStream.on("data", (chunk: unknown) => duplex.push(String(chunk)));
    lastStream.on("end", () => duplex.push(null));
    lastStream.on("error", (err: unknown) => {
        let errObj: Error = new Error(String(err));
        if (err instanceof Error) {
            errObj = err as Error;
        }
        duplex.destroy(errObj);
    });

    return duplex;
}

export class StreamPromises {
    _tag: string = "promises";

    finished(stream: unknown, options?: unknown): Promise<void> {
        return Promise.resolve(undefined) as Promise<void>;
    }

    pipeline(first?: unknown, second?: unknown, third?: unknown, fourth?: unknown, fifth?: unknown): Promise<unknown> {
        pipeline(first, second, third, fourth, fifth);
        return Promise.resolve(null);
    }
}

export class StreamConsumers {
    _tag: string = "consumers";

    async buffer(stream: unknown): Promise<Uint8Array> {
        let res = "";
        if (stream instanceof Stream) {
            const s = stream as Stream;
            s.on("data", (chunk: unknown) => {
                res = res + String(chunk);
            });
        }
        return Buffer.from(res);
    }

    async text(stream: unknown): Promise<string> {
        let res = "";
        if (stream instanceof Stream) {
            const s = stream as Stream;
            s.on("data", (chunk: unknown) => {
                res = res + String(chunk);
            });
        }
        return res;
    }

    async json(stream: unknown): Promise<unknown> {
        const t = await this.text(stream);
        return JSON.parse(t);
    }

    async arrayBuffer(stream: unknown): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async blob(stream: unknown): Promise<unknown> {
        return this.buffer(stream);
    }
}

export const promises: StreamPromises = new StreamPromises();
export const consumers: StreamConsumers = new StreamConsumers();
