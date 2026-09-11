// @dynamic
// @expect: 6
import { makeBox } from "./factory.js";

const box = makeBox() as { base: number; add: (value: number) => number };
console.log(box.add(2));
