// @dynamic
import { createCalc } from "./math.js";

const calc: any = createCalc(3);
console.log(calc.greet("ScriptGo"));
console.log(calc.scaleAndSum(1, 2, 3, 4, 5, 6));
