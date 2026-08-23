// ScriptGo Corpus: Circular References and Cycle Collection
// Consolidated test suite with inline assertions.

declare function gc(): void;

class Node {
    id: number;
    partner: Node | null;
    constructor(id: number) {
        this.id = id;
        this.partner = null;
    }
}

// 1. Direct Circular Reference (A <-> B)
// @expect: 20
// @expect: 10
const a = new Node(10);
const b = new Node(20);
a.partner = b;
b.partner = a;

if (a.partner !== null) {
    console.log(a.partner.id);
}
if (b.partner !== null) {
    console.log(b.partner.id);
}

// 2. Loop allocation of circular references followed by GC
// @expect: cycle cleanup done
function createCycles(): void {
    for (let i = 0; i < 500; i++) {
        const n1 = new Node(i);
        const n2 = new Node(i + 1000);
        n1.partner = n2;
        n2.partner = n1;
    }
}

createCycles();
gc();
console.log("cycle cleanup done");
