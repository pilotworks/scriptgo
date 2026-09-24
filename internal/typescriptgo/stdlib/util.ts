// ScriptGo Standard Library: node:util

import { deepStrictEqual } from "node:assert";
import { AbortController, AbortSignal } from "node:events";

export class MIMEParams {
    private _entries: Array<[string, string]> = [];

    constructor(init?: string) {
        if (init) {
            const parts = init.split(";");
            for (let i = 0; i < parts.length; i++) {
                const part = parts[i].trim();
                if (part.length === 0) continue;
                const eqIdx = part.indexOf("=");
                if (eqIdx >= 0) {
                    const key = part.slice(0, eqIdx).trim().toLowerCase();
                    const val = part.slice(eqIdx + 1).trim();
                    this.set(key, val);
                }
            }
        }
    }

    get(name: string): string | null {
        const k = name.toLowerCase();
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i][0] === k) {
                return this._entries[i][1];
            }
        }
        return null;
    }

    set(name: string, value: string): void {
        const k = name.toLowerCase();
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i][0] === k) {
                this._entries[i][1] = value;
                return;
            }
        }
        this._entries.push([k, value]);
    }

    has(name: string): boolean {
        const k = name.toLowerCase();
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i][0] === k) {
                return true;
            }
        }
        return false;
    }

    delete(name: string): void {
        const k = name.toLowerCase();
        const next: Array<[string, string]> = [];
        for (let i = 0; i < this._entries.length; i++) {
            if (this._entries[i][0] !== k) {
                next.push(this._entries[i]);
            }
        }
        this._entries = next;
    }

    keys(): string[] {
        const keys: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            keys.push(this._entries[i][0]);
        }
        return keys;
    }

    values(): string[] {
        const vals: string[] = [];
        for (let i = 0; i < this._entries.length; i++) {
            vals.push(this._entries[i][1]);
        }
        return vals;
    }

    entries(): Array<[string, string]> {
        return this._entries;
    }

    [Symbol.iterator](): Array<[string, string]> {
        return this._entries;
    }
}

export class MIMEType {
    type: string;
    subtype: string;
    params: MIMEParams;

    constructor(input: string) {
        const trimmed = input.trim();
        const semiIdx = trimmed.indexOf(";");
        const fullType = (semiIdx >= 0 ? trimmed.slice(0, semiIdx) : trimmed).trim().toLowerCase();
        const slashIdx = fullType.indexOf("/");

        if (slashIdx < 0) {
            this.type = fullType;
            this.subtype = "";
        } else {
            this.type = fullType.slice(0, slashIdx);
            this.subtype = fullType.slice(slashIdx + 1);
        }

        const paramString = semiIdx >= 0 ? trimmed.slice(semiIdx + 1) : "";
        this.params = new MIMEParams(paramString);
    }

    get essence(): string {
        return this.type + "/" + this.subtype;
    }

    toString(): string {
        let res = this.essence;
        const ents = this.params.entries();
        for (let i = 0; i < ents.length; i++) {
            res += ";" + ents[i][0] + "=" + ents[i][1];
        }
        return res;
    }

    toJSON(): string {
        return this.toString();
    }
}

export function isArray(object: unknown): boolean {
    return Array.isArray(object);
}

export function isBoolean(object: unknown): boolean {
    return typeof object === "boolean";
}

export function isNull(object: unknown): boolean {
    return object === null;
}

export function isNullOrUndefined(object: unknown): boolean {
    return object === null || object === undefined;
}

export function isNumber(object: unknown): boolean {
    return typeof object === "number";
}

export function isString(object: unknown): boolean {
    return typeof object === "string";
}

export function isSymbol(object: unknown): boolean {
    return typeof object === "symbol";
}

export function isUndefined(object: unknown): boolean {
    return object === undefined;
}

export function isRegExp(object: unknown): boolean {
    return object instanceof RegExp;
}

export function isObject(object: unknown): boolean {
    return object !== null && typeof object === "object";
}

export function isDate(object: unknown): boolean {
    return object instanceof Date;
}

export function isError(object: unknown): boolean {
    return object instanceof Error;
}

export function isFunction(object: unknown): boolean {
    return typeof object === "function";
}

