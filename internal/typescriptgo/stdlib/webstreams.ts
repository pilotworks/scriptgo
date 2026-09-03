// Node.js Web Streams module (node:stream/web or webstreams)
import {
    deflateSync,
    gzipSync,
    inflateSync,
    gunzipSync,
} from "node:zlib";

type StreamResult = { done: boolean; value: unknown };
type ByteStreamResult = { done: boolean; value: ArrayBufferView | undefined };

interface BlobOptions {
    type?: string;
}

interface ReadableSource {
    start?: (controller: ReadableStreamDefaultController) => unknown;
    pull?: (controller: ReadableStreamDefaultController) => unknown;
    cancel?: (reason?: unknown) => unknown;
}

interface WritableSink {
    start?: (controller: WritableStreamDefaultController) => unknown;
    write?: (chunk: unknown, controller: WritableStreamDefaultController) => unknown;
    close?: () => unknown;
    abort?: (reason?: unknown) => unknown;
}

interface TransformDefinition {
    start?: (controller: TransformStreamDefaultController) => unknown;
    transform?: (chunk: unknown, controller: TransformStreamDefaultController) => unknown;
    flush?: (controller: TransformStreamDefaultController) => unknown;
}

interface SyncIterator<T> {
    next(): { done: boolean; value: T };
}

interface SyncIterable<T> {
    [Symbol.iterator](): SyncIterator<T>;
}

interface AsyncIterableLike<T> {
    [Symbol.asyncIterator](): unknown;
}

function bytesFromString(value: string): Uint8Array {
    const bytes = new Uint8Array(value.length);
    for (let i = 0; i < value.length; i++) {
        bytes[i] = value.charCodeAt(i) & 0xff;
    }
    return bytes;
}

function copyBytes(value: Uint8Array): Uint8Array {
    const result = new Uint8Array(value.length);
    result.set(value);
    return result;
}

function bytesFromPart(part: unknown): Uint8Array {
    if (typeof part === "string") {
        return bytesFromString(part);
    }
    if (part instanceof Uint8Array) {
        return copyBytes(part);
    }
    if (part instanceof ArrayBuffer) {
        return copyBytes(new Uint8Array(part));
    }
    if (part !== null && typeof part === "object") {
        const view = part as ArrayBufferView;
        if (view.buffer instanceof ArrayBuffer) {
            return copyBytes(new Uint8Array(view.buffer, view.byteOffset, view.byteLength));
        }
    }
    return bytesFromString(String(part));
}

function concatBytes(parts: Uint8Array[]): Uint8Array {
    let length = 0;
    for (let i = 0; i < parts.length; i++) {
        length += parts[i].length;
    }
    const result = new Uint8Array(length);
    let offset = 0;
    for (let i = 0; i < parts.length; i++) {
        result.set(parts[i], offset);
        offset += parts[i].length;
    }
    return result;
}

function arrayBufferFromBytes(bytes: Uint8Array): ArrayBuffer {
    const result = new ArrayBuffer(bytes.length);
    new Uint8Array(result).set(bytes);
    return result;
}

function bytesFromChunk(chunk: unknown): Uint8Array {
    if (typeof chunk === "string") {
        return bytesFromString(chunk);
    }
    return bytesFromPart(chunk);
}

function textFromBytes(bytes: Uint8Array): string {
    return new TextDecoder().decode(bytes);
}

export class Blob {
    size: number = 0;
    type: string = "";
    private _bytes: Uint8Array = new Uint8Array(0);

    constructor(blobParts: unknown[] = [], options?: BlobOptions) {
        const parts: Uint8Array[] = [];
        for (let i = 0; i < blobParts.length; i++) {
            parts.push(bytesFromPart(blobParts[i]));
        }
        this._bytes = concatBytes(parts);
        this.size = this._bytes.length;
        this.type = options && typeof options.type === "string" ? options.type.toLowerCase() : "";
    }

    async arrayBuffer(): Promise<ArrayBuffer> {
        return arrayBufferFromBytes(this._bytes);
    }

    async text(): Promise<string> {
        return textFromBytes(this._bytes);
    }
}

export class ReadableStreamDefaultReader {
    closed: Promise<undefined> = Promise.resolve(undefined);
    private _stream: ReadableStream | null;

    constructor(stream?: ReadableStream) {
        this._stream = stream === undefined ? null : stream;
    }

    read(): Promise<StreamResult> {
        if (this._stream === null) {
            return Promise.resolve({ done: true, value: undefined });
        }
        return this._stream._readChunk();
    }

