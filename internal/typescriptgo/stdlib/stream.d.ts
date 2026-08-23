export class Stream {
    static _defaultHighWaterMark: number;
    static _defaultObjectModeHighWaterMark: number;
    readable: boolean;
    writable: boolean;
    destroyed: boolean;
    errored: Error | null;
    closed: boolean;
    constructor();
    addListener(event: string, listener: Function): this;
    on(event: string, listener: Function): this;
    once(event: string, listener: Function): this;
    prependListener(event: string, listener: Function): this;
    prependOnceListener(event: string, listener: Function): this;
    removeListener(event: string, listener: Function): this;
    off(event: string, listener: Function): this;
    removeAllListeners(event?: string): this;
    setMaxListeners(n: number): this;
    getMaxListeners(): number;
    listenerCount(event: string): number;
    listeners(event: string): Function[];
    rawListeners(event: string): Function[];
    eventNames(): string[];
    emit(event: string, arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): boolean;
    destroy(error?: unknown): this;
    write(chunk: unknown, encoding?: unknown, callback?: unknown): boolean;
    end(chunk?: unknown, encoding?: unknown, callback?: unknown): this;
    read(size?: number): unknown;
    pause(): this;
    resume(): this;
    pipe(destination: unknown, options?: unknown): unknown;
}
export function getDefaultHighWaterMark(objectMode?: boolean): number;
export function setDefaultHighWaterMark(objectMode: boolean, value: number): void;
export interface ReadableOptions {
    highWaterMark?: number;
    encoding?: string;
    objectMode?: boolean;
    autoDestroy?: boolean;
    read?: Function;
    destroy?: Function;
}
export class Readable extends Stream {
    readable: boolean;
    readableEnded: boolean;
    readableFlowing: boolean;
    readableHighWaterMark: number;
    readableLength: number;
    readableObjectMode: boolean;
    readableEncoding: string;
    destroyed: boolean;
    errored: Error | null;
    closed: boolean;
    _buffer: string[];
    _reading: boolean;
    _paused: boolean;
    _autoDestroy: boolean;
    _disturbed: boolean;
    _customRead: Function | null;
    _customDestroy: Function | null;
    _pipeDests: Stream[];
    constructor(options?: ReadableOptions);
    _read(size: number): void;
    read(size?: number): string | null;
    push(chunk: string | null, encoding?: string): boolean;
    unshift(chunk: string): boolean;
    on(event: string, listener: Function): this;
    addListener(event: string, listener: Function): this;
    pause(): this;
    resume(): this;
    isPaused(): boolean;
    setEncoding(encoding: string): this;
    pipe(destination: unknown, options?: PipeOptions): unknown;
    unpipe(destination?: unknown): this;
    wrap(oldStream: unknown): this;
    destroy(error?: unknown): this;
    static from(iterable: unknown, options?: unknown): Readable;
    static isDisturbed(stream: unknown): boolean;
    static fromWeb(readableStream: unknown, options?: unknown): Readable;
    static toWeb(streamReadable: unknown, options?: unknown): unknown;
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
    writable: boolean;
    writableEnded: boolean;
    writableFinished: boolean;
    writableHighWaterMark: number;
    writableLength: number;
    writableNeedDrain: boolean;
    writableObjectMode: boolean;
    writableCorked: number;
    destroyed: boolean;
    errored: Error | null;
    closed: boolean;
    _defaultEncoding: string;
    _writeBuffer: StreamWriteReq[];
    _autoDestroy: boolean;
    _customWrite: Function | null;
    _customWritev: Function | null;
    _customFinal: Function | null;
    _customDestroy: Function | null;
    constructor(options?: WritableOptions);
    _write(chunk: string, encoding: string, callback: Function): void;
    _writev(chunks: StreamWriteReq[], callback: Function): void;
    _final(callback: Function): void;
    setDefaultEncoding(encoding: string): this;
    cork(): void;
    uncork(): void;
    write(chunk: unknown, encodingOrCallback?: unknown, callback?: unknown): boolean;
    end(chunkOrCallback?: unknown, encodingOrCallback?: unknown, callback?: unknown): this;
    destroy(error?: unknown): this;
    static fromWeb(writableStream: unknown, options?: unknown): Writable;
    static toWeb(streamWritable: unknown, options?: unknown): unknown;
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
    writable: boolean;
    writableEnded: boolean;
    writableFinished: boolean;
    writableHighWaterMark: number;
    writableLength: number;
    writableNeedDrain: boolean;
    writableObjectMode: boolean;
    writableCorked: number;
    _defaultEncoding: string;
    _writeBuffer: StreamWriteReq[];
    _allowHalfOpen: boolean;
    _customWrite: Function | null;
    _customFinal: Function | null;
    constructor(options?: DuplexOptions);
    _write(chunk: string, encoding: string, callback: Function): void;
    _final(callback: Function): void;
    setDefaultEncoding(encoding: string): this;
    cork(): void;
    uncork(): void;
    write(chunk: unknown, encodingOrCallback?: unknown, callback?: unknown): boolean;
    end(chunkOrCallback?: unknown, encodingOrCallback?: unknown, callback?: unknown): this;
    static from(src: unknown, options?: unknown): Duplex;
    static fromWeb(pair: unknown, options?: unknown): Duplex;
    static toWeb(streamDuplex: unknown, options?: unknown): unknown;
}
export interface TransformOptions extends DuplexOptions {
    transform?: Function;
    flush?: Function;
}
export class Transform extends Duplex {
    _customTransform: Function | null;
    _customFlush: Function | null;
    constructor(options?: TransformOptions);
    _transform(chunk: string, encoding: string, callback: Function): void;
    _flush(callback: Function): void;
    _write(chunk: string, encoding: string, callback: Function): void;
    _final(callback: Function): void;
}
export class PassThrough extends Transform {
    _transform(chunk: string, encoding: string, callback: Function): void;
}
export function isReadable(stream: unknown): boolean;
export function isWritable(stream: unknown): boolean;
export function isErrored(stream: unknown): boolean;
export function addAbortSignal(signal: unknown, stream: unknown): unknown;
export function finished(stream: unknown, optionsOrCallback?: unknown, callback?: unknown): unknown;
export function pipeline(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown,
    fifth?: unknown
): unknown;
export function compose(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown
): Duplex;
export class StreamPromises {
    _tag: string;
    finished(stream: unknown, options?: unknown): Promise<void>;
    pipeline(first?: unknown, second?: unknown, third?: unknown, fourth?: unknown, fifth?: unknown): Promise<unknown>;
}
export class StreamConsumers {
    _tag: string;
    async buffer(stream: unknown): Promise<Uint8Array>;
    async text(stream: unknown): Promise<string>;
    async json(stream: unknown): Promise<unknown>;
    async arrayBuffer(stream: unknown): Promise<ArrayBuffer>;
    async blob(stream: unknown): Promise<unknown>;
}
export const promises: StreamPromises;
export const consumers: StreamConsumers;
