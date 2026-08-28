// @expect: A:circle:10
// @expect: B:square:20
// 1. In-operator property narrowing on disjoint object shapes
type ShapeA = { x: number; name: string };
type ShapeB = { y: number; title: string };

function inspectShape(item: ShapeA | ShapeB): string {
    if ("x" in item) {
        return "A:" + item.name + ":" + item.x;
    }
    return "B:" + item.title + ":" + item.y;
}

console.log(inspectShape({ x: 10, name: "circle" }));
console.log(inspectShape({ y: 20, title: "square" }));
