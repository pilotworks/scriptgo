// ES2024 Set Operations Benchmark: 2,500 element union, intersection, and difference

function run(): void {
    const n = 2500;
    const setA = new Set<number>();
    const setB = new Set<number>();

    for (let i = 0; i < n; i++) {
        setA.add(i);
        setB.add(i + n / 2);
    }

    // ES2024 Set methods
    const intersection = setA.intersection(setB);
    const union = setA.union(setB);
    const difference = setA.difference(setB);

    console.log("Set operations completed, sizes:", intersection.size, union.size, difference.size);
}

run();
