// @expect: 1 5
// A field read after a write observes the new value (no stale CSE).
class P { x = 1; }
const p = new P();
const a = p.x;
p.x = 5;
const b = p.x;
console.log(a, b);
