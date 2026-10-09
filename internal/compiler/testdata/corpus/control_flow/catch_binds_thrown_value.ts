// @expect: string plain
// @expect: number 42
// @expect: boolean false
// @expect: object true r
// @expect: x
// @expect: true TypeError Reduce of empty array with no initial value
// @expect: true TypeError
// catch (e) is unknown under strict TypeScript and binds exactly the thrown
// value; runtime-detected failures are thrown as built-in Error instances.
try { throw "plain"; } catch (e) { console.log(typeof e, e); }
try { throw 42; } catch (e) { console.log(typeof e, e); }
try { throw false; } catch (e) { console.log(typeof e, e); }
try { throw new RangeError("r"); } catch (e) { console.log(typeof e, e instanceof RangeError, (e as Error).message); }
try { throw new Error("x"); } catch (e: unknown) { if (e instanceof Error) console.log(e.message); }
try {
  const empty: number[] = [];
  empty.reduce((a, b) => a + b);
} catch (e) {
  console.log(e instanceof TypeError, (e as Error).name, (e as Error).message);
}
// Native difference: `as` on an unknown is a checked cast (SG4002) that throws.
const boxed: unknown = "s";
try {
  const n = boxed as number[];
  console.log(n.length);
} catch (e) {
  console.log(e instanceof TypeError, (e as Error).name);
}
