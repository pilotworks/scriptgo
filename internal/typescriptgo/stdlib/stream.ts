// ScriptGo Standard Library: node:stream

import { Blob, Buffer } from "node:buffer";
import { EventEmitter, AbortSignal } from "node:events";

export type StreamChunk = string | Buffer | Uint8Array;

export interface AbortSignalLike {
    aborted?: boolean;
    addEventListener?: Function;
}

export interface FinishedOptions {
    error?: boolean;
    readable?: boolean;
    writable?: boolean;
    signal?: AbortSignal | AbortSignalLike;
}

declare namespace __scriptgo {
    function streamGetDefaultHighWaterMark(objectMode: boolean): number;
    function streamSetDefaultHighWaterMark(objectMode: boolean, value: number): void;
}

class StreamWriteReq {
    chunk: StreamChunk;
    encoding: string;
    callback: Function;

    constructor(chunk: StreamChunk, encoding: string, callback: Function) {
        this.chunk = chunk;
        this.encoding = encoding;
        this.callback = callback;
    }
}

export class PipeOptions {
    end: boolean = true;
}

export class Stream extends EventEmitter {
    static _defaultHighWaterMark: number = 65536;
    static _defaultObjectModeHighWaterMark: number = 16;

    readable: boolean = true;
    writable: boolean = true;
    destroyed: boolean = false;
    errored: Error | null = null;
    closed: boolean = false;

    constructor() {
        super();
    }

