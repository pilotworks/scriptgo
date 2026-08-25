// @expect: 78
// @expect: 16
// @expect: 18
type Circle = { kind: "circle"; radius: number };
type Square = { kind: "square"; size: number };
type Rectangle = { kind: "rectangle"; width: number; height: number };

type Shape = Circle | Square | Rectangle;

function area(shape: Shape): number {
    switch (shape.kind) {
        case "circle":
            return Math.floor(3.14159 * shape.radius * shape.radius);
        case "square":
            return shape.size * shape.size;
        case "rectangle":
            return shape.width * shape.height;
    }
}

console.log(area({ kind: "circle", radius: 5 }));
console.log(area({ kind: "square", size: 4 }));
console.log(area({ kind: "rectangle", width: 3, height: 6 }));
