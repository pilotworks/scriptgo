import * as process from "process";

const args: string[] = process.argv;
console.log(args.length > 0);
console.log(process.cwd().length > 0);
process.exit(0);
console.log("unreachable");
