// ScriptGo Corpus: Set Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: set.add
// @expect: 1
const s_set_add_0 = new Set<number>(); s_set_add_0.add(10); s_set_add_0.add(10); console.log(s_set_add_0.size);

// @api: set.clear
// @expect: 0
const s_set_clear_1 = new Set<number>(); s_set_clear_1.add(10); s_set_clear_1.clear(); console.log(s_set_clear_1.size);

// @api: set.delete
// @expect: true
// @expect: 0
const s_set_delete_2 = new Set<number>(); s_set_delete_2.add(10); console.log(s_set_delete_2.delete(10)); console.log(s_set_delete_2.size);

// @api: set.forEach
// @expect: val=42
const s_set_forEach_3 = new Set<number>();
s_set_forEach_3.add(42);
s_set_forEach_3.forEach((val: number) => {
    console.log("val=" + val);
});

// @api: set.has
// @expect: true
// @expect: false
const s_set_has_4 = new Set<number>(); s_set_has_4.add(10); console.log(s_set_has_4.has(10)); console.log(s_set_has_4.has(20));

// @api: set.union
// @expect: 3
const setA = new Set<number>(); setA.add(1); setA.add(2);
const setB = new Set<number>(); setB.add(2); setB.add(3);
const setUnion = setA.union(setB);
console.log(setUnion.size);

// @api: set.intersection
// @expect: 1
// @expect: true
const setInter = setA.intersection(setB);
console.log(setInter.size);
console.log(setInter.has(2));

// @api: set.difference
// @expect: 1
// @expect: true
const setDiff = setA.difference(setB);
console.log(setDiff.size);
console.log(setDiff.has(1));

// @api: set.symmetricDifference
// @expect: 2
// @expect: true
// @expect: true
const setSymDiff = setA.symmetricDifference(setB);
console.log(setSymDiff.size);
console.log(setSymDiff.has(1));
console.log(setSymDiff.has(3));

// @api: set.isSubsetOf
// @expect: true
// @expect: false
const setSub = new Set<number>(); setSub.add(1);
console.log(setSub.isSubsetOf(setA));
console.log(setA.isSubsetOf(setSub));

// @api: set.isSupersetOf
// @expect: true
// @expect: false
console.log(setA.isSupersetOf(setSub));
console.log(setSub.isSupersetOf(setA));

// @api: set.isDisjointFrom
// @expect: true
// @expect: false
const setDisj = new Set<number>(); setDisj.add(99);
console.log(setA.isDisjointFrom(setDisj));
console.log(setA.isDisjointFrom(setB));

