// ScriptGo Standard Library: node:readline

class ReadlineListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class ReadlineEventBucket {
    name: string;
    listeners: ReadlineListenerEntry[];

    constructor(name: string, listeners: ReadlineListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export interface ReadLineOptions {
    input?: unknown;
    output?: unknown;
    completer?: Function;
    terminal?: boolean;
    history?: string[];
    historySize?: number;
    prompt?: string;
    crlfDelay?: number;
    removeHistoryDuplicates?: boolean;
    escapeCodeTimeout?: number;
    tabSize?: number;
    signal?: unknown;
}

export class InterfaceConstructor {
    line: string = "";
    cursor: number = 0;
    terminal: boolean = false;
    private _prompt: string = "> ";
    private _closed: boolean = false;
    private _paused: boolean = false;
    _buckets: ReadlineEventBucket[] = [];

    constructor(inputOrOptions: unknown = null, output: unknown = null, completer: Function | null = null, terminal: boolean = false) {
        if (inputOrOptions && typeof inputOrOptions === "object") {
            const opts = inputOrOptions as ReadLineOptions;
            if (opts.prompt) this._prompt = opts.prompt;
            if (opts.terminal !== undefined) this.terminal = opts.terminal;
        }
    }

    private _getBucket(name: string): ReadlineEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new ReadlineEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): InterfaceConstructor {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new ReadlineListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): InterfaceConstructor {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new ReadlineListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: ReadlineListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) {
                remaining.push(entry);
            }
        }
        bucket.listeners = remaining;
        return true;
    }

    setPrompt(prompt: string): void {
        this._prompt = prompt;
    }

    getPrompt(): string {
        return this._prompt;
    }

    prompt(preserveCursor: boolean = false): void {
        this.emit("prompt");
    }

    pause(): InterfaceConstructor {
        this._paused = true;
        this.emit("pause");
        return this;
    }

    resume(): InterfaceConstructor {
        this._paused = false;
        this.emit("resume");
        return this;
    }

    write(data: string, key: unknown = null): void {
        this.line += data;
        this.cursor += data.length;
    }

    getCursorPos(): { rows: number; cols: number } {
        return { rows: 0, cols: this.cursor };
    }

    close(): void {
        this._closed = true;
        this.emit("close");
    }

    question(query: string, callback?: unknown): void {
        if (typeof callback === "function") {
            (callback as Function)("");
        }
    }

    [Symbol.dispose](): void {
        this.close();
    }

    [Symbol.asyncIterator](): unknown {
        return {
            next: () => Promise.resolve({ done: true, value: "" })
        };
    }
}

export class Interface extends InterfaceConstructor {
    constructor(inputOrOptions: unknown = null, output: unknown = null, completer: Function | null = null, terminal: boolean = false) {
        super(inputOrOptions, output, completer, terminal);
    }
}

export namespace promises {
    export class Interface extends InterfaceConstructor {
        question(query: string, options: unknown = null): Promise<string> {
            return Promise.resolve("");
        }
    }

    export class Readline {
        stream: unknown = null;

        constructor(stream: unknown = null) {
            this.stream = stream;
        }

        clearLine(dir: number): Readline {
            return this;
        }

        clearScreenDown(): Readline {
            return this;
        }

        commit(): Promise<void> {
            return Promise.resolve(undefined);
        }

        cursorTo(x: number, y: number = 0): Readline {
            return this;
        }

        moveCursor(dx: number, dy: number): Readline {
            return this;
        }

        rollback(): Readline {
            return this;
        }
    }

    export function createInterface(inputOrOptions: unknown = null, output: unknown = null): Interface {
        return new Interface(inputOrOptions, output);
    }
}

export function createInterface(inputOrOptions: unknown, output: unknown = null, completer: Function | null = null, terminal: boolean = false): Interface {
    return new Interface(inputOrOptions, output, completer, terminal);
}

export function emitKeypressEvents(stream: unknown, rl: unknown = null): void {
}

export function clearLine(stream: unknown, dir: number, callback: unknown = null): boolean {
    if (typeof callback === "function") (callback as Function)();
    return true;
}

export function clearScreenDown(stream: unknown, callback: unknown = null): boolean {
    if (typeof callback === "function") (callback as Function)();
    return true;
}

export function cursorTo(stream: unknown, x: number, y: unknown = null, callback: unknown = null): boolean {
    if (typeof callback === "function") (callback as Function)();
    else if (typeof y === "function") (y as Function)();
    return true;
}

export function moveCursor(stream: unknown, dx: number, dy: number, callback: unknown = null): boolean {
    if (typeof callback === "function") (callback as Function)();
    return true;
}

export default {
    InterfaceConstructor,
    Interface,
    createInterface,
    emitKeypressEvents,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    promises,
};
