// ScriptGo Corpus: Console Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: console.assert
// @expect: assert passed
console.assert(true, "should not print");
console.log("assert passed");

// @api: console.clear
// @expect: [2J[0fcleared
console.clear();
console.log("cleared");

// @api: console.count
// @expect: default: 1
// @expect: default: 2
// @expect: count finished
console.count("default");
console.count("default");
console.countReset("default");
console.log("count finished");

// @api: console.debug
// @expect: debug message
console.debug("debug message");

// @api: console.dir
// @expect: dir message
console.dir("dir message");

// @api: console.error
// @expect: error output
console.error("error output");

// @api: console.group
// @expect: group1
// @expect:   inside
console.group("group1");
console.log("inside");
console.groupEnd();

// @api: console.groupEnd
// @expect: g
// @expect: ended
console.group("g");
console.groupEnd();
console.log("ended");

// @api: console.info
// @expect: info output
console.info("info output");

// @api: console.log
// @expect: log output
console.log("log output");

// @api: console.time
// @expect: timer running
console.time("timer");
console.log("timer running");

// @api: console.trace
// @expect: Trace: trace point
console.trace("trace point");

// @api: console.warn
// @expect: warn output
console.warn("warn output");
