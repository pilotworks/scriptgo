// Object Churn Benchmark: High-volume object allocations, nested closures & circular references

class Node {
    id: number;
    name: string;
    next: Node | null = null;
    prev: Node | null = null;
    action: () => number;

    constructor(id: number, name: string) {
        this.id = id;
        this.name = name;
        this.action = () => this.id * 2;
    }
}

function run(): void {
    const iterations = 100000;
    let total = 0;

    for (let i = 0; i < iterations; i++) {
        // Create circular reference pair
        const a = new Node(i, "node_a");
        const b = new Node(i + 1, "node_b");
        a.next = b;
        b.prev = a;

        total += a.action() + b.action();
    }

    console.log("Object churn completed, iterations:", iterations, "total:", total);
}

run();
