// ScriptGo Corpus: Scenario: Networking & Web APIs
// Consolidated test suite with inline assertions.

import { METHODS } from "node:http";
import { URL } from "node:url";

// --- Context Case: scenarios_fetch_headers ---
// @expect: application/json
// @expect: text/plain, text/html
// @expect: true
// @expect: false
const h_fetch_headers_0 = new Headers();
h_fetch_headers_0.set("Content-Type", "application/json");
h_fetch_headers_0.append("Accept", "text/plain");
h_fetch_headers_0.append("accept", "text/html");

console.log(h_fetch_headers_0.get("content-type"));
console.log(h_fetch_headers_0.get("Accept"));
console.log(h_fetch_headers_0.has("CONTENT-TYPE"));
console.log(h_fetch_headers_0.has("Authorization"));

h_fetch_headers_0.delete("content-type");

// --- Context Case: scenarios_fetch_node_prefix ---
// @expect: https://api.example.com/items
// @expect: POST
// @expect: secret123
// @expect: true
// @expect: 201
// @expect: true
// @expect: true
const h_fetch_node_prefix_1 = new Headers();
h_fetch_node_prefix_1.set("x-api-key", "secret123");

const req_fetch_node_prefix_1 = new Request("https://api.example.com/items", {
    method: "POST",
    headers: h_fetch_node_prefix_1,
    body: "item=1"
});

console.log(req_fetch_node_prefix_1.url);
console.log(req_fetch_node_prefix_1.method);
console.log(req_fetch_node_prefix_1.headers.get("x-api-key"));
console.log(req_fetch_node_prefix_1.body !== null);

const res_fetch_node_prefix_1 = new Response("done", {
    status: 201,
    statusText: "Created"
});

console.log(res_fetch_node_prefix_1.status);
console.log(res_fetch_node_prefix_1.statusText.length >= 0);
console.log(res_fetch_node_prefix_1.ok);

// --- Context Case: scenarios_fetch_response_json ---
// @expect: 200
// @expect: true
// @expect: true
// @expect: application/json
// @expect: true
// @expect: 301
// @expect: https://example.com/login
async function testResponse_fetch_response_json_2(): Promise<void> {
    const data_fetch_response_json_2 = { message: "hello", count: 42 };
    const resp_fetch_response_json_2 = Response.json(JSON.stringify(data_fetch_response_json_2));

    console.log(resp_fetch_response_json_2.status);
    console.log(resp_fetch_response_json_2.statusText.length >= 0);
    console.log(resp_fetch_response_json_2.ok);
    console.log(resp_fetch_response_json_2.headers.get("content-type"));

    const text_fetch_response_json_2 = await resp_fetch_response_json_2.text();
    console.log(text_fetch_response_json_2.length > 0);

    const redirect_fetch_response_json_2 = Response.redirect("https://example.com/login", 301);
    console.log(redirect_fetch_response_json_2.status);
    console.log(redirect_fetch_response_json_2.headers.get("location"));
}

await testResponse_fetch_response_json_2();

// --- Context Case: scenarios_http_constants ---
// @expect: true
// @expect: GET
console.log(METHODS.length > 10);
console.log(METHODS[6]);

// --- Context Case: scenarios_url_relative_resolution ---
// @expect: https://example.com/api/v1/posts
// @expect: https://example.com/root/path
// @expect: https://example.com/items?page=2
// @expect: 2
const base_url_relative_resolution_4: URL = new URL("https://example.com/api/v1/users");
const rel1_url_relative_resolution_4: URL = new URL("posts", base_url_relative_resolution_4.href);
console.log(rel1_url_relative_resolution_4.href);

const rel2_url_relative_resolution_4: URL = new URL("/root/path", base_url_relative_resolution_4.href);
console.log(rel2_url_relative_resolution_4.href);

const rel3_url_relative_resolution_4: URL = new URL("?page=2", "https://example.com/items");
console.log(rel3_url_relative_resolution_4.href);
console.log(rel3_url_relative_resolution_4.searchParams.get("page"));
