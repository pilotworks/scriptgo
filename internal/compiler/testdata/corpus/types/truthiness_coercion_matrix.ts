// @expect: 0: false
// @expect: -0: false
// @expect: 1: true
// @expect: NaN: false
// @expect: empty_str: false
// @expect: str_0: true
// @expect: str_false: true
// @expect: str_hello: true
// @expect: null: false
// @expect: undefined: false
// @expect: union_0: false
// @expect: union_str: true
// @expect: falsy
// @expect: truthy
// @expect: falsy
// @expect: truthy
// @expect: falsy
// @expect: falsy
// Systematic truthiness and coercion rules

function boolNum(v: number): boolean {
    return !!v;
}

function boolStr(v: string): boolean {
    return !!v;
}

function boolUnion(v: string | number | null | undefined): boolean {
    return !!v;
}

console.log("0:", boolNum(0));
console.log("-0:", boolNum(-0));
console.log("1:", boolNum(1));
console.log("NaN:", boolNum(NaN));

console.log("empty_str:", boolStr(""));
console.log("str_0:", boolStr("0"));
console.log("str_false:", boolStr("false"));
console.log("str_hello:", boolStr("hello"));

console.log("null:", boolUnion(null));
console.log("undefined:", boolUnion(undefined));
console.log("union_0:", boolUnion(0));
console.log("union_str:", boolUnion("hi"));

// Ternary truthiness checks
const numZero: number = 0;
console.log(numZero ? "truthy" : "falsy");

const numOne: number = 1;
console.log(numOne ? "truthy" : "falsy");

const strEmpty: string = "";
console.log(strEmpty ? "truthy" : "falsy");

const strNonEmpty: string = "abc";
console.log(strNonEmpty ? "truthy" : "falsy");

const nullVal: string | null = null;
console.log(nullVal ? "truthy" : "falsy");

const undefVal: string | undefined = undefined;
console.log(undefVal ? "truthy" : "falsy");
