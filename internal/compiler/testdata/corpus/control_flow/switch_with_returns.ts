// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 4
// @expect: 0
function evaluateLevel(level: string): number {
    switch (level) {
        case "low":
            return 1;
        case "medium":
            return 2;
        case "high":
            return 3;
        case "critical":
            return 4;
        default:
            return 0;
    }
}

console.log(evaluateLevel("low"));
console.log(evaluateLevel("medium"));
console.log(evaluateLevel("high"));
console.log(evaluateLevel("critical"));
console.log(evaluateLevel("unknown"));
