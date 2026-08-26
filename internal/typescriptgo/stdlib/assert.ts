// ScriptGo Standard Library: node:assert

export class AssertionErrorOptions {
    message?: string = "";
    actual?: unknown = undefined;
    expected?: unknown = undefined;
    operator?: string = "";
    stackStartFn?: Function = undefined;
    diff?: string = "";
}

export class AssertionError extends Error {
    name: string = "AssertionError";
    code: string = "ERR_ASSERTION";
    actual: unknown = undefined;
    expected: unknown = undefined;
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

export class CallTrackerReportInformation {
    message: string = "";
    actual: number = 0;
    expected: number = 0;
    operator: string = "";
    stack: string = "";
}

export class CallTrackerCall {
    thisArg: unknown = undefined;
    arguments: unknown[] = [];
}

class TrackedFunctionEntry {
    origFn: Function;
    wrappedFn: Function;
    expected: number = 0;
    actual: number = 0;
    calls: CallTrackerCall[] = [];
    name: string = "";

    constructor(origFn: Function, wrappedFn: Function, expected: number, name: string) {
        this.origFn = origFn;
        this.wrappedFn = wrappedFn;
        this.expected = expected;
        this.actual = 0;
        this.calls = [];
        this.name = name;
    }
}

export class CallTracker {
    private _entries: TrackedFunctionEntry[] = [];

    calls(fn?: Function, exact?: number): Function {
        const targetFn = fn || (() => {});
        const expected = typeof exact === "number" ? (exact as number) : 1;
        const fnName = "function";

        const entry: TrackedFunctionEntry = new TrackedFunctionEntry(targetFn, () => {}, expected, fnName);
        this._entries.push(entry);
        const wrapped = (arg1?: unknown, arg2?: unknown, arg3?: unknown, arg4?: unknown): unknown => {
            entry.actual = entry.actual + 1;
            const c = new CallTrackerCall();
            c.thisArg = undefined;
            if (arg1 !== undefined) c.arguments.push(arg1);
            if (arg2 !== undefined) c.arguments.push(arg2);
            if (arg3 !== undefined) c.arguments.push(arg3);
            if (arg4 !== undefined) c.arguments.push(arg4);
            entry.calls.push(c);
            if (arg1 === undefined) return targetFn();
            if (arg2 === undefined) return targetFn(arg1);
            if (arg3 === undefined) return targetFn(arg1, arg2);
            if (arg4 === undefined) return targetFn(arg1, arg2, arg3);
            return targetFn(arg1, arg2, arg3, arg4);
        };
        entry.wrappedFn = wrapped;

        return wrapped;
    }

    getCalls(fn?: Function): CallTrackerCall[] {
        if (fn !== undefined) {
            for (let i = 0; i < this._entries.length; i++) {
                if (this._entries[i].wrappedFn === fn || this._entries[i].origFn === fn) {
                    return this._entries[i].calls;
                }
            }
        }
        if (this._entries.length > 0) {
            return this._entries[0].calls;
        }
        return [];
    }

    report(): CallTrackerReportInformation[] {
        const res: CallTrackerReportInformation[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            const entry = this._entries[i];
            if (entry.actual !== entry.expected) {
                const info = new CallTrackerReportInformation();
                info.message = `Expected the ${entry.name} function to be executed ${entry.expected} time(s) but was executed ${entry.actual} time(s).`;
                info.actual = entry.actual;
                info.expected = entry.expected;
                info.operator = entry.name;
                info.stack = "";
                res.push(info);
            }
        }
        return res;
    }

    reset(fn?: Function): void {
        if (fn !== undefined) {
            for (let i = 0; i < this._entries.length; i++) {
                if (this._entries[i].wrappedFn === fn || this._entries[i].origFn === fn) {
                    this._entries[i].actual = 0;
                    this._entries[i].calls = [];
                }
            }
        } else {
            for (let i = 0; i < this._entries.length; i++) {
                this._entries[i].actual = 0;
                this._entries[i].calls = [];
            }
        }
    }

    verify(): void {
        const errors = this.report();
        if (errors.length > 0) {
            const first = errors[0];
            throw new AssertionError({
                message: first.message,
                actual: first.actual,
                expected: first.expected,
                operator: first.operator
            });
        }
    }
}

export class AssertOptions {
    diff?: "simple" | "full" = "simple";
    strict?: boolean = true;
}

export class Assert {
    private _diff: "simple" | "full" = "simple";
    private _strict: boolean = true;

