declare namespace __scriptgo {
    function setTimeout(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
    function clearTimeout(id: number | undefined): void;
    function setInterval(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
    function clearInterval(id: number | undefined): void;
    function setImmediate(callback: (...args: any[]) => void, ...args: any[]): number;
    function clearImmediate(id: number | undefined): void;
}

export function setTimeout(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number {
    return __scriptgo.setTimeout(callback, ms, ...args);
}

export function clearTimeout(id: number | undefined): void {
    __scriptgo.clearTimeout(id);
}

export function setInterval(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number {
    return __scriptgo.setInterval(callback, ms, ...args);
}

export function clearInterval(id: number | undefined): void {
    __scriptgo.clearInterval(id);
}

export function setImmediate(callback: (...args: any[]) => void, ...args: any[]): number {
    return __scriptgo.setImmediate(callback, ...args);
}

export function clearImmediate(id: number | undefined): void {
    __scriptgo.clearImmediate(id);
}
