function wrap(x: number): unknown {
    return x * 2;
}

function processUnknown(u: unknown): number {
    return (u as number) + 10;
}

const res: unknown = wrap(21);
console.log(res);
const finalNum: number = processUnknown(res);
console.log(finalNum);
