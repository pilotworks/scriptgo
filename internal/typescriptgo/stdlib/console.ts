// ScriptGo Standard Library: node:console

export function log(...data: unknown[]): void {
    console.log(...data);
}

export function info(...data: unknown[]): void {
    console.info(...data);
}

export function warn(...data: unknown[]): void {
    console.warn(...data);
}

export function error(...data: unknown[]): void {
    console.error(...data);
}

export function debug(...data: unknown[]): void {
    console.debug(...data);
}

export function assert(condition?: unknown, ...data: unknown[]): void {
    console.assert(condition, ...data);
}

export function clear(): void {
    console.clear();
}

export function count(label: string = "default"): void {
    console.count(label);
}

export function countReset(label: string = "default"): void {
    console.countReset(label);
}

export function time(label: string = "default"): void {
    console.time(label);
}

export function timeEnd(label: string = "default"): void {
    console.timeEnd(label);
}

export function timeLog(label: string = "default", ...data: unknown[]): void {
    console.timeLog(label, ...data);
}

export function trace(...data: unknown[]): void {
    console.trace(...data);
}

export function group(...data: unknown[]): void {
    console.group(...data);
}

export function groupCollapsed(...data: unknown[]): void {
    console.groupCollapsed(...data);
}

export function groupEnd(): void {
    console.groupEnd();
}

export function dir(item?: unknown, options?: unknown): void {
    console.dir(item, options);
}

export function dirxml(...data: unknown[]): void {
    console.dirxml(...data);
}

export function table(tabularData?: unknown, properties?: string[]): void {
    console.table(tabularData, properties);
}

export function profile(label?: string): void {
    console.profile(label);
}

export function profileEnd(label?: string): void {
    console.profileEnd(label);
}

export function timeStamp(label?: string): void {
    console.timeStamp(label);
}

export interface ConsoleConstructorOptions {
    stdout?: unknown;
    stderr?: unknown;
    ignoreErrors?: boolean;
    colorMode?: boolean | "auto";
    inspectOptions?: unknown;
    groupIndentation?: number;
}

export class Console {
    constructor(stdout?: unknown, stderr?: unknown, ignoreErrors?: boolean);
    constructor(options?: ConsoleConstructorOptions | unknown);
    constructor(stdoutOrOptions?: unknown, stderr?: unknown, ignoreErrors?: boolean) { }

    assert(condition?: unknown, ...data: unknown[]): void { console.assert(condition, ...data); }
    clear(): void { console.clear(); }
    count(label: string = "default"): void { console.count(label); }
    countReset(label: string = "default"): void { console.countReset(label); }
    debug(...data: unknown[]): void { console.debug(...data); }
    dir(item?: unknown, options?: unknown): void { console.dir(item, options); }
    dirxml(...data: unknown[]): void { console.dirxml(...data); }
    error(...data: unknown[]): void { console.error(...data); }
    group(...data: unknown[]): void { console.group(...data); }
    groupCollapsed(...data: unknown[]): void { console.groupCollapsed(...data); }
    groupEnd(): void { console.groupEnd(); }
    info(...data: unknown[]): void { console.info(...data); }
    log(...data: unknown[]): void { console.log(...data); }
    profile(label?: string): void { console.profile(label); }
    profileEnd(label?: string): void { console.profileEnd(label); }
    table(tabularData?: unknown, properties?: string[]): void { console.table(tabularData, properties); }
    time(label: string = "default"): void { console.time(label); }
    timeEnd(label: string = "default"): void { console.timeEnd(label); }
    timeLog(label: string = "default", ...data: unknown[]): void { console.timeLog(label, ...data); }
    timeStamp(label?: string): void { console.timeStamp(label); }
    trace(...data: unknown[]): void { console.trace(...data); }
    warn(...data: unknown[]): void { console.warn(...data); }
}

export default {
    assert,
    clear,
    count,
    countReset,
    debug,
    dir,
    dirxml,
    error,
    group,
    groupCollapsed,
    groupEnd,
    info,
    log,
    profile,
    profileEnd,
    table,
    time,
    timeEnd,
    timeLog,
    timeStamp,
    trace,
    warn,
    Console,
};
