// A replacer function receives match, captures, offset, the string and, with
// named groups, the groups object; groups is a null-prototype object whose
// unmatched groups are undefined.
console.log("2020-01-02".replace(/(\d+)-(\d+)-(\d+)/, (all, y, m, d, offset, str) => `${d}.${m}.${y}@${offset}/${str.length}`));
console.log("a-b c-d".replace(/(\w)-(\w)/g, (m, a, b, off) => `${b}${a}${off}`));
console.log("k=v".replace(/(?<key>\w)=(?<val>\w)/, (m, k, v, off, s, groups) => `${groups.val}=${groups.key}`));
console.log("x1y".replaceAll(/\d/g, (m, off, s) => `[${m}@${off}/${s}]`));

const g = /(?<a>x)|(?<b>y)/.exec("x")!.groups!;
console.log(g);
console.log(g.a, g.b, Object.keys(g), JSON.stringify(g), "b" in g);
console.log(Object.entries(g));
for (const r of "a1b".matchAll(/(?<l>[a-z])(?<d>\d)?/g)) console.log(r.groups);
// @expect: 02.01.2020@0/10
// @expect: ba0 dc4
// @expect: v=k
// @expect: x[1@1/x1y]y
// @expect: [Object: null prototype] { a: 'x', b: undefined }
// @expect: x undefined [ 'a', 'b' ] {"a":"x"} true
// @expect: [ [ 'a', 'x' ], [ 'b', undefined ] ]
// @expect: [Object: null prototype] { l: 'a', d: '1' }
// @expect: [Object: null prototype] { l: 'b', d: undefined }
