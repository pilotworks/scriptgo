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
}

declare var console: Console;

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

