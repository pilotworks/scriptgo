// ScriptGo Corpus: Tracing GC Cycle Collection
// Verifies mark-and-sweep reclamation of disconnected reference cycles
// while preserving surviving referenced subgraphs.

declare function gc(): number;

class Node {
    id: number;
    partner: Node | null;
    constructor(id: number) {
        this.id = id;
        this.partner = null;
    }
}

// 1. Allocate a live subgraph that remains reachable from stack
const liveRoot = new Node(100);
const livePartner = new Node(200);
liveRoot.partner = livePartner;
livePartner.partner = liveRoot;

// 2. Allocate cyclic garbage inside an isolated function scope
function allocateGarbageCycles(): void {
    for (let i = 0; i < 50; i++) {
        const a = new Node(i);
        const b = new Node(i + 1000);
        a.partner = b;
        b.partner = a;
    }
}

allocateGarbageCycles();

// 3. Trigger tracing garbage collection
// @expect: collected cycles successfully
const collected = gc();
if (collected > 0) {
    console.log("collected cycles successfully");
} else {
    console.log("gc returned: " + collected);
}

// 4. Verify live roots are intact and traversed correctly
// @expect: 200
// @expect: 100
if (liveRoot.partner !== null) {
    console.log(liveRoot.partner.id);
    if (liveRoot.partner.partner !== null) {
        console.log(liveRoot.partner.partner.id);
    }
}
