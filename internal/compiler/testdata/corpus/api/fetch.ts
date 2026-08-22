// ScriptGo Corpus: Fetch Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { Headers } from "node:http";
import { Response } from "node:http";

// @api: fetch.Headers
// @expect: application/json
// @expect: true
const h_fetch_Headers_0 = new Headers();
h_fetch_Headers_0.set("Content-Type", "application/json");
console.log(h_fetch_Headers_0.get("Content-Type"));
console.log(h_fetch_Headers_0.has("Content-Type"));

// @api: fetch.Response
// @expect: 200
// @expect: true
const r_fetch_Response_1 = new Response("hello response", { status: 200 });
console.log(r_fetch_Response_1.status);
console.log(r_fetch_Response_1.ok);