export function isPrimitive(object: unknown): boolean {
    return object === null || (typeof object !== "object" && typeof object !== "function");
}

export function isBuffer(object: unknown): boolean {
    return Buffer.isBuffer(object);
}

export function isDeepStrictEqual(val1: unknown, val2: unknown): boolean {
    try {
        deepStrictEqual(val1, val2);
        return true;
    } catch {
        return false;
    }
}

export namespace types {
    export function isDate(val: unknown): boolean { return val instanceof Date; }
    export function isRegExp(val: unknown): boolean { return val instanceof RegExp; }
    export function isNativeError(val: unknown): boolean { return val instanceof Error; }
    export function isArrayBuffer(val: unknown): boolean { return val instanceof ArrayBuffer; }
    export function isUint8Array(val: unknown): boolean { return val instanceof Uint8Array; }
    export function isMap(val: unknown): boolean { return val instanceof Map; }
    export function isSet(val: unknown): boolean { return val instanceof Set; }
    export function isAnyArrayBuffer(val: unknown): boolean { return val instanceof ArrayBuffer; }
    export function isBoxedPrimitive(val: unknown): boolean { return val !== null && typeof val === "object" && (val instanceof Number || val instanceof String || val instanceof Boolean); }
    export function isDataView(val: unknown): boolean { return val instanceof DataView; }
}

