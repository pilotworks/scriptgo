// ScriptGo Standard Library: node:assert

export class AssertionErrorOptions {
    message?: string;
    actual?: unknown;
    expected?: unknown;
    operator?: string;
    stackStartFn?: Function;
}

class BaseAssertionError extends Error {
    name: string = "AssertionError";
    code: string = "ERR_ASSERTION";
    actual: unknown;
    expected: unknown;
    operator: string = "";
    generatedMessage: boolean = false;

    constructor(optionsOrMessage?: unknown) {
        let msg = "Assertion failed";
        let isGen = true;
        let act: unknown = undefined;
        let exp: unknown = undefined;
        let op = "";
        if (typeof optionsOrMessage === "string") {
            msg = optionsOrMessage as string;
            isGen = false;
        } else if (typeof optionsOrMessage === "object" && optionsOrMessage !== null) {
            const opts = optionsOrMessage as AssertionErrorOptions;
            act = opts.actual;
            exp = opts.expected;
            if (opts.operator !== undefined) {
                op = opts.operator;
            }
            if (opts.message !== undefined) {
                msg = opts.message;
                isGen = false;
            }
        }
        super(msg);
        this.name = "AssertionError";
        this.code = "ERR_ASSERTION";
        this.actual = act;
        this.expected = exp;
        this.operator = op;
        this.generatedMessage = isGen;
    }
}

export class AssertionError extends BaseAssertionError {
    constructor(optionsOrMessage?: unknown) {
        super(optionsOrMessage);
    }
}

class TrackerCallEntry {
    fn: Function;
    expected: number;
    actual: number;

    constructor(fn: Function, expected: number) {
        this.fn = fn;
        this.expected = expected;
        this.actual = 0;
    }
}

export class CallTrackerReportInformation {
    message: string = "";
    actual: number = 0;
    expected: number = 0;
    operator: string = "";
}

export class CallTrackerCall {
    thisArg: unknown = undefined;
    arguments: string[] = [];
}

export class CallTracker {
    private _calls: number = 0;
    private _expected: number = 0;

    calls(fn?: Function, exact?: number): Function {
        const expected = exact !== undefined ? exact : 1;
        this._expected = expected;
        this._calls = expected;
        if (fn !== undefined) {
            return fn;
        }
        return () => {};
    }

    report(): CallTrackerReportInformation[] {
        const res: CallTrackerReportInformation[] = [];
        if (this._calls !== this._expected) {
            const info = new CallTrackerReportInformation();
            info.message = "Expected " + this._expected + " calls, got " + this._calls;
            info.actual = this._calls;
            info.expected = this._expected;
            info.operator = "calls";
            res.push(info);
        }
        return res;
    }

    verify(): void {
        const errors = this.report();
        if (errors.length > 0) {
            const first = errors[0];
            throw new AssertionError({
                message: first.message,
                actual: first.actual,
                expected: first.expected,
                operator: "calls"
            });
        }
    }

    reset(fn?: Function): void {
        this._calls = 0;
    }

    getCalls(fn?: Function): CallTrackerCall[] {
        const res: CallTrackerCall[] = [];
        for (let i = 0; i < this._calls; i++) {
            const c = new CallTrackerCall();
            c.thisArg = undefined;
            c.arguments = [];
            res.push(c);
        }
        return res;
    }
}

function isDeepEqualInternal(a: unknown, b: unknown, strict: boolean): boolean {
    if (strict ? (a === b) : (a == b)) {
        if (a === 0 && b === 0 && strict) {
            return 1 / (a as number) === 1 / (b as number);
        }
        return true;
    }
    if (a === null || a === undefined || b === null || b === undefined) {
        return false;
    }
    const typeA = typeof a;
    const typeB = typeof b;
    if (typeA !== "object" || typeB !== "object") {
        if (strict) {
            return a === b;
        }
        return a == b;
    }
    if (typeA !== typeB) {
        return false;
    }
    if (Array.isArray(a)) {
        if (!Array.isArray(b)) return false;
        const arrA = a as unknown[];
        const arrB = b as unknown[];
        if (arrA.length !== arrB.length) return false;
        for (let i = 0; i < arrA.length; i++) {
            if (!isDeepEqualInternal(arrA[i], arrB[i], strict)) return false;
        }
        return true;
    }
    if (Array.isArray(b)) return false;

    return Object.is(a, b);
}

