// @expect: 0
// @expect: 
// @expect: false
// @expect: null
// @expect: 100

const obj = {
  a: 0,
  b: "",
  c: false,
  d: null as string | null,
  e: undefined as number | undefined,
};

const {
  a = 10,
  b = "fallback",
  c = true,
  d = "fallback_d",
  e = 100,
} = obj;

console.log(a);
console.log(b);
console.log(c);
console.log(d);
console.log(e);
