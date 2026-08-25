// @expect: 10
// @expect: 20
// @expect: 0
// @expect: 5
// @expect: 15
// @expect: 25
type Point = {
    x: number;
    y: number;
    z?: number;
};

function printCoordinates({ x, y, z = 0 }: Point): void {
    console.log(x);
    console.log(y);
    console.log(z);
}

printCoordinates({ x: 10, y: 20 });
printCoordinates({ x: 5, y: 15, z: 25 });
