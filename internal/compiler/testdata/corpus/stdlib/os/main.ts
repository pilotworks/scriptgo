import * as os from "os";

console.log(os.platform().length > 0);
console.log(os.arch().length > 0);
console.log(os.homedir().length > 0);
console.log(os.uptime() > 0);
console.log(os.totalmem() > 0);
console.log(os.freemem() > 0);
console.log(os.type().length > 0);
console.log(os.release().length > 0);
