import { getStatusText, METHODS } from "node:http";

console.log(getStatusText(200));
console.log(getStatusText(404));
console.log(getStatusText(500));
console.log(METHODS.length > 10);
console.log(METHODS[6]);
