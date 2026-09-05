import { escape, unescape, stringify, encode, parse, decode } from "node:querystring";

// @api: querystring.escape
// @expect: hello%20world%20%26%20foo%3Dbar
console.log(escape("hello world & foo=bar"));

// @api: querystring.unescape
// @expect: hello world & foo=bar
console.log(unescape("hello%20world%20%26%20foo%3Dbar"));

// @api: querystring.stringify
// @expect: foo=bar&baz=qux&baz=quux&corge=
// @expect: foo:bar;baz:qux
const q1 = { foo: "bar", baz: ["qux", "quux"], corge: "" };
console.log(stringify(q1));

const q2 = { foo: "bar", baz: "qux" };
console.log(stringify(q2, ";", ":"));

// @api: querystring.encode
// @expect: a=1&b=2
const q3 = { a: "1", b: "2" };
console.log(encode(q3));

// @api: querystring.parse
// @expect: object
const parsed = parse("foo=bar&baz=qux");
console.log(typeof parsed);

// @api: querystring.decode
// @expect: object
const custom = decode("foo:bar;baz:1;baz:2", ";", ":");
console.log(typeof custom);

// @api: querystring default object
// @expect: a=b
const q4 = { a: "b" };
console.log(stringify(q4));
