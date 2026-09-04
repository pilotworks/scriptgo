import { Blob, Buffer } from "node:buffer";
import { consumers, Readable, Stream } from "node:stream";
import { ReadableStream } from "node:stream/web";

type ConsumableStream = Readable | Stream | ReadableStream | unknown;

export async function buffer(stream: ConsumableStream): Promise<Buffer> {
    if (stream instanceof ReadableStream) {
        const reader = stream.getReader();
        const chunks: Buffer[] = [];
        while (true) {
            const res = await reader.read();
            if (res.done) {
                break;
            }
            if (typeof res.value === "string") {
                chunks.push(Buffer.from(res.value));
            } else if (res.value instanceof Uint8Array) {
                chunks.push(Buffer.from(res.value));
            } else if (Buffer.isBuffer(res.value)) {
                chunks.push(res.value as Buffer);
            }
        }
        return Buffer.concat(chunks);
    }
    return consumers.buffer(stream as Stream | Readable);
}

export async function text(stream: ConsumableStream): Promise<string> {
    return (await buffer(stream)).toString();
}

export async function json(stream: ConsumableStream): Promise<unknown> {
    const t = await text(stream);
    return JSON.parse(t);
}

export async function arrayBuffer(stream: ConsumableStream): Promise<ArrayBuffer> {
    const bytes = await buffer(stream);
    const result = new ArrayBuffer(bytes.length);
    new Uint8Array(result).set(bytes);
    return result;
}

export async function blob(stream: ConsumableStream): Promise<Blob> {
    return new Blob([await buffer(stream)]);
}

export default {
    buffer,
    text,
    json,
    arrayBuffer,
    blob,
};
