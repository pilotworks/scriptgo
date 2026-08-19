const num: number = 42;
console.log(JSON.stringify(num));

const flag: boolean = true;
console.log(JSON.stringify(flag));

const text: string = "hello world";
console.log(JSON.stringify(text));

const arrNum: number[] = [1, 2, 3];
console.log(JSON.stringify(arrNum));

const arrStr: string[] = ["a", "b", "c"];
console.log(JSON.stringify(arrStr));

const parsed: string = JSON.parse("\"unquoted string\"");
console.log(parsed);
