import {
    AssertionError,
    SystemError,
    CustomError,
    env
} from "node:errors";

// @api: errors.errors.Error
// @api: errors.Error
// @api: new errors.Error
// @expect: err_err_inst: true
const err = new Error("msg", { cause: "root cause" });
console.log("err_err_inst: " + (err instanceof Error));

// @api: Error.message
// @expect: err_msg: msg
console.log("err_msg: " + err.message);

// @api: Error.cause
// @expect: err_cause: root cause
console.log("err_cause: " + (err as Error & { cause?: string }).cause);

// @api: Error.stack
// @expect: err_stack: true
console.log("err_stack: " + (typeof err.stack === "string"));

// @api: Error.code
// @expect: err_code: undefined
console.log("err_code: " + (err as Error & { code?: string }).code);

// @api: Error.captureStackTrace
// @expect: err_capStack: true
Error.captureStackTrace(err);
console.log("err_capStack: true");

// @api: Error.stackTraceLimit
// @expect: err_stackLimit: 10
console.log("err_stackLimit: " + Error.stackTraceLimit);

// @api: errors.errors.AssertionError
// @api: errors.AssertionError
// @api: new errors.AssertionError
// @expect: err_assert_inst: true
const assertErr = new AssertionError();
console.log("err_assert_inst: " + (assertErr instanceof AssertionError));

// @api: errors.errors.RangeError
// @api: errors.RangeError
// @api: new errors.RangeError
// @expect: err_range_inst: true
const rangeErr = new RangeError("out of range");
console.log("err_range_inst: " + (rangeErr instanceof RangeError));

// @api: errors.errors.ReferenceError
// @api: errors.ReferenceError
// @api: new errors.ReferenceError
// @expect: err_ref_inst: true
const refErr = new ReferenceError("not defined");
console.log("err_ref_inst: " + (refErr instanceof ReferenceError));

// @api: errors.errors.SyntaxError
// @api: errors.SyntaxError
// @api: new errors.SyntaxError
// @expect: err_syntax_inst: true
const synErr = new SyntaxError("unexpected token");
console.log("err_syntax_inst: " + (synErr instanceof SyntaxError));

// @api: errors.errors.TypeError
// @api: errors.TypeError
// @api: new errors.TypeError
// @expect: err_type_inst: true
const typErr = new TypeError("invalid type");
console.log("err_type_inst: " + (typErr instanceof TypeError));

// @api: errors.errors.SystemError
// @api: errors.SystemError
// @api: new errors.SystemError
// @expect: err_sys_inst: true
const sysErr = new SystemError("system failure");
console.log("err_sys_inst: " + (sysErr instanceof SystemError));

// @api: SystemError.code
// @expect: err_sys_code: ENOENT
sysErr.code = "ENOENT";
console.log("err_sys_code: " + sysErr.code);

// @api: SystemError.syscall
// @expect: err_sys_syscall: open
sysErr.syscall = "open";
console.log("err_sys_syscall: " + sysErr.syscall);

// @api: SystemError.errno
// @expect: err_sys_errno: -2
sysErr.errno = -2;
console.log("err_sys_errno: " + sysErr.errno);

// @api: SystemError.path
// @expect: err_sys_path: /tmp/file
sysErr.path = "/tmp/file";
console.log("err_sys_path: " + sysErr.path);

// @api: SystemError.dest
// @expect: err_sys_dest: /tmp/dest
sysErr.dest = "/tmp/dest";
console.log("err_sys_dest: " + sysErr.dest);

// @api: SystemError.address
// @expect: err_sys_addr: 127.0.0.1
sysErr.address = "127.0.0.1";
console.log("err_sys_addr: " + sysErr.address);

// @api: SystemError.port
// @expect: err_sys_port: 8080
sysErr.port = 8080;
console.log("err_sys_port: " + sysErr.port);

// @api: SystemError.info
// @expect: err_sys_info: true
sysErr.info = {};
console.log("err_sys_info: " + (typeof sysErr.info === "object"));

// @api: SystemError.message
// @expect: err_sys_msg: system failure
console.log("err_sys_msg: " + sysErr.message);

// @api: environment_variables.env
// @expect: env_type: object
console.log("env_type: " + typeof env);
