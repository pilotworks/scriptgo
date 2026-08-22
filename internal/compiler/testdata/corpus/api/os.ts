// ScriptGo Corpus: Os Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as os from "node:os";

// @api: os.arch
// @expect: string
console.log(typeof os.arch());

// @api: os.freemem
// @expect: true
console.log(os.freemem() > 0);

// @api: os.homedir
// @expect: string
console.log(typeof os.homedir());

// @api: os.platform
// @expect: string
console.log(typeof os.platform());

// @api: os.release
// @expect: true
console.log(os.release().length > 0);

// @api: os.tmpdir
// @expect: string
console.log(typeof os.tmpdir());

// @api: os.totalmem
// @expect: true
console.log(os.totalmem() > 0);

// @api: os.type
// @expect: true
console.log(os.type().length > 0);

// @api: os.uptime
// @expect: true
console.log(os.uptime() >= 0);
