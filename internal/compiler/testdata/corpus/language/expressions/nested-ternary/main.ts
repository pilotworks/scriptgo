function classifyScore(score: number): string {
    return score >= 90 ? "A" :
           score >= 80 ? "B" :
           score >= 70 ? "C" :
           score >= 60 ? "D" : "F";
}

console.log(classifyScore(95));
console.log(classifyScore(85));
console.log(classifyScore(75));
console.log(classifyScore(65));
console.log(classifyScore(50));

function clamp(val: number, min: number, max: number): number {
    return val < min ? min : val > max ? max : val;
}

console.log(clamp(5, 10, 20));
console.log(clamp(15, 10, 20));
console.log(clamp(25, 10, 20));
