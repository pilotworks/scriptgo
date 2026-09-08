// @dynamic
import { inspect, empty } from "./dynamic.js";

const result = inspect({ active: true, count: 2 });
console.log(result.nested.active);
console.log(result.values[0]);
console.log(result.values[1]);
console.log(result.values[2]);

const blank = empty();
console.log(blank.objectValue.value);
console.log(blank.arrayValue.length);