    releaseLock(): void {
        if (this._stream !== null) {
            this._stream._releaseReader();
            this._stream = null;
        }
    }

    async cancel(reason?: unknown): Promise<void> {
        if (this._stream !== null) {
            await this._stream.cancel(reason);
        }
    }
}

export class ReadableStreamBYOBRequest {
    view: ArrayBufferView | null = null;
    private _controller: ReadableByteStreamController | null;

    constructor(controller?: ReadableByteStreamController) {
        this.view = null;
        this._controller = controller === undefined ? null : controller;
    }

    respond(bytesWritten: number): void {
        if (this._controller !== null) {
            this._controller._respond(bytesWritten);
        }
    }

    respondWithNewView(view: ArrayBufferView): void {
        this.view = view;
    }
}

export class ReadableByteStreamController {
    byobRequest: ReadableStreamBYOBRequest | null = null;
    desiredSize: number = 0;
    private _stream: ReadableStream | null;

    constructor(stream?: ReadableStream) {
        this._stream = stream === undefined ? null : stream;
        this.byobRequest = null;
        this.desiredSize = 0;
    }

    close(): void {
        if (this._stream !== null) {
            this._stream._closeStream();
        }
    }

    enqueue(chunk: ArrayBufferView): void {
        if (this._stream !== null) {
            this._stream._enqueueChunk(chunk);
            this.desiredSize = this._stream._desiredSize();
        }
    }

    error(e?: unknown): void {
        if (this._stream !== null) {
            this._stream._errorStream(e);
        }
    }

    _respond(bytesWritten: number): void {
        this.desiredSize = this.desiredSize - bytesWritten;
    }
}

export class ReadableStreamBYOBReader {
    closed: Promise<undefined> = Promise.resolve(undefined);
    private _stream: ReadableStream | null;

    constructor(stream?: ReadableStream) {
        this._stream = stream === undefined ? null : stream;
    }

    async read(view: ArrayBufferView): Promise<ByteStreamResult> {
        if (this._stream === null) {
            return { done: true, value: undefined };
        }
        const result = await this._stream._readChunk();
        if (result.done || result.value === undefined) {
            return { done: result.done, value: undefined };
        }
        if (result.value instanceof Uint8Array && view instanceof Uint8Array) {
            const length = result.value.length < view.length ? result.value.length : view.length;
            for (let i = 0; i < length; i++) {
                view[i] = result.value[i];
            }
        }
        return { done: false, value: view };
    }

    releaseLock(): void {
        if (this._stream !== null) {
            this._stream._releaseReader();
            this._stream = null;
        }
    }

    async cancel(reason?: unknown): Promise<void> {
        if (this._stream !== null) {
            await this._stream.cancel(reason);
        }
    }
}

export class ReadableStreamDefaultController {
    desiredSize: number = 0;
    private _stream: ReadableStream | null;

    constructor(stream?: ReadableStream) {
        this._stream = stream === undefined ? null : stream;
        this.desiredSize = stream === undefined ? 0 : stream._desiredSize();
    }

    close(): void {
        if (this._stream !== null) {
            this._stream._closeStream();
        }
    }

    enqueue(chunk?: unknown): void {
        if (this._stream !== null) {
            this._stream._enqueueChunk(chunk);
            this.desiredSize = this._stream._desiredSize();
        }
    }

    error(e?: unknown): void {
        if (this._stream !== null) {
            this._stream._errorStream(e);
        }
    }
}

export class ReadableStream {
    locked: boolean = false;
    private _queue: unknown[] = [];
    private _closed: boolean = false;
    private _error: unknown = undefined;
    private _source: ReadableSource | null = null;
    private _controller: ReadableStreamDefaultController;

    constructor(underlyingSource?: ReadableSource, strategy?: unknown) {
        this.locked = false;
        this._queue = [];
        this._closed = false;
        this._error = undefined;
        this._source = underlyingSource === undefined ? null : underlyingSource;
        this._controller = new ReadableStreamDefaultController(this);
        if (this._source !== null && typeof this._source.start === "function") {
            this._source.start(this._controller);
        }
    }

    async cancel(reason?: unknown): Promise<void> {
        if (this._closed) {
            return;
        }
        this._closed = true;
        if (this._source !== null && typeof this._source.cancel === "function") {
            await this._source.cancel(reason);
        }
    }

