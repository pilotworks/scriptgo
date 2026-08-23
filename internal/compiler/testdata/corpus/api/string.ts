// ScriptGo Corpus: String Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: string.charAt
// @expect: e
const s_string_charAt_0: string = "hello"; console.log(s_string_charAt_0.charAt(1));

// @api: string.charCodeAt
// @expect: 65
const s_string_charCodeAt_1: string = "ABC"; console.log(s_string_charCodeAt_1.charCodeAt(0));

// @api: string.concat
// @expect: foobarbaz
const s_string_concat_2: string = "foo"; console.log(s_string_concat_2.concat("bar", "baz"));

// @api: string.endsWith
// @expect: true
// @expect: false
const s_string_endsWith_3: string = "hello world";
console.log(s_string_endsWith_3.endsWith("world"));
console.log(s_string_endsWith_3.endsWith("hello"));

// @api: string.includes
// @expect: true
// @expect: false
const s_string_includes_4: string = "hello world"; console.log(s_string_includes_4.includes("world")); console.log(s_string_includes_4.includes("foo"));

// @api: string.indexOf
// @expect: 6
// @expect: -1
const s_string_indexOf_5: string = "hello world"; console.log(s_string_indexOf_5.indexOf("world")); console.log(s_string_indexOf_5.indexOf("foo"));

// @api: string.lastIndexOf
// @expect: 6
const s_string_lastIndexOf_6: string = "hello hello";
console.log(s_string_lastIndexOf_6.lastIndexOf("hello"));

// @api: string.match
// @expect: 123
const s_string_match_7: string = "hello 123";
const m_string_match_7 = s_string_match_7.match(/[0-9]+/) as string[];
console.log(m_string_match_7[0]);

// @api: string.padEnd
// @expect: 500
const s_string_padEnd_8: string = "5";
console.log(s_string_padEnd_8.padEnd(3, "0"));

// @api: string.padStart
// @expect: 005
const s_string_padStart_9: string = "5";
console.log(s_string_padStart_9.padStart(3, "0"));

// @api: string.repeat
// @expect: abcabcabc
const s_string_repeat_10: string = "abc";
console.log(s_string_repeat_10.repeat(3));

// @api: string.replace
// @expect: hello there
const s_string_replace_11: string = "hello world"; console.log(s_string_replace_11.replace("world", "there"));

// @api: string.replaceAll
// @expect: x b x b
const s_string_replaceAll_12: string = "a b a b";
console.log(s_string_replaceAll_12.replaceAll("a", "x"));

// @api: string.search
// @expect: 6
const s_string_search_13: string = "hello 123";
console.log(s_string_search_13.search(/[0-9]+/));

// @api: string.slice
// @expect: hello
// @expect: world
const s_string_slice_14: string = "hello world"; console.log(s_string_slice_14.slice(0, 5)); console.log(s_string_slice_14.slice(6));

// @api: string.split
// @expect: a-b-c
const s_string_split_15: string = "a,b,c"; console.log(s_string_split_15.split(",").join("-"));

// @api: string.startsWith
// @expect: true
// @expect: false
const s_string_startsWith_16: string = "hello world";
console.log(s_string_startsWith_16.startsWith("hello"));
console.log(s_string_startsWith_16.startsWith("world"));

// @api: string.substring
// @expect: hello
const s_string_substring_17: string = "hello world"; console.log(s_string_substring_17.substring(0, 5));

// @api: string.toLowerCase
// @expect: hello
const s_string_toLowerCase_18: string = "HELLO"; console.log(s_string_toLowerCase_18.toLowerCase());

// @api: string.toUpperCase
// @expect: HELLO
const s_string_toUpperCase_19: string = "hello"; console.log(s_string_toUpperCase_19.toUpperCase());

// @api: string.trim
// @expect: hello
const s_string_trim_20: string = "  hello  "; console.log(s_string_trim_20.trim());

// @api: string.trimEnd
// @expect: !  hello
const s_string_trimEnd_21: string = "  hello  ";
console.log("!" + s_string_trimEnd_21.trimEnd());

// @api: string.trimStart
// @expect: hello  !
const s_string_trimStart_22: string = "  hello  ";
console.log(s_string_trimStart_22.trimStart() + "!");

// @api: string.codePointAt
// @expect: 65
console.log("ABC".codePointAt(0));

// @api: String.fromCodePoint
// @expect: A
console.log(String.fromCodePoint(65));

// @api: string.isWellFormed
// @expect: true
console.log("hello".isWellFormed());

// @api: string.toWellFormed
// @expect: hello
console.log("hello".toWellFormed());

// @api: string.matchAll
// @expect: 2
const s_string_matchAll = "test1test2";
const it_string_matchAll = Array.from(s_string_matchAll.matchAll(/test[0-9]/g));
console.log(it_string_matchAll.length);

// @api: string.anchor
// @expect: <a name="top">hello</a>
console.log("hello".anchor("top"));

// @api: string.at
// @expect: h
// @expect: o
console.log("hello".at(0));
console.log("hello".at(-1));

// @api: string.big
// @expect: <big>hello</big>
console.log("hello".big());

// @api: string.blink
// @expect: <blink>hello</blink>
console.log("hello".blink());

// @api: string.bold
// @expect: <b>hello</b>
console.log("hello".bold());

// @api: string.fixed
// @expect: <tt>hello</tt>
console.log("hello".fixed());

// @api: string.fontcolor
// @expect: <font color="red">hello</font>
console.log("hello".fontcolor("red"));

// @api: string.fontsize
// @expect: <font size="7">hello</font>
console.log("hello".fontsize(7));

// @api: String.fromCharCode
// @expect: AB
console.log(String.fromCharCode(65, 66));

// @api: string.italics
// @expect: <i>hello</i>
console.log("hello".italics());

// @api: string.link
// @expect: <a href="https://example.com">hello</a>
console.log("hello".link("https://example.com"));

// @api: string.localeCompare
// @expect: 0
console.log("a".localeCompare("a"));

// @api: string.normalize
// @expect: hello
console.log("hello".normalize());

// @api: String.raw
// @expect: hello
console.log(String.raw`hello`);

// @api: string.small
// @expect: <small>hello</small>
console.log("hello".small());

// @api: string.strike
// @expect: <strike>hello</strike>
console.log("hello".strike());

// @api: string.sub
// @expect: <sub>hello</sub>
console.log("hello".sub());

// @api: string.substr
// @expect: ell
console.log("hello".substr(1, 3));

// @api: string.sup
// @expect: <sup>hello</sup>
console.log("hello".sup());

// @api: string.toLocaleLowerCase
// @expect: hello
console.log("HELLO".toLocaleLowerCase());

// @api: string.toLocaleUpperCase
// @expect: HELLO
console.log("hello".toLocaleUpperCase());

// @api: string.toString
// @expect: hello
console.log("hello".toString());

// @api: string.trimLeft
// @expect: hello 
console.log("  hello ".trimLeft());

// @api: string.trimRight
// @expect:   hello
console.log("  hello ".trimRight());

// @api: string.valueOf
// @expect: hello
console.log("hello".valueOf());

// @api: String.new
// @expect: hello
console.log(String("hello"));

// @api: string.length
// @expect: 5
console.log("hello".length);

