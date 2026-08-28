interface ProcessEnv {
    [key: string]: string;
}

interface ProcessVersions {
    [key: string]: string;
}

interface Process {
    argv: string[];
    env: ProcessEnv;
    exit(code?: number): void;
    cwd(): string;
    pid: number;
    ppid: number;
    version: string;
    versions: ProcessVersions;
    platform: string;
    arch: string;
    uptime(): number;
    hrtime(time?: [number, number]): [number, number];
    nextTick(callback: (...args: unknown[]) => void, ...args: unknown[]): void;
}

declare var process: Process;

interface Console {
    log(...args: unknown[]): void;
    info(...args: unknown[]): void;
    warn(...args: unknown[]): void;
    error(...args: unknown[]): void;
    debug(...args: unknown[]): void;
    assert(condition?: unknown, ...data: unknown[]): void;
    clear(): void;
    count(label?: string): void;
    countReset(label?: string): void;
    time(label?: string): void;
    timeLog(label?: string, ...data: unknown[]): void;
    timeEnd(label?: string): void;
    trace(...data: unknown[]): void;
    dir(item?: unknown, options?: unknown): void;
    dirxml(...data: unknown[]): void;
    table(tabularData?: unknown, properties?: string[]): void;
    group(...data: unknown[]): void;
    groupCollapsed(...data: unknown[]): void;
    groupEnd(): void;
    profile(label?: string): void;
    profileEnd(label?: string): void;
    timeStamp(label?: string): void;
}

declare var console: Console;

interface ConsoleConstructorOptions {
    stdout?: unknown;
    stderr?: unknown;
    ignoreErrors?: boolean;
    colorMode?: boolean | "auto";
    inspectOptions?: unknown;
    groupIndentation?: number;
}

interface ConsoleConstructor {
    prototype: Console;
    new(stdout?: unknown, stderr?: unknown, ignoreErrors?: boolean): Console;
    new(options?: ConsoleConstructorOptions | unknown): Console;
}

declare var Console: ConsoleConstructor;

interface Performance {
    now(): number;
}

declare var performance: Performance;

declare function btoa(data: string): string;
declare function atob(data: string): string;
declare function queueMicrotask(callback: () => void): void;
declare function structuredClone<T>(value: T): T;

interface BigInt {
    toString(radix?: number): string;
}

interface BigIntConstructor {
    (value?: unknown): bigint;
    new(value?: unknown): BigInt;
    asIntN(bits: number, int: bigint): bigint;
    asUintN(bits: number, int: bigint): bigint;
}

declare var BigInt: BigIntConstructor;

interface RegExpExecArray extends Array<string> {
    index: number;
    input: string;
    groups?: { [key: string]: string };
    [index: number]: string;
}

interface RegExpMatchArray extends Array<string> {
    index?: number;
    input?: string;
    groups?: { [key: string]: string };
    [index: number]: string;
}

interface RegExp {
    readonly source: string;
    readonly flags: string;
    lastIndex: number;
    test(string: string): boolean;
    exec(string: string): RegExpExecArray | null;
}

interface RegExpConstructor {
    new(pattern: RegExp | string, flags?: string): RegExp;
    (pattern: RegExp | string, flags?: string): RegExp;
}

declare var RegExp: RegExpConstructor;

interface ObjectConstructor {
    getPrototypeOf(o: unknown): object;
    getOwnPropertyNames(o: unknown): string[];
    getOwnPropertySymbols(o: unknown): symbol[];
    getOwnPropertyDescriptor(o: unknown, p: string | symbol): PropertyDescriptor | undefined;
    getOwnPropertyDescriptors(o: unknown): Record<string, PropertyDescriptor>;
    keys(o: unknown): string[];
    values(o: unknown): unknown[];
    entries(o: unknown): [string, unknown][];
    assign<T extends Record<string, unknown>>(target: T, ...sources: unknown[]): T;
    is(value1: unknown, value2: unknown): boolean;
    hasOwn(o: unknown, v: string | symbol): boolean;
    create(o: object | null, properties?: unknown): object;
    freeze<T>(a: T): T;
    seal<T>(a: T): T;
    preventExtensions<T>(a: T): T;
    isFrozen(o: unknown): boolean;
    isSealed(o: unknown): boolean;
    isExtensible(o: unknown): boolean;
    setPrototypeOf(o: unknown, proto: object | null): unknown;
}

