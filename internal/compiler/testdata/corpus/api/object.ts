// ScriptGo Corpus: Object Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: object.assign
// @expect: 10,20
class Point_object_assign_0 {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}
const p_object_assign_0 = new Point_object_assign_0(1, 2);
Object.assign(p_object_assign_0, new Point_object_assign_0(10, 20));
console.log(p_object_assign_0.x + "," + p_object_assign_0.y);

// @api: object.hasOwn
// @expect: true
// @expect: false
const o_object_hasOwn_1 = { a: 1 }; console.log(Object.hasOwn(o_object_hasOwn_1, "a")); console.log(Object.hasOwn(o_object_hasOwn_1, "b"));

// @api: object.is
// @expect: true
// @expect: false
console.log(Object.is(1, 1)); console.log(Object.is(1, 2));

// @api: object.keys
// @expect: a,b
const o_object_keys_3 = { a: 1, b: 2 }; console.log(Object.keys(o_object_keys_3).join(","));

// @api: object.values
// @expect: 1,2
const o_object_values_4 = { a: 1, b: 2 }; console.log(Object.values(o_object_values_4).join(","));
