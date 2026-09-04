// ScriptGo Corpus: Regexp Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: regexp.constructor
// @api: RegExp.readonly source: string
// @api: RegExp.readonly flags: string
// @api: RegExp.readonly global: boolean
// @api: RegExp.readonly ignoreCase: boolean
// @api: RegExp.readonly multiline: boolean
// @api: RegExp.readonly dotAll: boolean
// @api: RegExp.readonly unicode: boolean
// @api: RegExp.readonly sticky: boolean
// @api: RegExp.readonly hasIndices: boolean
// @api: RegExp.readonly unicodeSets: boolean
// @api: RegExp.lastIndex: number
// @expect: foo
// @expect: gim
// @expect: true
// @expect: true
// @expect: true
// @expect: false
// @expect: false
// @expect: false
// @expect: false
// @expect: false
// @expect: 0
const r_regexp_constructor_0 = RegExp("foo", "gim");
console.log(r_regexp_constructor_0.source);
console.log(r_regexp_constructor_0.flags);
console.log(r_regexp_constructor_0.global);
console.log(r_regexp_constructor_0.ignoreCase);
console.log(r_regexp_constructor_0.multiline);
console.log(r_regexp_constructor_0.dotAll);
console.log(r_regexp_constructor_0.unicode);
console.log(r_regexp_constructor_0.sticky);
console.log(r_regexp_constructor_0.hasIndices);
console.log(r_regexp_constructor_0.unicodeSets);
console.log(r_regexp_constructor_0.lastIndex);

// @api: regexp.exec
// @expect: foo
const re_regexp_exec_1 = /foo/;
const m_regexp_exec_1 = re_regexp_exec_1.exec("foobar") as string[];
console.log(m_regexp_exec_1[0]);

// @api: regexp.test
// @expect: true
// @expect: false
const re_regexp_test_2 = /^[a-z]+$/;
console.log(re_regexp_test_2.test("hello"));
console.log(re_regexp_test_2.test("123"));

// @api: RegExp.compile(pattern: string, flags?: string): this
// @expect: true
const re_comp = /abc/;
re_comp.compile("xyz");
console.log(re_comp.test("xyz"));
