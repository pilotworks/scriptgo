// ScriptGo Standard Library: node:assert

export interface AssertionErrorOptions {
    message?: string;
    actual?: unknown;
    expected?: unknown;
    operator?: string;
    stackStartFn?: Function;
    diff?: "simple" | "full";
}

export class AssertionError extends Error {
    name: string = "AssertionError";
    code: string = "ERR_ASSERTION";
    actual: unknown;
    expected: unknown;
    operator: string;
    generatedMessage: boolean;

    constructor(options?: string | AssertionErrorOptions) {
        let msg = "Assertion failed";
        let isGen = true;
        let act: unknown = undefined;
        let exp: unknown = undefined;
        let op = "";

        if (typeof options === "string") {
            msg = options;
            isGen = false;
        } else if (options && typeof options === "object") {
            const opts = options as AssertionErrorOptions;
            act = opts.actual;
            exp = opts.expected;
            if (opts.operator !== undefined) {
                op = opts.operator;
            }
            if (opts.message !== undefined) {
                msg = opts.message;
                isGen = false;
            } else if (op) {
                msg = `${act} ${op} ${exp}`;
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

export interface CallTrackerReportInformation {
    message: string;
    actual: number;
    expected: number;
    operator: string;
    stack: string;
}

export interface CallTrackerCall {
    thisArg: unknown;
    arguments: unknown[];
}

class TrackedFunctionEntry {
    origFn: Function;
    wrappedFn: Function;
    expected: number;
    actual: number = 0;
    calls: CallTrackerCall[] = [];
    name: string;

    constructor(origFn: Function, wrappedFn: Function, expected: number, name: string) {
        this.origFn = origFn;
        this.wrappedFn = wrappedFn;
        this.expected = expected;
        this.name = name;
    }
}

export class CallTracker {
    private _entries: TrackedFunctionEntry[] = [];

    calls(fn?: Function, exact: number = 1): Function {
        const targetFn = fn || (() => {});
        const expected = typeof exact === "number" ? exact : 1;
        const fnName = "function";

        const entry = new TrackedFunctionEntry(targetFn, () => {}, expected, fnName);
        this._entries.push(entry);

        const wrapped = (a?: unknown, b?: unknown, c?: unknown): unknown => {
            entry.actual++;
            const argsList: unknown[] = [];
            if (a !== undefined) argsList.push(a);
            if (b !== undefined) argsList.push(b);
            if (c !== undefined) argsList.push(c);
            const call: CallTrackerCall = {
                thisArg: undefined,
                arguments: argsList
            };
            entry.calls.push(call);
            const callable = targetFn as (arg0?: unknown, arg1?: unknown, arg2?: unknown) => unknown;
            return callable(a, b, c);
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
            return [];
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
                res.push({
                    message: `Expected the ${entry.name} function to be executed ${entry.expected} time(s) but was executed ${entry.actual} time(s).`,
                    actual: entry.actual,
                    expected: entry.expected,
                    operator: entry.name,
                    stack: ""
                });
            }
        }
        return res;
    }

    reset(fn?: Function): void {
        for (let i = 0; i < this._entries.length; i++) {
            if (fn === undefined || this._entries[i].wrappedFn === fn || this._entries[i].origFn === fn) {
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

export interface AssertOptions {
    diff?: "simple" | "full";
    strict?: boolean;
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

    ok(value: unknown, message?: string | Error): void {
        okImpl(value, message);
    }
    equal(actual: unknown, expected: unknown, message?: string | Error): void {
        if (this._strict) strictEqualImpl(actual, expected, message);
        else equalImpl(actual, expected, message);
    }
    notEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        if (this._strict) notStrictEqualImpl(actual, expected, message);
        else notEqualImpl(actual, expected, message);
    }
    strictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        strictEqualImpl(actual, expected, message);
    }
    notStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notStrictEqualImpl(actual, expected, message);
    }
    deepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        if (this._strict) deepStrictEqualImpl(actual, expected, message);
        else deepEqualImpl(actual, expected, message);
    }
    notDeepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        if (this._strict) notDeepStrictEqualImpl(actual, expected, message);
        else notDeepEqualImpl(actual, expected, message);
    }
    deepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        deepStrictEqualImpl(actual, expected, message);
    }
    notDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notDeepStrictEqualImpl(actual, expected, message);
    }
    partialDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        partialDeepStrictEqualImpl(actual, expected, message);
    }
    throws(fn: Function, error?: unknown, message?: string | Error): void {
        throwsImpl(fn, error, message);
    }
    doesNotThrow(fn: Function, error?: unknown, message?: string | Error): void {
        doesNotThrowImpl(fn, error, message);
    }
    ifError(value: unknown): void {
        ifErrorImpl(value);
    }
    fail(actualOrMsg?: unknown, expected?: unknown, message?: string | Error, operator?: string, stackStartFn?: Function): never {
        return failImpl(actualOrMsg, expected, message, operator, stackStartFn);
    }
    match(str: string, regexp: RegExp, message?: string | Error): void {
        matchImpl(str, regexp, message);
    }
    doesNotMatch(str: string, regexp: RegExp, message?: string | Error): void {
        doesNotMatchImpl(str, regexp, message);
    }
    rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return rejectsImpl(asyncFn, error, message);
    }
    doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return doesNotRejectImpl(asyncFn, error, message);
    }
}

