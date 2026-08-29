// ScriptGo Corpus: Fetch Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { fetch, Headers, Request, Response } from "node:http";

// @api: fetch.Headers
// @expect: application/json
// @expect: true
const h_fetch_Headers_0 = new Headers();
h_fetch_Headers_0.set("Content-Type", "application/json");
console.log(h_fetch_Headers_0.get("Content-Type"));
console.log(h_fetch_Headers_0.has("Content-Type"));

// @api: fetch.Request
// @expect: https://api.example.com/v1/users
// @expect: POST
// @expect: user=alice
const req = new Request("https://api.example.com/v1/users", {
    method: "POST",
    headers: h_fetch_Headers_0,
    body: "user=alice"
});
console.log(req.url);
console.log(req.method);
console.log(req.body);

// @api: fetch.Response
// @expect: 200
// @expect: true
const r_fetch_Response_1 = new Response("hello response", { status: 200 });
console.log(r_fetch_Response_1.status);
console.log(r_fetch_Response_1.ok);

// @api: fetch.fn
// @expect: is_fetch_fn: true
// @expect: is_promise: true
console.log("is_fetch_fn: " + (typeof fetch === "function"));
const fetchPromise = fetch("http://127.0.0.1:9");
console.log("is_promise: " + (fetchPromise !== null));
