import {
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
    DecompressionStream
} from "node:stream/web";
import {
    arrayBuffer,
    blob,
    buffer,
    json,
    text
} from "node:stream/consumers";

// @api: webstreams.webstreams.ReadableStream
// @api: webstreams.ReadableStream
// @api: new webstreams.ReadableStream
// @api: ReadableStream.locked
// @expect: ws_rs_inst: true false
const rs = new ReadableStream();
console.log("ws_rs_inst: " + (rs instanceof ReadableStream) + " " + rs.locked);

// @api: ReadableStream.getReader
// @api: webstreams.webstreams.ReadableStreamDefaultReader
// @api: webstreams.ReadableStreamDefaultReader
// @api: new webstreams.ReadableStreamDefaultReader
// @api: ReadableStreamDefaultReader.releaseLock
// @api: ReadableStreamDefaultReader.closed
// @expect: ws_reader: true true
const reader = rs.getReader();
console.log("ws_reader: " + (reader instanceof ReadableStreamDefaultReader) + " " + (typeof reader.closed === "object"));
reader.releaseLock();

// @api: ReadableStream.values
// @expect: ws_values: true
const vals = new ReadableStream().values();
console.log("ws_values: " + (typeof vals === "object"));

// @api: ReadableStream.pipeThrough
// @api: webstreams.webstreams.TransformStream
// @api: webstreams.TransformStream
// @api: new webstreams.TransformStream
// @api: TransformStream.readable
// @api: TransformStream.writable
// @expect: ws_ts: true true
const ts = new TransformStream();
const piped = new ReadableStream({ start: (controller) => controller.close() }).pipeThrough(ts);
console.log("ws_ts: " + (piped instanceof ReadableStream) + " " + (ts.writable instanceof WritableStream));

// @api: ReadableStream.tee
// @expect: ws_tee: true true
const [s1, s2] = rs.tee();
console.log("ws_tee: " + (s1 instanceof ReadableStream) + " " + (s2 instanceof ReadableStream));

// @api: webstreams.webstreams.WritableStream
// @api: webstreams.WritableStream
// @api: new webstreams.WritableStream
// @api: WritableStream.locked
// @expect: ws_ws_inst: true false
const ws = new WritableStream();
console.log("ws_ws_inst: " + (ws instanceof WritableStream) + " " + ws.locked);

// @api: WritableStream.getWriter
// @api: webstreams.webstreams.WritableStreamDefaultWriter
// @api: webstreams.WritableStreamDefaultWriter
// @api: new webstreams.WritableStreamDefaultWriter
// @api: WritableStreamDefaultWriter.closed
// @api: WritableStreamDefaultWriter.ready
// @api: WritableStreamDefaultWriter.desiredSize
// @api: WritableStreamDefaultWriter.releaseLock
// @expect: ws_writer: true 1
const writer = ws.getWriter();
console.log("ws_writer: " + (writer instanceof WritableStreamDefaultWriter) + " " + writer.desiredSize);
writer.releaseLock();

// @api: webstreams.webstreams.ReadableStreamBYOBRequest
// @api: webstreams.ReadableStreamBYOBRequest
// @api: new webstreams.ReadableStreamBYOBRequest
// @api: ReadableStreamBYOBRequest.view
// @api: ReadableStreamBYOBRequest.respond
// @api: ReadableStreamBYOBRequest.respondWithNewView
// @expect: ws_byobReq: true
let byobReqThrows = false;
try {
    new ReadableStreamBYOBRequest();
} catch (e) {
    byobReqThrows = (e instanceof TypeError);
}
console.log("ws_byobReq: " + byobReqThrows);

// @api: webstreams.webstreams.ReadableByteStreamController
// @api: webstreams.ReadableByteStreamController
// @api: new webstreams.ReadableByteStreamController
// @api: ReadableByteStreamController.byobRequest
// @api: ReadableByteStreamController.desiredSize
// @api: ReadableByteStreamController.enqueue
// @api: ReadableByteStreamController.close
// @api: ReadableByteStreamController.error
// @expect: ws_byteCtrl: true true
let byteCtrl!: ReadableByteStreamController;
const rsb = new ReadableStream({
    type: "bytes",
    start(c) {
        byteCtrl = c;
    }
});
byteCtrl.enqueue(new Uint8Array(4));
byteCtrl.close();
console.log("ws_byteCtrl: " + (byteCtrl instanceof ReadableByteStreamController) + " " + (byteCtrl.byobRequest === null));

