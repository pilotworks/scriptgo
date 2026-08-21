const m = new Map<string, string>([["a", "1"], ["b", "2"]]);
m.set("c", "3").set("d", "4");
console.log(m.size);
console.log(m.get("c"));
console.log(m.get("d"));
console.log(m);

const s = new Set<string>(["alpha", "beta", "alpha"]);
s.add("gamma").add("delta");
console.log(s.size);
console.log(s.has("beta"));
console.log(s.has("omega"));
console.log(s);
