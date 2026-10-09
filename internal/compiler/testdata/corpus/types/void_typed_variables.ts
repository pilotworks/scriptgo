// @expect: undefined
// @expect: undefined
// @expect: undefined
// Variables typed undefined/void hold undefined (stored boxed, never as void).
let u: undefined = undefined;
console.log(u);
function noop(): void {}
const r = noop();
console.log(r);
function local(): void {
  let w: undefined;
  console.log(w);
}
local();