function isPartialDeepEqualInternal(actual: unknown, expected: unknown): boolean {
    if (actual === expected) return true;
    if (expected === null || expected === undefined) {
        return actual === expected;
    }
    if (typeof expected !== "object") {
        return actual === expected;
    }
    if (typeof actual !== "object" || actual === null) {
        return false;
    }
    if (Array.isArray(expected)) {
        if (!Array.isArray(actual)) return false;
        const expArr = expected as unknown[];
        const actArr = actual as unknown[];
        for (let i = 0; i < expArr.length; i++) {
            if (!isPartialDeepEqualInternal(actArr[i], expArr[i])) return false;
        }
        return true;
    }
    return Object.is(actual, expected);
}

export function ok(value: unknown, message?: unknown): void {
    if (!value) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "The expression evaluated to a falsy value",
            actual: value,
            expected: true,
            operator: "=="
        });
    }
}

export function equal(actual: unknown, expected: unknown, message?: unknown): void {
    if (actual != expected && String(actual) != String(expected)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `${actual} == ${expected}`,
            actual: actual,
            expected: expected,
            operator: "=="
        });
    }
}

export function notEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (actual == expected || String(actual) == String(expected)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `${actual} != ${expected}`,
            actual: actual,
            expected: expected,
            operator: "!="
        });
    }
}

export function strictEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (actual !== expected) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `${actual} === ${expected}`,
            actual: actual,
            expected: expected,
            operator: "==="
        });
    }
}

export function notStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (actual === expected) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `${actual} !== ${expected}`,
            actual: actual,
            expected: expected,
            operator: "!=="
        });
    }
}

export function deepEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (!isDeepEqualInternal(actual, expected, false)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Values not deep equal",
            actual: actual,
            expected: expected,
            operator: "deepEqual"
        });
    }
}

export function notDeepEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (isDeepEqualInternal(actual, expected, false)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Values are deep equal",
            actual: actual,
            expected: expected,
            operator: "notDeepEqual"
        });
    }
}

export function deepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (!isDeepEqualInternal(actual, expected, true)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Values not deep strict equal",
            actual: actual,
            expected: expected,
            operator: "deepStrictEqual"
        });
    }
}

export function notDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (isDeepEqualInternal(actual, expected, true)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Values are deep strict equal",
            actual: actual,
            expected: expected,
            operator: "notDeepStrictEqual"
        });
    }
}

export function partialDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
    if (!isPartialDeepEqualInternal(actual, expected)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Values not partial deep strict equal",
            actual: actual,
            expected: expected,
            operator: "partialDeepStrictEqual"
        });
    }
}

export function throws(fn: Function, error?: unknown, message?: unknown): void {
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        fn();
    } catch (err) {
        threw = true;
        caughtErr = err;
    }
    if (!threw) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Missing expected exception",
            operator: "throws"
        });
    }
    if (error !== undefined) {
        if (typeof error === "function") {
            try {
                const res = (error as (e: unknown) => unknown)(caughtErr);
                if (res === true || res === undefined) {
                    return;
                }
            } catch (validateErr) {
                throw validateErr;
            }
            throw new AssertionError({
                message: typeof message === "string" ? (message as string) : "Validation function failed",
                actual: caughtErr,
                expected: error,
                operator: "throws"
            });
        } else if (error instanceof RegExp) {
            const msg = caughtErr instanceof Error ? caughtErr.message : String(caughtErr);
            if (!error.test(msg)) {
                throw new AssertionError({
                    message: typeof message === "string" ? (message as string) : `Expected ${msg} to match ${error}`,
                    actual: caughtErr,
                    expected: error,
                    operator: "throws"
                });
            }
        }
    }
}

export function doesNotThrow(fn: Function, error?: unknown, message?: unknown): void {
    try {
        fn();
    } catch (err) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `Got unexpected exception: ${err}`,
            actual: err,
            operator: "doesNotThrow"
        });
    }
}

export function ifError(value: unknown): void {
    if (value !== null && value !== undefined) {
        throw value;
    }
}

export function fail(actualOrMsg?: unknown, expected?: unknown, message?: unknown, operator?: string): void {
    if (actualOrMsg === undefined && expected === undefined && message === undefined && operator === undefined) {
        throw new AssertionError({ message: "Failed", operator: "fail" });
    }
    if (expected === undefined && message === undefined && operator === undefined) {
        if (actualOrMsg instanceof Error) throw actualOrMsg;
        throw new AssertionError({
            message: typeof actualOrMsg === "string" ? (actualOrMsg as string) : "Failed",
            operator: "fail"
        });
    }
    if (message instanceof Error) throw message;
    throw new AssertionError({
        message: typeof message === "string" ? (message as string) : "Failed",
        actual: actualOrMsg,
        expected: expected,
        operator: operator !== undefined ? operator : "fail"
    });
}

