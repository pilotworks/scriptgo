// ScriptGo Corpus: Scenario: Collections, Math & String Manipulation
// Consolidated test suite with inline assertions.

// --- Context Case: scenarios_array_es_methods ---
// @expect: === Array.findIndex ===
// @expect: 1
// @expect: -1
// @expect: 1
// @expect: -1
// @expect: === Array.fill ===
// @expect: 0
// @expect: 0
// @expect: 0
// @expect: 0
// @expect: 1
// @expect: 9
// @expect: 9
// @expect: 4
// @expect: === Array.toReversed ===
// @expect: 3
// @expect: 2
// @expect: 1
// @expect: 1
// @expect: 2
// @expect: 3
// @expect: === Array.toSorted ===
// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 40
// @expect: 50
// @expect: 30
// @expect: 10
// @expect: 50
// @expect: 20
// @expect: 40
// @expect: apple
// @expect: banana
// @expect: cherry
function testFindIndex_array_es_methods_0(): void {
  console.log("=== Array.findIndex ===");
  const nums_array_es_methods_0 = [10, 20, 30];
  console.log(nums_array_es_methods_0.findIndex((x) => x > 15));
  console.log(nums_array_es_methods_0.findIndex((x) => x > 50));
  const words_array_es_methods_0 = ["apple", "banana", "cherry"];
  console.log(words_array_es_methods_0.findIndex((s) => s.startsWith("b")));
  console.log(words_array_es_methods_0.findIndex((s) => s === "date"));
}

function testFill_array_es_methods_0(): void {
  console.log("=== Array.fill ===");
  const a_array_es_methods_0 = [1, 2, 3, 4];
  a_array_es_methods_0.fill(0);
  for (const x of a_array_es_methods_0) {
    console.log(x);
  }
  const b_array_es_methods_0 = [1, 2, 3, 4];
  b_array_es_methods_0.fill(9, 1, 3);
  for (const x of b_array_es_methods_0) {
    console.log(x);
  }
}

function testToReversed_array_es_methods_0(): void {
  console.log("=== Array.toReversed ===");
  const original_array_es_methods_0 = [1, 2, 3];
  const reversed_array_es_methods_0 = original_array_es_methods_0.toReversed();
  for (const x of reversed_array_es_methods_0) {
    console.log(x);
  }
  for (const x of original_array_es_methods_0) {
    console.log(x);
  }
}

function testToSorted_array_es_methods_0(): void {
  console.log("=== Array.toSorted ===");
  const nums_array_es_methods_0 = [30, 10, 50, 20, 40];
  const sortedNums_array_es_methods_0 = nums_array_es_methods_0.toSorted();
  for (const x of sortedNums_array_es_methods_0) {
    console.log(x);
  }
  for (const x of nums_array_es_methods_0) {
    console.log(x);
  }
  const words_array_es_methods_0 = ["cherry", "apple", "banana"];
  const sortedWords_array_es_methods_0 = words_array_es_methods_0.toSorted();
  for (const s of sortedWords_array_es_methods_0) {
    console.log(s);
  }
}

function main_array_es_methods_0(): void {
  testFindIndex_array_es_methods_0();
  testFill_array_es_methods_0();
  testToReversed_array_es_methods_0();
  testToSorted_array_es_methods_0();
}

main_array_es_methods_0();

// --- Context Case: scenarios_date_formatting ---
// @expect: 1700000000000
// @expect: 2023-11-14T22:13:20.000Z
// @expect: true
// @expect: 1704067200000
// @expect: 2024-01-01T00:00:00.000Z
const d_date_formatting_1 = new Date(1700000000000);
console.log(d_date_formatting_1.getTime());
console.log(d_date_formatting_1.toISOString());
console.log(d_date_formatting_1.toString().length > 0);

const ts_date_formatting_1 = Date.parse("2024-01-01T00:00:00.000Z");
console.log(ts_date_formatting_1);

const d2_date_formatting_1 = new Date(ts_date_formatting_1);
console.log(d2_date_formatting_1.toISOString());

// --- Context Case: scenarios_map_advanced ---
// @expect: 4
// @expect: 3
// @expect: 4
// @expect: Map(4) { 'a' => '1', 'b' => '2', 'c' => '3', 'd' => '4' }
// @expect: 4
// @expect: true
// @expect: false
// @expect: Set(4) { 'alpha', 'beta', 'gamma', 'delta' }
const m_map_advanced_2 = new Map<string, string>([["a", "1"], ["b", "2"]]);
m_map_advanced_2.set("c", "3").set("d", "4");
console.log(m_map_advanced_2.size);
console.log(m_map_advanced_2.get("c"));
console.log(m_map_advanced_2.get("d"));
console.log(m_map_advanced_2);

const s_map_advanced_2 = new Set<string>(["alpha", "beta", "alpha"]);
s_map_advanced_2.add("gamma").add("delta");
console.log(s_map_advanced_2.size);
console.log(s_map_advanced_2.has("beta"));
console.log(s_map_advanced_2.has("omega"));
console.log(s_map_advanced_2);

// --- Context Case: scenarios_math_extended_api ---
// @expect: 3
// @expect: 0
// @expect: 3
// @expect: 10
// @expect: -1
// @expect: 0
// @expect: 1
// @expect: 0
// @expect: 0
// @expect: 13
// @expect: true
// @expect: true
// @expect: true
// @expect: true
console.log(Math.log10(1000));
console.log(Math.log10(1));
console.log(Math.log2(8));
console.log(Math.log2(1024));

