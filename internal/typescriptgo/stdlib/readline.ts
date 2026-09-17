// ScriptGo Standard Library: node:readline

import { EventEmitter } from "node:events";

declare namespace __scriptgo {
    function ttyReadLine(fd: number): string;
}

export class StreamLike extends EventEmitter {
    fd: number = 0;
    isTTY: boolean = false;
    _hasKeypressEvents: boolean = false;

    write(data: string): boolean {
        return true;
    }
    clearLine(dir: number, callback?: () => void): boolean {
        return true;
    }
    clearScreenDown(callback?: () => void): boolean {
        return true;
    }
    cursorTo(x: number, y?: number, callback?: () => void): boolean {
        return true;
    }
    moveCursor(dx: number, dy: number, callback?: () => void): boolean {
        return true;
    }
}

export interface ReadLineOptions {
    input: EventEmitter;
    output?: EventEmitter;
    completer?: unknown;
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

export class Interface extends EventEmitter {
    input: EventEmitter;
    output: EventEmitter | null = null;
    terminal: boolean;
    line: string = "";
    cursor: number = 0;
    history: string[] = [];
    historySize: number = 30;
    historyIndex: number = -1;
    paused: boolean = false;
    closed: boolean = false;
    private _promptStr: string = "> ";
    private _questionCb: ((answer: string) => void) | null = null;
    private _lineQueue: string[] = [];
    private _waitingResolvers: ((value: { value: string; done: boolean }) => void)[] = [];

    constructor(
        inputOrOptions: ReadLineOptions | EventEmitter,
        output?: EventEmitter,
        completer?: unknown,
        terminal?: boolean
    ) {
        super();
        let opts: ReadLineOptions;
        if (inputOrOptions instanceof EventEmitter) {
            opts = {
                input: inputOrOptions,
                output,
                completer,
                terminal
            };
        } else {
            opts = inputOrOptions as ReadLineOptions;
        }

        this.input = opts.input;
        if (opts.output !== undefined && opts.output !== null) {
            this.output = opts.output;
        }
        this.terminal = opts.terminal !== undefined ? !!opts.terminal : false;
        if (opts.prompt !== undefined) {
            this._promptStr = opts.prompt;
        }
        if (opts.historySize !== undefined) {
            this.historySize = opts.historySize;
        }
        if (opts.history) {
            this.history = opts.history.slice(0, this.historySize);
        }

        if (this.input) {
            let buffer = "";
            this.input.on("data", (data: unknown) => {
                const str = typeof data === "string" ? data : (data !== null && data !== undefined ? String(data) : "");
                buffer += str;
                const lines = buffer.split("\n");
                for (let i = 0; i < lines.length - 1; i++) {
                    let line = lines[i];
                    if (line.endsWith("\r")) {
                        line = line.slice(0, -1);
                    }
                    this._onLine(line);
                }
                buffer = lines[lines.length - 1];
            });
            this.input.on("end", () => {
                if (buffer.length > 0) {
                    let line = buffer;
                    if (line.endsWith("\r")) {
                        line = line.slice(0, -1);
                    }
                    this._onLine(line);
                    buffer = "";
                }
                this.close();
            });
        }
    }

    setPrompt(prompt: string): void {
        this._promptStr = prompt;
    }

    getPrompt(): string {
        return this._promptStr;
    }

    prompt(preserveCursor?: boolean): void {
        if (this.closed) return;
        if (this.output) {
            if (this.output instanceof StreamLike) {
                this.output.write(this._promptStr);
            } else {
                this.output.emit("data", this._promptStr);
            }
        }
    }

    question(query: string, callback: (answer: string) => void): void {
        if (this.closed) return;
        if (this.output) {
            if (this.output instanceof StreamLike) {
                this.output.write(query);
            } else {
                this.output.emit("data", query);
            }
        }
        this._questionCb = callback;

        if (!this.input || (this.input as unknown as { fd?: number }).fd === 0) {
            const line = __scriptgo.ttyReadLine(0);
            if (line.length > 0 && this._questionCb) {
                const cb = this._questionCb;
                this._questionCb = null;
                cb(line);
                this.emit("line", line);
            }
        }
    }