declare var Object: ObjectConstructor;

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

interface TypeError extends Error { }
interface TypeErrorConstructor {
    new(message?: string): TypeError;
    (message?: string): TypeError;
    readonly prototype: TypeError;
}
declare var TypeError: TypeErrorConstructor;

interface RangeError extends Error { }
interface RangeErrorConstructor {
    new(message?: string): RangeError;
    (message?: string): RangeError;
    readonly prototype: RangeError;
}
declare var RangeError: RangeErrorConstructor;

interface SyntaxError extends Error { }
interface SyntaxErrorConstructor {
    new(message?: string): SyntaxError;
    (message?: string): SyntaxError;
    readonly prototype: SyntaxError;
}
declare var SyntaxError: SyntaxErrorConstructor;

interface Date {
    toString(): string;
    toDateString(): string;
    toTimeString(): string;
    toISOString(): string;
    toUTCString(): string;
    toJSON(key?: unknown): string;
    toLocaleString(locales?: unknown, options?: unknown): string;
    toLocaleDateString(locales?: unknown, options?: unknown): string;
    toLocaleTimeString(locales?: unknown, options?: unknown): string;
    toTemporalInstant(): unknown;
    valueOf(): number;
    getTime(): number;
    getFullYear(): number;
    getUTCFullYear(): number;
    getMonth(): number;
    getUTCMonth(): number;
    getDate(): number;
    getUTCDate(): number;
    getDay(): number;
    getUTCDay(): number;
    getHours(): number;
    getUTCHours(): number;
    getMinutes(): number;
    getUTCMinutes(): number;
    getSeconds(): number;
    getUTCSeconds(): number;
    getMilliseconds(): number;
    getUTCMilliseconds(): number;
    getTimezoneOffset(): number;
    setTime(time: number): number;
    setMilliseconds(ms: number): number;
    setUTCMilliseconds(ms: number): number;
    setSeconds(sec: number, ms?: number): number;
    setUTCSeconds(sec: number, ms?: number): number;
    setMinutes(min: number, sec?: number, ms?: number): number;
    setUTCMinutes(min: number, sec?: number, ms?: number): number;
    setHours(hours: number, min?: number, sec?: number, ms?: number): number;
    setUTCHours(hours: number, min?: number, sec?: number, ms?: number): number;
    setDate(date: number): number;
    setUTCDate(date: number): number;
    setMonth(month: number, date?: number): number;
    setUTCMonth(month: number, date?: number): number;
    setFullYear(year: number, month?: number, date?: number): number;
    setUTCFullYear(year: number, month?: number, date?: number): number;
}

interface DateConstructor {
    new(): Date;
    new(value: number | string | Date): Date;
    (value?: unknown): string;
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
    isView(arg: unknown): boolean;
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

interface Int8Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Int8Array, offset?: number): void;
    subarray(begin?: number, end?: number): Int8Array;
    slice(begin?: number, end?: number): Int8Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Int8ArrayConstructor {
    new(length: number): Int8Array;
    new(array: ArrayLike<number> | Array<number>): Int8Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Int8Array;
    readonly prototype: Int8Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Int8Array;
    of(...items: number[]): Int8Array;
}
declare var Int8Array: Int8ArrayConstructor;

interface Uint8ClampedArray {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint8ClampedArray, offset?: number): void;
    subarray(begin?: number, end?: number): Uint8ClampedArray;
    slice(begin?: number, end?: number): Uint8ClampedArray;
    fill(value: number, start?: number, end?: number): this;
}

interface Uint8ClampedArrayConstructor {
    new(length: number): Uint8ClampedArray;
    new(array: ArrayLike<number> | Array<number>): Uint8ClampedArray;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Uint8ClampedArray;
    readonly prototype: Uint8ClampedArray;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Uint8ClampedArray;
    of(...items: number[]): Uint8ClampedArray;
}
declare var Uint8ClampedArray: Uint8ClampedArrayConstructor;

