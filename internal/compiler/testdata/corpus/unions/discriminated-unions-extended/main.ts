interface Square {
  kind: "square";
  size: number;
}

interface Rectangle {
  kind: "rectangle";
  width: number;
  height: number;
}

interface Circle {
  kind: "circle";
  radius: number;
}

type Shape = Square | Rectangle | Circle;

function area(s: Shape): number {
  if (s.kind === "square") {
    const sq = s as Square;
    return sq.size * sq.size;
  }
  if (s.kind === "rectangle") {
    const rect = s as Rectangle;
    return rect.width * rect.height;
  }
  const circ = s as Circle;
  return 3.14 * circ.radius * circ.radius;
}

const s1: Square = { kind: "square", size: 5 };
const s2: Rectangle = { kind: "rectangle", width: 4, height: 6 };
const s3: Circle = { kind: "circle", radius: 10 };

console.log(area(s1));
console.log(area(s2));
console.log(area(s3));
