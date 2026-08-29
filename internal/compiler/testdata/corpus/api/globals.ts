// @api: globals.AbortController
// @api: AbortController.abort
// @api: AbortController.signal
// @expect: gl_abort: true
console.log("gl_abort: " + (typeof AbortController === "function" || typeof AbortController === "object"));

// @api: globals.Blob
// @expect: gl_blob: true
console.log("gl_blob: " + (typeof Blob === "function" || typeof Blob === "object"));

// @api: globals.Buffer
// @expect: gl_buf: true
console.log("gl_buf: " + (typeof Buffer === "function" || typeof Buffer === "object"));

// @api: globals.ByteLengthQueuingStrategy
// @expect: gl_blStrat: true
console.log("gl_blStrat: " + (typeof ByteLengthQueuingStrategy === "function" || typeof ByteLengthQueuingStrategy === "object"));

// @api: globals.BroadcastChannel
// @expect: gl_bc: true
console.log("gl_bc: " + (typeof BroadcastChannel === "function" || typeof BroadcastChannel === "object"));

// @api: globals.CompressionStream
// @expect: gl_compStr: true
console.log("gl_compStr: " + (typeof CompressionStream === "function" || typeof CompressionStream === "object"));

// @api: globals.CountQueuingStrategy
// @expect: gl_cntStrat: true
console.log("gl_cntStrat: " + (typeof CountQueuingStrategy === "function" || typeof CountQueuingStrategy === "object"));

// @api: globals.Crypto
// @api: globals.CryptoKey
// @api: globals.SubtleCrypto
// @expect: gl_crypto: true true true
console.log("gl_crypto: " + (typeof Crypto === "function" || typeof Crypto === "object") + " " + (typeof CryptoKey === "function" || typeof CryptoKey === "object") + " " + (typeof SubtleCrypto === "function" || typeof SubtleCrypto === "object"));

// @api: globals.CustomEvent
// @api: globals.Event
// @api: globals.EventSource
// @api: globals.EventTarget
// @expect: gl_events: true true true true
console.log("gl_events: " + (typeof CustomEvent === "function" || typeof CustomEvent === "object") + " " + (typeof Event === "function" || typeof Event === "object") + " " + (typeof EventSource === "function" || typeof EventSource === "object") + " " + (typeof EventTarget === "function" || typeof EventTarget === "object"));

// @api: globals.DecompressionStream
// @expect: gl_decompStr: true
console.log("gl_decompStr: " + (typeof DecompressionStream === "function" || typeof DecompressionStream === "object"));

// @api: globals.File
// @api: globals.FormData
// @api: globals.Headers
// @api: globals.Request
// @api: globals.Response
// @expect: gl_fetchTypes: true true true true true
console.log("gl_fetchTypes: " + (typeof File === "function" || typeof File === "object") + " " + (typeof FormData === "function" || typeof FormData === "object") + " " + (typeof Headers === "function" || typeof Headers === "object") + " " + (typeof Request === "function" || typeof Request === "object") + " " + (typeof Response === "function" || typeof Response === "object"));

// @api: globals.MessageChannel
// @api: globals.MessageEvent
// @api: globals.MessagePort
// @expect: gl_msg: true true true
console.log("gl_msg: " + (typeof MessageChannel === "function" || typeof MessageChannel === "object") + " " + (typeof MessageEvent === "function" || typeof MessageEvent === "object") + " " + (typeof MessagePort === "function" || typeof MessagePort === "object"));

// @api: globals.Navigator
// @expect: gl_nav: true
console.log("gl_nav: " + (typeof Navigator === "function" || typeof Navigator === "object"));

// @api: globals.PerformanceEntry
// @api: globals.PerformanceMark
// @api: globals.PerformanceMeasure
// @api: globals.PerformanceObserver
// @api: globals.PerformanceObserverEntryList
// @api: globals.PerformanceResourceTiming
// @expect: gl_perf: true true true true true true
console.log("gl_perf: " + (typeof PerformanceEntry === "function" || typeof PerformanceEntry === "object") + " " + (typeof PerformanceMark === "function" || typeof PerformanceMark === "object") + " " + (typeof PerformanceMeasure === "function" || typeof PerformanceMeasure === "object") + " " + (typeof PerformanceObserver === "function" || typeof PerformanceObserver === "object") + " " + (typeof PerformanceObserverEntryList === "function" || typeof PerformanceObserverEntryList === "object") + " " + (typeof PerformanceResourceTiming === "function" || typeof PerformanceResourceTiming === "object"));

