// @expect: 3 3
// @expect: 2
// @expect: 2 1
// continue in finally overrides break; code after the try is unreachable.
let runs = 0;
let finals = 0;
do {
  try {
    runs += 1;
    break;
  } finally {
    finals += 1;
    continue;
  }
} while (runs < 3);
console.log(runs, finals);
let w = 0;
while (w < 2) {
  try {
    w += 1;
  } finally {
    continue;
  }
}
console.log(w);
let c2 = 0;
let fin2 = 0;
do {
  try {
    throw new Error("ex1");
  } catch (error) {
    c2 += 1;
    break;
  } finally {
    fin2 = 1;
    continue;
  }
  c2 += 2;
  fin2 = -1;
} while (c2 < 2);
console.log(c2, fin2);
