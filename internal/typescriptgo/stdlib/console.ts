export function log(msg: string): void {
    console.log(msg);
}

export function info(msg: string): void {
    console.info(msg);
}

export function warn(msg: string): void {
    console.warn(msg);
}

export function error(msg: string): void {
    console.error(msg);
}

export function debug(msg: string): void {
    console.debug(msg);
}

export class Console {
    private _tag: number = 0;

    constructor() {}

    log(msg: string): void { console.log(msg); }
    info(msg: string): void { console.info(msg); }
    warn(msg: string): void { console.warn(msg); }
    error(msg: string): void { console.error(msg); }
    debug(msg: string): void { console.debug(msg); }
    clear(): void { console.clear(); }
    count(label: string = ""): void { console.count(label); }
    countReset(label: string = ""): void { console.countReset(label); }
    time(label: string = ""): void { console.time(label); }
    timeEnd(label: string = ""): void { console.timeEnd(label); }
    trace(msg: string = ""): void { console.trace(msg); }
    group(): void { console.group(); }
    groupEnd(): void { console.groupEnd(); }
}
