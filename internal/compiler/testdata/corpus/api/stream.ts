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
    isErrored
} from "node:stream";

// @api: stream.getDefaultHighWaterMark(objectMode)
// @expect: 65536
// @expect: 16
console.log(getDefaultHighWaterMark(false));
console.log(getDefaultHighWaterMark(true));

// @api: stream.setDefaultHighWaterMark(objectMode, value)
// @expect: 32768
// @expect: 32
setDefaultHighWaterMark(false, 32768);
setDefaultHighWaterMark(true, 32);
console.log(getDefaultHighWaterMark(false));
console.log(getDefaultHighWaterMark(true));
// Restore defaults
setDefaultHighWaterMark(false, 65536);
setDefaultHighWaterMark(true, 16);

// @api: readable.push('')
// @expect: push_result: true
// @expect: chunk: hello
const r1 = new Readable({ read: () => {} });
const pushRes = r1.push("hello");
console.log("push_result: " + pushRes);
const readChunk = r1.read();
console.log("chunk: " + readChunk);

// @api: readable.read(0)
// @expect: read0_result: null
const r2 = new Readable({ read: () => {} });
const read0 = r2.read(0);
console.log("read0_result: " + read0);

// @api: stream.Readable.isDisturbed(stream)
// @expect: disturbed_before: false
// @expect: disturbed_after: true
const rDist = new Readable({ read: () => {} });
console.log("disturbed_before: " + Readable.isDisturbed(rDist));
rDist.push("test");
rDist.read(1);
console.log("disturbed_after: " + Readable.isDisturbed(rDist));

// @api: stream.Readable.fromWeb(readableStream[, options])
// @expect: fromWeb_created: true
const rFromWeb = Readable.fromWeb(new ReadableStream());
console.log("fromWeb_created: " + (rFromWeb !== null));

// @api: stream.Readable.toWeb(streamReadable[, options])
// @expect: toWeb_created: true
const rToWeb = Readable.toWeb(new Readable({ read: () => {} }));
console.log("toWeb_created: " + (rToWeb !== null));

// @api: stream.Writable.fromWeb(writableStream[, options])
// @expect: writableFromWeb_created: true
const wFromWeb = Writable.fromWeb(new WritableStream());
console.log("writableFromWeb_created: " + (wFromWeb !== null));

// @api: stream.Writable.toWeb(streamWritable)
// @expect: writableToWeb_created: true
const wToWeb = Writable.toWeb(new Writable({ write: (c: unknown, e: string, cb: Function) => cb() }));
console.log("writableToWeb_created: " + (wToWeb !== null));

// @api: stream.Duplex.from(src)
// @expect: duplexFrom_created: true
const dFrom = Duplex.from(new PassThrough());
console.log("duplexFrom_created: " + (dFrom !== null));

// @api: stream.Duplex.fromWeb(pair[, options])
// @expect: duplexFromWeb_created: true
const dFromWeb = Duplex.fromWeb({ readable: new ReadableStream(), writable: new WritableStream() });
console.log("duplexFromWeb_created: " + (dFromWeb !== null));

// @api: stream.Duplex.toWeb(streamDuplex)
// @expect: duplexToWeb_created: true
const dToWeb = Duplex.toWeb(new Duplex());
console.log("duplexToWeb_created: " + (dToWeb !== null));

// @api: stream.Transform
// @expect: transform_out: TRANSFORMED: abc
const t1 = new Transform({
    transform: (chunk: string, enc: string, cb: (err?: Error | null, data?: string) => void) => {
        cb(null, "TRANSFORMED: " + chunk);
    }
});
t1.on("data", (chunk: string) => {
    console.log("transform_out: " + chunk);
});
t1.write("abc");
t1.end();

// @api: stream.PassThrough
// @expect: pass_out: direct_data
const pt = new PassThrough();
pt.on("data", (chunk: string) => {
    console.log("pass_out: " + chunk);
});
pt.write("direct_data");
pt.end();

// @api: stream.isReadable(stream)
// @expect: isReadable_true: true
// @expect: isReadable_false: false
console.log("isReadable_true: " + (isReadable(new Readable({ read: () => {} })) ? "true" : "false"));
console.log("isReadable_false: " + (isReadable(null) ? "true" : "false"));

// @api: stream.isWritable(stream)
// @expect: isWritable_true: true
// @expect: isWritable_false: false
console.log("isWritable_true: " + (isWritable(new Writable({ write: (c: unknown, e: string, cb: Function) => cb() })) ? "true" : "false"));
console.log("isWritable_false: " + (isWritable(null) ? "true" : "false"));

// @api: stream.isErrored(stream)
// @expect: isErrored_false: false
// @expect: isErrored_true: true
const rErr = new Readable({ read: () => {} });
rErr.on("error", () => {});
console.log("isErrored_false: " + (rErr.errored !== null ? "true" : "false"));
rErr.destroy(new Error("test error"));
console.log("isErrored_true: " + (rErr.errored !== null ? "true" : "false"));

// @api: stream.addAbortSignal(signal, stream)
// @expect: addAbortSignal_destroyed: true
const rSig = new Readable({ read: () => {} });
rSig.on("error", () => {});
addAbortSignal({ aborted: true }, rSig);
console.log("addAbortSignal_destroyed: " + (rSig.destroyed ? "true" : "false"));

// @api: stream.promises.finished(stream[, options])
// @expect: finished_promise: true
const rFinP = new Readable({ read: () => {} });
const pFin = promises.finished(rFinP);
console.log("finished_promise: " + (pFin !== null));
rFinP.push(null);

// @api: stream.promises.pipeline(source, ...transforms, destination)
// @expect: pipeline_multi_promise: true
const pSrc2 = new Readable({ read: () => {} });
pSrc2._customRead = (size: number) => {
    pSrc2.push("data2");
    pSrc2.push(null);
};
const pPass2 = new PassThrough();
const pDest2 = new Writable({
    write: (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
        cb(null);
    }
});
const pPipelineProm = promises.pipeline(pSrc2, pPass2, pDest2);
console.log("pipeline_multi_promise: " + (pPipelineProm !== null));

// @api: stream.promises.pipeline(streams[, options])
// @expect: pipeline_array_opt: true
const pSrcArr2 = new Readable({ read: () => {} });
pSrcArr2._customRead = (size: number) => {
    pSrcArr2.push("arr2");
    pSrcArr2.push(null);
};
const pDestArr2 = new Writable({
    write: (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
        cb(null);
    }
});
promises.pipeline([pSrcArr2, pDestArr2], { end: true });
console.log("pipeline_array_opt: true");

// @api: stream.compose(...streams)
// @expect: compose_created: true
const comp1 = new PassThrough();
const comp2 = new PassThrough();
const composed = compose(comp1, comp2);
console.log("compose_created: " + (composed !== null));
