// ScriptGo Corpus: Request Standard Builtin APIs
// Consolidated test suite with inline assertions.

import "node:http";
import { Blob } from "node:buffer";

// @api: request.url
// @api: request.method
// @api: request.headers
// @api: request.body
// @api: request.constructor
// @api: request.text
// @api: request.json
// @api: request.arrayBuffer
// @api: request.blob
// @api: request.bytes
// @api: request.formData
// @api: request.clone
// @expect: https://example.com/api
// @expect: POST
// @expect: application/json
// @expect: true
// @expect: cloned: POST, https://example.com/api, hello
// @expect: text: hello
// @expect: json: world
// @expect: arrayBuffer: 5
// @expect: blob: 5
// @expect: bytes: 5
// @expect: fd: admin
// @expect: search: 42
async function testRequest() {
    const h = new Headers();
    h.set("Content-Type", "application/json");
    const req = new Request("https://example.com/api", {
        method: "POST",
        headers: h,
        body: '{"hello":"world"}'
    });

    console.log(req.url);
    console.log(req.method);
    console.log(req.headers.get("Content-Type"));
    console.log(req.body !== null);

    const reqSimple = new Request("https://example.com/api", { method: "POST", body: "hello" });
    const cloned = reqSimple.clone();
    console.log("cloned: " + cloned.method + ", " + cloned.url + ", " + (await cloned.text()));
    console.log("text: " + (await reqSimple.text()));

    const jsonReq = new Request("https://example.com/data", { method: "POST", body: '{"key":"world"}' });
    const data = await jsonReq.json<{ key: string }>();
    console.log("json: " + data.key);

    const abReq = new Request("https://example.com/ab", { method: "POST", body: "hello" });
    const ab = await abReq.arrayBuffer();
    console.log("arrayBuffer: " + ab.byteLength);

    const blobReq = new Request("https://example.com/blob", { method: "POST", body: "hello" });
    const bl = await blobReq.blob();
    console.log("blob: " + bl.size);

    const bytesReq = new Request("https://example.com/bytes", { method: "POST", body: "hello" });
    const bArr = await bytesReq.bytes();
    console.log("bytes: " + bArr.length);

    const fd = new FormData();
    fd.append("username", "admin");
    const fdReq = new Request("https://example.com/login", { method: "POST", body: fd });
    const parsedFd = await fdReq.formData();
    console.log("fd: " + parsedFd.get("username"));

    const sp = new URLSearchParams("id=42");
    const spReq = new Request("https://example.com/query", { method: "POST", body: sp });
    const parsedSpFd = await spReq.formData();
    console.log("search: " + parsedSpFd.get("id"));
}
testRequest();