export function inspect(
    object: unknown,
    optionsOrShowHidden?: unknown,
    depth?: number,
    colors?: boolean
): string {
    if (object === null) return "null";
    if (object === undefined) return "undefined";
    if (typeof object === "string") {
        return "'" + (object as string).replace(/\\/g, "\\\\").replace(/'/g, "\\'").replace(/\n/g, "\\n").replace(/\r/g, "\\r").replace(/\t/g, "\\t") + "'";
    }
    if (typeof object === "number") return String(object);
    if (typeof object === "boolean") return String(object);
    if (typeof object === "symbol") return String(object);
    if (typeof object === "function") return "[Function]";
    if (object instanceof Error) return (object as Error).message;
    if (object instanceof Date) return (object as Date).toISOString();
    if (object instanceof RegExp) return (object as RegExp).toString();
    if (Array.isArray(object)) {
        const items = (object as unknown[]).map((x) => inspect(x, optionsOrShowHidden, depth, colors));
        return "[ " + items.join(", ") + " ]";
    }
    if (typeof object === "object") {
        const keys = Object.keys(object as object);
        const entries = keys.map((k) => k + ": " + inspect((object as Record<string, unknown>)[k], optionsOrShowHidden, depth, colors));
        return "{ " + entries.join(", ") + " }";
    }
    return String(object);
}

export function format(formatStr?: unknown, ...args: unknown[]): string {
    if (typeof formatStr !== "string") {
        const all = [formatStr, ...args].filter((x) => x !== undefined);
        const mapped = all.map((x) => (typeof x === "string" ? (x as string) : inspect(x)));
        return mapped.join(" ");
    }

    let str = formatStr as string;
    let argIdx = 0;
    let res = "";
    let i = 0;

    while (i < str.length) {
        if (str[i] === "%" && i + 1 < str.length) {
            const spec = str[i + 1];
            if (spec === "%") {
                res += "%";
                i += 2;
                continue;
            }
            if (argIdx < args.length) {
                const arg = args[argIdx++];
                if (spec === "s") {
                    res += String(arg);
                } else if (spec === "d" || spec === "i") {
                    res += String(parseInt(String(arg), 10));
                } else if (spec === "f") {
                    res += String(parseFloat(String(arg)));
                } else if (spec === "j") {
                    res += inspect(arg);
                } else if (spec === "o" || spec === "O") {
                    res += inspect(arg);
                } else {
                    res += "%" + spec;
                }
                i += 2;
                continue;
            }
        }
        res += str[i];
        i++;
    }

    while (argIdx < args.length) {
        const arg = args[argIdx++];
        res += " " + (typeof arg === "string" ? (arg as string) : inspect(arg));
    }

    return res;
}

export function formatWithOptions(inspectOptions: unknown, formatStr?: unknown, a1?: unknown, a2?: unknown, a3?: unknown, a4?: unknown): string {
    if (a4 !== undefined) return format(formatStr, a1, a2, a3, a4);
    if (a3 !== undefined) return format(formatStr, a1, a2, a3);
    if (a2 !== undefined) return format(formatStr, a1, a2);
    if (a1 !== undefined) return format(formatStr, a1);
    return format(formatStr);
}

export function promisify(original: Function): Function {
    return (a?: unknown, b?: unknown) => {
        return new Promise((resolve, reject) => {
            (original as (a?: unknown, b?: unknown, cb?: (err: unknown, res: unknown) => void) => void)(
                a,
                b,
                (err: unknown, result: unknown) => {
                    if (err) {
                        reject(err);
                    } else {
                        resolve(result);
                    }
                }
            );
        });
    };
}

export function callbackify(original: Function): Function {
    return (a?: unknown, b?: unknown, cb?: unknown) => {
        const callback = (typeof cb === "function" ? cb : typeof b === "function" ? b : a) as (err: unknown, res?: unknown) => void;
        Promise.resolve((original as (a?: unknown, b?: unknown) => unknown)(a, b)).then(
            (res) => callback(null, res),
            (err) => callback(err)
        );
    };
}

export function deprecate<T extends Function>(fn: T, msg: string, code?: string): T {
    const prefix = typeof code === "string" && code.length > 0 ? code + ": " : "";
    console.warn(`[DEPRECATION] ${prefix}${msg}`);
    return fn;
}

export function _extend(target: Record<string, unknown>, source: Record<string, unknown>): Record<string, unknown> {
    return Object.assign(target, source);
}

export function log(string: string): void {
    const d = new Date();
    console.log(`${d.toLocaleDateString()} ${d.toLocaleTimeString()} - ${string}`);
}

export function debug(string: string): void {
    console.error(string);
}

export function debuglog(section: string, callback?: (fn: (...args: unknown[]) => void) => void): ((...args: unknown[]) => void) & { enabled: boolean } {
    let envVal = "";
    if (typeof process !== "undefined" && typeof process.env !== "undefined" && typeof process.env.NODE_DEBUG === "string") {
        envVal = process.env.NODE_DEBUG;
    }
    const reg = new RegExp(`\\b${section}\\b`, "i");
    const enabled = reg.test(envVal);
    const fn = ((...args: unknown[]): void => {
        if (enabled) {
            let pid = 0;
            if (typeof process !== "undefined" && typeof process.pid === "number") {
                pid = process.pid;
            }
            const prefix = `${section.toUpperCase()} ${pid}: `;
            console.error(prefix + args.map(a => String(a)).join(" "));
        }
    }) as ((...args: unknown[]) => void) & { enabled: boolean };
    fn.enabled = enabled;
    if (callback) {
        callback(fn);
    }
    return fn;
}

export function inherits(ctor: Function, superCtor: Function): void {
    if (ctor === undefined || ctor === null) {
        throw new TypeError('The "ctor" argument must be of type Function.');
    }
    if (superCtor === undefined || superCtor === null) {
        throw new TypeError('The "superCtor" argument must be of type Function.');
    }
    (ctor as unknown as Record<string, unknown>).super_ = superCtor;
}

let _traceSigInt = false;
export function setTraceSigInt(enable: boolean): void {
    _traceSigInt = enable;
}

export function transferableAbortController(): AbortController {
    return new AbortController();
}

export function transferableAbortSignal(signal: AbortSignal): AbortSignal {
    if (signal === undefined || signal === null) {
        throw new TypeError('The "signal" argument must be an instance of AbortSignal');
    }
    return signal;
}

export function aborted(signal: AbortSignal, resource: object): Promise<void> {
    if (signal === undefined || signal === null) {
        throw new TypeError('The "signal" argument must be an instance of AbortSignal');
    }
    if (resource === undefined || resource === null || typeof resource !== "object") {
        throw new TypeError('The "resource" argument must be of type object');
    }
    if (signal.aborted) {
        return Promise.resolve();
    }
    return new Promise<void>((resolve) => {
        signal.addEventListener("abort", () => {
            resolve();
        }, { once: true });
    });
}

export interface CallSite {
    functionName: string;
    scriptName: string;
    scriptId: string;
    lineNumber: number;
    columnNumber: number;
    column?: number;
}

export function getCallSites(frameCount: number = 10, options?: { sourceMap?: boolean }): CallSite[] {
    const err = new Error();
    const stack = err.stack || "";
    const lines = stack.split("\n");
    const sites: CallSite[] = [];
    let idCounter = 1;

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i].trim();
        if (!line.startsWith("at ")) continue;
        const entry = line.slice(3).trim();
        if (entry.includes("getCallSites")) continue;

        let fnName = "";
        let loc = entry;
        const parenIdx = entry.indexOf("(");
        if (parenIdx !== -1 && entry.endsWith(")")) {
            fnName = entry.slice(0, parenIdx).trim();
            loc = entry.slice(parenIdx + 1, entry.length - 1).trim();
        }

        let scriptName = loc;
        let lineNum = 1;
        let colNum = 1;

        const parts = loc.split(":");
        if (parts.length >= 3) {
            colNum = parseInt(parts[parts.length - 1], 10) || 1;
            lineNum = parseInt(parts[parts.length - 2], 10) || 1;
            scriptName = parts.slice(0, parts.length - 2).join(":");
        } else if (parts.length === 2) {
            lineNum = parseInt(parts[1], 10) || 1;
            scriptName = parts[0];
        }

        sites.push({
            functionName: fnName,
            scriptName: scriptName,
            scriptId: String(idCounter++),
            lineNumber: lineNum,
            columnNumber: colNum,
            column: colNum,
        });

        if (sites.length >= frameCount) break;
    }

    if (sites.length === 0) {
        sites.push({
            functionName: "",
            scriptName: "main.ts",
            scriptId: "1",
            lineNumber: 1,
            columnNumber: 1,
            column: 1,
        });
    }

    return sites;
}

