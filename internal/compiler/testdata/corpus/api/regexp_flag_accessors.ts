// RegExp flag accessors read the flags the regex was created with.
// @expect: true false false false
// @expect: false true true true true false
// @expect: true true false
// @expect: false q true
const a = /x/g;
const b = /y/imsu;
console.log(a.global, a.ignoreCase, a.multiline, a.sticky);
console.log(b.global, b.ignoreCase, b.multiline, b.dotAll, b.unicode, b.hasIndices);
const c = new RegExp("z", "dy");
console.log(c.sticky, c.hasIndices, c.global);
c.compile("q");
console.log(c.sticky, c.source, c.flags === "");