console.log(Math.sign(-50));
console.log(Math.sign(0));
console.log(Math.sign(42));

console.log(Math.atan(0));
console.log(Math.atan2(0, 5));
console.log(Math.hypot(5, 12));

console.log(Math.SQRT2 > 1.41 && Math.SQRT2 < 1.42);
console.log(Math.SQRT1_2 > 0.70 && Math.SQRT1_2 < 0.71);
console.log(Math.LN2 > 0.69 && Math.LN2 < 0.70);
console.log(Math.LN10 > 2.30 && Math.LN10 < 2.31);

// --- Context Case: scenarios_math_trig_extended ---
// @expect: 42
// @expect: 4
// @expect: 5
// @expect: 4
// @expect: -4
// @expect: 0
// @expect: 1
// @expect: 0
// @expect: 0
// @expect: 1
// @expect: 5
// @expect: 0
console.log(Math.abs(-42));
console.log(Math.floor(4.9));
console.log(Math.ceil(4.1));
console.log(Math.trunc(4.9));
console.log(Math.trunc(-4.9));
console.log(Math.sin(0));
console.log(Math.cos(0));
console.log(Math.tan(0));
console.log(Math.log(1));
console.log(Math.exp(0));
console.log(Math.hypot(3, 4));
console.log(Math.atan2(0, 1));

// --- Context Case: scenarios_number_static_methods ---
// @expect: true
// @expect: false
// @expect: false
// @expect: false
// @expect: true
// @expect: false
// @expect: false
// @expect: true
// @expect: false
// @expect: 12345
// @expect: 3.14159
// @expect: 9007199254740991
// @expect: -9007199254740991
// @expect: true
console.log(Number.isInteger(42));
console.log(Number.isInteger(42.5));
console.log(Number.isInteger(NaN));
console.log(Number.isInteger(Infinity));

console.log(Number.isFinite(100));
console.log(Number.isFinite(Infinity));
console.log(Number.isFinite(NaN));

console.log(Number.isNaN(NaN));
console.log(Number.isNaN(123));

console.log(Number.parseInt("12345"));
console.log(Number.parseFloat("3.14159"));

console.log(Number.MAX_SAFE_INTEGER);
console.log(Number.MIN_SAFE_INTEGER);
console.log(Number.EPSILON > 0 && Number.EPSILON < 1e-15);

// --- Context Case: scenarios_string_advanced_methods ---
// @expect: 00042
// @expect: 42----
// @expect: [leading and trailing   ]
// @expect: [   leading and trailing]
// @expect: abc-abc-abc-
// @expect: H
// @expect: o
// @expect: 72
// @expect: 101
const id_string_advanced_methods_6: string = "42";
console.log(id_string_advanced_methods_6.padStart(5, "0"));
console.log(id_string_advanced_methods_6.padEnd(6, "-"));

const raw_string_advanced_methods_6: string = "   leading and trailing   ";
console.log(`[${raw_string_advanced_methods_6.trimStart()}]`);
console.log(`[${raw_string_advanced_methods_6.trimEnd()}]`);

const rep_string_advanced_methods_6: string = "abc-";
console.log(rep_string_advanced_methods_6.repeat(3));

const greeting_string_advanced_methods_6: string = "Hello";
console.log(greeting_string_advanced_methods_6.charAt(0));
console.log(greeting_string_advanced_methods_6.charAt(4));
console.log(greeting_string_advanced_methods_6.charCodeAt(0));
console.log(greeting_string_advanced_methods_6.charCodeAt(1));

// --- Context Case: scenarios_string_extended ---
// @expect: hello world
// @expect: foo baz
// @expect: bcd
const str_string_extended_7: string = "  hello world  ";
console.log(str_string_extended_7.trim());
console.log("foo bar".replace("bar", "baz"));
console.log("abcdef".substring(1, 4));

// --- Context Case: scenarios_string_manipulation ---
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: 6
// @expect: -1
// @expect: 23
// @expect: S
// @expect: 83
// @expect: abcabcabc
// @expect: padded  
// @expect:   padded
// @expect: scriptgo compiler engine
// @expect: SCRIPTGO COMPILER ENGINE
const text_string_manipulation_8 = "ScriptGo Compiler Engine";

console.log(text_string_manipulation_8.startsWith("Script"));
console.log(text_string_manipulation_8.startsWith("Engine"));
console.log(text_string_manipulation_8.endsWith("Engine"));
console.log(text_string_manipulation_8.endsWith("Script"));

console.log(text_string_manipulation_8.includes("Compiler"));
console.log(text_string_manipulation_8.includes("xyz"));

console.log(text_string_manipulation_8.indexOf("Go"));
console.log(text_string_manipulation_8.indexOf("xyz"));
console.log(text_string_manipulation_8.lastIndexOf("e"));

console.log(text_string_manipulation_8.charAt(0));
console.log(text_string_manipulation_8.charCodeAt(0));

console.log("abc".repeat(3));
console.log("  padded  ".trimStart());
console.log("  padded  ".trimEnd());
console.log(text_string_manipulation_8.toLowerCase());
console.log(text_string_manipulation_8.toUpperCase());