function createAssertionError(
    defaultMessage: string,
    actual: unknown,
    expected: unknown,
    operator: string,
    message?: string | Error
): AssertionError {
    if (message instanceof Error) {
        throw message;
    }
    let msg = defaultMessage;
    if (typeof message === "string") {
        msg = message;
    }
    return new AssertionError({
        message: msg,
        actual: actual,
        expected: expected,
        operator: operator
    });
}

function normalizeErrorAndMessage(
    error?: unknown,
    message?: string | Error
): { expectedError: unknown; customMessage?: string | Error } {
    if (typeof error === "string" && message === undefined) {
        return { expectedError: undefined, customMessage: error };
    }
    return { expectedError: error, customMessage: message };
}

function isDeepEqualInternal(a: unknown, b: unknown, strict: boolean): boolean {
    if (strict) {
        if (Object.is(a, b)) return true;
    } else {
        if (a == b) {
            if (a === null || a === undefined || b === null || b === undefined) {
                return true;
            }
            if (typeof a !== "object" && typeof b !== "object") {
                return true;
            }
        }
    }

    if (a === null || a === undefined || b === null || b === undefined) {
        return false;
    }

    const typeA = typeof a;
    const typeB = typeof b;

    if (typeA !== typeB) {
        return false;
    }

    if (typeA !== "object") {
        return strict ? Object.is(a, b) : a == b;
    }

    if (a instanceof Date && b instanceof Date) {
        return a.getTime() === b.getTime();
    }
    if (a instanceof Date || b instanceof Date) {
        return false;
    }

    if (a instanceof RegExp && b instanceof RegExp) {
        return a.source === b.source && a.flags === b.flags;
    }
    if (a instanceof RegExp || b instanceof RegExp) {
        return false;
    }

    if (a instanceof Error && b instanceof Error) {
        return a.name === b.name && a.message === b.message;
    }
    if (a instanceof Error || b instanceof Error) {
        return false;
    }

    if (Array.isArray(a)) {
        if (!Array.isArray(b)) return false;
        if (a.length !== b.length) return false;
        for (let i = 0; i < a.length; i++) {
            if (!isDeepEqualInternal(a[i], b[i], strict)) return false;
        }
        return true;
    }
    if (Array.isArray(b)) return false;

    if (a instanceof Uint8Array && b instanceof Uint8Array) {
        if (a.length !== b.length) return false;
        for (let i = 0; i < a.length; i++) {
            if (a[i] !== b[i]) return false;
        }
        return true;
    }
    if (a instanceof Uint8Array || b instanceof Uint8Array) return false;

    if (a instanceof Map && b instanceof Map) {
        if (a.size !== b.size) return false;
        for (const [key, val] of a.entries()) {
            if (!b.has(key) || !isDeepEqualInternal(val, b.get(key), strict)) {
                return false;
            }
        }
        return true;
    }
    if (a instanceof Map || b instanceof Map) return false;

    if (a instanceof Set && b instanceof Set) {
        if (a.size !== b.size) return false;
        for (const val of a.values()) {
            if (!b.has(val)) {
                let found = false;
                for (const otherVal of b.values()) {
                    if (isDeepEqualInternal(val, otherVal, strict)) {
                        found = true;
                        break;
                    }
                }
                if (!found) return false;
            }
        }
        return true;
    }
    if (a instanceof Set || b instanceof Set) return false;

    if (strict) {
        const protoA = Object.getPrototypeOf(a);
        const protoB = Object.getPrototypeOf(b);
        if (protoA !== protoB && protoA !== null && protoB !== null) {
            return false;
        }
    }

    const keysA = Object.keys(a);
    const keysB = Object.keys(b);
    if (keysA.length !== keysB.length) return false;

    for (let i = 0; i < keysA.length; i++) {
        const key = keysA[i];
        if (!Reflect.has(b, key)) {
            return false;
        }
        const valA = Reflect.get(a, key);
        const valB = Reflect.get(b, key);
        if (!isDeepEqualInternal(valA, valB, strict)) {
            return false;
        }
    }

    return true;
}

