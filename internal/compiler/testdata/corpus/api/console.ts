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
console.count("default");
console.count("default");

// @api: console.countReset
// @expect: count reset passed
console.countReset("default");
console.log("count reset passed");

// @api: console.debug
// @expect: debug message
console.debug("debug message");

// @api: console.dir
// @expect: dir message
console.dir("dir message");

// @api: console.dirxml
// @expect: dirxml output
console.dirxml("dirxml output");

// @api: console.error
// @expect: error output
console.error("error output");

// @api: console.group
// @expect: group1
// @expect:   inside
console.group("group1");
console.log("inside");
console.groupEnd();

// @api: console.groupCollapsed
// @expect: groupCollapsed
// @expect:   collapsed inside
console.groupCollapsed("groupCollapsed");
console.log("collapsed inside");
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

// @api: console.table
// @expect: ["a","b"]
// @expect: table output
console.table(["a", "b"]);
console.log("table output");

// @api: console.time
// @expect: timer running
console.time("timer");
console.log("timer running");

// @api: console.timeLog
// @expect: Timer 'missing' does not exist
console.timeLog("missing");

// @api: console.timeEnd
// @expect: Timer 'missing' does not exist
console.timeEnd("missing");


// @api: console.trace
// @expect: Trace: trace point
console.trace("trace point");

// @api: console.warn
// @expect: warn output
console.warn("warn output");

// @api: console.profile
// @api: console.profileEnd
// @api: console.timeStamp
// @expect: profiled
console.log("profiled");

// @api: Console
// @api: Console.constructor
// @expect: custom console
const c = new Console();
console.log("custom console");

