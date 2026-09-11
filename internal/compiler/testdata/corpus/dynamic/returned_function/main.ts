// @dynamic
// @expect: 11
import { makeIncrement } from "./factory.js";

const increment = makeIncrement() as (value: number) => number;
console.log(increment(10));
