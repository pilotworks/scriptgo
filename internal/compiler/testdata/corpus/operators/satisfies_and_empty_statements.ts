// @expect: 42
// @expect: hello
// @expect: red
// @expect: true
const a = 42 satisfies number;;;
const b = "hello" satisfies string;
;;
interface ColorObj {
    name: string;
}
const c = { name: "red" } satisfies ColorObj;
;
console.log(a);
console.log(b);
console.log(c.name);
;;;
console.log(true satisfies boolean);
;