interface Int16Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Int16Array, offset?: number): void;
    subarray(begin?: number, end?: number): Int16Array;
    slice(begin?: number, end?: number): Int16Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Int16ArrayConstructor {
    new(length: number): Int16Array;
    new(array: ArrayLike<number> | Array<number>): Int16Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Int16Array;
    readonly prototype: Int16Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Int16Array;
    of(...items: number[]): Int16Array;
}
declare var Int16Array: Int16ArrayConstructor;

interface Uint16Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint16Array, offset?: number): void;
    subarray(begin?: number, end?: number): Uint16Array;
    slice(begin?: number, end?: number): Uint16Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Uint16ArrayConstructor {
    new(length: number): Uint16Array;
    new(array: ArrayLike<number> | Array<number>): Uint16Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Uint16Array;
    readonly prototype: Uint16Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Uint16Array;
    of(...items: number[]): Uint16Array;
}
declare var Uint16Array: Uint16ArrayConstructor;

interface Uint32Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint32Array, offset?: number): void;
    subarray(begin?: number, end?: number): Uint32Array;
    slice(begin?: number, end?: number): Uint32Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Uint32ArrayConstructor {
    new(length: number): Uint32Array;
    new(array: ArrayLike<number> | Array<number>): Uint32Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Uint32Array;
    readonly prototype: Uint32Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Uint32Array;
    of(...items: number[]): Uint32Array;
}
declare var Uint32Array: Uint32ArrayConstructor;

interface Float32Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Float32Array, offset?: number): void;
    subarray(begin?: number, end?: number): Float32Array;
    slice(begin?: number, end?: number): Float32Array;
    fill(value: number, start?: number, end?: number): this;
}

interface Float32ArrayConstructor {
    new(length: number): Float32Array;
    new(array: ArrayLike<number> | Array<number>): Float32Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): Float32Array;
    readonly prototype: Float32Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<number> | Array<number>): Float32Array;
    of(...items: number[]): Float32Array;
}
declare var Float32Array: Float32ArrayConstructor;

interface BigInt64Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: bigint;
    set(array: ArrayLike<bigint> | Array<bigint> | BigInt64Array, offset?: number): void;
    subarray(begin?: number, end?: number): BigInt64Array;
    slice(begin?: number, end?: number): BigInt64Array;
    fill(value: bigint, start?: number, end?: number): this;
}

interface BigInt64ArrayConstructor {
    new(length: number): BigInt64Array;
    new(array: ArrayLike<bigint> | Array<bigint>): BigInt64Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): BigInt64Array;
    readonly prototype: BigInt64Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<bigint> | Array<bigint>): BigInt64Array;
    of(...items: bigint[]): BigInt64Array;
}
declare var BigInt64Array: BigInt64ArrayConstructor;

interface BigUint64Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    readonly BYTES_PER_ELEMENT: number;
    [index: number]: bigint;
    set(array: ArrayLike<bigint> | Array<bigint> | BigUint64Array, offset?: number): void;
    subarray(begin?: number, end?: number): BigUint64Array;
    slice(begin?: number, end?: number): BigUint64Array;
    fill(value: bigint, start?: number, end?: number): this;
}

interface BigUint64ArrayConstructor {
    new(length: number): BigUint64Array;
    new(array: ArrayLike<bigint> | Array<bigint>): BigUint64Array;
    new(buffer: ArrayBuffer, byteOffset?: number, length?: number): BigUint64Array;
    readonly prototype: BigUint64Array;
    readonly BYTES_PER_ELEMENT: number;
    from(arrayLike: ArrayLike<bigint> | Array<bigint>): BigUint64Array;
    of(...items: bigint[]): BigUint64Array;
}
declare var BigUint64Array: BigUint64ArrayConstructor;

