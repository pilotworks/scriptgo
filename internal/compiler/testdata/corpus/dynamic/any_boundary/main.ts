// @dynamic
const numberValue: any = 42;
console.log(numberValue);

const objectValue: any = { name: "Ada" };
console.log(objectValue.name);

function echo(value: any): any {
  return value;
}

console.log(echo("boxed"));
