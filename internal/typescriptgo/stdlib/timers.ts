declare namespace __scriptgo {
    function setTimeout(callback: (...args: unknown[]) => void, ms?: number): number;
    function clearTimeout(id: number): void;
    function setInterval(callback: (...args: unknown[]) => void, ms?: number): number;
    function clearInterval(id: number): void;
    function setImmediate(callback: (...args: unknown[]) => void): number;
    function clearImmediate(id: number): void;
}

export class Immediate {
    _id: number = 0;
    _refed: boolean = true;

    constructor(id: number = 0) {
        this._id = id;
    }

    hasRef(): boolean {
        return this._refed;
    }

    ref(): this {
        this._refed = true;
        return this;
    }

    unref(): this {
        this._refed = false;
        return this;
    }

    [Symbol.dispose](): void {
        clearImmediate(this._id);
    }
}

export class Timeout {
    _id: number = 0;
    _refed: boolean = true;

    constructor(id: number = 0) {
        this._id = id;
    }

    hasRef(): boolean {
        return this._refed;
    }

    ref(): this {
        this._refed = true;
        return this;
    }

    unref(): this {
        this._refed = false;
        return this;
    }

    refresh(): this {
        return this;
    }

    close(): this {
        clearTimeout(this._id);
        return this;
    }

    [Symbol.toPrimitive](): number {
        return this._id;
    }

    [Symbol.dispose](): void {
        this.close();
    }
}

export function setTimeout(callback: (...args: unknown[]) => void, ms?: number): number {
    return __scriptgo.setTimeout(callback, ms);
}

export function clearTimeout(id: Timeout | number | undefined | null): void {
    if (typeof id === "number") {
        __scriptgo.clearTimeout(id);
    } else if (id instanceof Timeout) {
        __scriptgo.clearTimeout(id._id);
    }
}

export function setInterval(callback: (...args: unknown[]) => void, ms?: number): number {
    return __scriptgo.setInterval(callback, ms);
}

export function clearInterval(id: Timeout | number | undefined | null): void {
    if (typeof id === "number") {
        __scriptgo.clearInterval(id);
    } else if (id instanceof Timeout) {
        __scriptgo.clearInterval(id._id);
    }
}

export function setImmediate(callback: (...args: unknown[]) => void): number {
    return __scriptgo.setImmediate(callback);
}

export function clearImmediate(id: Immediate | number | undefined | null): void {
    if (typeof id === "number") {
        __scriptgo.clearImmediate(id);
    } else if (id instanceof Immediate) {
        __scriptgo.clearImmediate(id._id);
    }
}

export class Scheduler {
    wait(delay: number = 0, options?: unknown): Promise<void> {
        return new Promise((resolve) => {
            __scriptgo.setTimeout(() => {
                resolve();
            }, delay);
        });
    }

    yield(): Promise<void> {
        return new Promise((resolve) => {
            __scriptgo.setImmediate(() => {
                resolve();
            });
        });
    }
}

export const scheduler: Scheduler = new Scheduler();

export class TimersPromises {
    scheduler: Scheduler = scheduler;

    setTimeout(delay: number = 1, value?: unknown): Promise<unknown> {
        return new Promise((resolve) => {
            __scriptgo.setTimeout(() => {
                resolve(value);
            }, delay);
        });
    }

    setImmediate(value?: unknown): Promise<unknown> {
        return new Promise((resolve) => {
            __scriptgo.setImmediate(() => {
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
    scheduler,
    promises,
};