export type DiffResult = Array<[number, string]>;

export function diff(actual: string, expected: string): DiffResult {
    const aSeq: string[] = actual.split("");
    const bSeq: string[] = expected.split("");

    const n = aSeq.length;
    const m = bSeq.length;

    if (n === m) {
        let same = true;
        for (let i = 0; i < n; i++) {
            if (aSeq[i] !== bSeq[i]) {
                same = false;
                break;
            }
        }
        if (same) return [];
    }

    const max = n + m;
    const offset = max;
    const history: number[][] = [];
    const v: number[] = new Array<number>(2 * max + 1);
    for (let i = 0; i < v.length; i++) v[i] = 0;

    let foundD = -1;
    for (let d = 0; d <= max; d++) {
        const vCopy = v.slice();
        history.push(vCopy);
        for (let k = -d; k <= d; k += 2) {
            let x: number;
            if (k === -d || (k !== d && v[offset + k - 1] < v[offset + k + 1])) {
                x = v[offset + k + 1];
            } else {
                x = v[offset + k - 1] + 1;
            }
            let y = x - k;
            while (x < n && y < m && aSeq[x] === bSeq[y]) {
                x++;
                y++;
            }
            v[offset + k] = x;
            if (x >= n && y >= m) {
                foundD = d;
                break;
            }
        }
        if (foundD !== -1) break;
    }

    const result: DiffResult = [];
    let curX = n;
    let curY = m;

    for (let d = foundD; d > 0; d--) {
        const prevV = history[d];
        const k = curX - curY;
        let prevK: number;
        if (k === -d || (k !== d && prevV[offset + k - 1] < prevV[offset + k + 1])) {
            prevK = k + 1;
        } else {
            prevK = k - 1;
        }
        const prevX = prevV[offset + prevK];
        const prevY = prevX - prevK;

        while (curX > prevX && curY > prevY) {
            result.unshift([0, aSeq[curX - 1]]);
            curX--;
            curY--;
        }

        if (d > 0) {
            if (curX === prevX) {
                result.unshift([-1, bSeq[prevY]]);
                curY = prevY;
            } else {
                result.unshift([1, aSeq[prevX]]);
                curX = prevX;
            }
        }
    }

    while (curX > 0 && curY > 0) {
        result.unshift([0, aSeq[curX - 1]]);
        curX--;
        curY--;
    }

    return result;
}

