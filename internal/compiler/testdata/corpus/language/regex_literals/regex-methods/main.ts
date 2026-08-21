const re = /([0-9]+)/;
const text = "order 12345 placed";

const match = text.match(re)!;
console.log(match.length);
console.log(match[0]);
console.log(match[1]);

const execRes = re.exec("test 999 done")!;
console.log(execRes.length);
console.log(execRes[0]);
console.log(execRes[1]);
