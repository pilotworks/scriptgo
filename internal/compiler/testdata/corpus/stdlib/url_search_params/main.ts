import { URLSearchParams } from "node:url";

const p: URLSearchParams = new URLSearchParams("foo=1&bar=2&foo=3");
console.log(p.get("foo"));
console.log(p.get("bar"));
console.log(p.get("baz"));

const fooAll: string[] = p.getAll("foo");
console.log(fooAll.join(","));

console.log(p.has("bar"));
console.log(p.has("baz"));

p.append("baz", "4");
console.log(p.get("baz"));

p.set("foo", "100");
console.log(p.get("foo"));
console.log(p.getAll("foo").length);

p.delete("bar");
console.log(p.has("bar"));

p.sort();
console.log(p.toString());
console.log(p.size);