function isPartialDeepEqualInternal(actual: unknown, expected: unknown): boolean {
    if (Object.is(actual, expected)) return true;
    if (expected === null || expected === undefined || typeof expected !== "object") {
        return Object.is(actual, expected);
    }
    if (actual === null || actual === undefined || typeof actual !== "object") {
        return false;
    }
    if (expected instanceof Date) {
        return actual instanceof Date && actual.getTime() === expected.getTime();
    }
    if (expected instanceof RegExp) {
        return actual instanceof RegExp && actual.source === expected.source && actual.flags === expected.flags;
    }
    if (expected instanceof Error) {
        if (!(actual instanceof Error)) return false;
        if (expected.name && actual.name !== expected.name) return false;
        if (expected.message && actual.message !== expected.message) return false;
        return true;
    }
    if (Array.isArray(expected)) {
        if (!Array.isArray(actual)) return false;
        if (actual.length < expected.length) return false;
        for (let i = 0; i < expected.length; i++) {
            if (!isPartialDeepEqualInternal(actual[i], expected[i])) return false;
        }
        return true;
    }
    if (expected instanceof Map && actual instanceof Map) {
        for (const [k, v] of expected.entries()) {
            if (!actual.has(k) || !isPartialDeepEqualInternal(actual.get(k), v)) return false;
        }
        return true;
    }
    if (expected instanceof Set && actual instanceof Set) {
        for (const v of expected.values()) {
            if (!actual.has(v)) return false;
        }
        return true;
    }

    const expObj = expected as Record<string, unknown>;
    const actObj = actual as Record<string, unknown>;
    const expKeys = Object.keys(expObj);

    for (let i = 0; i < expKeys.length; i++) {
        const key = expKeys[i];
        if (!Reflect.has(actObj, key)) return false;
        const actVal = Reflect.get(actObj, key);
        const expVal = Reflect.get(expObj, key);
        if (!isPartialDeepEqualInternal(actVal, expVal)) return false;
    }
    return true;
}

function validateExpectedError(caughtErr: unknown, expectedError: unknown, op: string, customMessage?: string | Error): void {
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
        throw createAssertionError(
            "Validation function failed",
            caughtErr,
            expectedError,
            op,
            customMessage
        );
    } else if (expectedError instanceof RegExp) {
        const msg = caughtErr instanceof Error ? caughtErr.message : String(caughtErr);
        if (!expectedError.test(msg)) {
            throw createAssertionError(
                `The input did not match the regular expression ${expectedError}. Input: ${msg}`,
                caughtErr,
                expectedError,
                op,
                customMessage
            );
        }
    } else if (typeof expectedError === "object" && expectedError !== null) {
        if (!isPartialDeepEqualInternal(caughtErr, expectedError)) {
            throw createAssertionError(
                "Error object does not match expected pattern",
                caughtErr,
                expectedError,
                op,
                customMessage
            );
        }
    }
}

function okImpl(value: unknown, message?: string | Error): void {
    if (!value) {
        throw createAssertionError(
            "The expression evaluated to a falsy value",
            value,
            true,
            "==",
            message
        );
    }
}

function equalImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (actual != expected && String(actual) != String(expected)) {
        throw createAssertionError(
            `${actual} == ${expected}`,
            actual,
            expected,
            "==",
            message
        );
    }
}

function notEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (actual == expected || String(actual) == String(expected)) {
        throw createAssertionError(
            `${actual} != ${expected}`,
            actual,
            expected,
            "!=",
            message
        );
    }
}

function strictEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (!Object.is(actual, expected)) {
        throw createAssertionError(
            `Expected values to be strictly equal:\n${actual} !== ${expected}`,
            actual,
            expected,
            "strictEqual",
            message
        );
    }
}

function notStrictEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (Object.is(actual, expected)) {
        throw createAssertionError(
            `Expected values to be strictly not equal:\n${actual} === ${expected}`,
            actual,
            expected,
            "notStrictEqual",
            message
        );
    }
}

function deepEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (!isDeepEqualInternal(actual, expected, false)) {
        throw createAssertionError(
            "Values not deep equal",
            actual,
            expected,
            "deepEqual",
            message
        );
    }
}

function notDeepEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (isDeepEqualInternal(actual, expected, false)) {
        throw createAssertionError(
            "Values are deep equal",
            actual,
            expected,
            "notDeepEqual",
            message
        );
    }
}

function deepStrictEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (!isDeepEqualInternal(actual, expected, true)) {
        throw createAssertionError(
            "Values not deep strict equal",
            actual,
            expected,
            "deepStrictEqual",
            message
        );
    }
}

function notDeepStrictEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (isDeepEqualInternal(actual, expected, true)) {
        throw createAssertionError(
            "Values are deep strict equal",
            actual,
            expected,
            "notDeepStrictEqual",
            message
        );
    }
}

function partialDeepStrictEqualImpl(actual: unknown, expected: unknown, message?: string | Error): void {
    if (!isPartialDeepEqualInternal(actual, expected)) {
        throw createAssertionError(
            "Values not partial deep strict equal",
            actual,
            expected,
            "partialDeepStrictEqual",
            message
        );
    }
}

function throwsImpl(fn: Function, error?: unknown, message?: string | Error): void {
    const norm = normalizeErrorAndMessage(error, message);
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        fn();
    } catch (err) {
        threw = true;
        caughtErr = err;
    }
    if (!threw) {
        throw createAssertionError(
            "Missing expected exception",
            undefined,
            norm.expectedError,
            "throws",
            norm.customMessage
        );
    }
    validateExpectedError(caughtErr, norm.expectedError, "throws", norm.customMessage);
}

function doesNotThrowImpl(fn: Function, error?: unknown, message?: string | Error): void {
    const norm = normalizeErrorAndMessage(error, message);
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        fn();
    } catch (err) {
        threw = true;
        caughtErr = err;
    }
    if (threw) {
        if (norm.expectedError !== undefined) {
            let matches = false;
            if (typeof norm.expectedError === "function") {
                try {
                    const checkFn = norm.expectedError as (e: unknown) => boolean;
                    matches = checkFn(caughtErr) === true;
                } catch {
                    matches = false;
                }
            } else if (norm.expectedError instanceof RegExp) {
                const msg = caughtErr instanceof Error ? caughtErr.message : String(caughtErr);
                matches = norm.expectedError.test(msg);
            }
            if (matches) {
                throw createAssertionError(
                    `Got unwanted exception: ${caughtErr}`,
                    caughtErr,
                    undefined,
                    "doesNotThrow",
                    norm.customMessage
                );
            }
            throw caughtErr;
        }
        throw createAssertionError(
            `Got unexpected exception: ${caughtErr}`,
            caughtErr,
            undefined,
            "doesNotThrow",
            norm.customMessage
        );
    }
}

function ifErrorImpl(value: unknown): void {
    if (value !== null && value !== undefined) {
        throw value;
    }
}

