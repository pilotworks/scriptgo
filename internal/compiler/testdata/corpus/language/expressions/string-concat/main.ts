const greeting: string = 'hello' + ', ' + 'world';
console.log(greeting);

const numConcat1: string = "count: " + 42;
const numConcat2: string = 100 + " ms";
console.log(numConcat1);
console.log(numConcat2);

const boolConcat1: string = "isReady: " + true;
const boolConcat2: string = false + " is false";
console.log(boolConcat1);
console.log(boolConcat2);

let compoundStr: string = "total: ";
compoundStr += 50;
console.log(compoundStr);
