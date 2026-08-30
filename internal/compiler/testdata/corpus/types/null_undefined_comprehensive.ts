// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
// @expect: false
// @expect: false
// @expect: false
// @expect: object
// @expect: undefined
// @expect: null
// @expect: undefined
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: false
// @expect: true
// @expect: true
// @expect: false
// @expect: x=10, y=20
// @expect: x=10, y=20
// @expect: x=null, y=null
// @expect: x=5, y=50
// @expect: null
// @expect: undefined
// @expect: true
// @expect: true
// @expect: undefined
// @expect: null
// @expect: Anonymous
// @expect: 18
// @expect: null
// @expect: undefined
// @expect: 42
// @expect: true
// @expect: true

// 1. Literal comparisons (strict and loose equality)
console.log(null === null);
console.log(undefined === undefined);
console.log(null === undefined);
console.log(null !== undefined);
console.log(null == undefined);
console.log(null == 0);
console.log(undefined == 0);
console.log(null == "");
console.log(undefined == false);

// 2. Typeof operator
console.log(typeof null);
console.log(typeof undefined);

// 3. Stringification / console output
console.log(null);
console.log(undefined);

// 4. Union variable comparisons
const numNull: number | null = null;
const numUndef: number | undefined = undefined;
const numVal: number | null = 42;
console.log(numNull === null);
console.log(numNull === undefined);
console.log(numUndef === undefined);
console.log(numUndef === null);
console.log(numVal === null);

const strNull: string | null = null;
const strUndef: string | undefined = undefined;
const strVal: string | null = "hello";
console.log(strNull === null);
console.log(strUndef === undefined);
console.log(strVal === strNull);

// 5. Default parameters (undefined triggers default, null preserves null)
function fnDefaults(x: number | null = 10, y: number | null = 20): string {
    return "x=" + x + ", y=" + y;
}
console.log(fnDefaults());
console.log(fnDefaults(undefined, undefined));
console.log(fnDefaults(null, null));
console.log(fnDefaults(5, 50));

// 6. Function return values
function getNull(): string | null {
    return null;
}
function getUndef(): string | undefined {
    return undefined;
}
console.log(getNull());
console.log(getUndef());
console.log(getNull() === null);
console.log(getUndef() === undefined);

// 7. Optional chaining & Nullish coalescing
type UserProfile = {
    name?: string;
    age?: number | null;
};
const user: UserProfile = { age: null };
console.log(user.name);
console.log(user.age);
console.log(user.name ?? "Anonymous");
console.log(user.age ?? 18);

// 8. Arrays containing null and undefined
const arr = [null, undefined, 42];
console.log(arr[0]);
console.log(arr[1]);
console.log(arr[2]);
console.log(arr[0] === null);
console.log(arr[1] === undefined);
