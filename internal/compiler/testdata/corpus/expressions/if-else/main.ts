let resultStatus: string = "pending";
const score: number = 85;
if (score >= 80) {
    resultStatus = "passed";
} else {
    resultStatus = "failed";
}
console.log(resultStatus);
