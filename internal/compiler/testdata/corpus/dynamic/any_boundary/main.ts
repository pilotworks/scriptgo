// @dynamic
const numberValue: any = 42;
console.log(numberValue);

const objectValue: any = { name: "Ada" };
console.log(objectValue.name);

const arrayValue: any = [1, 2, 3];
console.log(arrayValue[1]);

function echo(value: any): any {
  return value;
}

console.log(echo("boxed"));