    getReader(): ReadableStreamDefaultReader;
    getReader(options: { mode: "byob" }): ReadableStreamBYOBReader;
    getReader(options?: { mode?: string }): ReadableStreamDefaultReader | ReadableStreamBYOBReader {
        if (this.locked) {
            throw new TypeError("ReadableStream is locked");
        }
        this.locked = true;
        if (options && options.mode === "byob") {
            return new ReadableStreamBYOBReader(this);
        }
        return new ReadableStreamDefaultReader(this);
    }

    pipeThrough(transform: TransformStream, options?: unknown): ReadableStream {
        this.pipeTo(transform.writable, options);
        return transform.readable;
    }

    async pipeTo(destination: WritableStream, options?: unknown): Promise<void> {
        const reader = this.getReader();
        const writer = destination.getWriter();
        while (true) {
            const result = await reader.read();
            if (result.done) {
                break;
            }
            await writer.write(result.value);
        }
        await writer.close();
        reader.releaseLock();
        writer.releaseLock();
    }

    tee(): [ReadableStream, ReadableStream] {
        const first = new ReadableStream();
        const second = new ReadableStream();
        for (let i = 0; i < this._queue.length; i++) {
            first._enqueueChunk(this._queue[i]);
            second._enqueueChunk(this._queue[i]);
        }
        if (this._closed) {
            first._closeStream();
            second._closeStream();
        }
        return [first, second];
    }

    values(options?: unknown): AsyncIterableIterator<unknown> {
        const reader = this.getReader();
        return {
            [Symbol.asyncIterator]() {
                return this;
            },
            async next() {
                const result = await reader.read();
                return { done: result.done, value: result.value };
            },
        };
    }

    async _readChunk(): Promise<StreamResult> {
        if (this._queue.length > 0) {
            return { done: false, value: this._queue.shift() };
        }
        if (!this._closed && this._source !== null && typeof this._source.pull === "function") {
            await this._source.pull(this._controller);
            if (this._queue.length > 0) {
                return { done: false, value: this._queue.shift() };
            }
        }
        if (this._error !== undefined) {
            throw this._error;
        }
        // A stream without an underlying source has no producer to wait for.
        return { done: this._closed || this._source === null, value: undefined };
    }

    _enqueueChunk(chunk: unknown): void {
        if (!this._closed) {
            this._queue.push(chunk);
        }
    }

    _closeStream(): void {
        this._closed = true;
    }

    _errorStream(error?: unknown): void {
        this._error = error === undefined ? new Error("ReadableStream error") : error;
        this._closed = true;
    }

    _releaseReader(): void {
        this.locked = false;
    }

    _desiredSize(): number {
        return this._closed ? 0 : 1 - this._queue.length;
    }
}

export class StreamAbortSignal {
    aborted: boolean = false;
    reason: unknown = undefined;
}

export class WritableStreamDefaultController {
    signal: StreamAbortSignal = new StreamAbortSignal();
    private _stream: WritableStream | null;

    constructor(stream?: WritableStream) {
        this._stream = stream === undefined ? null : stream;
        this.signal = new StreamAbortSignal();
    }

    error(e?: unknown): void {
        this.signal.aborted = true;
        this.signal.reason = e;
        if (this._stream !== null) {
            this._stream._errorStream(e);
        }
    }
}

export class WritableStreamDefaultWriter {
    closed: Promise<undefined> = Promise.resolve(undefined);
    ready: Promise<undefined> = Promise.resolve(undefined);
    desiredSize: number = 0;
    private _stream: WritableStream | null;

    constructor(stream?: WritableStream) {
        this._stream = stream === undefined ? null : stream;
        this.closed = Promise.resolve(undefined);
        this.ready = Promise.resolve(undefined);
        this.desiredSize = stream === undefined ? 0 : stream._desiredSize();
    }

    async abort(reason?: unknown): Promise<void> {
        if (this._stream !== null) {
            await this._stream.abort(reason);
        }
    }

    async close(): Promise<void> {
        if (this._stream !== null) {
            await this._stream.close();
        }
    }

    releaseLock(): void {
        if (this._stream !== null) {
            this._stream._releaseWriter();
            this._stream = null;
        }
    }

    async write(chunk?: unknown): Promise<void> {
        if (this._stream !== null) {
            await this._stream._writeChunk(chunk);
            this.desiredSize = this._stream._desiredSize();
        }
    }
}

export class WritableStream {
    locked: boolean = false;
    private _queue: unknown[] = [];
    private _closed: boolean = false;
    private _error: unknown = undefined;
    private _sink: WritableSink | null = null;
    private _controller: WritableStreamDefaultController;

