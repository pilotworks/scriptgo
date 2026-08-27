// @expect: 32
// @expect: [5, 7, 9]
// @expect: 3
type Vector3 = [number, number, number];

function dotProduct(v1: Vector3, v2: Vector3): number {
    return v1[0] * v2[0] + v1[1] * v2[1] + v1[2] * v2[2];
}

function addVectors(v1: Vector3, v2: Vector3): Vector3 {
    return [v1[0] + v2[0], v1[1] + v2[1], v1[2] + v2[2]];
}

function manhattanDistance(v1: Vector3, v2: Vector3): number {
    return Math.abs(v1[0] - v2[0]) + Math.abs(v1[1] - v2[1]) + Math.abs(v1[2] - v2[2]);
}

const u: Vector3 = [1, 2, 3];
const v: Vector3 = [4, 5, 6];

console.log(dotProduct(u, v));
const sum = addVectors(u, v);
console.log("[" + sum[0] + ", " + sum[1] + ", " + sum[2] + "]");
console.log(manhattanDistance(u, [2, 3, 4]));
