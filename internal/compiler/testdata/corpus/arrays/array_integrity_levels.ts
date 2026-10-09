// Arrays honor Object.freeze, seal and preventExtensions.
// @expect: Cannot assign to read only property '0' of object '[object Array]'
// @expect: loop Cannot assign to read only property '0' of object '[object Array]'
// @expect: Cannot add property 3, object is not extensible
// @expect: Cannot delete property '2' of [object Array]
// @expect: Cannot assign to read only property '0' of object '[object Array]'
// @expect: 1,2,3 true true false
// @expect: Cannot add property 2, object is not extensible
// @expect: Cannot delete property '1' of [object Array]
// @expect: 5,7
// @expect: Cannot add property 1, object is not extensible
// @expect: 1
// @expect: 1,2,3
const a = Object.freeze([1, 2, 3]) as number[];
try { a[0] = 9; } catch (e) { console.log((e as Error).message); }
for (let i = 0; i < a.length; i++) { try { a[i] = 0; } catch (e) { console.log("loop", (e as Error).message); break; } }
try { a.push(4); } catch (e) { console.log((e as Error).message); }
try { a.pop(); } catch (e) { console.log((e as Error).message); }
try { a.sort(); } catch (e) { console.log((e as Error).message); }
console.log(a.join(","), Object.isFrozen(a), Object.isSealed(a), Object.isExtensible(a));
const s = Object.seal([5, 6]) as number[];
s[1] = 7;
try { s.push(8); } catch (e) { console.log((e as Error).message); }
try { s.pop(); } catch (e) { console.log((e as Error).message); }
console.log(s.join(","));
const n = Object.preventExtensions([1, 2]) as number[];
n.pop();
try { n.push(3); } catch (e) { console.log((e as Error).message); }
console.log(n.join(","));
const fresh = [1, 2];
fresh[2] = 3;
console.log(fresh.join(","));
