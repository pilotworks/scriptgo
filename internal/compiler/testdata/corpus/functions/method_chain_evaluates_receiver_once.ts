// @expect: 3 2
// A method call on a closure-valued property evaluates its receiver once,
// so a call chain runs each call once.
let calls = 0;
const k = {
  n: 0,
  self() { calls++; return k; },
  bump(): number { k.n++; return k.n; },
};
k.self().bump();
k.self().self().bump();
console.log(calls, k.n);