    constructor(options?: AssertOptions) {
        if (options) {
            if (options.diff) this._diff = options.diff;
            if (options.strict !== undefined) this._strict = options.strict;
        }
    }

    ok(value: unknown, message?: unknown): void { ok(value, message); }
    equal(actual: unknown, expected: unknown, message?: unknown): void {
        if (this._strict) strictEqual(actual, expected, message);
        else equal(actual, expected, message);
    }
    notEqual(actual: unknown, expected: unknown, message?: unknown): void {
        if (this._strict) notStrictEqual(actual, expected, message);
        else notEqual(actual, expected, message);
    }
    strictEqual(actual: unknown, expected: unknown, message?: unknown): void { strictEqual(actual, expected, message); }
    notStrictEqual(actual: unknown, expected: unknown, message?: unknown): void { notStrictEqual(actual, expected, message); }
    deepEqual(actual: unknown, expected: unknown, message?: unknown): void {
        if (this._strict) deepStrictEqual(actual, expected, message);
        else deepEqual(actual, expected, message);
    }
    notDeepEqual(actual: unknown, expected: unknown, message?: unknown): void {
        if (this._strict) notDeepStrictEqual(actual, expected, message);
        else notDeepEqual(actual, expected, message);
    }
    deepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void { deepStrictEqual(actual, expected, message); }
    notDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void { notDeepStrictEqual(actual, expected, message); }
    partialDeepStrictEqual(actual: unknown, expected: unknown, message?: unknown): void { partialDeepStrictEqual(actual, expected, message); }
    throws(fn: Function, error?: unknown, message?: unknown): void { throws(fn, error, message); }
    doesNotThrow(fn: Function, error?: unknown, message?: unknown): void { doesNotThrow(fn, error, message); }
    ifError(value: unknown): void { ifError(value); }
    fail(actualOrMsg?: unknown, expected?: unknown, message?: unknown, operator?: string): void { fail(actualOrMsg, expected, message, operator); }
    match(str: string, regexp: RegExp, message?: unknown): void { match(str, regexp, message); }
    doesNotMatch(str: string, regexp: RegExp, message?: unknown): void { doesNotMatch(str, regexp, message); }
    rejects(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> { return rejects(asyncFn, error, message); }
    doesNotReject(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> { return doesNotReject(asyncFn, error, message); }
}

function isDeepEqualInternal(a: unknown, b: unknown, strict: boolean): boolean {
    if (a === null || a === undefined || b === null || b === undefined) {
        return a === b;
    }
    const typeA = typeof a;
    const typeB = typeof b;
    if (typeA !== typeB) {
        return false;
    }
    if (typeA !== "object") {
        if (strict) {
            if (a === 0 && b === 0) {
                return 1 / (a as number) === 1 / (b as number);
            }
            if (typeof a === "number" && typeof b === "number" && Number.isNaN(a as number) && Number.isNaN(b as number)) {
                return true;
            }
            return a === b;
        }
        return a == b;
    }
    if (a instanceof Date && b instanceof Date) {
        const da: Date = a as Date;
        const db: Date = b as Date;
        return da.getTime() === db.getTime();
    }
    if (a instanceof RegExp && b instanceof RegExp) {
        const ra: RegExp = a as RegExp;
        const rb: RegExp = b as RegExp;
        return ra.source === rb.source && ra.flags === rb.flags;
    }
    if (a instanceof Error && b instanceof Error) {
        return a.name === b.name && a.message === b.message;
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

    if (a instanceof Map && b instanceof Map) {
        const mapA = a as Map<unknown, unknown>;
        const mapB = b as Map<unknown, unknown>;
        if (mapA.size !== mapB.size) return false;
        let mapMatch = true;
        mapA.forEach((v: unknown, k: unknown) => {
            if (!mapMatch) return;
            if (!mapB.has(k) || !isDeepEqualInternal(v, mapB.get(k), strict)) {
                mapMatch = false;
            }
        });
        return mapMatch;
    }

    if (a instanceof Set && b instanceof Set) {
        const setA = a as Set<unknown>;
        const setB = b as Set<unknown>;
        if (setA.size !== setB.size) return false;
        let setMatch = true;
        setA.forEach((v: unknown) => {
            if (!setMatch) return;
            if (!setB.has(v)) {
                setMatch = false;
            }
        });
        return setMatch;
    }

    return Object.is(a, b);
}

function isPartialDeepEqualInternal(actual: unknown, expected: unknown): boolean {
    if (actual === expected) return true;
    if (typeof actual === "number" && typeof expected === "number" && Number.isNaN(actual as number) && Number.isNaN(expected as number)) {
        return true;
    }
    if (expected === null || expected === undefined) {
        return actual === expected;
    }
    if (typeof expected !== "object") {
        return actual === expected;
    }
    if (typeof actual !== "object" || actual === null) {
        return false;
    }
    if (expected instanceof Date) {
        if (!(actual instanceof Date)) return false;
        const da: Date = actual as Date;
        const de: Date = expected as Date;
        return da.getTime() === de.getTime();
    }
    if (expected instanceof RegExp) {
        if (!(actual instanceof RegExp)) return false;
        const ra: RegExp = actual as RegExp;
        const re: RegExp = expected as RegExp;
        return ra.source === re.source && ra.flags === re.flags;
    }
    if (Array.isArray(expected)) {
        if (!Array.isArray(actual)) return false;
        const expArr = expected as unknown[];
        const actArr = actual as unknown[];
        if (actArr.length < expArr.length) return false;
        for (let i = 0; i < expArr.length; i++) {
            if (!isPartialDeepEqualInternal(actArr[i], expArr[i])) return false;
        }
        return true;
    }
    const expObj = expected as Record<string, unknown>;
    const actObj = actual as Record<string, unknown>;
    const keys = Object.keys(expObj);
    const actKeys = Object.keys(actObj);
    for (let i = 0; i < keys.length; i++) {
        const k = keys[i];
        if (actKeys.indexOf(k) < 0) return false;
    }
    const actJson = JSON.stringify(actual);
    for (let i = 0; i < keys.length; i++) {
        const k = keys[i];
        const keyPattern = '"' + k + '":';
        if (actJson.indexOf(keyPattern) < 0) return false;
    }
    return true;
}

function validateExpectedError(caughtErr: unknown, expectedError: unknown, op: string, customMessage?: unknown): void {
    if (expectedError === undefined) return;
    if (typeof expectedError === "function") {
        try {
            const fn = expectedError as (e: unknown) => unknown;
            const res = fn(caughtErr);
            if (res === true || res === undefined) {
                return;
            }
        } catch (validateErr) {
            throw validateErr;
        }
        throw new AssertionError({
            message: typeof customMessage === "string" ? (customMessage as string) : "Validation function failed",
            actual: caughtErr,
            expected: expectedError,
            operator: op
        });
    } else if (expectedError instanceof RegExp) {
        const msg = caughtErr instanceof Error ? caughtErr.message : String(caughtErr);
        if (!expectedError.test(msg)) {
            throw new AssertionError({
                message: typeof customMessage === "string" ? (customMessage as string) : `Expected ${msg} to match ${expectedError}`,
                actual: caughtErr,
                expected: expectedError,
                operator: op
            });
        }
    } else if (typeof expectedError === "object" && expectedError !== null) {
        if (!isPartialDeepEqualInternal(caughtErr, expectedError)) {
            throw new AssertionError({
                message: typeof customMessage === "string" ? (customMessage as string) : "Error object does not match expected pattern",
                actual: caughtErr,
                expected: expectedError,
                operator: op
            });
        }
    }
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
    validateExpectedError(caughtErr, error, "throws", message);
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

export async function rejects(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> {
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        if (typeof asyncFn === "function") {
            const fn = asyncFn as () => Promise<unknown>;
            await fn();
        } else {
            await (asyncFn as Promise<unknown>);
        }
    } catch (err) {
        threw = true;
        caughtErr = err;
    }

    if (!threw) {
        if (message instanceof Error) throw message;
        throw new AssertionError({
            message: typeof message === "string" ? (message as string) : "Missing expected rejection",
            operator: "rejects"
        });
    }

    validateExpectedError(caughtErr, error, "rejects", message);
}

export async function doesNotReject(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> {
    try {
        if (typeof asyncFn === "function") {
            const fn = asyncFn as () => Promise<unknown>;
            await fn();
        } else {
            await (asyncFn as Promise<unknown>);
        }
    } catch (err) {
        if (message instanceof Error) throw message;
        const msg = typeof message === "string" ? (message as string) : "Got unexpected rejection";
        throw new AssertionError({
            message: msg,
            actual: err,
            operator: "doesNotReject"
        });
    }
}

export function assert(value: unknown, message?: unknown): void {
    ok(value, message);
}

export const strict = new Assert({ strict: true });

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
    export function rejects(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> {
        return rejects(asyncFn, error, message);
    }
    export function doesNotReject(asyncFn: unknown, error?: unknown, message?: unknown): Promise<void> {
        return doesNotReject(asyncFn, error, message);
    }
}