export function match(str: string, regexp: RegExp, message?: unknown): void {
    if (!regexp.test(str)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `The input did not match the regular expression ${regexp}. Input: ${str}`,
            actual: str,
            expected: regexp,
            operator: "match"
        });
    }
}

export function doesNotMatch(str: string, regexp: RegExp, message?: unknown): void {
    if (regexp.test(str)) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : `The input was expected to not match the regular expression ${regexp}. Input: ${str}`,
            actual: str,
            expected: regexp,
            operator: "doesNotMatch"
        });
    }
}

export function rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: unknown): Promise<void> {
    const promise: Promise<unknown> = typeof asyncFn === "function" ? (asyncFn as () => Promise<unknown>)() : (asyncFn as Promise<unknown>);
    return promise.then(
        () => {
            if (message instanceof Error) throw message;
            throw new AssertionError({
                message: typeof message === "string" ? (message as string) : "Missing expected rejection",
                operator: "rejects"
            });
        },
        (err) => {
            if (error !== undefined) {
                if (typeof error === "function") {
                    try {
                        const res = (error as (e: unknown) => unknown)(err);
                        if (res === true || res === undefined) {
                            return;
                        }
                    } catch (validateErr) {
                        throw validateErr;
                    }
                    throw new AssertionError({
                        message: typeof message === "string" ? (message as string) : "Validation function failed",
                        actual: err,
                        expected: error,
                        operator: "rejects"
                    });
                } else if (error instanceof RegExp) {
                    const msg = err instanceof Error ? err.message : String(err);
                    if (!error.test(msg)) {
                        throw new AssertionError({
                            message: typeof message === "string" ? (message as string) : `Expected ${msg} to match ${error}`,
                            actual: err,
                            expected: error,
                            operator: "rejects"
                        });
                    }
                }
            }
        }
    );
}

export function doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: unknown): Promise<void> {
    const promise: Promise<unknown> = typeof asyncFn === "function" ? (asyncFn as () => Promise<unknown>)() : (asyncFn as Promise<unknown>);
    return promise.then(
        () => {},
        (err) => {
            if (message instanceof Error) throw message;
            throw new AssertionError({
                message: typeof message === "string" ? (message as string) : `Got unexpected rejection: ${err}`,
                actual: err,
                operator: "doesNotReject"
            });
        }
    );
}

export function assert(value: unknown, message?: unknown): void {
    ok(value, message);
}

export namespace assert {
    export function ok(value: unknown, message?: unknown): void {
        ok(value, message);
    }
    export function equal(actual: unknown, expected: unknown, message?: unknown): void {
        equal(actual, expected, message);
    }
    export function notEqual(actual: unknown, expected: unknown, message?: unknown): void {
        notEqual(actual, expected, message);
    }
    export function strictEqual(actual: unknown, expected: unknown, message?: unknown): void {
        strictEqual(actual, expected, message);
    }
    export function notStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
        notStrictEqual(actual, expected, message);
    }
    export function deepEqual(actual: unknown, expected: unknown, message?: unknown): void {
        deepEqual(actual, expected, message);
    }
    export function notDeepEqual(actual: unknown, expected: unknown, message?: unknown): void {
        notDeepEqual(actual, expected, message);
    }
    export function deepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
        deepStrictEqual(actual, expected, message);
    }
    export function notDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
        notDeepStrictEqual(actual, expected, message);
    }
    export function partialDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void {
        partialDeepStrictEqual(actual, expected, message);
    }
    export function throws(fn: Function, error?: unknown, message?: unknown): void {
        throws(fn, error, message);
    }
    export function doesNotThrow(fn: Function, error?: unknown, message?: unknown): void {
        doesNotThrow(fn, error, message);
    }
    export function ifError(value: unknown): void {
        ifError(value);
    }
    export function fail(actualOrMsg?: unknown, expected?: unknown, message?: unknown, operator?: string): void {
        fail(actualOrMsg, expected, message, operator);
    }
    export function match(str: string, regexp: RegExp, message?: unknown): void {
        match(str, regexp, message);
    }
    export function doesNotMatch(str: string, regexp: RegExp, message?: unknown): void {
        doesNotMatch(str, regexp, message);
    }
    export function rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: unknown): Promise<void> {
        return rejects(asyncFn, error, message);
    }
    export function doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: unknown): Promise<void> {
        return doesNotReject(asyncFn, error, message);
    }
    export function Assert(value: unknown, message?: unknown): void {
        ok(value, message);
    }
    export class AssertionError extends BaseAssertionError {}
    export class CallTracker {
        private _tracker: unknown;
        constructor() {
            this._tracker = new CallTracker();
        }
    }
}
