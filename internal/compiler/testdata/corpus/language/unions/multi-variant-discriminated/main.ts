interface Circle {
    kind: "circle";
    radius: number;
}

interface Rectangle {
    kind: "rectangle";
    width: number;
    height: number;
}

interface Square {
    kind: "square";
    size: number;
}

type Shape = Circle | Rectangle | Square;

function calculateArea(s: Shape): number {
    if (s.kind === "circle") {
        return Math.floor(Math.PI * s.radius * s.radius);
    }
    if (s.kind === "rectangle") {
        return s.width * s.height;
    }
    if (s.kind === "square") {
        return s.size * s.size;
    }
    return 0;
}

function describeShape(s: Shape): string {
    switch (s.kind) {
        case "circle":
            return `Circle(r=${s.radius})`;
        case "rectangle":
            return `Rectangle(w=${s.width}, h=${s.height})`;
        case "square":
            return `Square(s=${s.size})`;
    }
}

const c: Shape = { kind: "circle", radius: 5 };
const r: Shape = { kind: "rectangle", width: 10, height: 4 };
const sq: Shape = { kind: "square", size: 6 };

console.log(describeShape(c));
console.log(calculateArea(c));

console.log(describeShape(r));
console.log(calculateArea(r));

console.log(describeShape(sq));
console.log(calculateArea(sq));
