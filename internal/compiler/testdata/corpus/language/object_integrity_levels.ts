// Frozen, sealed and non-extensible objects reject writes, additions and
// deletions with V8's TypeErrors.
// Strict mode (as in ES modules), where integrity violations throw.
"use strict";

// @expect: TypeError Cannot assign to read only property 'x' of object '#<Object>' 1
// @expect: 2 true true false
// @expect: Cannot assign to read only property 'v' of object '#<P>' 3
// @expect: Cannot add property b, object is not extensible
// @expect: 7
// @expect: Cannot delete property 'k' of #<Object>
const o = Object.freeze({ x: 1 });
try { (o as { x: number }).x = 5; console.log("no throw", o.x); } catch (e) { console.log((e as Error).name, (e as Error).message, o.x); }
const s = Object.seal({ y: 1 });
(s as { y: number }).y = 2;
console.log(s.y, Object.isFrozen(o), Object.isSealed(s), Object.isExtensible(s));
class P { constructor(public v: number) { Object.freeze(this); } bump() { this.v++; } }
const p = new P(3);
try { p.bump(); } catch (e) { console.log((e as Error).message, p.v); }
const d: Record<string, number> = { a: 1 };
Object.preventExtensions(d);
try { d["b"] = 2; } catch (e) { console.log((e as Error).message); }
d["a"] = 7;
console.log(d["a"]);
const sealed = Object.seal({ k: 1 }) as { k?: number };
try { delete sealed.k; } catch (e) { console.log((e as Error).message); }
