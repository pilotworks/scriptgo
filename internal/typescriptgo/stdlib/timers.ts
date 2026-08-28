declare namespace __scriptgo {
    function setTimeout(callback: (...args: unknown[]) => void, ms?: number): number;
    function clearTimeout(id: number): void;
    function setInterval(callback: (...args: unknown[]) => void, ms?: number): number;
    function clearInterval(id: number): void;
    function setImmediate(callback: (...args: unknown[]) => void): number;
    function clearImmediate(id: number): void;
}

export function setTimeout(callback: (...args: unknown[]) => void, ms?: number): number {
    return __scriptgo.setTimeout(callback, ms);
}

export function clearTimeout(id: number | undefined): void {
    if (typeof id === "number") {
        __scriptgo.clearTimeout(id);
    }
}

export function setInterval(callback: (...args: unknown[]) => void, ms?: number): number {
    return __scriptgo.setInterval(callback, ms);
}

export function clearInterval(id: number | undefined): void {
    if (typeof id === "number") {
        __scriptgo.clearInterval(id);
    }
}

export function setImmediate(callback: (...args: unknown[]) => void): number {
    return __scriptgo.setImmediate(callback);
}

export function clearImmediate(id: number | undefined): void {
    if (typeof id === "number") {
        __scriptgo.clearImmediate(id);
    }
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
