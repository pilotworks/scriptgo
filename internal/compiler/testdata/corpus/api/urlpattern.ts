// ScriptGo Corpus: WHATWG URLPattern Standard API
// Consolidated test suite with inline assertions.

import { URLPattern } from "node:url";

// --- 1. Basic Pathname Pattern with Named Parameter ---
// @api: urlpattern.constructor
// @api: urlpattern.pathname
// @api: urlpattern.test
// @api: urlpattern.exec
// @native.expected: /books/:id
// @native.expected: true
// @native.expected: false
// @native.expected: 123
const p1 = new URLPattern({ pathname: "/books/:id" });
console.log(p1.pathname);
console.log(p1.test("https://example.com/books/123"));
console.log(p1.test("https://example.com/authors/123"));

const res1 = p1.exec("https://example.com/books/123");
if (res1 !== null) {
    console.log(res1.pathname.groups["id"]);
}

// --- 2. Multiple Named Parameters in Pathname ---
// @native.expected: v2
// @native.expected: 42
const p2 = new URLPattern({ pathname: "/api/:version/users/:userId" });
const res2 = p2.exec("https://api.example.com/api/v2/users/42");
if (res2 !== null) {
    console.log(res2.pathname.groups["version"]);
    console.log(res2.pathname.groups["userId"]);
}

// --- 3. Wildcard Pattern ---
// @native.expected: css/theme.css
const p3 = new URLPattern({ pathname: "/static/*" });
const res3 = p3.exec("/static/css/theme.css");
if (res3 !== null) {
    console.log(res3.pathname.groups["0"]);
}

// --- 4. Relative Pattern with baseURL ---
// @native.expected: https
// @native.expected: example.com
// @native.expected: 99
const p4 = new URLPattern("/items/:itemId", "https://example.com");
console.log(p4.protocol);
console.log(p4.hostname);
const res4 = p4.exec("https://example.com/items/99");
if (res4 !== null) {
    console.log(res4.pathname.groups["itemId"]);
}

// --- 5. Search Query Parameter Matching ---
// @native.expected: scriptgo
const p5 = new URLPattern({ pathname: "/search", search: "q=:query" });
const res5 = p5.exec("https://example.com/search?q=scriptgo");
if (res5 !== null) {
    console.log(res5.search.groups["query"]);
}

// --- 6. Optional Parameters ---
// @native.expected: with_id: 101
// @native.expected: no_id: true
const p6 = new URLPattern({ pathname: "/posts/:id?" });
const res6a = p6.exec("/posts/101");
if (res6a !== null) {
    console.log("with_id: " + res6a.pathname.groups["id"]);
}
const res6b = p6.exec("/posts");
if (res6b !== null) {
    console.log("no_id: " + (res6b.pathname.groups["id"] === undefined));
}

// --- 7. Custom Regular Expression Constraint ---
// @native.expected: digits_only: true
// @native.expected: non_digits: false
// @native.expected: true
const p7 = new URLPattern({ pathname: "/orders/:orderId(\\d+)" });
console.log("digits_only: " + p7.test("/orders/12345"));
console.log("non_digits: " + p7.test("/orders/abcde"));
console.log(p7.hasRegExpGroups);