function failImpl(
    actualOrMsg?: unknown,
    expected?: unknown,
    message?: string | Error,
    operator?: string,
    _stackStartFn?: Function
): never {
    if (actualOrMsg === undefined && expected === undefined && message === undefined && operator === undefined) {
        throw new AssertionError({ message: "Failed", operator: "fail" });
    }
    if (expected === undefined && message === undefined && operator === undefined) {
        if (actualOrMsg instanceof Error) throw actualOrMsg;
        let failMsg = "Failed";
        if (typeof actualOrMsg === "string") {
            failMsg = actualOrMsg;
        }
        throw new AssertionError({
            message: failMsg,
            operator: "fail"
        });
    }
    if (message instanceof Error) throw message;
    let failMsg = "Failed";
    if (typeof message === "string") {
        failMsg = message;
    }
    let op = "fail";
    if (operator !== undefined) {
        op = operator;
    }
    throw new AssertionError({
        message: failMsg,
        actual: actualOrMsg,
        expected: expected,
        operator: op
    });
}

function matchImpl(str: string, regexp: RegExp, message?: string | Error): void {
    if (!regexp.test(str)) {
        throw createAssertionError(
            `The input did not match the regular expression ${regexp}. Input: ${str}`,
            str,
            regexp,
            "match",
            message
        );
    }
}

function doesNotMatchImpl(str: string, regexp: RegExp, message?: string | Error): void {
    if (regexp.test(str)) {
        throw createAssertionError(
            `The input was expected to not match the regular expression ${regexp}. Input: ${str}`,
            str,
            regexp,
            "doesNotMatch",
            message
        );
    }
}

async function rejectsImpl(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
    const norm = normalizeErrorAndMessage(error, message);
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        if (typeof asyncFn === "function") {
            const callable = asyncFn as () => Promise<unknown>;
            const promise = callable();
            await promise;
        } else {
            await (asyncFn as Promise<unknown>);
        }
    } catch (err) {
        threw = true;
        caughtErr = err;
    }

    if (!threw) {
        throw createAssertionError(
            "Missing expected rejection",
            undefined,
            norm.expectedError,
            "rejects",
            norm.customMessage
        );
    }

    validateExpectedError(caughtErr, norm.expectedError, "rejects", norm.customMessage);
}

async function doesNotRejectImpl(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
    const norm = normalizeErrorAndMessage(error, message);
    let threw = false;
    let caughtErr: unknown = undefined;
    try {
        if (typeof asyncFn === "function") {
            const callable = asyncFn as () => Promise<unknown>;
            const promise = callable();
            await promise;
        } else {
            await (asyncFn as Promise<unknown>);
        }
    } catch (err) {
        threw = true;
        caughtErr = err;
    }

    if (threw) {
        throw createAssertionError(
            `Got unexpected rejection: ${caughtErr}`,
            caughtErr,
            undefined,
            "doesNotReject",
            norm.customMessage
        );
    }
}

export function ok(value: unknown, message?: string | Error): void {
    okImpl(value, message);
}

export function equal(actual: unknown, expected: unknown, message?: string | Error): void {
    equalImpl(actual, expected, message);
}

export function notEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    notEqualImpl(actual, expected, message);
}

export function strictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    strictEqualImpl(actual, expected, message);
}

export function notStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    notStrictEqualImpl(actual, expected, message);
}

export function deepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    deepEqualImpl(actual, expected, message);
}

export function notDeepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    notDeepEqualImpl(actual, expected, message);
}

export function deepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    deepStrictEqualImpl(actual, expected, message);
}

export function notDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    notDeepStrictEqualImpl(actual, expected, message);
}

export function partialDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
    partialDeepStrictEqualImpl(actual, expected, message);
}

export function throws(fn: Function, error?: unknown, message?: string | Error): void {
    throwsImpl(fn, error, message);
}

export function doesNotThrow(fn: Function, error?: unknown, message?: string | Error): void {
    doesNotThrowImpl(fn, error, message);
}

export function ifError(value: unknown): void {
    ifErrorImpl(value);
}

export function fail(
    actualOrMsg?: unknown,
    expected?: unknown,
    message?: string | Error,
    operator?: string,
    stackStartFn?: Function
): never {
    return failImpl(actualOrMsg, expected, message, operator, stackStartFn);
}

export function match(str: string, regexp: RegExp, message?: string | Error): void {
    matchImpl(str, regexp, message);
}

export function doesNotMatch(str: string, regexp: RegExp, message?: string | Error): void {
    doesNotMatchImpl(str, regexp, message);
}

export function rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
    return rejectsImpl(asyncFn, error, message);
}

export function doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
    return doesNotRejectImpl(asyncFn, error, message);
}

