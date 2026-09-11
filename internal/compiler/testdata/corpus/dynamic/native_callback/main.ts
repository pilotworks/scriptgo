// @dynamic
// @expect: 11
import { apply } from "./callback.js";

const increment = (value: number): number => value + 1;
const result = apply(10, increment) as number;
console.log(result);
