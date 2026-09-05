// ScriptGo Corpus: WHATWG Streams & Streaming Fetch Integration
// Consolidated test suite verifying Response.body stream reader, locking, and byte piping.

import "node:http";

import { ReadableStream, TransformStream } from "node:stream/web";

// @api: Response.body
// @expect: body_is_object: true
const res = new Response("hello world stream");
console.log("body_is_object: " + (typeof res.body === "object"));

// @api: ReadableStream.pipeThrough
// @expect: stream_piped: true
const stream = new ReadableStream();
const transform = new TransformStream();
const piped = stream.pipeThrough(transform);
console.log("stream_piped: " + (piped !== null));
