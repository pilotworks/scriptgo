import { Blob, Buffer } from "node:buffer";
import { consumers, Readable } from "node:stream";

type ConsumableStream = Readable;

export function buffer(stream: ConsumableStream): Promise<Buffer> {
    return consumers.buffer(stream);
}

export function text(stream: ConsumableStream): Promise<string> {
    return consumers.text(stream);
}

export function json(stream: ConsumableStream): Promise<unknown> {
    return consumers.json(stream);
}

export function arrayBuffer(stream: ConsumableStream): Promise<ArrayBuffer> {
    return consumers.arrayBuffer(stream);
}

export function blob(stream: ConsumableStream): Promise<Blob> {
    return consumers.blob(stream);
}

export default {
    buffer,
    text,
    json,
    arrayBuffer,
    blob,
};
