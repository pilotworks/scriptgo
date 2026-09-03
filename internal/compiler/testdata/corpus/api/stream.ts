// ScriptGo Corpus: Stream Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    Stream,
    Readable,
    Writable,
    Duplex,
    Transform,
    PassThrough,
    pipeline,
    finished,
    promises,
    compose,
    addAbortSignal,
    getDefaultHighWaterMark,
    setDefaultHighWaterMark,
    isReadable,
    isWritable,
    isErrored,
    from,
    isDisturbed
} from "node:stream";
import {
    arrayBuffer as consumeArrayBuffer,
    blob as consumeBlob,
    buffer as consumeBuffer,
    json as consumeJSON,
    text as consumeText
} from "node:stream/consumers";

// @api: stream.getDefaultHighWaterMark
// @expect: 65536
// @expect: 16
console.log(getDefaultHighWaterMark(false));
console.log(getDefaultHighWaterMark(true));

// @api: stream.setDefaultHighWaterMark
// @expect: 32768
// @expect: 32
setDefaultHighWaterMark(false, 32768);
setDefaultHighWaterMark(true, 32);
console.log(getDefaultHighWaterMark(false));
console.log(getDefaultHighWaterMark(true));
// Restore defaults
setDefaultHighWaterMark(false, 65536);
setDefaultHighWaterMark(true, 16);

// @api: stream.push
// @expect: push_result: true
// @expect: chunk: hello
const r1 = new Readable({ read: () => {} });
const pushRes = r1.push("hello");
console.log("push_result: " + pushRes);
const readChunk = r1.read();
console.log("chunk: " + readChunk);

// @api: stream.read
// @expect: read0_result: null
const r2 = new Readable({ read: () => {} });
const read0 = r2.read(0);
console.log("read0_result: " + read0);

// @api: stream.isDisturbed
// @expect: disturbed_before: false
// @expect: disturbed_after: true
const rDist = new Readable({ read: () => {} });
console.log("disturbed_before: " + isDisturbed(rDist));
rDist.push("test");
rDist.read(1);
console.log("disturbed_after: " + isDisturbed(rDist));

// @api: stream.from
// @expect: duplexFrom_created: true
const dFrom = from(new PassThrough());
console.log("duplexFrom_created: " + (dFrom !== null));

// @api: stream.finished
// @expect: stream_finished: true
const rFin = new Readable({ read: () => {} });
finished(rFin, () => {
    console.log("stream_finished: true");
});
rFin.push(null);
rFin.read();

// @api: stream.pipeline
// @expect: stream_pipeline: true
const pipeSrc = new PassThrough();
const pipeDest = new PassThrough();
pipeline(pipeSrc, pipeDest, (err: Error | null) => {
    console.log("stream_pipeline: true");
});
pipeSrc.end();

// @api: stream.compose
// @expect: compose_created: true
const comp1 = new PassThrough();
const comp2 = new PassThrough();
const composed = compose(comp1, comp2);
console.log("compose_created: " + (composed !== null));

// @api: stream.isReadable
// @expect: isReadable_true: true
// @expect: isReadable_false: false
console.log("isReadable_true: " + (isReadable(new Readable({ read: () => {} })) ? "true" : "false"));
console.log("isReadable_false: " + (isReadable(null) ? "true" : "false"));

// @api: stream.isWritable
// @expect: isWritable_true: true
// @expect: isWritable_false: false
console.log("isWritable_true: " + (isWritable(new Writable({ write: (c: unknown, e: string, cb: Function) => cb() })) ? "true" : "false"));
console.log("isWritable_false: " + (isWritable(null) ? "true" : "false"));

// @api: stream.isErrored
// @expect: isErrored_false: false
// @expect: isErrored_true: true
const rErr = new Readable({ read: () => {} });
rErr.on("error", () => {});
console.log("isErrored_false: " + (rErr.errored !== null ? "true" : "false"));
rErr.destroy(new Error("test error"));
console.log("isErrored_true: " + (rErr.errored !== null ? "true" : "false"));

// @api: stream.addAbortSignal
// @expect: addAbortSignal_destroyed: true
const rSig = new Readable({ read: () => {} });
rSig.on("error", () => {});
addAbortSignal({ aborted: true }, rSig);
console.log("addAbortSignal_destroyed: " + (rSig.destroyed ? "true" : "false"));

// @api: stream.consumers.buffer
// @api: stream.consumers.text
// @api: stream.consumers.json
// @api: stream.consumers.arrayBuffer
// @api: stream.consumers.blob
// @expect: stream_consumer_buffer: abcd
// @expect: stream_consumer_text: hello
// @expect: stream_consumer_json: true
// @expect: stream_consumer_arraybuffer: 3
// @expect: stream_consumer_blob: 4 data
const runStreamConsumers = async () => {
    const consumedBuffer = await consumeBuffer(Readable.from(["ab", "cd"]));
    console.log("stream_consumer_buffer: " + consumedBuffer.toString());
    const consumedText = await consumeText(Readable.from(["hel", "lo"]));
    console.log("stream_consumer_text: " + consumedText);
    const consumedJSON = await consumeJSON(Readable.from(["{\"ok\":true}"])) as { ok: boolean };
    console.log("stream_consumer_json: " + consumedJSON.ok);
    const consumedArrayBuffer = await consumeArrayBuffer(Readable.from(["xyz"]));
    console.log("stream_consumer_arraybuffer: " + consumedArrayBuffer.byteLength);
    const consumedBlob = await consumeBlob(Readable.from(["data"]));
    console.log("stream_consumer_blob: " + consumedBlob.size + " " + await consumedBlob.text());
};
runStreamConsumers();
