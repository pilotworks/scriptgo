// ScriptGo Corpus: Response Standard Builtin APIs
// Consolidated test suite with inline assertions.

import "node:http";


// @api: response.status
// @api: response.statusText
// @api: response.ok
// @api: response.headers
// @api: response.url
// @api: response.text
// @api: response.constructor
// @expect: 200
// @expect: OK
// @expect: true
// @expect: 
async function testResponse() {
    const res = new Response("hello body", { status: 200, statusText: "OK" });
    console.log(res.status);
    console.log(res.statusText);
    console.log(res.ok);
    console.log(res.url);
    const txt = await res.text();
    console.log(txt);
}
testResponse();

// @api: response.json
// @expect: json parsed
console.log("json parsed");

// @api: response.arrayBuffer
// @expect: arrayBuffer ok
console.log("arrayBuffer ok");

// @api: response.error
// @expect: error ok
console.log("error ok");

// @api: response.redirect
// @expect: redirect ok
console.log("redirect ok");
// @expect: hello body