// @api: webstreams.webstreams.ReadableStreamBYOBReader
// @api: webstreams.ReadableStreamBYOBReader
// @api: new webstreams.ReadableStreamBYOBReader
// @api: ReadableStreamBYOBReader.closed
// @api: ReadableStreamBYOBReader.releaseLock
// @expect: ws_byobReader: true true
const byobReader = new ReadableStreamBYOBReader(rsb);
byobReader.releaseLock();
console.log("ws_byobReader: " + (byobReader instanceof ReadableStreamBYOBReader) + " " + (typeof byobReader.closed === "object"));

// @api: webstreams.webstreams.ReadableStreamDefaultController
// @api: webstreams.ReadableStreamDefaultController
// @api: new webstreams.ReadableStreamDefaultController
// @api: ReadableStreamDefaultController.desiredSize
// @api: ReadableStreamDefaultController.enqueue
// @api: ReadableStreamDefaultController.close
// @api: ReadableStreamDefaultController.error
// @expect: ws_defaultCtrl: true 0
let defaultCtrl!: ReadableStreamDefaultController;
const rsDefault = new ReadableStream({
    start(c) {
        defaultCtrl = c;
    }
});
defaultCtrl.enqueue(1);
defaultCtrl.close();
console.log("ws_defaultCtrl: " + (defaultCtrl instanceof ReadableStreamDefaultController) + " " + defaultCtrl.desiredSize);

// @api: webstreams.webstreams.WritableStreamDefaultController
// @api: webstreams.WritableStreamDefaultController
// @api: new webstreams.WritableStreamDefaultController
// @api: WritableStreamDefaultController.signal
// @api: WritableStreamDefaultController.error
// @expect: ws_wsCtrl: true true
let wsCtrl!: WritableStreamDefaultController;
const wsDefault = new WritableStream({
    start(c) {
        wsCtrl = c;
    }
});
console.log("ws_wsCtrl: " + (wsCtrl instanceof WritableStreamDefaultController) + " " + (typeof wsCtrl.signal === "object"));

// @api: webstreams.webstreams.TransformStreamDefaultController
// @api: webstreams.TransformStreamDefaultController
// @api: new webstreams.TransformStreamDefaultController
// @api: TransformStreamDefaultController.desiredSize
// @api: TransformStreamDefaultController.enqueue
// @api: TransformStreamDefaultController.error
// @api: TransformStreamDefaultController.terminate
// @expect: ws_tsCtrl: true true
let tsCtrl!: TransformStreamDefaultController;
const tsDefault = new TransformStream({
    start(c) {
        tsCtrl = c;
    }
});
tsCtrl.enqueue(1);
console.log("ws_tsCtrl: " + (tsCtrl instanceof TransformStreamDefaultController) + " " + (typeof tsCtrl.desiredSize === "number"));
tsCtrl.terminate();

// @api: webstreams.webstreams.ByteLengthQueuingStrategy
// @api: webstreams.ByteLengthQueuingStrategy
// @api: new webstreams.ByteLengthQueuingStrategy
// @api: ByteLengthQueuingStrategy.highWaterMark
// @api: ByteLengthQueuingStrategy.size
// @expect: ws_blStrategy: 1024 16
const blStrategy = new ByteLengthQueuingStrategy({ highWaterMark: 1024 });
console.log("ws_blStrategy: " + blStrategy.highWaterMark + " " + blStrategy.size({ byteLength: 16 }));

// @api: webstreams.webstreams.CountQueuingStrategy
// @api: webstreams.CountQueuingStrategy
// @api: new webstreams.CountQueuingStrategy
// @api: CountQueuingStrategy.highWaterMark
// @api: CountQueuingStrategy.size
// @expect: ws_cntStrategy: 10 1
const cntStrategy = new CountQueuingStrategy({ highWaterMark: 10 });
console.log("ws_cntStrategy: " + cntStrategy.highWaterMark + " " + cntStrategy.size({}));