interface DataView {
    readonly buffer: ArrayBuffer;
    readonly byteLength: number;
    readonly byteOffset: number;
    getInt8(byteOffset: number): number;
    setUint8(byteOffset: number, value: number): void;
    getUint8(byteOffset: number): number;
    setInt8(byteOffset: number, value: number): void;
    getInt16(byteOffset: number, littleEndian?: boolean): number;
    setUint16(byteOffset: number, value: number, littleEndian?: boolean): void;
    getUint16(byteOffset: number, littleEndian?: boolean): number;
    setInt16(byteOffset: number, value: number, littleEndian?: boolean): void;
    getInt32(byteOffset: number, littleEndian?: boolean): number;
    setUint32(byteOffset: number, value: number, littleEndian?: boolean): void;
    getUint32(byteOffset: number, littleEndian?: boolean): number;
    setInt32(byteOffset: number, value: number, littleEndian?: boolean): void;
    getFloat32(byteOffset: number, littleEndian?: boolean): number;
    setFloat32(byteOffset: number, value: number, littleEndian?: boolean): void;
    getFloat64(byteOffset: number, littleEndian?: boolean): number;
    setFloat64(byteOffset: number, value: number, littleEndian?: boolean): void;
    getBigInt64(byteOffset: number, littleEndian?: boolean): bigint;
    setBigInt64(byteOffset: number, value: bigint, littleEndian?: boolean): void;
    getBigUint64(byteOffset: number, littleEndian?: boolean): bigint;
    setBigUint64(byteOffset: number, value: bigint, littleEndian?: boolean): void;
}

interface DataViewConstructor {
    new(buffer: ArrayBuffer, byteOffset?: number, byteLength?: number): DataView;
    readonly prototype: DataView;
}
declare var DataView: DataViewConstructor;

interface Map<K = unknown, V = unknown> {
    readonly size: number;
    get(key: K): V;
    set(key: K, value: V): this;
    has(key: K): boolean;
    delete(key: K): boolean;
    clear(): void;
    forEach(callbackfn: (value: V, key: K, map: Map<K, V>) => void): void;
    keys(): K[];
    values(): V[];
    entries(): [K, V][];
}

interface MapConstructor {
    new <K = unknown, V = unknown>(): Map<K, V>;
    new <K = unknown, V = unknown>(entries?: readonly (readonly [K, V])[] | null): Map<K, V>;
    readonly prototype: Map<unknown, unknown>;
}
declare var Map: MapConstructor;

interface Set<T = unknown> {
    readonly size: number;
    add(value: T): this;
    has(value: T): boolean;
    delete(value: T): boolean;
    clear(): void;
    forEach(callbackfn: (value: T, value2: T, set: Set<T>) => void): void;
    keys(): T[];
    values(): T[];
    entries(): [T, T][];
}

interface SetConstructor {
    new <T = unknown>(): Set<T>;
    new <T = unknown>(values?: readonly T[] | null): Set<T>;
    readonly prototype: Set<unknown>;
}
declare var Set: SetConstructor;

interface TextEncoderEncodeIntoResult {
    read: number;
    written: number;
}

interface TextEncoder {
    readonly encoding: string;
    encode(input?: string): Uint8Array;
    encodeInto(source: string, destination: Uint8Array): TextEncoderEncodeIntoResult;
}

interface TextEncoderConstructor {
    new(): TextEncoder;
    readonly prototype: TextEncoder;
}
declare var TextEncoder: TextEncoderConstructor;

interface TextDecoderOptions {
    fatal?: boolean;
    ignoreBOM?: boolean;
}

interface TextDecodeOptions {
    stream?: boolean;
}

interface TextDecoder {
    readonly encoding: string;
    readonly fatal: boolean;
    readonly ignoreBOM: boolean;
    decode(input?: ArrayBufferView | ArrayBuffer, options?: TextDecodeOptions): string;
}

interface TextDecoderConstructor {
    new(label?: string, options?: TextDecoderOptions): TextDecoder;
    readonly prototype: TextDecoder;
}
declare var TextDecoder: TextDecoderConstructor;

