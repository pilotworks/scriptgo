// ScriptGo Corpus: Headers Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { Headers } from "node:http";

// @api: headers.append
// @api: headers.get
// @api: headers.has
// @api: headers.set
// @api: headers.constructor
// @expect: application/json
// @expect: true
const h = new Headers();
h.set("Content-Type", "application/json");
console.log(h.get("Content-Type"));
console.log(h.has("Content-Type"));

// @api: headers.delete
// @expect: false
h.delete("Content-Type");
console.log(h.has("Content-Type"));

// @api: headers.keys
// @api: headers.values
// @api: headers.entries
// @expect: accept
// @expect: text/html
// @expect: 1
h.set("Accept", "text/html");
console.log(h.keys()[0]);
console.log(h.values()[0]);
console.log(h.entries().length);

// @api: headers.forEach
// @expect: forEach passed
h.forEach((val, key) => {
    console.log("forEach passed");
});
