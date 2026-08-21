export function log(msg: string): void;
export function info(msg: string): void;
export function warn(msg: string): void;
export function error(msg: string): void;
export function debug(msg: string): void;
export class Console {
    constructor();
    log(msg: string): void;
    info(msg: string): void;
    warn(msg: string): void;
    error(msg: string): void;
    debug(msg: string): void;
    clear(): void;
    count(label?: string): void;
    countReset(label?: string): void;
    time(label?: string): void;
    timeEnd(label?: string): void;
    trace(msg?: string): void;
    group(): void;
    groupEnd(): void;
}
