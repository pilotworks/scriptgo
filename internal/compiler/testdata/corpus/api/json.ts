// ScriptGo Corpus: Json Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: json.parse
// @expect: hello
const s_json_parse_0: string = JSON.parse("\"hello\""); console.log(s_json_parse_0);

// @api: json.stringify
// @expect: 42
// @expect: "hello"
// @expect: true
console.log(JSON.stringify(42)); console.log(JSON.stringify("hello")); console.log(JSON.stringify(true));
