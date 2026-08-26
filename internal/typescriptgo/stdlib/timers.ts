declare namespace __scriptgo {
    function setTimeout(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number;
    function clearTimeout(id: number | undefined): void;
    function setInterval(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number;
    function clearInterval(id: number | undefined): void;
    function setImmediate(callback: (...args: unknown[]) => void, ...args: unknown[]): number;
    function clearImmediate(id: number | undefined): void;
}

export function setTimeout(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number {
    return __scriptgo.setTimeout(callback, ms, ...args);
}

export function clearTimeout(id: number | undefined): void {
    __scriptgo.clearTimeout(id);
}

export function setInterval(callback: (...args: unknown[]) => void, ms?: number, ...args: unknown[]): number {
    return __scriptgo.setInterval(callback, ms, ...args);
}

export function clearInterval(id: number | undefined): void {
    __scriptgo.clearInterval(id);
}

export function setImmediate(callback: (...args: unknown[]) => void, ...args: unknown[]): number {
    return __scriptgo.setImmediate(callback, ...args);
}

export function clearImmediate(id: number | undefined): void {
    __scriptgo.clearImmediate(id);
}

export class TimersPromises {
    async setTimeout(delay: number = 1, value?: unknown): Promise<unknown> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve(value);
            }, delay);
        });
    }

    async setImmediate(value?: unknown): Promise<unknown> {
        return new Promise((resolve) => {
            setImmediate(() => {
                resolve(value);
            });
        });
    }

    async *setInterval(delay: number = 1, value?: unknown): AsyncIterableIterator<unknown> {
        while (true) {
            await this.setTimeout(delay);
            yield value;
        }
    }
}

export const promises: TimersPromises = new TimersPromises();

export default {
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval,
    setImmediate,
    clearImmediate,
    promises,
};
