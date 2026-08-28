// @expect: Circle: 62
// @expect: Square: 32
// @expect: Triangle: 12
type CircleShape = { kind: "circle"; radius: number };
type SquareShape = { kind: "square"; side: number };
type TriangleShape = { kind: "triangle"; base: number; height: number };

type AnyShape = CircleShape | SquareShape | TriangleShape;

function computePerimeterOrArea(s: AnyShape): number {
    switch (s.kind) {
        case "circle":
            return Math.floor(2 * 3.14159 * s.radius);
        case "square":
            return 4 * s.side;
        case "triangle":
            return Math.floor(0.5 * s.base * s.height);
        default:
            return -1;
    }
}

const c: AnyShape = { kind: "circle", radius: 10 };
const sq: AnyShape = { kind: "square", side: 8 };
const tr: AnyShape = { kind: "triangle", base: 6, height: 4 };

console.log("Circle: " + computePerimeterOrArea(c));
console.log("Square: " + computePerimeterOrArea(sq));
console.log("Triangle: " + computePerimeterOrArea(tr));
