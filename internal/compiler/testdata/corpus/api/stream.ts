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
    compose,
    addAbortSignal,
    getDefaultHighWaterMark,
    setDefaultHighWaterMark,
    isReadable,
    isWritable,
    isErrored
} from "node:stream";

// @api: stream.getDefaultHighWaterMark(objectMode)
// @expect: 16384
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
setDefaultHighWaterMark(false, 16384);
setDefaultHighWaterMark(true, 16);

// @api: readable.push('')
// @expect: push_result: true
// @expect: chunk: hello
const r1 = new Readable();
const pushRes = r1.push("hello");
console.log("push_result: " + pushRes);
const readChunk = r1.read();
console.log("chunk: " + readChunk);

// @api: readable.read(0)
// @expect: read0_result: null
const r2 = new Readable();
const read0 = r2.read(0);
console.log("read0_result: " + read0);

// @api: stream.Readable.from(iterable[, options])
// @expect: from_data: 1
// @expect: from_data: 2
// @expect: from_data: 3
// @expect: from_ended
const rFrom = Readable.from(["1", "2", "3"]);
rFrom.on("data", (chunk: string) => {
    console.log("from_data: " + chunk);
});
rFrom.on("end", () => {
    console.log("from_ended");
});

// @api: stream.Readable.isDisturbed(stream)
// @expect: disturbed_before: false
// @expect: disturbed_after: true
const rDist = new Readable();
console.log("disturbed_before: " + Readable.isDisturbed(rDist));
rDist.read(0);
console.log("disturbed_after: " + Readable.isDisturbed(rDist));

// @api: stream.Readable.fromWeb(readableStream[, options])
// @expect: fromWeb_created: true
const fakeWebReadable = {
    getReader: () => ({
        read: () => Promise.resolve({ value: "web-chunk", done: true })
    })
};
const rFromWeb = Readable.fromWeb(fakeWebReadable);
console.log("fromWeb_created: " + (rFromWeb !== null));

// @api: stream.Readable.toWeb(streamReadable[, options])
// @expect: toWeb_created: true
const rToWeb = Readable.toWeb(new Readable());
console.log("toWeb_created: " + (rToWeb !== null));

// @api: stream.Writable.fromWeb(writableStream[, options])
// @expect: writableFromWeb_created: true
const fakeWebWritable = {
    getWriter: () => ({
        write: (chunk: unknown) => Promise.resolve(null),
        close: () => Promise.resolve(null)
    })
};
const wFromWeb = Writable.fromWeb(fakeWebWritable);
console.log("writableFromWeb_created: " + (wFromWeb !== null));

// @api: stream.Writable.toWeb(streamWritable)
// @expect: writableToWeb_created: true
const wToWeb = Writable.toWeb(new Writable());
console.log("writableToWeb_created: " + (wToWeb !== null));

// @api: stream.Duplex.from(src)
// @expect: duplexFrom_created: true
const dFrom = Duplex.from(new PassThrough());
console.log("duplexFrom_created: " + (dFrom !== null));

// @api: stream.Duplex.fromWeb(pair[, options])
// @expect: duplexFromWeb_created: true
const dFromWeb = Duplex.fromWeb({ readable: fakeWebReadable, writable: fakeWebWritable });
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
console.log("isReadable_true: " + isReadable(new Readable()));
console.log("isReadable_false: " + isReadable(null));

// @api: stream.isWritable(stream)
// @expect: isWritable_true: true
// @expect: isWritable_false: false
console.log("isWritable_true: " + isWritable(new Writable()));
console.log("isWritable_false: " + isWritable(null));

// @api: stream.isErrored(stream)
// @expect: isErrored_false: false
// @expect: isErrored_true: true
const rErr = new Readable();
console.log("isErrored_false: " + isErrored(rErr));
rErr.destroy(new Error("test error"));
console.log("isErrored_true: " + isErrored(rErr));

// @api: stream.addAbortSignal(signal, stream)
// @expect: addAbortSignal_destroyed: true
const rSig = new Readable();
addAbortSignal({ aborted: true }, rSig);
console.log("addAbortSignal_destroyed: " + isErrored(rSig));

// @api: stream.finished(stream, callback)
// @expect: finished_callback: true
const rFin = new Readable();
finished(rFin, (err?: unknown) => {
    console.log("finished_callback: true");
});
rFin.push(null);

// @api: stream.finished(stream[, options])
// @expect: finished_promise: true
const rFinP = new Readable();
const pFin = finished(rFinP);
console.log("finished_promise: " + (pFin !== null));
rFinP.push(null);

// @api: stream.pipeline(source, ...transforms, destination, callback)
// @expect: pipeline_multi_cb: true
const pSource = new Readable();
pSource._customRead = (size: number) => {
    pSource.push("data");
    pSource.push(null);
};
const pPass = new PassThrough();
const pDest = new Writable();
pDest._customWrite = (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
    cb(null);
};
pipeline(pSource, pPass, pDest, (err?: unknown) => {
    console.log("pipeline_multi_cb: true");
});

// @api: stream.pipeline(source, ...transforms, destination)
// @expect: pipeline_multi_promise: true
const pSrc2 = new Readable();
pSrc2._customRead = (size: number) => {
    pSrc2.push("data2");
    pSrc2.push(null);
};
const pPass2 = new PassThrough();
const pDest2 = new Writable();
pDest2._customWrite = (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
    cb(null);
};
const pPipelineProm = pipeline(pSrc2, pPass2, pDest2);
console.log("pipeline_multi_promise: " + (pPipelineProm !== null));

// @api: stream.pipeline(streams, callback)
// @expect: pipeline_array_cb: true
const pSrcArr = new Readable();
pSrcArr._customRead = (size: number) => {
    pSrcArr.push("arr");
    pSrcArr.push(null);
};
const pDestArr = new Writable();
pDestArr._customWrite = (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
    cb(null);
};
pipeline([pSrcArr, pDestArr], (err?: unknown) => {
    console.log("pipeline_array_cb: true");
});

// @api: stream.pipeline(streams[, options])
// @expect: pipeline_array_opt: true
const pSrcArr2 = new Readable();
pSrcArr2._customRead = (size: number) => {
    pSrcArr2.push("arr2");
    pSrcArr2.push(null);
};
const pDestArr2 = new Writable();
pDestArr2._customWrite = (chunk: string, enc: string, cb: (error?: Error | null) => void) => {
    cb(null);
};
pipeline([pSrcArr2, pDestArr2], { end: true });
console.log("pipeline_array_opt: true");

// @api: stream.compose(...streams)
// @expect: compose_created: true
const comp1 = new PassThrough();
const comp2 = new PassThrough();
const composed = compose(comp1, comp2);
console.log("compose_created: " + (composed !== null));
