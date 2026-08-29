// ScriptGo Corpus: Node.js OS Module (Strict 1:1 Parity Tests)
import * as os from "node:os";

// @api: os.arch
// @expect: true
console.log(os.arch().length > 0);

// @api: os.platform
// @expect: true
console.log(os.platform().length > 0);

// @api: os.type
// @expect: true
console.log(os.type().length > 0);

// @api: os.release
// @expect: true
console.log(os.release().length > 0);

// @api: os.version
// @expect: true
console.log(os.version().length > 0);

// @api: os.machine
// @expect: true
console.log(os.machine().length > 0);

// @api: os.homedir
// @expect: true
console.log(os.homedir().length > 0);

// @api: os.tmpdir
// @expect: true
console.log(os.tmpdir().length > 0);

// @api: os.hostname
// @expect: true
console.log(os.hostname().length > 0);

// @api: os.endianness
// @expect: true
console.log(os.endianness() === "LE" || os.endianness() === "BE");

// @api: os.uptime
// @expect: true
console.log(os.uptime() >= 0);

// @api: os.freemem
// @expect: true
console.log(os.freemem() > 0);

// @api: os.totalmem
// @expect: true
console.log(os.totalmem() > 0);

// @api: os.loadavg
// @expect: true
console.log(os.loadavg().length === 3);

// @api: os.availableParallelism
// @expect: true
console.log(os.availableParallelism() > 0);

// @api: os.cpus
// @expect: true
const cpusList = os.cpus();
console.log(cpusList.length > 0 && cpusList[0].model.length > 0);

// @api: os.networkInterfaces
// @expect: true
const netInterfaces = os.networkInterfaces();
console.log(typeof netInterfaces === "object");

// @api: os.userInfo
// @expect: true
const uInfo = os.userInfo();
console.log(uInfo.username.length > 0);

// @api: os.getPriority
// @expect: true
console.log(typeof os.getPriority() === "number");

// @api: os.setPriority
// @expect: true
os.setPriority(0, 0);
console.log(true);

// @api: os.EOL
// @expect: true
console.log(os.EOL === "\n" || os.EOL === "\r\n");

// @api: os.devNull
// @expect: true
console.log(os.devNull === "/dev/null" || os.devNull === "\\\\.\\nul");

// @api: os.constants
// @expect: true
console.log(os.constants.signals.SIGINT === 2);
