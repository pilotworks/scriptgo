// Object-valued field initializers run for any class (no name-based cases).
// @expect: 1 false n 3
class Node2 { kids: Map<string, Node2> = new Map(); end: boolean = false; tag = "n"; }
class Tree { root: Node2 = new Node2(); count = 3; }
const t = new Tree();
t.root.kids.set("a", new Node2());
console.log(t.root.kids.size, t.root.end, t.root.tag, t.count);
