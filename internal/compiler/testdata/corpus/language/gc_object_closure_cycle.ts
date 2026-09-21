// ScriptGo Corpus: Tracing GC Object and Closure Cycle Collection
// Verifies mark-and-sweep cycle collection for circular references
// between objects and closures, while preserving live references.

declare function gc(): void;

class Holder {
    value: number;
    action: (() => number) | null;
    partner: Holder | null;

    constructor(value: number) {
        this.value = value;
        this.action = null;
        this.partner = null;
    }
}

// 1. Create a live root holding an object with a closure capturing the object
const liveHolder = new Holder(42);
liveHolder.action = () => liveHolder.value * 2;

// 2. Allocate cyclic garbage inside an isolated function scope:
//    - Objects capturing closures that capture the objects back
//    - Closures mutually capturing each other
function allocateGarbageObjectClosureCycles(): void {
    for (let i = 0; i < 50; i++) {
        // Object -> Closure -> Object cycle
        const h = new Holder(i + 1);
        h.action = () => h.value + 10;

        // Mutual closure cycle across objects
        const h1 = new Holder(i * 2);
        const h2 = new Holder(i * 2 + 1);
        h1.partner = h2;
        h2.partner = h1;
        h1.action = () => {
            if (h2.partner !== null) {
                return h2.value;
            }
            return 0;
        };
        h2.action = () => {
            if (h1.partner !== null) {
                return h1.value;
            }
            return 0;
        };
    }
}

allocateGarbageObjectClosureCycles();

// 3. Trigger tracing garbage collection to reclaim disconnected cycles
gc();

// @expect: 84
if (liveHolder.action !== null) {
    console.log(liveHolder.action());
}

// @expect: object closure cycle collection ok
console.log("object closure cycle collection ok");
