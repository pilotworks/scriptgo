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