// @api: globals.ReadableByteStreamController
// @api: globals.ReadableStream
// @api: globals.ReadableStreamBYOBReader
// @api: globals.ReadableStreamBYOBRequest
// @api: globals.ReadableStreamDefaultController
// @api: globals.ReadableStreamDefaultReader
// @expect: gl_rsTypes: true true true true true true
console.log("gl_rsTypes: " + (typeof ReadableByteStreamController === "function" || typeof ReadableByteStreamController === "object") + " " + (typeof ReadableStream === "function" || typeof ReadableStream === "object") + " " + (typeof ReadableStreamBYOBReader === "function" || typeof ReadableStreamBYOBReader === "object") + " " + (typeof ReadableStreamBYOBRequest === "function" || typeof ReadableStreamBYOBRequest === "object") + " " + (typeof ReadableStreamDefaultController === "function" || typeof ReadableStreamDefaultController === "object") + " " + (typeof ReadableStreamDefaultReader === "function" || typeof ReadableStreamDefaultReader === "object"));

// @api: globals.Storage
// @expect: gl_storage: true
console.log("gl_storage: " + (typeof Storage === "function" || typeof Storage === "object"));

// @api: globals.DOMException
// @expect: gl_domEx: true
console.log("gl_domEx: " + (typeof DOMException === "function" || typeof DOMException === "object"));

// @api: globals.TextDecoder
// @api: globals.TextDecoderStream
// @api: globals.TextEncoder
// @api: globals.TextEncoderStream
// @expect: gl_text: true true true true
console.log("gl_text: " + (typeof TextDecoder === "function" || typeof TextDecoder === "object") + " " + (typeof TextDecoderStream === "function" || typeof TextDecoderStream === "object") + " " + (typeof TextEncoder === "function" || typeof TextEncoder === "object") + " " + (typeof TextEncoderStream === "function" || typeof TextEncoderStream === "object"));

// @api: globals.TransformStream
// @api: globals.TransformStreamDefaultController
// @expect: gl_ts: true true
console.log("gl_ts: " + (typeof TransformStream === "function" || typeof TransformStream === "object") + " " + (typeof TransformStreamDefaultController === "function" || typeof TransformStreamDefaultController === "object"));

// @api: globals.URL
// @api: globals.URLSearchParams
// @expect: gl_url: true true
console.log("gl_url: " + (typeof URL === "function" || typeof URL === "object") + " " + (typeof URLSearchParams === "function" || typeof URLSearchParams === "object"));

// @api: globals.WebAssembly
// @expect: gl_wasm: true
console.log("gl_wasm: " + (typeof WebAssembly === "object"));

// @api: globals.WebSocket
// @expect: gl_ws: true
console.log("gl_ws: " + (typeof WebSocket === "function" || typeof WebSocket === "object"));

// @api: globals.WritableStream
// @api: globals.WritableStreamDefaultController
// @api: globals.WritableStreamDefaultWriter
// @expect: gl_wsTypes: true true true
console.log("gl_wsTypes: " + (typeof WritableStream === "function" || typeof WritableStream === "object") + " " + (typeof WritableStreamDefaultController === "function" || typeof WritableStreamDefaultController === "object") + " " + (typeof WritableStreamDefaultWriter === "function" || typeof WritableStreamDefaultWriter === "object"));

// @api: globals.atob
// @api: globals.btoa
// @expect: gl_base64: aGVsbG8= hello
const b = btoa("hello");
const a = atob(b);
console.log("gl_base64: " + b + " " + a);

// @api: globals.structuredClone
// @expect: gl_clone: 123
const sc = structuredClone({ a: 123 });
console.log("gl_clone: " + sc.a);

// @api: globals.queueMicrotask
// @api: globals.setImmediate
// @api: globals.clearImmediate
// @api: globals.setTimeout
// @api: globals.clearTimeout
// @api: globals.setInterval
// @api: globals.clearInterval
// @api: globals.require
// @expect: gl_timers: true
queueMicrotask(() => {});
const imm = setImmediate(() => {});
clearImmediate(imm);
const tm = setTimeout(() => {}, 1000);
clearTimeout(tm);
const iv = setInterval(() => {}, 1000);
clearInterval(iv);
console.log("gl_timers: true");
