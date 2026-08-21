export function setTimeout(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
export function clearTimeout(id: number | undefined): void;
export function setInterval(callback: (...args: any[]) => void, ms?: number, ...args: any[]): number;
export function clearInterval(id: number | undefined): void;
export function setImmediate(callback: (...args: any[]) => void, ...args: any[]): number;
export function clearImmediate(id: number | undefined): void;
