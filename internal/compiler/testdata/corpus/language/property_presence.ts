// in, Object.hasOwn, Object.keys and JSON.stringify see the keys an object
// actually has: absent optional fields and deleted keys are not present;
// methods of a class instance are found through its prototype.

type Opt = { a?: number; b: string; c?: boolean };

// @expect: false false false b {"b":"x"}
const o: Opt = { b: "x" };
const k: string = ["a", "b"][0];
console.log("a" in o, k in o, Object.hasOwn(o, "a"), Object.keys(o).join(","), JSON.stringify(o));

// @expect: false false b,c {"b":"y","c":true}
const p: Opt = { a: 1, b: "y", c: true };
delete p.a;
console.log("a" in p, Object.hasOwn(p, "a"), Object.keys(p).join(","), JSON.stringify(p));

// @expect: true true 5 b,c,a
p.a = 5;
console.log("a" in p, k in p, p.a, Object.keys(p).join(","));

// @expect: {"y":2} false 2
const d: Record<string, number> = {};
d["x"] = 1;
d["y"] = 2;
delete d["x"];
console.log(JSON.stringify(d), "x" in d, d["y"]);

// @expect: true true false
class G { v = 1; greet() { return 1; } }
const g = new G();
const method: string = "greet";
console.log(method in g, "greet" in g, Object.hasOwn(g, "greet"));
