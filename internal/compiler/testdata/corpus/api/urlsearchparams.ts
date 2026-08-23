// ScriptGo Corpus: URLSearchParams Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { URLSearchParams } from "node:url";

// @api: urlsearchparams.append
// @api: urlsearchparams.get
// @api: urlsearchparams.has
// @api: urlsearchparams.set
// @api: urlsearchparams.size
// @api: urlsearchparams.constructor
// @expect: 1
// @expect: true
// @expect: 2
const sp = new URLSearchParams("a=1");
console.log(sp.get("a"));
sp.append("b", "2");
console.log(sp.has("b"));
console.log(sp.size);

// @api: urlsearchparams.getAll
// @expect: 1,99
sp.append("a", "99");
console.log(sp.getAll("a").join(","));

// @api: urlsearchparams.delete
// @expect: false
sp.delete("a");
console.log(sp.has("a"));

// @api: urlsearchparams.keys
// @api: urlsearchparams.values
// @api: urlsearchparams.entries
// @expect: b
// @expect: 2
// @expect: 1
console.log(sp.keys().join(","));
console.log(sp.values().join(","));
console.log(sp.entries().length);

// @api: urlsearchparams.forEach
// @expect: 2 b
sp.forEach((v, k) => {
    console.log(v + " " + k);
});

// @api: urlsearchparams.sort
// @api: urlsearchparams.toString
// @expect: x=1&y=2
const sp2 = new URLSearchParams("y=2&x=1");
sp2.sort();
console.log(sp2.toString());
