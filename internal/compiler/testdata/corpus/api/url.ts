// ScriptGo Corpus: Url Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { URLSearchParams } from "node:url";
import { URL } from "node:url";

// @api: url.URLSearchParams
// @expect: 1
// @expect: 2
const p_url_URLSearchParams_0 = new URLSearchParams("a=1&b=2");
console.log(p_url_URLSearchParams_0.get("a"));
console.log(p_url_URLSearchParams_0.get("b"));

// @api: url.append
// @expect: a=1&b=2
const params_url_append_1 = new URLSearchParams("a=1");
params_url_append_1.append("b", "2");
console.log(params_url_append_1.toString());

// @api: url.canParse
// @expect: true
// @expect: false
console.log(URL.canParse("https://example.com"));
console.log(URL.canParse("not a valid url:::"));

// @api: url.delete
// @expect: b=2
const params_url_delete_3 = new URLSearchParams("a=1&b=2");
params_url_delete_3.delete("a");
console.log(params_url_delete_3.toString());

// @api: url.get
// @expect: bar
const params_url_get_4 = new URLSearchParams("foo=bar");
console.log(params_url_get_4.get("foo"));

// @api: url.getAll
// @expect: a,b
const params_url_getAll_5 = new URLSearchParams("tag=a&tag=b");
const tags_url_getAll_5: string[] = params_url_getAll_5.getAll("tag");
console.log(tags_url_getAll_5.join(","));

// @api: url.has
// @expect: true
// @expect: false
const params_url_has_6 = new URLSearchParams("key=val");
console.log(params_url_has_6.has("key"));
console.log(params_url_has_6.has("other"));

// @api: url.set
// @expect: 99
const params_url_set_7 = new URLSearchParams("x=1");
params_url_set_7.set("x", "99");
console.log(params_url_set_7.get("x"));

// @api: url.sort
// @expect: a=1&b=2&c=3
const params_url_sort_8 = new URLSearchParams("c=3&a=1&b=2");
params_url_sort_8.sort();
console.log(params_url_sort_8.toString());

// @api: url.toJSON
// @expect: https://example.com/path
const u_url_toJSON_9 = new URL("https://example.com/path");
console.log(u_url_toJSON_9.toJSON());

// @api: url.url
// @expect: https://example.com/path
const u_url_url_10 = new URL("https://example.com/path");
console.log(u_url_url_10.toString());