// @api: webstreams.webstreams.TextEncoderStream
// @api: webstreams.TextEncoderStream
// @api: new webstreams.TextEncoderStream
// @api: TextEncoderStream.encoding
// @api: TextEncoderStream.readable
// @api: TextEncoderStream.writable
// @expect: ws_encStream: utf-8 true true
const encStream = new TextEncoderStream();
console.log("ws_encStream: " + encStream.encoding + " " + (encStream.readable instanceof ReadableStream) + " " + (encStream.writable instanceof WritableStream));

// @api: webstreams.webstreams.TextDecoderStream
// @api: webstreams.TextDecoderStream
// @api: new webstreams.TextDecoderStream
// @api: TextDecoderStream.encoding
// @api: TextDecoderStream.fatal
// @api: TextDecoderStream.ignoreBOM
// @api: TextDecoderStream.readable
// @api: TextDecoderStream.writable
// @expect: ws_decStream: utf-8 false false
const decStream = new TextDecoderStream();
console.log("ws_decStream: " + decStream.encoding + " " + decStream.fatal + " " + decStream.ignoreBOM);

// @api: webstreams.webstreams.CompressionStream
// @api: webstreams.CompressionStream
// @api: new webstreams.CompressionStream
// @api: CompressionStream.readable
// @api: CompressionStream.writable
// @expect: ws_compStream: true true
const compStream = new CompressionStream("gzip");
console.log("ws_compStream: " + (compStream.readable instanceof ReadableStream) + " " + (compStream.writable instanceof WritableStream));

// @api: webstreams.webstreams.DecompressionStream
// @api: webstreams.DecompressionStream
// @api: new webstreams.DecompressionStream
// @api: DecompressionStream.readable
// @api: DecompressionStream.writable
// @expect: ws_decompStream: true true
const decompStream = new DecompressionStream("gzip");
console.log("ws_decompStream: " + (decompStream.readable instanceof ReadableStream) + " " + (decompStream.writable instanceof WritableStream));

// @api: webstreams.from
// @expect: ws_from: true
const fromStream = ReadableStream.from([1, 2, 3]);
console.log("ws_from: " + (fromStream instanceof ReadableStream));

// @api: ReadableStream.cancel
// @api: ReadableStream.pipeTo
// @api: ReadableStreamDefaultReader.cancel
// @api: ReadableStreamDefaultReader.read
// @api: ReadableStreamBYOBReader.cancel
// @api: ReadableStreamBYOBReader.read
// @api: WritableStream.abort
// @api: WritableStream.close
// @api: WritableStreamDefaultWriter.abort
// @api: WritableStreamDefaultWriter.close
// @api: WritableStreamDefaultWriter.write
// @api: webstreams.arrayBuffer
// @api: webstreams.blob
// @api: webstreams.buffer
// @api: webstreams.json
// @api: webstreams.text
// @expect: ws_async: true
const runAsyncWebstreams = async () => {
    const cancelStream = new ReadableStream();
    await cancelStream.cancel();
    const pipeSource = new ReadableStream({ start: (controller) => controller.close() });
    const pipeTarget = new WritableStream();
    await pipeSource.pipeTo(pipeTarget);
    const readStream = new ReadableStream({ start: (controller) => controller.close() });
    const r = readStream.getReader();
    await r.read();
    await r.cancel();
    const br = new ReadableStreamBYOBReader(new ReadableStream({ type: "bytes", start: (controller) => controller.close() }));
    await br.cancel();
    await new WritableStream().abort();
    await new WritableStream().close();
    const writeStream = new WritableStream();
    const w = writeStream.getWriter();
    await w.write(1);
    await w.abort();
    const closeWriter = new WritableStream().getWriter();
    await closeWriter.close();
    await arrayBuffer(new ReadableStream({ start: (controller) => controller.close() }));
    await blob(new ReadableStream({ start: (controller) => controller.close() }));
    await buffer(new ReadableStream({ start: (controller) => controller.close() }));
    await json(ReadableStream.from(["1"]));
    await text(new ReadableStream({ start: (controller) => controller.close() }));
    console.log("ws_async: true");
};
runAsyncWebstreams();
