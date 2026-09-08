// @dynamic
import { summarize, bump } from "./dynamic.js";

const summary = summarize({ name: "Ada", count: 2 });
console.log(summary.name);
console.log(summary.count);

const values = bump([1, 2, 3]);
console.log(values[0]);
console.log(values[2]);
