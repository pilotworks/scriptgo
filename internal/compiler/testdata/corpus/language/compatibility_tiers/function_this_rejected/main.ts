class Box {
  v = 5;
  handler: ((this: Box, n: number) => number) | undefined;
}
const box = new Box();
box.handler = function (this: Box, n: number): number {
  return this.v + n;
};
console.log(box.handler(1));
