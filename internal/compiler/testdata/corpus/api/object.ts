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

// @api: object.entries
// @expect: 2
const o_object_entries = { a: 1, b: 2 };
console.log(Object.entries(o_object_entries).length);

// @api: object.fromEntries
// @expect: 1
const o_fromEntries = Object.fromEntries([["a", 1]]);
console.log(o_fromEntries.a);

// @api: object.groupBy
// @expect: 2
const grouped = Object.groupBy([1, 2, 3, 4], (x: number) => x % 2 === 0 ? "even" : "odd");
console.log(grouped.even!.length);

// @api: object.create
// @expect: true
const obj_create: object = Object.create(null);
console.log(typeof obj_create === "object");

// @api: object.defineProperties
// @expect: true
const obj_defProps = Object.defineProperties({ a: 1 }, { a: { value: 2 } });
console.log(typeof obj_defProps === "object");

// @api: object.defineProperty
// @expect: true
const obj_defProp = Object.defineProperty({ a: 1 }, "a", { value: 2 });
console.log(typeof obj_defProp === "object");

// @api: object.freeze
// @expect: true
const obj_freeze = Object.freeze({ x: 1 });
console.log(typeof obj_freeze === "object");

// @api: object.getOwnPropertyDescriptor
// @expect: true
const obj_desc = Object.getOwnPropertyDescriptor({ a: 1 }, "a");
console.log(typeof obj_desc === "object");

// @api: object.getOwnPropertyDescriptors
// @expect: true
const obj_descs = Object.getOwnPropertyDescriptors({ a: 1 });
console.log(typeof obj_descs === "object");

// @api: object.getOwnPropertyNames
// @expect: a,b
console.log(Object.getOwnPropertyNames({ a: 1, b: 2 }).join(","));

// @api: object.getOwnPropertySymbols
// @expect: 0
console.log(Object.getOwnPropertySymbols({ a: 1 }).length);

// @api: object.getPrototypeOf
// @expect: true
const proto: object = Object.getPrototypeOf({ a: 1 });
console.log(typeof proto === "object");

// @api: object.hasOwnProperty
// @expect: true
console.log(({ a: 1 }).hasOwnProperty("a"));

// @api: object.isExtensible
// @expect: true
console.log(Object.isExtensible({ a: 1 }));

// @api: object.isFrozen
// @expect: true
console.log(Object.isFrozen({ a: 1 }));

// @api: object.isPrototypeOf
// @expect: false
console.log(({ a: 1 }).isPrototypeOf({ b: 2 }));

// @api: object.isSealed
// @expect: true
console.log(Object.isSealed({ a: 1 }));

// @api: object.preventExtensions
// @expect: true
console.log(typeof Object.preventExtensions({ a: 1 }) === "object");

// @api: object.propertyIsEnumerable
// @expect: true
console.log(({ a: 1 }).propertyIsEnumerable("a"));

// @api: object.seal
// @expect: true
console.log(typeof Object.seal({ a: 1 }) === "object");

// @api: object.setPrototypeOf
// @expect: true
console.log(typeof Object.setPrototypeOf({ a: 1 }, null) === "object");

// @api: object.toLocaleString
// @expect: [object Object]
console.log(({ a: 1 }).toLocaleString());

// @api: object.toString
// @expect: [object Object]
console.log(({ a: 1 }).toString());

// @api: object.valueOf
// @expect: true
console.log(typeof ({ a: 1 }).valueOf() === "object");

// @api: Objectconstructor
// @expect: true
console.log(typeof Object() === "object");

