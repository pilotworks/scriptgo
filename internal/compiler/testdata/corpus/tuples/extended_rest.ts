// @expect: scores
// @expect: 80
// @expect: 90
type NamedScores = [string, ...number[]];

const s1: NamedScores = ["scores", 80, 90, 100];
console.log(s1[0]);
console.log(s1[1]);
console.log(s1[2]);
