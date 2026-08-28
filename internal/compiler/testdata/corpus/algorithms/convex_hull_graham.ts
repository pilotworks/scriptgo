// @expect: 4
// @expect: 0,0
// @expect: 3,0
// @expect: 3,3
// @expect: 0,3
interface Point {
    x: number;
    y: number;
}

function orientation(p: Point, q: Point, r: Point): number {
    const val = (q.y - p.y) * (r.x - q.x) - (q.x - p.x) * (r.y - q.y);
    if (val === 0) return 0;
    return val > 0 ? 1 : 2;
}

function distSq(p1: Point, p2: Point): number {
    return (p1.x - p2.x) * (p1.x - p2.x) + (p1.y - p2.y) * (p1.y - p2.y);
}

function convexHull(points: Point[]): Point[] {
    const n = points.length;
    if (n < 3) return points;

    let ymin = points[0].y;
    let minIdx = 0;
    for (let i = 1; i < n; i++) {
        const y = points[i].y;
        if (y < ymin || (ymin === y && points[i].x < points[minIdx].x)) {
            ymin = points[i].y;
            minIdx = i;
        }
    }

    const temp = points[0];
    points[0] = points[minIdx];
    points[minIdx] = temp;

    const p0 = points[0];

    const sorted: Point[] = [];
    for (let i = 1; i < n; i++) {
        sorted.push(points[i]);
    }

    sorted.sort((p1: Point, p2: Point) => {
        const o = orientation(p0, p1, p2);
        if (o === 0) {
            return distSq(p0, p2) >= distSq(p0, p1) ? -1 : 1;
        }
        return o === 2 ? -1 : 1;
    });

    const stack: Point[] = [p0, sorted[0]];

    for (let i = 1; i < sorted.length; i++) {
        while (stack.length > 1 && orientation(stack[stack.length - 2], stack[stack.length - 1], sorted[i]) !== 2) {
            stack.pop();
        }
        stack.push(sorted[i]);
    }

    return stack;
}

const pts: Point[] = [
    { x: 0, y: 3 },
    { x: 2, y: 2 },
    { x: 1, y: 1 },
    { x: 2, y: 1 },
    { x: 3, y: 0 },
    { x: 0, y: 0 },
    { x: 3, y: 3 }
];

const hull = convexHull(pts);
console.log(hull.length);
for (const p of hull) {
    console.log(`${p.x},${p.y}`);
}
