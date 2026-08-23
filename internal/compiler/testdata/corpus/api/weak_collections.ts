// ScriptGo Corpus: Weak Collections (WeakMap, WeakSet, WeakRef) & GC
// Consolidated test suite with inline assertions.

declare function gc(): void;

// 1. WeakRef
// @expect: 42
class DataHolder {
    val: number;
    constructor(val: number) {
        this.val = val;
    }
}

const holder = new DataHolder(42);
const wref = new WeakRef(holder);
const derefed = wref.deref();
if (derefed) {
    console.log(derefed.val);
} else {
    console.log("deref failed");
}

// 2. WeakMap
// @expect: true
// @expect: 999
// @expect: true
const wm = new WeakMap<DataHolder, number>();
wm.set(holder, 999);
console.log(wm.has(holder));
console.log(wm.get(holder));
console.log(wm.delete(holder));

// 3. WeakSet
// @expect: true
// @expect: true
const ws = new WeakSet<DataHolder>();
ws.add(holder);
console.log(ws.has(holder));
console.log(ws.delete(holder));

// 4. GC invocation
// @expect: gc ok
gc();
console.log("gc ok");