    constructor(underlyingSink?: WritableSink, strategy?: unknown) {
        this.locked = false;
        this._queue = [];
        this._closed = false;
        this._error = undefined;
        this._sink = underlyingSink === undefined ? null : underlyingSink;
        this._controller = new WritableStreamDefaultController(this);
        if (this._sink !== null && typeof this._sink.start === "function") {
            this._sink.start(this._controller);
        }
    }

    async abort(reason?: unknown): Promise<void> {
        this._closed = true;
        if (this._sink !== null && typeof this._sink.abort === "function") {
            await this._sink.abort(reason);
        }
    }

    async close(): Promise<void> {
        if (this._closed) {
            return;
        }
        this._closed = true;
        if (this._sink !== null && typeof this._sink.close === "function") {
            await this._sink.close();
        }
    }

    getWriter(): WritableStreamDefaultWriter {
        if (this.locked) {
            throw new TypeError("WritableStream is locked");
        }
        this.locked = true;
        return new WritableStreamDefaultWriter(this);
    }

    async _writeChunk(chunk?: unknown): Promise<void> {
        if (this._closed) {
            throw new TypeError("WritableStream is closed");
        }
        this._queue.push(chunk);
        if (this._sink !== null && typeof this._sink.write === "function") {
            await this._sink.write(chunk, this._controller);
        }
        if (this._error !== undefined) {
            throw this._error;
        }
    }

    _errorStream(error?: unknown): void {
        this._error = error === undefined ? new Error("WritableStream error") : error;
    }

    _releaseWriter(): void {
        this.locked = false;
    }

    _desiredSize(): number {
        return this._closed ? 0 : 1 - this._queue.length;
    }
}

export class TransformStreamDefaultController {
    desiredSize: number = 0;
    private _readableController: ReadableStreamDefaultController | null;

    constructor(controller?: ReadableStreamDefaultController) {
        this.desiredSize = 0;
        this._readableController = controller === undefined ? null : controller;
    }

    enqueue(chunk?: unknown): void {
        if (this._readableController !== null) {
            this._readableController.enqueue(chunk);
            this.desiredSize = this._readableController.desiredSize;
        }
    }

    error(reason?: unknown): void {
        if (this._readableController !== null) {
            this._readableController.error(reason);
        }
    }

    terminate(): void {
        if (this._readableController !== null) {
            this._readableController.close();
        }
    }
}

export class TransformStream {
    readable: ReadableStream;
    writable: WritableStream;
    private _definition: TransformDefinition | null;
    private _controller: TransformStreamDefaultController | null = null;

    constructor(transformer?: TransformDefinition, writableStrategy?: unknown, readableStrategy?: unknown) {
        this._definition = transformer === undefined ? null : transformer;
        let readableController: ReadableStreamDefaultController | null = null;
        this.readable = new ReadableStream({
            start: (controller) => {
                readableController = controller;
            },
        });
        this._controller = new TransformStreamDefaultController(readableController === null ? undefined : readableController);
        const definition = this._definition;
        const controller = this._controller;
        this.writable = new WritableStream({
            write: async (chunk) => {
                if (definition !== null && typeof definition.transform === "function") {
                    await definition.transform(chunk, controller);
                } else {
                    controller.enqueue(chunk);
                }
            },
            close: () => {
                if (definition !== null && typeof definition.flush === "function") {
                    definition.flush(controller);
                }
            },
        });
        if (definition !== null && typeof definition.start === "function") {
            definition.start(controller);
        }
    }
}

export class ByteLengthQueuingStrategy {
    highWaterMark: number;

    constructor(options: { highWaterMark: number }) {
        this.highWaterMark = options.highWaterMark;
    }

    size(chunk: { byteLength?: number }): number {
        return chunk.byteLength === undefined ? 0 : chunk.byteLength;
    }
}

export class CountQueuingStrategy {
    highWaterMark: number;

    constructor(options: { highWaterMark: number }) {
        this.highWaterMark = options.highWaterMark;
    }

    size(chunk: unknown): number {
        return 1;
    }
}

export class TextEncoderStream {
    encoding: string = "utf-8";
    readable: ReadableStream;
    writable: WritableStream;

    constructor() {
        this.encoding = "utf-8";
        const stream = new TransformStream({
            transform: (chunk, controller) => {
                controller.enqueue(new TextEncoder().encode(typeof chunk === "string" ? chunk : String(chunk)));
            },
        });
        this.readable = stream.readable;
        this.writable = stream.writable;
    }
}

