// @expect: 100
// @expect: 200
// @expect: 300
type Point3D = [x: number, y: number, z: number];

const p: Point3D = [100, 200, 300];
console.log(p[0]);
console.log(p[1]);
console.log(p[2]);
