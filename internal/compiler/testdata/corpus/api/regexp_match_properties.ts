// @expect: 42 42 6 order 42 and 77 2
// @expect: [ '42', '42', index: 6, input: 'order 42 and 77', groups: undefined ]
// @expect: [ '42', '77' ]
// @expect: order 42 order 42 0 order 42 and 77
// @expect: 2024 05 06 0
// @expect: null
// @expect: 42 4 6
// @expect: 77 7 13
// @expect: 4 3 2024 05 06 06
// @expect: 7 36 13
// @expect: [ 'ac', undefined, index: 0, input: 'ac', groups: undefined ] true
// @expect: [ 1, 2, 5, 6 ]
// @expect: 3 [ '1@1', '22@3', '333@6' ] 333
// @expect: [ 0, 1, 2, 3 ]
// @expect: 1 null [ '1', index: 1, input: 'x1', groups: undefined ]
// @expect: 1 null
// @expect: null
// @expect: 1 null
// RegExp match arrays carry index (in UTF-16 code units), input and groups
// (named captures) like JavaScript's, and console.log shows them after the
// elements. matchAll yields one match array per match, advancing past empty
// matches; a global match() returns the matched strings.
const s = "order 42 and 77";
const m = s.match(/(\d+)/);
if (m) console.log(m[0], m[1], m.index, m.input, m.length);
console.log(m);
const g = s.match(/\d+/g);
console.log(g);
const e = /(\w+) (\d+)/.exec(s);
if (e) console.log(e[0], e[1], e[2], e.index, e.input);
const named = "2024-05-06".match(/(?<y>\d{4})-(?<m>\d{2})-(?<d>\d{2})/);
if (named && named.groups) console.log(named.groups.y, named.groups.m, named.groups.d, named.index);
console.log("x".match(/y/));
for (const mm of s.matchAll(/(\d)(\d)/g)) console.log(mm[0], mm[1], mm.index);
const dated = "on 2024-05-06".match(/(?<y>\d{4})-(?<m>\d{2})-(?<d>\d{2})/);
if (dated && dated.groups) console.log(dated.length, dated.index, dated.groups.y, dated.groups.m, dated.groups.d, dated[3]);
const viet = "Hà Nội 36 phố".match(/\d+/);
if (viet) console.log(viet.index, viet[0], viet.input?.length);
const opt = "ac".match(/a(b)?c/);
console.log(opt, opt?.[1] === undefined);
const reO = /o/g;
let hit: RegExpExecArray | null;
const found: number[] = [];
while ((hit = reO.exec("foo boo")) !== null) found.push(hit.index);
console.log(found);
const all = [..."a1b22c333".matchAll(/(?<n>\d+)/g)];
console.log(all.length, all.map((m) => m[0] + "@" + m.index), all[2].groups?.n);
const empty = [..."abc".matchAll(/x*/g)].map((m) => m.index);
console.log(empty);
console.log("A-B".match(/-/)?.index, "none".match(/z/g), "x1".match(/\d/));
console.log(1, "none".match(/z/g));
console.log("none".match(/z/g));
console.log(1, "none".match(/z/));
