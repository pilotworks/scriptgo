// ScriptGo Standard Library: node:readline/promises

import { EventEmitter, AbortSignal } from "node:events";
import {
    Interface as CallbackInterface,
    ReadLineOptions,
    StreamLike,
    createInterface as callbackCreateInterface,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents
} from "node:readline";

export class QuestionOptions {
    signal?: AbortSignal;
}

export class Interface extends EventEmitter {
    terminal: boolean;
    line: string = "";
    cursor: number = 0;
    private _rl: CallbackInterface;

    constructor(
        inputOrOptions: ReadLineOptions | EventEmitter,
        output?: EventEmitter,
        completer?: unknown,
        terminal?: boolean
    ) {
        super();
        this._rl = new CallbackInterface(inputOrOptions, output, completer, terminal);
        this.terminal = this._rl.terminal;
        this.line = this._rl.line;
        this.cursor = this._rl.cursor;

        this._rl.on("line", (line: unknown) => {
            this.line = this._rl.line;
            this.emit("line", line);
        });
        this._rl.on("close", () => this.emit("close"));
        this._rl.on("pause", () => this.emit("pause"));
        this._rl.on("resume", () => this.emit("resume"));
        this._rl.on("history", (h: unknown) => this.emit("history", h));
    }

    question(query: string, options?: QuestionOptions): Promise<string> {
        return new Promise<string>((resolve, reject) => {
            if (options && options.signal && options.signal.aborted) {
                return reject(new Error("The operation was aborted"));
            }
            if (options && options.signal && typeof options.signal.addEventListener === "function") {
                options.signal.addEventListener("abort", () => {
                    reject(new Error("The operation was aborted"));
                });
            }
            this._rl.question(query, (answer: string) => {
                resolve(answer);
            });
        });
    }

    prompt(preserveCursor?: boolean): void {
        this._rl.prompt(preserveCursor);
    }

    setPrompt(prompt: string): void {
        this._rl.setPrompt(prompt);
    }

    getPrompt(): string {
        return this._rl.getPrompt();
    }

    pause(): this {
        this._rl.pause();
        return this;
    }

    resume(): this {
        this._rl.resume();
        return this;
    }

    close(): void {
        this._rl.close();
    }

    write(data: string, key?: unknown): void {
        this._rl.write(data, key);
    }

    getCursorPos(): { rows: number; cols: number } {
        return this._rl.getCursorPos();
    }

    [Symbol.dispose](): void {
        this.close();
    }

    [Symbol.asyncIterator](): AsyncIterableIterator<string> {
        return this._rl[Symbol.asyncIterator]();
    }
}

export class Readline {
    private _stream: EventEmitter;
    private _actions: string[] = [];

    constructor(stream: EventEmitter, options?: { autoCommit?: boolean }) {
        this._stream = stream;
    }

    clearLine(dir: number): this {
        let seq = "\x1b[2K";
        if (dir < 0) {
            seq = "\x1b[1K";
        } else if (dir > 0) {
            seq = "\x1b[0K";
        }
        this._actions.push(seq);
        return this;
    }

    clearScreenDown(): this {
        this._actions.push("\x1b[0J");
        return this;
    }

    cursorTo(x: number, y?: number): this {
        let seq = `\x1b[${x + 1}G`;
        if (typeof y === "number") {
            seq = `\x1b[${y + 1};${x + 1}H`;
        }
        this._actions.push(seq);
        return this;
    }

    moveCursor(dx: number, dy: number): this {
        let seq = "";
        if (dx < 0) {
            seq += `\x1b[${-dx}D`;
        } else if (dx > 0) {
            seq += `\x1b[${dx}C`;
        }
        if (dy < 0) {
            seq += `\x1b[${-dy}A`;
        } else if (dy > 0) {
            seq += `\x1b[${dy}B`;
        }
        if (seq.length > 0) {
            this._actions.push(seq);
        }
        return this;
    }

    commit(): Promise<void> {
        return new Promise<void>((resolve) => {
            const data = this._actions.join("");
            this._actions = [];
            if (data.length > 0) {
                if (this._stream instanceof StreamLike) {
                    this._stream.write(data);
                } else {
                    this._stream.emit("data", data);
                }
            }
            resolve();
        });
    }

    rollback(): this {
        this._actions = [];
        return this;
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

export {
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents
};

export default {
    QuestionOptions,
    Interface,
    Readline,
    createInterface,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents,
};
