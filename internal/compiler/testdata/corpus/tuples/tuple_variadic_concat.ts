// @expect: Swapped: 42, alpha
// @expect: Triple(alpha, 42, true)
// @expect: Rest tuple len: 4, first: 3
type Pair = [string, number];
type Triple = [string, number, boolean];

function swapPair(p: Pair): [number, string] {
    const [str, num] = p;
    return [num, str];
}

function expandPair(p: Pair, flag: boolean): Triple {
    return [p[0], p[1], flag];
}

function summarizeTuple(t: Triple): string {
    return "Triple(" + t[0] + ", " + t[1] + ", " + t[2] + ")";
}

const p1: Pair = ["alpha", 42];
const swapped = swapPair(p1);
console.log("Swapped: " + swapped[0] + ", " + swapped[1]);

const t1 = expandPair(p1, true);
console.log(summarizeTuple(t1));

const t2: [number, ...string[]] = [3, "apple", "banana", "cherry"];
console.log("Rest tuple len: " + t2.length + ", first: " + t2[0]);
