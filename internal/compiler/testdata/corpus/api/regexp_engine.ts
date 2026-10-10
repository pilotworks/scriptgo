// @expect: true false 42
// @expect: 17 ll "hi"
// @expect: true true false true
// @expect: true 123 a
// @expect: Đường true false
// @expect: true 6 false 0
// @expect: 2 3
// @expect: 6 7
// @expect: 06/05/2024 a[b|a|c]c a$b
// @expect: Smith, John
// @expect: price: USD 10 $0 100 $
// @expect: bbb x 3 -1
// @expect: a+b+c 01/2020
// @expect: a[b@1]c x2y44
// @expect: -a-a-a- .a.b.c. ĐÀ NẴng
// @expect: [ 'a', 'b', '' ] [ 'a', 'b' ] [ '' ] [] [ 'a', 'b', 'c' ]
// @expect: [ 'one', ' ', ' ', 'two' ] [ 'Hà', 'Nội' ] [ '😀', '😀' ]
// @expect: [ 'T', 'q', 'b', 'f' ] 1
// @expect: SyntaxError
// @expect: SyntaxError
// @expect: TypeError String.prototype.replaceAll called with a non-global RegExp argument
// Regular expressions run on an ECMAScript engine (QuickJS-ng's libregexp):
// lookaround, backreferences, named groups, the d/g/i/m/s/u/v/y flags and
// Unicode property escapes behave as in JavaScript. Indices and lastIndex
// count UTF-16 code units, test and exec advance lastIndex for g and y, an
// invalid pattern throws SyntaxError, and split, replace and replaceAll take
// a RegExp (with $ patterns or a replacer function).
console.log(/foo(?=bar)/.test("foobar"), /foo(?!bar)/.test("foobar"), /(?<=\$)\d+/.exec("cost $42")?.[0]);
console.log(/(?<!\$)\b\d+/.exec("$42 or 17")?.[0], /(\w)\1/.exec("hello")?.[0], /(?<q>['"]).*?\k<q>/.exec(`say "hi" now`)?.[0]);
console.log(/^b/m.test("a\nb"), /a.b/s.test("a\nb"), /a.b/.test("a\nb"), /HELLO/i.test("hello"));
console.log(/\bword\b/.test("a word here"), /\d{2,3}/.exec("x12345")?.[0], /a+?/.exec("aaa")?.[0]);
console.log(/\p{L}+/u.exec("123 Đường phố")?.[0], /^.$/u.test("😀"), /^.$/.test("😀"));

const sticky = /foo/y;
sticky.lastIndex = 3;
console.log(sticky.test("barfoo"), sticky.lastIndex, sticky.test("barfoo"), sticky.lastIndex);
const g = /ố/g;
let m: RegExpExecArray | null;
while ((m = g.exec("phố phố")) !== null) console.log(m.index, g.lastIndex);

console.log("2024-05-06".replace(/(\d+)-(\d+)-(\d+)/, "$3/$2/$1"), "abc".replace(/b/, "[$&|$`|$']"), "a.b".replace(/\./, "$$"));
console.log("John Smith".replace(/(?<first>\w+)\s(?<last>\w+)/, "$<last>, $<first>"));
console.log("price: 10 USD".replace(/(?<n>\d+) (?<cur>\w+)/, "$<cur> $<n> $0 $10 $$"));
console.log("aaa".replace(/a/g, "b"), "x".replace(/y/, "z"), "Hà Nội".search(/N/), "test".search(/z/));
console.log("a-b-c".replaceAll(/-/g, "+"), "2020-01-02".replace(/(\d+)-(\d+)-(\d+)/, (all, y, mo) => `${mo}/${y}`));
console.log("abc".replace(/b/, (s, off) => `[${s}@${off}]`), "x1y22".replace(/\d+/g, (s) => String(Number(s) * 2)));
console.log("aaa".replace(/a*?/g, "-"), "abc".replace(/(?:)/g, "."), "Đà Nẵng".replace(/[àẵ]/g, (c) => c.toUpperCase()));

console.log("a1b2".split(/\d/), "a, b,c".split(/\s*,\s*/, 2), "".split(/x/), "".split(/(?:)/), "abc".split(/(?:)/));
console.log("one  two".split(/(\s)(\s)?/), "Hà-Nội".split(/-/), "😀a😀".split(/a/u));
console.log("The quick brown fox".match(/\b\w/g), "x1".match(/\d/)?.index);

try { new RegExp("(", ""); } catch (e) { console.log((e as Error).name); }
try { new RegExp("a", "gg"); } catch (e) { console.log((e as Error).name); }
try { "a".replaceAll(/a/, "b"); } catch (e) { console.log((e as Error).name, (e as Error).message); }
