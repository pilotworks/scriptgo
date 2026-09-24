// ScriptGo Corpus: Response Standard Builtin APIs
// Consolidated test suite with inline assertions.

import "node:http";
import { Blob } from "node:buffer";

// @api: response.status
// @api: response.statusText
// @api: response.ok
// @api: response.headers
// @api: response.url
// @api: response.text
// @api: response.constructor
// @api: response.json
// @api: response.arrayBuffer
// @api: response.blob
// @api: response.bytes
// @api: response.formData
// @api: response.clone
// @api: response.error
// @api: response.redirect
// @expect: 200
// @expect: OK
// @expect: true
// @expect: hello body
// @expect: cloned: 200, hello body
// @expect: json: 123
// @expect: arrayBuffer: 5
// @expect: blob: 5
// @expect: bytes: 5
// @expect: fd: alice
// @expect: err_status: 0
// @expect: redir_status: 302
// @expect: redir_loc: https://example.com/login
// @expect: static_json: 42
async function testResponse() {
    const res = new Response("hello body", { status: 200, statusText: "OK" });
    console.log(res.status);
    console.log(res.statusText);
    console.log(res.ok);
    const cloned = res.clone();
    console.log(await res.text());
    console.log("cloned: " + cloned.status + ", " + (await cloned.text()));

    const jsonRes = new Response('{"count":123}');
    const j = await jsonRes.json<{ count: number }>();
    console.log("json: " + j.count);

    const abRes = new Response("hello");
    const ab = await abRes.arrayBuffer();
    console.log("arrayBuffer: " + ab.byteLength);

    const blobRes = new Response("hello", { headers: { "content-type": "text/plain" } });
    const bl = await blobRes.blob();
    console.log("blob: " + bl.size);

    const bytesRes = new Response("hello");
    const by = await bytesRes.bytes();
    console.log("bytes: " + by.length);

    const fd = new FormData();
    fd.append("name", "alice");
    const fdRes = new Response(fd);
    const parsedFd = await fdRes.formData();
    console.log("fd: " + parsedFd.get("name"));

    const errRes = Response.error();
    console.log("err_status: " + errRes.status);

    const redirRes = Response.redirect("https://example.com/login", 302);
    console.log("redir_status: " + redirRes.status);
    console.log("redir_loc: " + redirRes.headers.get("location"));

    const staticJsonRes = Response.json({ score: 42 });
    const sj = await staticJsonRes.json<{ score: number }>();
    console.log("static_json: " + sj.score);
}
testResponse();
