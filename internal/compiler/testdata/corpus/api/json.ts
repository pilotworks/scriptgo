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

// @api: json.stringify dynamic object
// @expect: {"ok":"yes","count":42}
const dynamicJson: unknown = JSON.parse("{\"ok\":\"yes\",\"count\":42}");
console.log(JSON.stringify(dynamicJson));

// @api: json.stringify array of numbers
// @expect: [1,2,3,4]
console.log(JSON.stringify([1, 2, 3, 4]));

// @api: json.stringify array of objects and roundtrip
// @expect: [{"id":1,"name":"alice"},{"id":2,"name":"bob"}]
// @expect: [{"id":1,"name":"alice"},{"id":2,"name":"bob"}]
interface Member {
    id: number;
    name: string;
}
const members: Member[] = [
    { id: 1, name: "alice" },
    { id: 2, name: "bob" },
];
const serializedMembers: string = JSON.stringify(members);
console.log(serializedMembers);
const parsedMembers: unknown = JSON.parse(serializedMembers);
console.log(JSON.stringify(parsedMembers));