export class TextDecoderStream {
    encoding: string = "utf-8";
    fatal: boolean = false;
    ignoreBOM: boolean = false;
    readable: ReadableStream;
    writable: WritableStream;

    constructor(label: string = "utf-8", options?: { fatal?: boolean; ignoreBOM?: boolean }) {
        this.encoding = label.toLowerCase();
        this.fatal = options !== undefined && options.fatal === true;
        this.ignoreBOM = options !== undefined && options.ignoreBOM === true;
        const stream = new TransformStream({
            transform: (chunk, controller) => {
                controller.enqueue(textFromBytes(bytesFromChunk(chunk)));
            },
        });
        this.readable = stream.readable;
        this.writable = stream.writable;
    }
}

function compressionTransform(format: string, compress: boolean): (chunk: unknown) => Uint8Array {
    if (format !== "gzip" && format !== "deflate") {
        throw new TypeError("Unsupported compression format: " + format);
    }
    return (chunk: unknown) => {
        const input = bytesFromChunk(chunk);
        if (format === "gzip") {
            return compress ? gzipSync(input) : gunzipSync(input);
        }
        return compress ? deflateSync(input) : inflateSync(input);
    };
}

export class CompressionStream {
    readable: ReadableStream;
    writable: WritableStream;

    constructor(format: string) {
        const transform = compressionTransform(format, true);
        const stream = new TransformStream({
            transform: (chunk, controller) => controller.enqueue(transform(chunk)),
        });
        this.readable = stream.readable;
        this.writable = stream.writable;
    }
}

export class DecompressionStream {
    readable: ReadableStream;
    writable: WritableStream;

    constructor(format: string) {
        const transform = compressionTransform(format, false);
        const stream = new TransformStream({
            transform: (chunk, controller) => controller.enqueue(transform(chunk)),
        });
        this.readable = stream.readable;
        this.writable = stream.writable;
    }
}

export function from<T>(iterable: SyncIterable<T> | AsyncIterableLike<T>): ReadableStream {
    const values: unknown[] = [];
    if (Array.isArray(iterable)) {
        for (let i = 0; i < iterable.length; i++) {
            values.push(iterable[i]);
        }
    }
    return new ReadableStream({
        start: (controller) => {
            for (let i = 0; i < values.length; i++) {
                controller.enqueue(values[i]);
            }
            controller.close();
        },
    });
}

async function readAll(stream: ReadableStream): Promise<unknown[]> {
    const reader = stream.getReader();
    const chunks: unknown[] = [];
    while (true) {
        const result = await reader.read();
        if (result.done) {
            break;
        }
        chunks.push(result.value);
    }
    reader.releaseLock();
    return chunks;
}

export async function arrayBuffer(stream: ReadableStream): Promise<ArrayBuffer> {
    const chunks = await readAll(stream);
    const parts: Uint8Array[] = [];
    for (let i = 0; i < chunks.length; i++) {
        parts.push(bytesFromChunk(chunks[i]));
    }
    return arrayBufferFromBytes(concatBytes(parts));
}

export async function blob(stream: ReadableStream): Promise<Blob> {
    const bytes = new Uint8Array(await arrayBuffer(stream));
    return new Blob([bytes]);
}

export async function buffer(stream: ReadableStream): Promise<Buffer> {
    const bytes = new Uint8Array(await arrayBuffer(stream));
    return Buffer.from(bytes);
}

export async function json(stream: ReadableStream): Promise<unknown> {
    return JSON.parse(await text(stream));
}

export async function text(stream: ReadableStream): Promise<string> {
    const chunks = await readAll(stream);
    const parts: string[] = [];
    for (let i = 0; i < chunks.length; i++) {
        const chunk = chunks[i];
        if (typeof chunk === "string") {
            parts.push(chunk);
        } else {
            parts.push(textFromBytes(bytesFromChunk(chunk)));
        }
    }
    return parts.join("");
}

export default {
    Blob,
    ReadableStream,
    ReadableStreamDefaultReader,
    ReadableStreamBYOBReader,
    ReadableStreamDefaultController,
    ReadableByteStreamController,
    ReadableStreamBYOBRequest,
    WritableStream,
    WritableStreamDefaultWriter,
    WritableStreamDefaultController,
    TransformStream,
    TransformStreamDefaultController,
    ByteLengthQueuingStrategy,
    CountQueuingStrategy,
    TextEncoderStream,
    TextDecoderStream,
    CompressionStream,
    DecompressionStream,
    from,
    arrayBuffer,
    blob,
    buffer,
    json,
    text,
};
