// @expect: A
// @expect: B
// @expect: C
// @expect: D
// @expect: F
function grade(score: number): string {
    return score >= 90 ? "A" :
           score >= 80 ? "B" :
           score >= 70 ? "C" :
           score >= 60 ? "D" : "F";
}

console.log(grade(95));
console.log(grade(85));
console.log(grade(72));
console.log(grade(65));
console.log(grade(45));