export interface ParseArgsOptionConfig {
    type: "string" | "boolean";
    short?: string;
    multiple?: boolean;
    default?: string | boolean | string[] | boolean[];
}

export interface ParseArgsConfig {
    args?: string[];
    options?: Record<string, ParseArgsOptionConfig>;
    strict?: boolean;
    allowPositionals?: boolean;
    tokens?: boolean;
}

export interface ParseArgsToken {
    kind: "option" | "positional" | "option-terminator";
    index: number;
    name?: string;
    rawName?: string;
    value?: string | boolean | undefined;
    inlineValue?: boolean;
}

export interface ParseArgsResult {
    values: Record<string, unknown>;
    positionals: string[];
    tokens?: ParseArgsToken[];
}

export function parseArgs(config?: ParseArgsConfig): ParseArgsResult {
    const conf = config || {};
    const args: string[] = conf.args !== undefined ? conf.args : (typeof process !== "undefined" && Array.isArray(process.argv) ? process.argv.slice(2) : []);
    const options = conf.options || {};
    const strict = conf.strict !== false;
    const allowPositionals = conf.allowPositionals !== undefined ? conf.allowPositionals : !strict;
    const returnTokens = conf.tokens === true;

    const shortToLong: Record<string, string> = {};
    for (const optName in options) {
        const opt = options[optName];
        if (opt.short) {
            shortToLong[opt.short] = optName;
        }
    }

    const values: Record<string, unknown> = {};
    for (const optName in options) {
        const opt = options[optName];
        if (opt.default !== undefined) {
            values[optName] = opt.default;
        } else if (opt.multiple) {
            values[optName] = [];
        }
    }

    const positionals: string[] = [];
    const tokens: ParseArgsToken[] = [];
    let parsingOptions = true;
    let i = 0;

    while (i < args.length) {
        const arg = args[i];

        if (parsingOptions && arg === "--") {
            parsingOptions = false;
            if (returnTokens) {
                tokens.push({ kind: "option-terminator", index: i });
            }
            i++;
            continue;
        }

        if (parsingOptions && arg.startsWith("--") && arg.length > 2) {
            const eqIdx = arg.indexOf("=");
            let rawOptName: string;
            let inlineVal: string | undefined = undefined;
            if (eqIdx !== -1) {
                rawOptName = arg.slice(2, eqIdx);
                inlineVal = arg.slice(eqIdx + 1);
            } else {
                rawOptName = arg.slice(2);
            }

            const optDef = options[rawOptName];
            if (strict && !optDef) {
                throw new TypeError(`Unknown option '--${rawOptName}'`);
            }

            const optType = optDef ? optDef.type : (inlineVal !== undefined ? "string" : "boolean");
            let val: string | boolean;

            if (optType === "boolean") {
                if (inlineVal !== undefined) {
                    if (inlineVal === "true") val = true;
                    else if (inlineVal === "false") val = false;
                    else if (strict) throw new TypeError(`Option '--${rawOptName}' does not take a value`);
                    else val = true;
                } else {
                    val = true;
                }
            } else {
                if (inlineVal !== undefined) {
                    val = inlineVal;
                } else if (i + 1 < args.length && !args[i + 1].startsWith("-")) {
                    i++;
                    val = args[i];
                } else if (strict) {
                    throw new TypeError(`Option '--${rawOptName}' requires a value`);
                } else {
                    val = "";
                }
            }

            if (optDef && optDef.multiple) {
                const arr = (values[rawOptName] as unknown[]) || [];
                arr.push(val);
                values[rawOptName] = arr;
            } else {
                values[rawOptName] = val;
            }

            if (returnTokens) {
                tokens.push({
                    kind: "option",
                    name: rawOptName,
                    rawName: arg.slice(0, eqIdx !== -1 ? eqIdx : arg.length),
                    index: i,
                    value: val,
                    inlineValue: inlineVal !== undefined,
                });
            }
            i++;
            continue;
        }

        if (parsingOptions && arg.startsWith("-") && arg.length > 1) {
            const shortStr = arg.slice(1);
            const eqIdx = shortStr.indexOf("=");
            if (eqIdx !== -1) {
                const s = shortStr.slice(0, eqIdx);
                const inlineVal = shortStr.slice(eqIdx + 1);
                const longName = shortToLong[s] || s;
                const optDef = options[longName];
                if (strict && !optDef) throw new TypeError(`Unknown option '-${s}'`);
                const val = optDef && optDef.type === "boolean" ? (inlineVal === "true") : inlineVal;
                if (optDef && optDef.multiple) {
                    const arr = (values[longName] as unknown[]) || [];
                    arr.push(val);
                    values[longName] = arr;
                } else {
                    values[longName] = val;
                }
                if (returnTokens) {
                    tokens.push({
                        kind: "option",
                        name: longName,
                        rawName: `-${s}`,
                        index: i,
                        value: val,
                        inlineValue: true,
                    });
                }
                i++;
                continue;
            }

            for (let cIdx = 0; cIdx < shortStr.length; cIdx++) {
                const s = shortStr[cIdx];
                const longName = shortToLong[s] || s;
                const optDef = options[longName];
                if (strict && !optDef) throw new TypeError(`Unknown option '-${s}'`);

                const optType = optDef ? optDef.type : "boolean";
                if (optType === "boolean") {
                    if (optDef && optDef.multiple) {
                        const arr = (values[longName] as unknown[]) || [];
                        arr.push(true);
                        values[longName] = arr;
                    } else {
                        values[longName] = true;
                    }
                    if (returnTokens) {
                        tokens.push({ kind: "option", name: longName, rawName: `-${s}`, index: i, value: true });
                    }
                } else {
                    let val: string;
                    if (cIdx + 1 < shortStr.length) {
                        val = shortStr.slice(cIdx + 1);
                        cIdx = shortStr.length;
                    } else if (i + 1 < args.length) {
                        i++;
                        val = args[i];
                    } else if (strict) {
                        throw new TypeError(`Option '-${s}' requires a value`);
                    } else {
                        val = "";
                    }
                    if (optDef && optDef.multiple) {
                        const arr = (values[longName] as unknown[]) || [];
                        arr.push(val);
                        values[longName] = arr;
                    } else {
                        values[longName] = val;
                    }
                    if (returnTokens) {
                        tokens.push({ kind: "option", name: longName, rawName: `-${s}`, index: i, value: val });
                    }
                    break;
                }
            }
            i++;
            continue;
        }

        if (strict && !allowPositionals) {
            throw new TypeError(`Unexpected positional argument: '${arg}'`);
        }
        positionals.push(arg);
        if (returnTokens) {
            tokens.push({ kind: "positional", index: i, value: arg });
        }
        i++;
    }

    const res: ParseArgsResult = { values, positionals };
    if (returnTokens) {
        res.tokens = tokens;
    }
    return res;
}

