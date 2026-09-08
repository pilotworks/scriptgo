// @dynamic
import { inspect, empty } from "./dynamic.js";

const result = inspect({ active: true, count: 2 }) as {
  nested: { active: boolean };
  values: (number | null | undefined)[];
};
console.log(result.nested.active);
console.log(result.values[0]);
console.log(result.values[1]);
console.log(result.values[2]);

const blank = empty() as { objectValue: { value?: string }; arrayValue: number[] };
console.log(blank.objectValue.value);
console.log(blank.arrayValue.length);
