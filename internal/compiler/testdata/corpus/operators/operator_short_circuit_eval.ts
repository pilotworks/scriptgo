// @expect: --- AND short-circuit ---
// @expect: EFFECT: A1
// @expect: Result a: false, sideEffects: 1
// @expect: --- OR short-circuit ---
// @expect: EFFECT: B1
// @expect: Result b: true, sideEffects: 2
// @expect: --- Nullish coalescing short-circuit ---
// @expect: EFFECT_VAL: C1
// @expect: Result c: existing, sideEffects: 3
// @expect: EFFECT_VAL: D1
// @expect: EFFECT_VAL: D2
// @expect: Result d: fallback_used, sideEffects: 5
let sideEffects = 0;

function effect(val: boolean, tag: string): boolean {
    sideEffects++;
    console.log(`EFFECT: ${tag}`);
    return val;
}

console.log("--- AND short-circuit ---");
const a = effect(false, "A1") && effect(true, "A2");
console.log(`Result a: ${a}, sideEffects: ${sideEffects}`);

console.log("--- OR short-circuit ---");
const b = effect(true, "B1") || effect(false, "B2");
console.log(`Result b: ${b}, sideEffects: ${sideEffects}`);

console.log("--- Nullish coalescing short-circuit ---");
function effectVal<T>(val: T, tag: string): T {
    sideEffects++;
    console.log(`EFFECT_VAL: ${tag}`);
    return val;
}

const c = effectVal("existing", "C1") ?? effectVal("fallback", "C2");
console.log(`Result c: ${c}, sideEffects: ${sideEffects}`);

const d = effectVal(null, "D1") ?? effectVal("fallback_used", "D2");
console.log(`Result d: ${d}, sideEffects: ${sideEffects}`);
