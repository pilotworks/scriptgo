import * as http from "node:http";
console.log(http.getStatusText(200));
console.log(http.getStatusText(404));