export function strict(value: unknown, message?: string | Error): void {
    okImpl(value, message);
}

export namespace strict {
    export function ok(value: unknown, message?: string | Error): void {
        okImpl(value, message);
    }
    export function equal(actual: unknown, expected: unknown, message?: string | Error): void {
        strictEqualImpl(actual, expected, message);
    }
    export function notEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notStrictEqualImpl(actual, expected, message);
    }
    export function strictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        strictEqualImpl(actual, expected, message);
    }
    export function notStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notStrictEqualImpl(actual, expected, message);
    }
    export function deepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        deepStrictEqualImpl(actual, expected, message);
    }
    export function notDeepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notDeepStrictEqualImpl(actual, expected, message);
    }
    export function deepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        deepStrictEqualImpl(actual, expected, message);
    }
    export function notDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notDeepStrictEqualImpl(actual, expected, message);
    }
    export function partialDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        partialDeepStrictEqualImpl(actual, expected, message);
    }
    export function throws(fn: Function, error?: unknown, message?: string | Error): void {
        throwsImpl(fn, error, message);
    }
    export function doesNotThrow(fn: Function, error?: unknown, message?: string | Error): void {
        doesNotThrowImpl(fn, error, message);
    }
    export function ifError(value: unknown): void {
        ifErrorImpl(value);
    }
    export function fail(
        actualOrMsg?: unknown,
        expected?: unknown,
        message?: string | Error,
        operator?: string,
        stackStartFn?: Function
    ): never {
        return failImpl(actualOrMsg, expected, message, operator, stackStartFn);
    }
    export function match(str: string, regexp: RegExp, message?: string | Error): void {
        matchImpl(str, regexp, message);
    }
    export function doesNotMatch(str: string, regexp: RegExp, message?: string | Error): void {
        doesNotMatchImpl(str, regexp, message);
    }
    export function rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return rejectsImpl(asyncFn, error, message);
    }
    export function doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return doesNotRejectImpl(asyncFn, error, message);
    }
}

export function assert(value: unknown, message?: string | Error): void {
    okImpl(value, message);
}

export namespace assert {
    export function ok(value: unknown, message?: string | Error): void {
        okImpl(value, message);
    }
    export function equal(actual: unknown, expected: unknown, message?: string | Error): void {
        equalImpl(actual, expected, message);
    }
    export function notEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notEqualImpl(actual, expected, message);
    }
    export function strictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        strictEqualImpl(actual, expected, message);
    }
    export function notStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notStrictEqualImpl(actual, expected, message);
    }
    export function deepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        deepEqualImpl(actual, expected, message);
    }
    export function notDeepEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notDeepEqualImpl(actual, expected, message);
    }
    export function deepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        deepStrictEqualImpl(actual, expected, message);
    }
    export function notDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        notDeepStrictEqualImpl(actual, expected, message);
    }
    export function partialDeepStrictEqual(actual: unknown, expected: unknown, message?: string | Error): void {
        partialDeepStrictEqualImpl(actual, expected, message);
    }
    export function throws(fn: Function, error?: unknown, message?: string | Error): void {
        throwsImpl(fn, error, message);
    }
    export function doesNotThrow(fn: Function, error?: unknown, message?: string | Error): void {
        doesNotThrowImpl(fn, error, message);
    }
    export function ifError(value: unknown): void {
        ifErrorImpl(value);
    }
    export function fail(
        actualOrMsg?: unknown,
        expected?: unknown,
        message?: string | Error,
        operator?: string,
        stackStartFn?: Function
    ): never {
        return failImpl(actualOrMsg, expected, message, operator, stackStartFn);
    }
    export function match(str: string, regexp: RegExp, message?: string | Error): void {
        matchImpl(str, regexp, message);
    }
    export function doesNotMatch(str: string, regexp: RegExp, message?: string | Error): void {
        doesNotMatchImpl(str, regexp, message);
    }
    export function rejects(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return rejectsImpl(asyncFn, error, message);
    }
    export function doesNotReject(asyncFn: Function | Promise<unknown>, error?: unknown, message?: string | Error): Promise<void> {
        return doesNotRejectImpl(asyncFn, error, message);
    }
}

export default assert;
