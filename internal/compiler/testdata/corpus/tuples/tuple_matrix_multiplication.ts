// @expect: vRes: 17, 39
// @expect: mRes r0: 4, 4
// @expect: mRes r1: 10, 8
type Vec2 = [number, number];
type Mat2x2 = [Vec2, Vec2];

function multMatVec(m: Mat2x2, v: Vec2): Vec2 {
    const [r0, r1] = m;
    const [x, y] = v;
    return [
        r0[0] * x + r0[1] * y,
        r1[0] * x + r1[1] * y
    ];
}

function multMatMat(a: Mat2x2, b: Mat2x2): Mat2x2 {
    const [a0, a1] = a;
    const [b0, b1] = b;
    return [
        [a0[0] * b0[0] + a0[1] * b1[0], a0[0] * b0[1] + a0[1] * b1[1]],
        [a1[0] * b0[0] + a1[1] * b1[0], a1[0] * b0[1] + a1[1] * b1[1]]
    ];
}

const m1: Mat2x2 = [
    [1, 2],
    [3, 4]
];

const m2: Mat2x2 = [
    [2, 0],
    [1, 2]
];

const v: Vec2 = [5, 6];

const vRes = multMatVec(m1, v);
console.log(`vRes: ${vRes[0]}, ${vRes[1]}`);

const mRes = multMatMat(m1, m2);
console.log(`mRes r0: ${mRes[0][0]}, ${mRes[0][1]}`);
console.log(`mRes r1: ${mRes[1][0]}, ${mRes[1][1]}`);
