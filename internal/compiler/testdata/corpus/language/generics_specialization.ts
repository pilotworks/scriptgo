// ScriptGo Corpus: Language: Generics Specialization and Advanced Patterns
// Consolidated test suite with inline assertions.

// @expect: Alice
// @expect: 42
// @expect: Bob
// @expect: 99
// @expect: true
// @expect: false
// @expect: iife-result-100
// @expect: opt-default-0
// @expect: opt-custom-50

interface Entry<V> {
  key: string;
  value: V;
}

class Node<V> {
  public prev: Node<V> | null = null;
  public next: Node<V> | null = null;

  constructor(
    public key: string,
    public entry: Entry<V>
  ) {}
}

interface User {
  name: string;
  age: number;
}

const e1: Entry<User> = { key: "u1", value: { name: "Alice", age: 42 } };
const n1 = new Node<User>("u1", e1);

console.log(n1.entry.value.name);
console.log(n1.entry.value.age);

n1.entry.value.name = "Bob";
n1.entry.value.age = 99;
console.log(n1.entry.value.name);
console.log(n1.entry.value.age);

// Test unary ! on nullable object
console.log(!n1.prev);
console.log(!n1);

// Test arrow IIFE
const iifeRes = (() => {
  return "iife-result-100";
})();
console.log(iifeRes);

// Test optional parameter default
function testOptParam(name: string, count?: number): void {
  const c = count ?? 0;
  console.log(name + "-" + c);
}

testOptParam("opt-default");
testOptParam("opt-custom", 50);
