// ScriptGo Corpus: Request Standard Builtin APIs
// Consolidated test suite with inline assertions.


// @api: request.url
// @api: request.method
// @api: request.headers
// @api: request.body
// @api: request.constructor
// @expect: https://example.com/api
// @expect: POST
// @expect: application/json
// @expect: {"hello":"world"}
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
console.log(req.body);