export function stripVTControlCharacters(str: string): string {
    let res = "";
    let i = 0;
    while (i < str.length) {
        if (str.charCodeAt(i) === 27 && i + 1 < str.length && str[i + 1] === "[") {
            i += 2;
            while (i < str.length && ((str.charCodeAt(i) >= 48 && str.charCodeAt(i) <= 57) || str[i] === ";")) {
                i++;
            }
            if (i < str.length) {
                i++;
            }
        } else {
            res += str[i];
            i++;
        }
    }
    return res;
}

export function toUSVString(string: string): string {
    let res = "";
    for (let i = 0; i < string.length; i++) {
        const c = string.charCodeAt(i);
        if (c >= 0xD800 && c <= 0xDBFF) {
            if (i + 1 < string.length) {
                const next = string.charCodeAt(i + 1);
                if (next >= 0xDC00 && next <= 0xDFFF) {
                    res += string[i] + string[i + 1];
                    i++;
                    continue;
                }
            }
            res += "\uFFFD";
        } else if (c >= 0xDC00 && c <= 0xDFFF) {
            res += "\uFFFD";
        } else {
            res += string[i];
        }
    }
    return res;
}

export function styleText(formatName: string | string[], text: string, options?: unknown): string {
    let open = "";
    let close = "";
    if (typeof formatName === "string") {
        const name = formatName as string;
        if (name === "bold") { open = "\x1b[1m"; close = "\x1b[22m"; }
        else if (name === "italic") { open = "\x1b[3m"; close = "\x1b[23m"; }
        else if (name === "underline") { open = "\x1b[4m"; close = "\x1b[24m"; }
        else if (name === "red") { open = "\x1b[31m"; close = "\x1b[39m"; }
        else if (name === "green") { open = "\x1b[32m"; close = "\x1b[39m"; }
        else if (name === "yellow") { open = "\x1b[33m"; close = "\x1b[39m"; }
        else if (name === "blue") { open = "\x1b[34m"; close = "\x1b[39m"; }
        else if (name === "cyan") { open = "\x1b[36m"; close = "\x1b[39m"; }
        else if (name === "white") { open = "\x1b[37m"; close = "\x1b[39m"; }
        return open + text + close;
    }
    return text;
}