    write(data: string, key?: unknown): void {
        if (this.output) {
            if (this.output instanceof StreamLike) {
                this.output.write(data);
            } else {
                this.output.emit("data", data);
            }
        }
    }

    pause(): this {
        this.paused = true;
        this.emit("pause");
        return this;
    }

    resume(): this {
        this.paused = false;
        this.emit("resume");
        return this;
    }

    close(): void {
        if (this.closed) return;
        this.closed = true;
        this.emit("close");
        while (this._waitingResolvers.length > 0) {
            const resolve = this._waitingResolvers.shift()!;
            resolve({ value: "", done: true });
        }
    }

    getCursorPos(): { rows: number; cols: number } {
        return { rows: 0, cols: this.cursor };
    }

    private _onLine(line: string): void {
        if (this.closed) return;
        this.line = line;
        if (line.length > 0) {
            this.history.unshift(line);
            if (this.history.length > this.historySize) {
                this.history.pop();
            }
            this.emit("history", this.history);
        }
        if (this._questionCb) {
            const cb = this._questionCb;
            this._questionCb = null;
            cb(line);
        }
        this.emit("line", line);

        if (this._waitingResolvers.length > 0) {
            const resolve = this._waitingResolvers.shift()!;
            resolve({ value: line, done: false });
        } else {
            this._lineQueue.push(line);
        }
    }

    [Symbol.asyncIterator](): AsyncIterableIterator<string> {
        const self = this;
        return {
            next(): Promise<IteratorResult<string>> {
                if (self._lineQueue.length > 0) {
                    const value = self._lineQueue.shift()!;
                    return Promise.resolve({ value, done: false });
                }
                if (self.closed) {
                    return Promise.resolve({ value: "", done: true });
                }
                return new Promise<IteratorResult<string>>((resolve) => {
                    self._waitingResolvers.push(resolve);
                });
            },
            [Symbol.asyncIterator]() {
                return this;
            }
        };
    }
}

export function createInterface(
    inputOrOptions: ReadLineOptions | EventEmitter,
    output?: EventEmitter,
    completer?: unknown,
    terminal?: boolean
): Interface {
    return new Interface(inputOrOptions, output, completer, terminal);
}

export function clearLine(stream: EventEmitter | null | undefined, dir: number, callback?: () => void): boolean {
    if (stream && stream instanceof StreamLike) {
        return stream.clearLine(dir, callback);
    }
    if (callback) callback();
    return true;
}

export function clearScreenDown(stream: EventEmitter | null | undefined, callback?: () => void): boolean {
    if (stream && stream instanceof StreamLike) {
        return stream.clearScreenDown(callback);
    }
    if (callback) callback();
    return true;
}

export function cursorTo(stream: EventEmitter | null | undefined, x: number, y?: number, callback?: () => void): boolean {
    if (stream && stream instanceof StreamLike) {
        return stream.cursorTo(x, y, callback);
    }
    if (callback) callback();
    return true;
}

export function moveCursor(stream: EventEmitter | null | undefined, dx: number, dy: number, callback?: () => void): boolean {
    if (stream && stream instanceof StreamLike) {
        return stream.moveCursor(dx, dy, callback);
    }
    if (callback) callback();
    return true;
}

export function emitKeypressEvents(stream: EventEmitter | null | undefined, iface?: Interface): void {
    if (!stream) return;
    stream.on("data", (b: unknown) => {
        const s = typeof b === "string" ? b : (b !== null && b !== undefined ? String(b) : "");
        for (let i = 0; i < s.length; i++) {
            const ch = s[i];
            stream.emit("keypress", ch);
        }
    });
}

export default {
    StreamLike,
    Interface,
    createInterface,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents,
};
