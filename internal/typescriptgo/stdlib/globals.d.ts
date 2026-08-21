interface ProcessEnv {
    [key: string]: string | undefined;
}

interface Process {
    argv: string[];
    env: ProcessEnv;
    exit(code?: number): void;
    cwd(): string;
}

declare var process: Process;

interface Console {
    log(...args: any[]): void;
    info(...args: any[]): void;
    warn(...args: any[]): void;
    error(...args: any[]): void;
    debug(...args: any[]): void;
    assert(condition?: boolean, ...data: any[]): void;
    clear(): void;
    count(label?: string): void;
    countReset(label?: string): void;
    time(label?: string): void;
    timeLog(label?: string, ...data: any[]): void;
    timeEnd(label?: string): void;
    trace(...data: any[]): void;
    dir(item?: any, options?: any): void;
    dirxml(...data: any[]): void;
    table(tabularData?: any, properties?: string[]): void;
    group(...data: any[]): void;
    groupCollapsed(...data: any[]): void;
    groupEnd(): void;
}

declare var console: Console;

interface ConsoleConstructor {
    prototype: Console;
    new(): Console;
}

declare var Console: ConsoleConstructor;

interface Performance {
    now(): number;
}

declare var performance: Performance;

declare function btoa(data: string): string;
declare function atob(data: string): string;
declare function queueMicrotask(callback: () => void): void;

interface BigInt {
    toString(radix?: number): string;
}

interface BigIntConstructor {
    (value?: any): bigint;
    new(value?: any): BigInt;
    asIntN(bits: number, int: bigint): bigint;
    asUintN(bits: number, int: bigint): bigint;
}

declare var BigInt: BigIntConstructor;

interface RegExp {
    readonly source: string;
    readonly flags: string;
    test(string: string): boolean;
    exec(string: string): string[] | null;
}

interface RegExpConstructor {
    new(pattern: RegExp | string, flags?: string): RegExp;
    (pattern: RegExp | string, flags?: string): RegExp;
}

declare var RegExp: RegExpConstructor;

interface String {
    match(matcher: RegExp | string): string[] | null;
    search(searcher: RegExp | string): number;
    replace(searchValue: RegExp | string, replaceValue: string): string;
}

interface Symbol {
    readonly description: string | undefined;
    toString(): string;
    valueOf(): symbol;
}

interface SymbolConstructor {
    (description?: string | number): symbol;
    readonly iterator: symbol;
    readonly asyncIterator: symbol;
    readonly hasInstance: symbol;
    readonly isConcatSpreadable: symbol;
    readonly match: symbol;
    readonly replace: symbol;
    readonly search: symbol;
    readonly species: symbol;
    readonly split: symbol;
    readonly toPrimitive: symbol;
    readonly toStringTag: symbol;
    readonly unscopables: symbol;
    for(key: string): symbol;
    keyFor(sym: symbol): string | undefined;
}

declare var Symbol: SymbolConstructor;

interface TemplateStringsArray extends Array<string> {
    readonly raw: readonly string[];
}

interface Error {
    name: string;
    message: string;
    stack?: string;
    toString(): string;
}

interface ErrorConstructor {
    new(message?: string): Error;
    (message?: string): Error;
    readonly prototype: Error;
}

declare var Error: ErrorConstructor;

interface TypeError extends Error {}
interface TypeErrorConstructor {
    new(message?: string): TypeError;
    (message?: string): TypeError;
    readonly prototype: TypeError;
}
declare var TypeError: TypeErrorConstructor;

interface RangeError extends Error {}
interface RangeErrorConstructor {
    new(message?: string): RangeError;
    (message?: string): RangeError;
    readonly prototype: RangeError;
}
declare var RangeError: RangeErrorConstructor;

interface SyntaxError extends Error {}
interface SyntaxErrorConstructor {
    new(message?: string): SyntaxError;
    (message?: string): SyntaxError;
    readonly prototype: SyntaxError;
}
declare var SyntaxError: SyntaxErrorConstructor;

interface Date {
    toString(): string;
    toISOString(): string;
    toUTCString(): string;
    getTime(): number;
    getFullYear(): number;
    getMonth(): number;
    getDate(): number;
    getDay(): number;
    getHours(): number;
    getMinutes(): number;
    getSeconds(): number;
    getMilliseconds(): number;
    getTimezoneOffset(): number;
}

interface DateConstructor {
    new(): Date;
    new(value: number | string | Date): Date;
    (value?: any): string;
    now(): number;
    parse(s: string): number;
    UTC(year: number, monthIndex?: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): number;
}

declare var Date: DateConstructor;

interface ArrayBuffer {
    readonly byteLength: number;
    slice(begin?: number, end?: number): ArrayBuffer;
}

interface ArrayBufferConstructor {
    new(byteLength: number): ArrayBuffer;
    readonly prototype: ArrayBuffer;
    isView(arg: any): boolean;
}

declare var ArrayBuffer: ArrayBufferConstructor;

interface ArrayLike<T> {
    readonly length: number;
    readonly [n: number]: T;
}

interface Uint8Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint8Array, offset?: number): void;
    subarray(begin?: number, end?: number): Uint8Array;
    slice(begin?: number, end?: number): Uint8Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Uint8ArrayConstructor {
    new(length: number): Uint8Array;
    new(array: ArrayLike<number> | Array<number>): Uint8Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Uint8Array;
    readonly prototype: Uint8Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Uint8Array;
    of(...items: number[]): Uint8Array;
}

declare var Uint8Array: Uint8ArrayConstructor;

interface Int32Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Int32Array, offset?: number): void;
    subarray(begin?: number, end?: number): Int32Array;
    slice(begin?: number, end?: number): Int32Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Int32ArrayConstructor {
    new(length: number): Int32Array;
    new(array: ArrayLike<number> | Array<number>): Int32Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Int32Array;
    readonly prototype: Int32Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Int32Array;
    of(...items: number[]): Int32Array;
}

declare var Int32Array: Int32ArrayConstructor;

interface Float64Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Float64Array, offset?: number): void;
    subarray(begin?: number, end?: number): Float64Array;
    slice(begin?: number, end?: number): Float64Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Float64ArrayConstructor {
    new(length: number): Float64Array;
    new(array: ArrayLike<number> | Array<number>): Float64Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Float64Array;
    readonly prototype: Float64Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Float64Array;
    of(...items: number[]): Float64Array;
}

declare var Float64Array: Float64ArrayConstructor;

declare function setTimeout(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
declare function clearTimeout(id: number | undefined): void;
declare function setInterval(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
declare function clearInterval(id: number | undefined): void;
declare function setImmediate(callback: (...args: any[]) => void, ...args: any[]): number;
declare function clearImmediate(id: number | undefined): void;
