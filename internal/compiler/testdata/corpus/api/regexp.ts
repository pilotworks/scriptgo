// ScriptGo Corpus: Regexp Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: regexp.constructor
// @expect: foo
// @expect: i
const r_regexp_constructor_0 = RegExp("foo", "i"); console.log(r_regexp_constructor_0.source); console.log(r_regexp_constructor_0.flags);

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
