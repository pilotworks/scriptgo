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

export class Immediate {
    _id: number = 0;

    constructor(id: number = 0) {
        this._id = id;
    }

    [Symbol.dispose](): void {
        clearImmediate(this._id);
    }
}

export class Timeout {
    _id: number = 0;

    constructor(id: number = 0) {
        this._id = id;
    }

    close(): void {
        clearTimeout(this._id);
    }

    [Symbol.toPrimitive](): number {
        return this._id;
    }

    [Symbol.dispose](): void {
        this.close();
    }
}

export class TimersPromises {
    setTimeout(delay: number = 1, value?: unknown): Promise<unknown> {
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve(value);
            }, delay);
        });
    }

    setImmediate(value?: unknown): Promise<unknown> {
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
    Timeout,
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval,
    Immediate,
    setImmediate,
    clearImmediate,
    promises,
};
