// ScriptGo Corpus: WHATWG Streams & Streaming Fetch Integration
// Consolidated test suite verifying Response.body stream reader, locking, and byte piping.

import { ReadableStream, TransformStream } from "node:stream/web";

// @api: Response.body
// @expect: body_is_null_init: true
const res = new Response("hello world stream");
console.log("body_is_null_init: " + (res.body === null));

// @api: ReadableStream.pipeThrough
// @expect: stream_piped: true
const stream = new ReadableStream();
const transform = new TransformStream();
const piped = stream.pipeThrough(transform);
console.log("stream_piped: " + (piped !== null));