let _sysErrorMap: Map<number, [string, string]> | null = null;

export function getSystemErrorMap(): Map<number, [string, string]> {
    if (_sysErrorMap !== null) {
        return _sysErrorMap;
    }
    const map = new Map<number, [string, string]>();
    map.set(-1, ["EPERM", "Operation not permitted"]);
    map.set(-2, ["ENOENT", "No such file or directory"]);
    map.set(-3, ["ESRCH", "No such process"]);
    map.set(-4, ["EINTR", "Interrupted system call"]);
    map.set(-5, ["EIO", "I/O error"]);
    map.set(-9, ["EBADF", "Bad file descriptor"]);
    map.set(-12, ["ENOMEM", "Cannot allocate memory"]);
    map.set(-13, ["EACCES", "Permission denied"]);
    map.set(-22, ["EINVAL", "Invalid argument"]);
    _sysErrorMap = map;
    return _sysErrorMap;
}

export function getSystemErrorName(err: number): string {
    const entry = getSystemErrorMap().get(err);
    if (entry) {
        return entry[0];
    }
    return "";
}

export function getSystemErrorMessage(err: number): string {
    const entry = getSystemErrorMap().get(err);
    if (entry) {
        return entry[1];
    }
    return "";
}

export function parseEnv(content: string): Record<string, string> {
    const result: Record<string, string> = {};
    const lines = content.split("\n");
    for (let i = 0; i < lines.length; i++) {
        let line = lines[i].trim();
        if (line.length === 0 || line.startsWith("#")) continue;
        const eqIdx = line.indexOf("=");
        if (eqIdx >= 0) {
            const key = line.slice(0, eqIdx).trim();
            let val = line.slice(eqIdx + 1).trim();
            if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
                val = val.slice(1, val.length - 1);
            }
            result[key] = val;
        }
    }
    return result;
}

export class TextDecoder {
    readonly encoding: string = "utf-8";
    readonly fatal: boolean = false;
    readonly ignoreBOM: boolean = false;
    constructor(label?: string, options?: { fatal?: boolean; ignoreBOM?: boolean }) {}
    decode(input?: Uint8Array, options?: { stream?: boolean }): string {
        return "";
    }
}

export class TextEncoder {
    readonly encoding: string = "utf-8";
    encode(input?: string): Uint8Array {
        return new Uint8Array(0);
    }
}

export default {
    MIMEParams,
    MIMEType,
    isArray,
    isBoolean,
    isNull,
    isNullOrUndefined,
    isNumber,
    isString,
    isSymbol,
    isUndefined,
    isRegExp,
    isObject,
    isDate,
    isError,
    isFunction,
    isPrimitive,
    isBuffer,
    isDeepStrictEqual,
    types,
    inspect,
    format,
    formatWithOptions,
    deprecate,
    callbackify,
    promisify,
    toUSVString,
    stripVTControlCharacters,
    styleText,
    getSystemErrorMap,
    getSystemErrorName,
    getSystemErrorMessage,
    parseEnv,
    diff,
    parseArgs,
    debuglog,
    debug,
    log,
    inherits,
    getCallSites,
    setTraceSigInt,
    transferableAbortController,
    transferableAbortSignal,
    aborted,
    TextDecoder,
    TextEncoder,
};
