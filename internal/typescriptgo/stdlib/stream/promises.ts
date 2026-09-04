// ScriptGo Standard Library: node:stream/promises
import { promises, Stream, FinishedOptions } from "node:stream";

export function pipeline(
    first?: unknown,
    second?: unknown,
    third?: unknown,
    fourth?: unknown,
    fifth?: unknown
): Promise<Stream | null> {
    return promises.pipeline(first, second, third, fourth, fifth);
}

export function finished(stream: Stream, options?: FinishedOptions): Promise<void> {
    return promises.finished(stream, options);
}

export default {
    pipeline,
    finished,
};
