function evaluateTier(tier: string): string {
    let result: string = "";
    switch (tier) {
        case "bronze":
            result += "bronze-";
            // fallthrough
        case "silver":
            result += "silver-";
            // fallthrough
        case "gold":
            result += "gold-tier";
            break;
        case "platinum":
            result += "platinum-special";
            break;
        default:
            result = "unknown-tier";
            break;
    }
    return result;
}

console.log(evaluateTier("bronze"));
console.log(evaluateTier("silver"));
console.log(evaluateTier("gold"));
console.log(evaluateTier("platinum"));
console.log(evaluateTier("diamond"));

function getDaysInMonth(month: number): number {
    let days: number = 0;
    switch (month) {
        case 1:
        case 3:
        case 5:
        case 7:
        case 8:
        case 10:
        case 12:
            days = 31;
            break;
        case 4:
        case 6:
        case 9:
        case 11:
            days = 30;
            break;
        case 2:
            days = 28;
            break;
        default:
            days = -1;
            break;
    }
    return days;
}

console.log(getDaysInMonth(1));
console.log(getDaysInMonth(2));
console.log(getDaysInMonth(4));
console.log(getDaysInMonth(12));
console.log(getDaysInMonth(13));