interface Buffer extends Uint8Array {
    readonly length: number;
    readonly byteLength: number;
    readonly byteOffset: number;
    readonly buffer: ArrayBuffer;
    [index: number]: number;
    set(array: ArrayLike<number> | Array<number> | Uint8Array, offset?: number): void;
    toString(encoding?: string, start?: number, end?: number): string;
    subarray(begin?: number, end?: number): Buffer;
    slice(begin?: number, end?: number): Buffer;
    copy(target: Buffer | Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number;
    fill(value: number, start?: number, end?: number): this;
    equals(other: Buffer | Uint8Array): boolean;
    compare(other: Buffer | Uint8Array): number;
    indexOf(value: string | number | Uint8Array, byteOffset?: number, encoding?: string): number;

    readUInt8(offset: number): number;
    writeUInt8(value: number, offset: number): number;
    readInt8(offset: number): number;
    writeInt8(value: number, offset: number): number;
    readUInt16LE(offset: number): number;
    readUInt16BE(offset: number): number;
    writeUInt16LE(value: number, offset: number): number;
    writeUInt16BE(value: number, offset: number): number;
    readInt16LE(offset: number): number;
    readInt16BE(offset: number): number;
    writeInt16LE(value: number, offset: number): number;
    writeInt16BE(value: number, offset: number): number;
    readUInt32LE(offset: number): number;
    readUInt32BE(offset: number): number;
    writeUInt32LE(value: number, offset: number): number;
    writeUInt32BE(value: number, offset: number): number;
    readInt32LE(offset: number): number;
    readInt32BE(offset: number): number;
    writeInt32LE(value: number, offset: number): number;
    writeInt32BE(value: number, offset: number): number;
    readFloatLE(offset: number): number;
    readFloatBE(offset: number): number;
    writeFloatLE(value: number, offset: number): number;
    writeFloatBE(value: number, offset: number): number;
    readDoubleLE(offset: number): number;
    readDoubleBE(offset: number): number;
    writeDoubleLE(value: number, offset: number): number;
    writeDoubleBE(value: number, offset: number): number;
    readBigInt64LE(offset: number): bigint;
    readBigInt64BE(offset: number): bigint;
    writeBigInt64LE(value: bigint, offset: number): number;
    writeBigInt64BE(value: bigint, offset: number): number;
    readBigUInt64LE(offset: number): bigint;
    readBigUInt64BE(offset: number): bigint;
    writeBigUInt64LE(value: bigint, offset: number): number;
    writeBigUInt64BE(value: bigint, offset: number): number;
}

interface BufferConstructor {
    alloc(size: number, fill?: number | string): Buffer;
    allocUnsafe(size: number): Buffer;
    from(str: string, encoding?: string): Buffer;
    from(array: ArrayLike<number> | Array<number> | Uint8Array | ArrayBuffer): Buffer;
    concat(list: (Buffer | Uint8Array)[], totalLength?: number): Buffer;
    isBuffer(obj: unknown): boolean;
    byteLength(string: string, encoding?: string): number;
    readonly prototype: Buffer;
}

declare var Buffer: BufferConstructor;

declare function setTimeout(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number;
declare function clearTimeout(id: number | undefined): void;
declare function setInterval(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number;
declare function clearInterval(id: number | undefined): void;
declare function setImmediate(callback: (...args: unknown[]) => void, ...args: unknown[]): number;
declare function clearImmediate(id: number | undefined): void;

interface URLSearchParams {
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    getAll(name: string): string[];
    has(name: string): boolean;
    set(name: string, value: string): void;
    sort(): void;
    toString(): string;
    readonly size: number;
}

interface URLSearchParamsConstructor {
    new(init?: string): URLSearchParams;
    readonly prototype: URLSearchParams;
}

declare var URLSearchParams: URLSearchParamsConstructor;

interface URL {
    href: string;
    origin: string;
    protocol: string;
    username: string;
    password: string;
    host: string;
    hostname: string;
    port: string;
    pathname: string;
    search: string;
    searchParams: URLSearchParams;
    hash: string;
    toString(): string;
    toJSON(): string;
}

interface URLConstructor {
    new(url: string, base?: string | URL): URL;
    readonly prototype: URL;
    canParse(url: string, base?: string | URL): boolean;
}

declare var URL: URLConstructor;

interface Headers {
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    has(name: string): boolean;
    set(name: string, value: string): void;
    forEach(callback: (value: string, name: string, parent: Headers) => void): void;
    entries(): [string, string][];
    keys(): string[];
    values(): string[];
}

interface HeadersConstructor {
    new(init?: Record<string, string> | [string, string][] | Headers): Headers;
    readonly prototype: Headers;
}

declare var Headers: HeadersConstructor;

interface RequestInit {
    method?: string;
    headers?: Headers | Record<string, string> | [string, string][];
    body?: string | null;
}

interface Request {
    readonly url: string;
    readonly method: string;
    readonly headers: Headers;
    readonly body: string | null;
}

interface RequestConstructor {
    new(input: string | Request, init?: RequestInit): Request;
    readonly prototype: Request;
}

declare var Request: RequestConstructor;

interface ResponseInit {
    status?: number;
    statusText?: string;
    headers?: Headers | Record<string, string> | [string, string][];
}

interface Response {
    readonly ok: boolean;
    readonly status: number;
    readonly statusText: string;
    readonly headers: Headers;
    readonly url: string;
    text(): Promise<string>;
    json<T = unknown>(): Promise<T>;
    arrayBuffer(): Promise<ArrayBuffer>;
}

interface ResponseConstructor {
    new(body?: string | null, init?: ResponseInit): Response;
    readonly prototype: Response;
    json(data: unknown, init?: ResponseInit): Response;
    error(): Response;
    redirect(url: string, status?: number): Response;
}

declare var Response: ResponseConstructor;

interface ReadableStream {
    readonly locked: boolean;
    cancel(reason?: unknown): Promise<void>;
    getReader(): unknown;
}

interface ReadableStreamConstructor {
    new(underlyingSource?: unknown, strategy?: unknown): ReadableStream;
    readonly prototype: ReadableStream;
}

declare var ReadableStream: ReadableStreamConstructor;

interface WritableStream {
    readonly locked: boolean;
    abort(reason?: unknown): Promise<void>;
    close(): Promise<void>;
    getWriter(): unknown;
}

interface WritableStreamConstructor {
    new(underlyingSink?: unknown, strategy?: unknown): WritableStream;
    readonly prototype: WritableStream;
}

declare var WritableStream: WritableStreamConstructor;

declare function fetch(input: string | Request, init?: RequestInit): Promise<Response>;

declare namespace Reflect {
    function apply(target: unknown, thisArgument: unknown, argumentsList: unknown[]): unknown;
    function construct(target: unknown, argumentsList: unknown[], newTarget?: unknown): unknown;
    function defineProperty(target: unknown, propertyKey: string | symbol, attributes: unknown): boolean;
    function deleteProperty(target: unknown, propertyKey: string | symbol): boolean;
    function get(target: unknown, propertyKey: string | symbol, receiver?: unknown): unknown;
    function getOwnPropertyDescriptor(target: unknown, propertyKey: string | symbol): PropertyDescriptor | undefined;
    function getPrototypeOf(target: unknown): unknown;
    function has(target: unknown, propertyKey: string | symbol): boolean;
    function isExtensible(target: unknown): boolean;
    function ownKeys(target: unknown): string[];
    function preventExtensions(target: unknown): boolean;
    function set(target: unknown, propertyKey: string | symbol, value: unknown, receiver?: unknown): boolean;
    function setPrototypeOf(target: unknown, proto: unknown): boolean;
    function getMetadata(metadataKey: unknown, target: unknown, propertyKey?: string | symbol): unknown;
    function getOwnMetadata(metadataKey: unknown, target: unknown, propertyKey?: string | symbol): unknown;
    function hasMetadata(metadataKey: unknown, target: unknown, propertyKey?: string | symbol): boolean;
    function hasOwnMetadata(metadataKey: unknown, target: unknown, propertyKey?: string | symbol): boolean;
    function defineMetadata(metadataKey: unknown, metadataValue: unknown, target: unknown, propertyKey?: string | symbol): void;
    function metadata(metadataKey: unknown, metadataValue: unknown): (target: unknown, propertyKey?: unknown) => void;
}

interface ArrayConstructor {
    fromAsync<T>(iterableOrArrayLike: AsyncIterable<T> | Iterable<T> | ArrayLike<T>, mapfn?: (value: unknown, index: number) => unknown): Promise<unknown[]>;
}

interface IteratorResult<T, TReturn = unknown> {
    done: boolean;
    value: T;
}

interface IteratorObject<T, TReturn = undefined, TNext = unknown> {
    [Symbol.iterator](): IteratorObject<T, TReturn, TNext>;
    next(...args: [] | [TNext]): IteratorResult<T, TReturn>;
    return?(value?: TReturn): IteratorResult<T, TReturn>;
    throw?(e?: unknown): IteratorResult<T, TReturn>;
    map<U>(callbackfn: (value: T, index: number) => U): IteratorObject<U, undefined, unknown>;
    filter<S extends T>(predicate: (value: T, index: number) => value is S): IteratorObject<S, undefined, unknown>;
    filter(predicate: (value: T, index: number) => unknown): IteratorObject<T, undefined, unknown>;
    take(limit: number): IteratorObject<T, undefined, unknown>;
    drop(limit: number): IteratorObject<T, undefined, unknown>;
    flatMap<U>(callback: (value: T, index: number) => Iterator<U, unknown, undefined> | Iterable<U, unknown, undefined>): IteratorObject<U, undefined, unknown>;
    reduce(callbackfn: (previousValue: T, currentValue: T, currentIndex: number) => T): T;
    reduce<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number) => U, initialValue: U): U;
    toArray(): T[];
    forEach(callbackfn: (value: T, index: number) => void): void;
    some(predicate: (value: T, index: number) => unknown): boolean;
    every(predicate: (value: T, index: number) => unknown): boolean;
    find<S extends T>(predicate: (value: T, index: number) => value is S): S | undefined;
    find(predicate: (value: T, index: number) => unknown): T | undefined;
}

interface IteratorConstructor {
    from<T>(value: Iterator<T, unknown, undefined> | Iterable<T, unknown, undefined>): IteratorObject<T, undefined, unknown>;
}

declare var Iterator: IteratorConstructor;

interface WebSocket {
    readonly url: string;
    readonly readyState: number;
    close(code?: number, reason?: string): void;
    send(data: string): void;
}

interface WebSocketConstructor {
    new(url: string, protocols?: string | string[]): WebSocket;
    readonly prototype: WebSocket;
    readonly CONNECTING: 0;
    readonly OPEN: 1;
    readonly CLOSING: 2;
    readonly CLOSED: 3;
}

declare var WebSocket: WebSocketConstructor;

interface EventInit {
    bubbles?: boolean;
    cancelable?: boolean;
    composed?: boolean;
}

interface CustomEventInit<T = unknown> extends EventInit {
    detail?: T;
}

interface Event {
    readonly type: string;
    readonly target: unknown;
    readonly currentTarget: unknown;
    readonly bubbles: boolean;
    readonly cancelable: boolean;
    readonly defaultPrevented: boolean;
    readonly timeStamp: number;
    preventDefault(): void;
    stopPropagation(): void;
    stopImmediatePropagation(): void;
}

interface EventConstructor {
    new(type: string, eventInitDict?: EventInit): Event;
    readonly prototype: Event;
}

declare var Event: EventConstructor;

interface CustomEvent<T = unknown> extends Event {
    readonly detail: T;
}

interface CustomEventConstructor {
    new <T = unknown>(type: string, eventInitDict?: CustomEventInit<T>): CustomEvent<T>;
    readonly prototype: CustomEvent;
}

declare var CustomEvent: CustomEventConstructor;

interface EventListenerObject {
    handleEvent(object: Event): void;
}

interface AddEventListenerOptions {
    once?: boolean;
    passive?: boolean;
    capture?: boolean;
    signal?: AbortSignal;
}

interface EventTarget {
    addEventListener(type: string, callback: ((evt: Event) => void) | EventListenerObject | Function | null, options?: AddEventListenerOptions | boolean): void;
    removeEventListener(type: string, callback: ((evt: Event) => void) | EventListenerObject | Function | null, options?: boolean | unknown): void;
    dispatchEvent(event: Event): boolean;
}

interface EventTargetConstructor {
    new(): EventTarget;
    readonly prototype: EventTarget;
}

declare var EventTarget: EventTargetConstructor;

interface DOMException extends Error {
    readonly name: string;
    readonly message: string;
    readonly code: number;
}

interface DOMExceptionConstructor {
    new(message?: string, name?: string): DOMException;
    readonly prototype: DOMException;
}

declare var DOMException: DOMExceptionConstructor;

interface AbortSignal extends EventTarget {
    readonly aborted: boolean;
    readonly reason: unknown;
    onabort: ((this: AbortSignal, ev: Event) => unknown) | null;
    throwIfAborted(): void;
}

interface AbortSignalConstructor {
    readonly prototype: AbortSignal;
    abort(reason?: unknown): AbortSignal;
    timeout(milliseconds: number): AbortSignal;
    any(signals: AbortSignal[]): AbortSignal;
}

declare var AbortSignal: AbortSignalConstructor;

interface AbortController {
    readonly signal: AbortSignal;
    abort(reason?: unknown): void;
}

interface AbortControllerConstructor {
    new(): AbortController;
    readonly prototype: AbortController;
}

declare var AbortController: AbortControllerConstructor;

interface StringDecoder {
    encoding: string;
    write(buffer: Buffer | Uint8Array | string): string;
    end(buffer?: Buffer | Uint8Array | string): string;
}

interface StringDecoderConstructor {
    new(encoding?: string): StringDecoder;
    readonly prototype: StringDecoder;
}

declare var StringDecoder: StringDecoderConstructor;

interface PunycodeUcs2 {
    decode(string: string): number[];
    encode(codePoints: number[]): string;
}

interface Punycode {
    decode(input: string): string;
    encode(input: string): string;
    toASCII(domain: string): string;
    toUnicode(domain: string): string;
    ucs2: PunycodeUcs2;
    version: string;
}

declare var punycode: Punycode;

declare function setTimeout(callback: (...args: unknown[]) => void, ms?: number): number;
declare function clearTimeout(id: number | undefined): void;
declare function setInterval(callback: (...args: unknown[]) => void, ms?: number): number;
declare function clearInterval(id: number | undefined): void;
declare function setImmediate(callback: (...args: unknown[]) => void): number;
declare function clearImmediate(id: number | undefined): void;

interface SharedArrayBuffer {
    readonly byteLength: number;
    slice(begin?: number, end?: number): SharedArrayBuffer;
}

interface SharedArrayBufferConstructor {
    readonly prototype: SharedArrayBuffer;
    new(byteLength: number): SharedArrayBuffer;
}

declare var SharedArrayBuffer: SharedArrayBufferConstructor;

interface Atomics {
    isLockFree(size: number): boolean;
    add(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    sub(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    and(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    or(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    xor(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    load(typedArray: Int32Array | Uint32Array, index: number): number;
    store(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    exchange(typedArray: Int32Array | Uint32Array, index: number, value: number): number;
    compareExchange(typedArray: Int32Array | Uint32Array, index: number, expected: number, replacement: number): number;
    wait(typedArray: Int32Array, index: number, value: number, timeout?: number): "ok" | "not-equal" | "timed-out";
    notify(typedArray: Int32Array, index: number, count?: number): number;
}

declare var Atomics: Atomics;

interface WeakRef<T extends object> {
    deref(): T | undefined;
}

interface WeakRefConstructor {
    readonly prototype: WeakRef<object>;
    new<T extends object>(target: T): WeakRef<T>;
}

declare var WeakRef: WeakRefConstructor;

interface FinalizationRegistry<T> {
    register(target: object, heldValue: T, unregisterToken?: object): void;
    unregister(unregisterToken: object): boolean;
}

interface FinalizationRegistryConstructor {
    readonly prototype: FinalizationRegistry<unknown>;
    new<T>(cleanupCallback: (heldValue: T) => void): FinalizationRegistry<T>;
}

declare var FinalizationRegistry: FinalizationRegistryConstructor;

