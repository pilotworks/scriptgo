import { escape, unescape, stringify, encode, parse, decode, querystring } from "node:querystring";

class ParsedUrlQuery {
    [key: string]: unknown;
    set(key: string, value: string | string[]): void {
        this[key] = value;
    }
}

// @api: querystring.escape
// @expect: hello%20world%20%26%20foo%3Dbar
console.log(escape("hello world & foo=bar"));

// @api: querystring.unescape
// @expect: hello world & foo=bar
// @expect: hello world
console.log(unescape("hello%20world%20%26%20foo%3Dbar"));
console.log(unescape("hello+world"));

// @api: querystring.stringify
// @expect: foo=bar&baz=qux&baz=quux&corge=
// @expect: foo:bar;baz:qux
const q1 = new ParsedUrlQuery();
q1.set("foo", "bar");
q1.set("baz", ["qux", "quux"]);
q1.set("corge", "");
console.log(stringify(q1));

const q2 = new ParsedUrlQuery();
q2.set("foo", "bar");
q2.set("baz", "qux");
console.log(stringify(q2, ";", ":"));

// @api: querystring.encode
// @expect: a=1&b=2
const q3 = new ParsedUrlQuery();
q3.set("a", "1");
q3.set("b", "2");
console.log(encode(q3));

// @api: querystring.parse
// @expect: bar
// @expect: qux
const parsed = parse("foo=bar&baz=qux");
console.log(parsed.get("foo") as string);
console.log(parsed.get("baz") as string);

// @api: querystring.decode
// @expect: bar
// @expect: 2
const custom = decode("foo:bar;baz:1;baz:2", ";", ":");
console.log(custom.get("foo") as string);
const bazArr = custom.get("baz") as string[];
console.log(bazArr.length);

// @api: querystring default object
// @expect: a=b
const q4 = new ParsedUrlQuery();
q4.set("a", "b");
console.log(stringify(q4));