    destroy(error?: Error | null): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error instanceof Error ? error : new Error(String(error));
            if (this.listenerCount("error") > 0) {
                this.emit("error", this.errored);
            }
        }
        this.emit("close");
        return this;
    }

    write(chunk: StreamChunk, encoding?: string | Function, callback?: Function): boolean {
        if (this instanceof Writable) {
            return (this as Writable).write(chunk, encoding, callback);
        }
        if (this instanceof Duplex) {
            return (this as Duplex).write(chunk, encoding, callback);
        }
        return true;
    }

    end(chunk?: StreamChunk | Function, encoding?: string | Function, callback?: Function): this {
        if (this instanceof Writable) {
            (this as Writable).end(chunk, encoding, callback);
        } else if (this instanceof Duplex) {
            (this as Duplex).end(chunk, encoding, callback);
        }
        return this;
    }

    read(size: number = -1): StreamChunk | null {
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

    pipe(destination: Writable | Stream, options?: PipeOptions): Writable | Stream {
        if (this instanceof Readable) {
            return (this as Readable).pipe(destination, options);
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
    signal?: AbortSignal | AbortSignalLike;
    read?: Function;
    destroy?: Function;
}

class NodeReadableStream {
    locked: boolean = false;
    _stream: Readable;

    constructor(stream: Readable) {
        this.locked = false;
        this._stream = stream;
    }
}

export class Readable extends Stream {
    readable: boolean = true;
    readableEnded: boolean = false;
    readableFlowing: boolean | null = null;
    readableHighWaterMark: number = 16384;
    readableLength: number = 0;
    readableObjectMode: boolean = false;
    readableEncoding: string | null = null;
    readableDidRead: boolean = false;
    readableAborted: boolean = false;
    destroyed: boolean = false;
    errored: Error | null = null;
    closed: boolean = false;

    _buffer: StreamChunk[] = [];
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
            if (options.signal) {
                addAbortSignal(options.signal, this);
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
            const item = this._buffer[i];
            if (typeof item === "string") {
                total += (item as string).length;
            } else if (item && typeof (item as { length?: number }).length === "number") {
                total += (item as { length: number }).length;
            } else {
                total += 1;
            }
        }
        return total;
    }

    _read(size: number): void {
        if (typeof this._customRead === "function") {
            const fn = this._customRead as Function;
            fn(size);
        }
    }

    read(size: number = -1): StreamChunk | null {
        this._disturbed = true;
        this.readableDidRead = true;
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

    push(chunk: StreamChunk | null, encoding?: string): boolean {
        this._disturbed = true;
        if (chunk === null) {
            this.readableEnded = true;
            if (this._buffer.length === 0) {
                this.emit("end");
                if (this._autoDestroy) {
                    this.destroy();
                }
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
            if (this._buffer.length === 0 && this.readableEnded) {
                this.emit("end");
                if (this._autoDestroy) {
                    this.destroy();
                }
            }
        }

        return this.readableLength < this.readableHighWaterMark;
    }

    unshift(chunk: StreamChunk): boolean {
        this._buffer.unshift(chunk);
        this.readableLength = this._calcLength();
        return true;
    }

    on(event: string | symbol, listener: Function): this {
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

    addListener(event: string | symbol, listener: Function): this {
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
            if (this._autoDestroy) {
                this.destroy();
            }
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

    pipe(destination: Writable | Stream, options?: PipeOptions): Writable | Stream {
        let endOption = true;
        if (options !== undefined && options !== null) {
            endOption = options.end;
        }

        if (destination instanceof Writable) {
            const dest = destination as Writable;
            this._pipeDests.push(destination);
            this.on("data", (chunk: StreamChunk) => {
                const canContinue = dest.write(chunk);
                if (!canContinue) {
                    this.pause();
                }
            });
            dest.on("drain", () => {
                this.resume();
            });
            if (endOption) {
                this.once("end", () => {
                    dest.end();
                });
            }
        } else if (destination instanceof Duplex) {
            const dest = destination as Duplex;
            this._pipeDests.push(destination);
            this.on("data", (chunk: StreamChunk) => {
                const canContinue = dest.write(chunk);
                if (!canContinue) {
                    this.pause();
                }
            });
            dest.on("drain", () => {
                this.resume();
            });
            if (endOption) {
                this.once("end", () => {
                    dest.end();
                });
            }
        } else if (destination instanceof Stream) {
            const dest = destination as Stream;
            this._pipeDests.push(destination);
            this.on("data", (chunk: StreamChunk) => {
                dest.write(chunk);
            });
            if (endOption) {
                this.once("end", () => {
                    dest.end();
                });
            }
        }

        if (this._pipeDests.length === 1 && !this._paused) {
            this.resume();
        }
        return destination;
    }

    unpipe(destination?: Stream): this {
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

    wrap(oldStream: Stream | EventEmitter): this {
        if (oldStream instanceof Stream) {
            const old = oldStream as Stream;
            old.on("data", (chunk: StreamChunk) => {
                this.push(chunk);
            });
            old.on("end", () => {
                this.push(null);
            });
            old.on("error", (err: Error) => {
                this.destroy(err);
            });
        }
        return this;
    }

    destroy(error?: Error | null): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.readable = false;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error instanceof Error ? error : new Error(String(error));
            this.readableAborted = true;
            if (this.listenerCount("error") > 0) {
                this.emit("error", this.errored);
            }
        }
        if (typeof this._customDestroy === "function") {
            const fn = this._customDestroy as Function;
            fn(error !== undefined ? error : null, (err: Error | null) => {
                if (err && this.listenerCount("error") > 0) {
                    this.emit("error", err);
                }
            });
        }
        this.emit("close");
        return this;
    }

    [Symbol.asyncIterator](): AsyncIterableIterator<StreamChunk> {
        const self = this;
        return {
            [Symbol.asyncIterator]() {
                return this;
            },
            async next(): Promise<IteratorResult<StreamChunk>> {
                const chunk = self.read();
                if (chunk !== null) {
                    return { done: false, value: chunk };
                }
                if (self.readableEnded && self._buffer.length === 0) {
                    return { done: true, value: undefined as unknown as StreamChunk };
                }
                if (self.destroyed) {
                    return { done: true, value: undefined as unknown as StreamChunk };
                }
                return new Promise<IteratorResult<StreamChunk>>((resolve, reject) => {
                    const onData = (val: StreamChunk) => {
                        cleanup();
                        resolve({ done: false, value: val });
                    };
                    const onEnd = () => {
                        cleanup();
                        resolve({ done: true, value: undefined as unknown as StreamChunk });
                    };
                    const onError = (err: Error) => {
                        cleanup();
                        reject(err);
                    };
                    const cleanup = () => {
                        self.removeListener("data", onData);
                        self.removeListener("end", onEnd);
                        self.removeListener("error", onError);
                    };
                    self.once("data", onData);
                    self.once("end", onEnd);
                    self.once("error", onError);
                    self.resume();
                });
            },
            async return(): Promise<IteratorResult<StreamChunk>> {
                self.destroy();
                return { done: true, value: undefined as unknown as StreamChunk };
            },
            async throw(err?: Error): Promise<IteratorResult<StreamChunk>> {
                self.destroy(err);
                return Promise.reject(err);
            }
        };
    }

    static from(iterable: unknown, options?: ReadableOptions): Readable {
        let stream: Readable = new Readable();
        if (options !== undefined && options !== null) {
            stream = new Readable(options);
        }
        if (iterable !== null && iterable !== undefined) {
            if (Array.isArray(iterable)) {
                const arr = iterable as StreamChunk[];
                let idx = 0;
                stream._customRead = () => {
                    while (idx < arr.length && !stream.isPaused()) {
                        stream.push(arr[idx]);
                        idx++;
                    }
                    if (idx >= arr.length) {
                        stream.push(null);
                    }
                };
            } else if (typeof iterable === "string") {
                const str = iterable as string;
                let sent = false;
                stream._customRead = () => {
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

    static fromWeb(src: unknown, options?: ReadableOptions): Readable {
        if (src instanceof NodeReadableStream) {
            return (src as NodeReadableStream)._stream;
        }
        return new Readable(options);
    }

    static toWeb(streamReadable: Readable, options?: unknown): unknown {
        return new NodeReadableStream(streamReadable);
    }

    static isDisturbed(stream: Stream | null | undefined): boolean {
        if (stream instanceof Readable) {
            return (stream as Readable)._disturbed;
        }
        return false;
    }
}

export interface WritableOptions {
    highWaterMark?: number;
    decodeStrings?: boolean;
    defaultEncoding?: string;
    objectMode?: boolean;
    autoDestroy?: boolean;
    signal?: AbortSignal | AbortSignalLike;
    write?: Function;
    writev?: Function;
    final?: Function;
    destroy?: Function;
}

class NodeWritableStream {
    locked: boolean = false;
    _stream: Writable;

    constructor(stream: Writable) {
        this.locked = false;
        this._stream = stream;
    }
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
            if (options.signal) {
                addAbortSignal(options.signal, this);
            }
        } else {
            this.writableHighWaterMark = getDefaultHighWaterMark(false);
        }
    }

    _write(chunk: StreamChunk, encoding: string, callback: Function): void {
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

    write(chunk: StreamChunk, encodingOrCallback?: string | Function, callback?: Function): boolean {
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

        let chunkLen = 1;
        if (typeof chunk === "string") {
            chunkLen = (chunk as string).length;
        } else if (chunk && typeof (chunk as { length?: number }).length === "number") {
            chunkLen = (chunk as { length: number }).length;
        }
        this.writableLength += chunkLen;

        const self = this;
        const onDone = (err?: Error | null) => {
            self.writableLength = Math.max(0, self.writableLength - chunkLen);
            if (err) {
                self.destroy(err);
                if (cb !== null) {
                    cb(err);
                }
                return;
            }
            if (cb !== null) {
                cb(null);
            }
            if (self.writableNeedDrain && self.writableLength < self.writableHighWaterMark) {
                self.writableNeedDrain = false;
                self.emit("drain");
            }
        };

        if (this.writableCorked > 0) {
            this._writeBuffer.push(new StreamWriteReq(chunk, encoding, onDone));
        } else {
            this._write(chunk, encoding, onDone);
        }

        const canWrite = this.writableLength < this.writableHighWaterMark;
        if (!canWrite) {
            this.writableNeedDrain = true;
        }
        return canWrite;
    }

    end(chunkOrCallback?: StreamChunk | Function, encodingOrCallback?: string | Function, callback?: Function): this {
        let chunk: StreamChunk | null = null;
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof chunkOrCallback === "function") {
            cb = chunkOrCallback as Function;
        } else if (chunkOrCallback !== undefined && chunkOrCallback !== null) {
            chunk = chunkOrCallback as StreamChunk;
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
                if (cb !== null) {
                    cb(err);
                }
                return;
            }
            self.writableFinished = true;
            self.emit("finish");
            if (cb !== null) {
                cb(null);
            }
            if (self._autoDestroy) {
                self.destroy();
            }
        });

        return this;
    }

    destroy(error?: Error | null): this {
        if (this.destroyed) {
            return this;
        }
        this.destroyed = true;
        this.writable = false;
        this.closed = true;
        if (error !== undefined && error !== null) {
            this.errored = error instanceof Error ? error : new Error(String(error));
            if (this.listenerCount("error") > 0) {
                this.emit("error", this.errored);
            }
        }
        if (typeof this._customDestroy === "function") {
            const fn = this._customDestroy as Function;
            fn(error !== undefined ? error : null, (err: Error | null) => {
                if (err && this.listenerCount("error") > 0) {
                    this.emit("error", err);
                }
            });
        }
        this.emit("close");
        return this;
    }

    static fromWeb(src: unknown, options?: WritableOptions): Writable {
        if (src instanceof NodeWritableStream) {
            return (src as NodeWritableStream)._stream;
        }
        return new Writable(options);
    }

    static toWeb(streamWritable: Writable): unknown {
        return new NodeWritableStream(streamWritable);
    }

    static isDisturbed(stream: Stream | null | undefined): boolean {
        if (stream instanceof Writable) {
            return (stream as Writable).writableEnded || (stream as Writable).destroyed;
        }
        return false;
    }
}

export interface DuplexOptions extends ReadableOptions, WritableOptions {
    allowHalfOpen?: boolean;
    readable?: boolean;
    writable?: boolean;
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
    _customWritev: Function | null = null;
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
            if (options.writev !== undefined) {
                this._customWritev = options.writev as Function;
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

    _write(chunk: StreamChunk, encoding: string, callback: Function): void {
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

    _writev(chunks: StreamWriteReq[], callback: Function): void {
        if (typeof this._customWritev === "function") {
            const fn = this._customWritev as Function;
            fn(chunks, callback);
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

    write(chunk: StreamChunk, encodingOrCallback?: string | Function, callback?: Function): boolean {
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

        let chunkLen = 1;
        if (typeof chunk === "string") {
            chunkLen = (chunk as string).length;
        } else if (chunk && typeof (chunk as { length?: number }).length === "number") {
            chunkLen = (chunk as { length: number }).length;
        }
        this.writableLength += chunkLen;

        const self = this;
        const onDone = (err?: Error | null) => {
            self.writableLength = Math.max(0, self.writableLength - chunkLen);
            if (err) {
                self.destroy(err);
                if (cb !== null) {
                    cb(err);
                }
                return;
            }
            if (cb !== null) {
                cb(null);
            }
            if (self.writableNeedDrain && self.writableLength < self.writableHighWaterMark) {
                self.writableNeedDrain = false;
                self.emit("drain");
            }
        };

        if (this.writableCorked > 0) {
            this._writeBuffer.push(new StreamWriteReq(chunk, encoding, onDone));
        } else {
            this._write(chunk, encoding, onDone);
        }

        const canWrite = this.writableLength < this.writableHighWaterMark;
        if (!canWrite) {
            this.writableNeedDrain = true;
        }
        return canWrite;
    }

    end(chunkOrCallback?: StreamChunk | Function, encodingOrCallback?: string | Function, callback?: Function): this {
        let chunk: StreamChunk | null = null;
        let encoding: string = this._defaultEncoding;
        let cb: Function | null = null;

        if (typeof chunkOrCallback === "function") {
            cb = chunkOrCallback as Function;
        } else if (chunkOrCallback !== undefined && chunkOrCallback !== null) {
            chunk = chunkOrCallback as StreamChunk;
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
                if (cb !== null) {
                    cb(err);
                }
                return;
            }
            self.writableFinished = true;
            self.emit("finish");
            if (cb !== null) {
                cb(null);
            }
            if (!self._allowHalfOpen && self._autoDestroy) {
                self.destroy();
            }
        });

        return this;
    }

    static from(src: unknown, options?: ReadableOptions): Duplex {
        if (src instanceof Duplex) {
            return src as Duplex;
        }
        let duplex: Duplex = new Duplex();
        if (options !== undefined && options !== null) {
            duplex = new Duplex(options);
        }
        if (src instanceof Stream) {
            const s = src as Stream;
            const res = new PassThrough();
            s.pipe(res);
            return res;
        }
        return duplex;
    }

    static fromWeb(src: unknown, options?: ReadableOptions): Duplex {
        const duplex = new Duplex(options);
        if (src && typeof src === "object") {
            const pair = src as { readable?: unknown; writable?: unknown };
            if (pair.readable) {
                const r = Readable.fromWeb(pair.readable, options);
                r.on("data", (chunk: StreamChunk) => duplex.push(chunk));
                r.once("end", () => duplex.push(null));
                r.once("error", (err: Error) => duplex.destroy(err));
            }
            if (pair.writable) {
                const w = Writable.fromWeb(pair.writable, options);
                duplex._customWrite = (chunk: StreamChunk, enc: string, cb: Function) => {
                    w.write(chunk, enc, cb);
                };
                duplex._customFinal = (cb: Function) => {
                    w.end(() => cb(null));
                };
            }
        }
        return duplex;
    }

    static toWeb(stream: unknown, options?: unknown): unknown {
        const duplex = stream as Duplex;
        return {
            readable: Readable.toWeb(duplex),
            writable: Writable.toWeb(duplex)
        };
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

    _transform(chunk: StreamChunk, encoding: string, callback: Function): void {
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

    _write(chunk: StreamChunk, encoding: string, callback: Function): void {
        const self = this;
        this._transform(chunk, encoding, (err?: Error | null, data?: StreamChunk) => {
            if (err) {
                callback(err);
                return;
            }
            if (data !== undefined && data !== null) {
                self.push(data);
            }
            callback(null);
        });
    }

    _final(callback: Function): void {
        const self = this;
        this._flush((err?: Error | null, data?: StreamChunk) => {
            if (err) {
                callback(err);
                return;
            }
            if (data !== undefined && data !== null) {
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
    _transform(chunk: StreamChunk, encoding: string, callback: Function): void {
        callback(null, chunk);
    }
}

export function isReadable(stream: Stream | null | undefined): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.readable !== false && !s.destroyed;
    }
    return false;
}

export function isWritable(stream: Stream | null | undefined): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.writable !== false && !s.destroyed;
    }
    return false;
}

export function isErrored(stream: Stream | null | undefined): boolean {
    if (stream === null || stream === undefined) {
        return false;
    }
    if (stream instanceof Stream) {
        const s = stream as Stream;
        return s.errored !== null && s.errored !== undefined;
    }
    return false;
}

export function addAbortSignal(signal: AbortSignal | AbortSignalLike | null | undefined, stream: Stream): Stream {
    if (signal !== null && signal !== undefined && stream instanceof Stream) {
        const sig = signal as { aborted?: boolean; addEventListener?: Function };
        const str = stream as Stream;
        if (sig.aborted) {
            str.destroy(new Error("The operation was aborted"));
        } else if (typeof sig.addEventListener === "function") {
            sig.addEventListener("abort", () => {
                str.destroy(new Error("The operation was aborted"));
            });
        }
    }
    return stream;
}

export function finished(
    stream: Stream,
    optionsOrCallback?: FinishedOptions | Function,
    callback?: Function
): Function | Promise<void> | null {
    let cb: Function | null = null;
    let signal: AbortSignal | AbortSignalLike | undefined = undefined;

    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as Function;
    } else if (typeof callback === "function") {
        cb = callback as Function;
        if (optionsOrCallback && typeof optionsOrCallback === "object") {
            signal = (optionsOrCallback as FinishedOptions).signal;
        }
    } else if (optionsOrCallback && typeof optionsOrCallback === "object") {
        signal = (optionsOrCallback as FinishedOptions).signal;
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
            const onDone = (err?: Error | null) => {
                if (!done) {
                    done = true;
                    queueMicrotask(() => {
                        fn(err !== undefined ? err : null);
                    });
                }
            };
            s.once("finish", () => onDone());
            s.once("end", () => onDone());
            s.once("close", () => onDone());
            s.once("error", (err: Error) => onDone(err));
            if (signal) {
                addAbortSignal(signal, s);
            }
        }
        return null;
    }

    return new Promise<void>((resolve, reject) => {
        finished(stream, optionsOrCallback, (err: Error | null) => {
            if (err) {
                reject(err);
            } else {
                resolve();
            }
        });
    });
}

function _extractStreams(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown,
    fifth?: unknown
): Stream[] {
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

function _extractCallback(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown,
    fifth?: unknown
): Function | null {
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
): Stream | Promise<Stream | null> | null {
    const streams: Stream[] = _extractStreams(first, second, third, fourth, fifth);
    const cb: Function | null = _extractCallback(first, second, third, fourth, fifth);

    if (cb !== null) {
        const fn = cb as Function;
        if (streams.length === 0) {
            fn(null);
            return null;
        }

        let done = false;
        const destroyAll = (err?: Error | null) => {
            if (!done) {
                done = true;
                for (let i = 0; i < streams.length; i++) {
                    streams[i].destroy(err);
                }
                queueMicrotask(() => {
                    fn(err !== undefined ? err : null);
                });
            }
        };

        const last = streams[streams.length - 1];
        const onFinish = () => {
            if (!done) {
                done = true;
                queueMicrotask(() => {
                    fn(null);
                });
            }
        };
        const onError = (err: Error) => {
            destroyAll(err);
        };

        last.once("finish", onFinish);
        last.once("end", onFinish);
        for (let i = 0; i < streams.length; i++) {
            streams[i].once("error", onError);
        }

        for (let i = 0; i < streams.length - 1; i++) {
            const current = streams[i];
            const next = streams[i + 1];
            current.pipe(next);
        }

        return last;
    }

    return new Promise<Stream | null>((resolve, reject) => {
        if (streams.length === 0) {
            resolve(null);
            return;
        }

        let done = false;
        const destroyAll = (err?: Error | null) => {
            if (!done) {
                done = true;
                for (let i = 0; i < streams.length; i++) {
                    streams[i].destroy(err);
                }
                reject(err || new Error("stream pipeline failed"));
            }
        };

        const last = streams[streams.length - 1];
        const onFinish = () => {
            if (!done) {
                done = true;
                resolve(last);
            }
        };
        const onError = (err: Error) => {
            destroyAll(err);
        };

        last.once("finish", onFinish);
        last.once("end", onFinish);
        for (let i = 0; i < streams.length; i++) {
            streams[i].once("error", onError);
        }

        for (let i = 0; i < streams.length - 1; i++) {
            const current = streams[i];
            const next = streams[i + 1];
            current.pipe(next);
        }
    });
}

function _processComposeArg(arg: Stream | Stream[] | undefined, list: Stream[]): void {
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
    first?: Stream | Stream[],
    second?: Stream,
    third?: Stream,
    fourth?: Stream
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
    duplex._customWrite = (chunk: StreamChunk, encoding: string, callback: Function) => {
        firstStream.write(chunk);
        callback(null);
    };
    duplex._customFinal = (callback: Function) => {
        firstStream.end();
        callback(null);
    };

    lastStream.on("data", (chunk: StreamChunk) => duplex.push(chunk));
    lastStream.on("end", () => duplex.push(null));
    lastStream.on("error", (err: Error) => {
        duplex.destroy(err);
    });

    return duplex;
}

export class StreamPromises {
    _tag: string = "promises";

    finished(stream: Stream, options?: FinishedOptions): Promise<void> {
        return new Promise<void>((resolve, reject) => {
            finished(stream, options, (err: Error | null) => {
                if (err) {
                    reject(err);
                } else {
                    resolve();
                }
            });
        });
    }

    pipeline(
        first?: unknown,
        second?: unknown,
        third?: unknown,
        fourth?: unknown,
        fifth?: unknown
    ): Promise<Stream | null> {
        return pipeline(first, second, third, fourth, fifth) as Promise<Stream | null>;
    }
}

export class StreamConsumers {
    _tag: string = "consumers";

    private _chunkBuffer(chunk: StreamChunk): Buffer {
        if (typeof chunk === "string") {
            return Buffer.from(chunk);
        }
        if (Buffer.isBuffer(chunk)) {
            return chunk as Buffer;
        }
        return Buffer.from(chunk);
    }

    async buffer(stream: Stream | Readable): Promise<Buffer> {
        if (!stream) {
            return Buffer.alloc(0);
        }

        if (stream instanceof Readable) {
            const readable = stream as Readable;
            return new Promise<Buffer>((resolve, reject) => {
                const chunks: Buffer[] = [];
                readable.once("error", (err: Error) => reject(err));
                readable.once("end", () => resolve(Buffer.concat(chunks)));
                readable.on("data", (chunk: StreamChunk) => {
                    chunks.push(this._chunkBuffer(chunk));
                });
            });
        }

        if (stream instanceof Stream) {
            const s = stream as Stream;
            return new Promise<Buffer>((resolve, reject) => {
                const chunks: Buffer[] = [];
                s.once("error", (err: Error) => reject(err));
                s.once("end", () => resolve(Buffer.concat(chunks)));
                s.on("data", (chunk: StreamChunk) => {
                    chunks.push(this._chunkBuffer(chunk));
                });
            });
        }

        return Buffer.alloc(0);
    }

    async text(stream: Stream | Readable): Promise<string> {
        return (await this.buffer(stream)).toString();
    }

    async json(stream: Stream | Readable): Promise<unknown> {
        const t = await this.text(stream);
        return JSON.parse(t);
    }

    async arrayBuffer(stream: Stream | Readable): Promise<ArrayBuffer> {
        const bytes = await this.buffer(stream);
        const result = new ArrayBuffer(bytes.length);
        new Uint8Array(result).set(bytes);
        return result;
    }

    async blob(stream: Stream | Readable): Promise<Blob> {
        return new Blob([await this.buffer(stream)]);
    }
}

export function destroy(stream: Stream, err?: Error | null): void {
    if (stream instanceof Stream) {
        stream.destroy(err);
    }
}

export function isDisturbed(stream: Stream | null | undefined): boolean {
    return Readable.isDisturbed(stream);
}

export const promises: StreamPromises = new StreamPromises();
export const consumers: StreamConsumers = new StreamConsumers();
export const from = Readable.from;
export const fromWeb = Readable.fromWeb;
export const toWeb = Readable.toWeb;

export default {
    Stream,
    Readable,
    Writable,
    Duplex,
    Transform,
    PassThrough,
    pipeline,
    finished,
    destroy,
    compose,
    isDisturbed,
    from,
    fromWeb,
    toWeb,
    getDefaultHighWaterMark,
    setDefaultHighWaterMark,
    promises,
    consumers,
};
