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

    [Symbol.asyncIterator](): AsyncIterableIterator<string> {
        return this._rl[Symbol.asyncIterator]();
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
    createInterface,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents,
};
