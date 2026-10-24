// @expect: 5 c=5
// @expect: 2 101 8 101
// @expect: x-y-z
// @expect: inner in outer outer
// @expect: id0,id1,id2
// @expect: sym!
// `this` in an object literal method is the object it is defined on: per
// instance from a factory, through arrow functions, in a returned nested
// literal with its own methods, in a loop, and under a symbol key.
const counter = {
  count: 0,
  label: "c",
  inc(by: number): number {
    this.count += by;
    return this.count;
  },
  describe(): string {
    return this.label + "=" + this.count;
  },
};
counter.inc(2);
console.log(counter.inc(3), counter.describe());
function makeCounter(start: number) {
  return {
    value: start,
    inc(): number {
      this.value++;
      return this.value;
    },
    addAll(items: number[]): number {
      items.forEach((n) => {
        this.value += n;
      });
      return this.value;
    },
  };
}
const a = makeCounter(1);
const b = makeCounter(100);
a.inc();
b.inc();
console.log(a.value, b.value, a.addAll([1, 2, 3]), b.value);

const builder = {
  parts: [] as string[],
  add(part: string) {
    this.parts.push(part);
    return this;
  },
  build(): string {
    return this.parts.join("-");
  },
};
console.log(builder.add("x").add("y").add("z").build());

const outer = {
  name: "outer",
  makeInner() {
    const outerName = this.name;
    return {
      name: "inner",
      describe(): string {
        return this.name + " in " + outerName;
      },
      viaArrow: () => this.name,
    };
  },
};
const inner = outer.makeInner();
console.log(inner.describe(), inner.viaArrow());

const objects: { id: number; show(): string }[] = [];
for (let i = 0; i < 3; i++) {
  objects.push({
    id: i,
    show(): string {
      return "id" + this.id;
    },
  });
}
console.log(objects.map((o) => o.show()).join(","));

const tag = Symbol("tag");
const tagged = {
  label: "sym",
  [tag]() {
    return this.label + "!";
  },
};
console.log(tagged[tag]());
