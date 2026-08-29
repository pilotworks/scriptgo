// Node.js Web Streams module (node:stream/web or webstreams)

export class Blob {
    size: number = 0;
    type: string = "";

    constructor(blobParts?: unknown[], options?: unknown) {
        this.size = 0;
        this.type = "";
    }

    async arrayBuffer(): Promise<ArrayBuffer> {
        return new ArrayBuffer(0);
    }

    async text(): Promise<string> {
        return "";
    }
}

export class ReadableStreamDefaultReader {
    closed: Promise<undefined> = Promise.resolve(undefined);

    constructor(stream?: ReadableStream) {
        this.closed = Promise.resolve(undefined);
    }

    async read(): Promise<{ done: boolean; value: unknown }> {
        return { done: true, value: undefined };
    }

    releaseLock(): void {}

    async cancel(reason?: unknown): Promise<void> {}
}

export class ReadableStreamBYOBRequest {
    view: ArrayBufferView | null = null;

    constructor() {
        this.view = null;
    }

    respond(bytesWritten: number): void {}
    respondWithNewView(view: ArrayBufferView): void {}
}

export class ReadableByteStreamController {
    byobRequest: ReadableStreamBYOBRequest | null = null;
    desiredSize: number = 0;

    constructor() {
        this.byobRequest = null;
        this.desiredSize = 0;
    }

    close(): void {}
    enqueue(chunk: ArrayBufferView): void {}
    error(e?: unknown): void {}
}

export class ReadableStreamBYOBReader {
    closed: Promise<undefined> = Promise.resolve(undefined);

    constructor(stream?: ReadableStream) {
        this.closed = Promise.resolve(undefined);
    }

    async read(view: ArrayBufferView): Promise<{ done: boolean; value: ArrayBufferView | undefined }> {
        return { done: true, value: undefined };
    }

    releaseLock(): void {}

    async cancel(reason?: unknown): Promise<void> {}
}

export class ReadableStreamDefaultController {
    desiredSize: number = 0;

    constructor() {
        this.desiredSize = 0;
    }

    close(): void {}
    enqueue(chunk?: unknown): void {}
    error(e?: unknown): void {}
}

export class TransformStream {
    readable: ReadableStream = new ReadableStream();
    writable: WritableStream = new WritableStream();

    constructor(transformer?: unknown, writableStrategy?: unknown, readableStrategy?: unknown) {
        this.readable = new ReadableStream();
        this.writable = new WritableStream();
    }
}

export class ReadableStream {
    locked: boolean = false;

    constructor(underlyingSource?: unknown, strategy?: unknown) {
        this.locked = false;
    }

    async cancel(reason?: unknown): Promise<void> {}

    getReader(): ReadableStreamDefaultReader {
        this.locked = true;
        return new ReadableStreamDefaultReader(this);
    }

    pipeThrough(transform: TransformStream, options?: unknown): ReadableStream {
        return transform.readable;
    }

    async pipeTo(destination: WritableStream, options?: unknown): Promise<void> {}

    tee(): [ReadableStream, ReadableStream] {
        return [new ReadableStream(), new ReadableStream()];
    }

    values(options?: unknown): AsyncIterableIterator<unknown> {
        const reader = this.getReader();
        return {
            [Symbol.asyncIterator]() {
                return this;
            },
            async next() {
                const res = await reader.read();
                return { done: res.done, value: res.value };
            },
        };
    }
}

export class StreamAbortSignal {
    aborted: boolean = false;
    reason: unknown = undefined;
}

export class WritableStreamDefaultController {
    signal: StreamAbortSignal = new StreamAbortSignal();

    error(e?: unknown): void {}
}

export class WritableStreamDefaultWriter {
    closed: Promise<undefined> = Promise.resolve(undefined);
    ready: Promise<undefined> = Promise.resolve(undefined);
    desiredSize: number = 0;

    constructor(stream?: WritableStream) {
        this.closed = Promise.resolve(undefined);
        this.ready = Promise.resolve(undefined);
        this.desiredSize = 0;
    }

    async abort(reason?: unknown): Promise<void> {}
    async close(): Promise<void> {}
    releaseLock(): void {}
    async write(chunk?: unknown): Promise<void> {}
}

export class WritableStream {
    locked: boolean = false;

    constructor(underlyingSink?: unknown, strategy?: unknown) {
        this.locked = false;
    }

    async abort(reason?: unknown): Promise<void> {}
    async close(): Promise<void> {}

    getWriter(): WritableStreamDefaultWriter {
        this.locked = true;
        return new WritableStreamDefaultWriter(this);
    }
}

export class TransformStreamDefaultController {
    desiredSize: number = 0;

    enqueue(chunk?: unknown): void {}
    error(reason?: unknown): void {}
    terminate(): void {}
}

export class ByteLengthQueuingStrategy {
    highWaterMark: number;

    constructor(options: { highWaterMark: number }) {
        this.highWaterMark = options.highWaterMark;
    }

    size(chunk: { byteLength?: number }): number {
        return 16;
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
    readable: ReadableStream = new ReadableStream();
    writable: WritableStream = new WritableStream();

    constructor() {
        this.encoding = "utf-8";
        this.readable = new ReadableStream();
        this.writable = new WritableStream();
    }
}

export class TextDecoderStream {
    encoding: string = "utf-8";
    fatal: boolean = false;
    ignoreBOM: boolean = false;
    readable: ReadableStream = new ReadableStream();
    writable: WritableStream = new WritableStream();

    constructor(label: string = "utf-8", options?: { fatal?: boolean; ignoreBOM?: boolean }) {
        this.encoding = label;
        this.fatal = options?.fatal || false;
        this.ignoreBOM = options?.ignoreBOM || false;
        this.readable = new ReadableStream();
        this.writable = new WritableStream();
    }
}

export class CompressionStream {
    readable: ReadableStream = new ReadableStream();
    writable: WritableStream = new WritableStream();

    constructor(format: string) {
        this.readable = new ReadableStream();
        this.writable = new WritableStream();
    }
}

export class DecompressionStream {
    readable: ReadableStream = new ReadableStream();
    writable: WritableStream = new WritableStream();

    constructor(format: string) {
        this.readable = new ReadableStream();
        this.writable = new WritableStream();
    }
}

export function from<T>(iterable: AsyncIterable<T> | Iterable<T>): ReadableStream {
    return new ReadableStream();
}

export async function arrayBuffer(stream: ReadableStream): Promise<ArrayBuffer> {
    return new ArrayBuffer(0);
}

export async function blob(stream: ReadableStream): Promise<Blob> {
    return new Blob();
}

export async function buffer(stream: ReadableStream): Promise<Buffer> {
    return Buffer.alloc(0);
}

export async function json(stream: ReadableStream): Promise<unknown> {
    return {};
}

export async function text(stream: ReadableStream): Promise<string> {
    return "";
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
