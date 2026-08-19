import * as os from "node:os";

console.log(os.platform().length > 0);
console.log(os.arch().length > 0);
console.log(os.homedir().length > 0);
