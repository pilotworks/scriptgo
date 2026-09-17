// ScriptGo Standard Library: node:tty

import { EventEmitter } from "node:events";

declare namespace __scriptgo {
    function ttyIsatty(fd: number): boolean;
    function ttyGetWindowSize(fd: number): [number, number];
    function ttySetRawMode(fd: number, mode: boolean): boolean;
    function ttyRead(fd: number, maxLen?: number): string;
    function ttyReadLine(fd: number): string;
    function ttyWrite(fd: number, data: string, len?: number): number;
}

export function isatty(fd: number): boolean {
    if (typeof fd !== "number" || isNaN(fd) || fd < 0) {
        return false;
    }
    return __scriptgo.ttyIsatty(fd);
}

export class ReadStream extends EventEmitter {
    fd: number;
    isRaw: boolean = false;
    isTTY: boolean;
    readable: boolean = true;

    constructor(fd: number, options?: Record<string, unknown>) {
        super();
        this.fd = fd;
        this.isTTY = isatty(fd);
    }

    setRawMode(mode: boolean): this {
        this.isRaw = !!mode;
        __scriptgo.ttySetRawMode(this.fd, this.isRaw);
        return this;
    }

    read(size?: number): string | null {
        const data = __scriptgo.ttyRead(this.fd, size);
        return data.length > 0 ? data : null;
    }

    destroy(): this {
        this.readable = false;
        this.emit("close");
        return this;
    }
}

export class WriteStream extends EventEmitter {
    fd: number;
    isTTY: boolean;
    writable: boolean = true;
    columns: number;
    rows: number;

    constructor(fd: number) {
        super();
        this.fd = fd;
        this.isTTY = isatty(fd);
        const ws = this.getWindowSize();
        this.columns = ws[0];
        this.rows = ws[1];
    }

    getWindowSize(): [number, number] {
        const size = __scriptgo.ttyGetWindowSize(this.fd);
        this.columns = size[0];
        this.rows = size[1];
        return size;
    }

    write(chunk: unknown, encoding?: string | Function, callback?: Function): boolean {
        const str = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        __scriptgo.ttyWrite(this.fd, str);
        if (typeof encoding === "function") {
            encoding();
        } else if (typeof callback === "function") {
            callback();
        }
        return true;
    }

    clearLine(dir: number, callback?: () => void): boolean {
        let seq = "\x1b[2K";
        if (dir < 0) {
            seq = "\x1b[1K";
        } else if (dir > 0) {
            seq = "\x1b[0K";
        }
        this.write(seq);
        if (callback) callback();
        return true;
    }

    clearScreenDown(callback?: () => void): boolean {
        this.write("\x1b[0J");
        if (callback) callback();
        return true;
    }

    cursorTo(x: number, y?: number, callback?: () => void): boolean {
        if (typeof x !== "number") {
            if (callback) callback();
            return false;
        }
        let seq = `\x1b[${x + 1}G`;
        if (typeof y === "number") {
            seq = `\x1b[${y + 1};${x + 1}H`;
        }
        this.write(seq);
        if (callback) callback();
        return true;
    }

    moveCursor(dx: number, dy: number, callback?: () => void): boolean {
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
            this.write(seq);
        }
        if (callback) callback();
        return true;
    }

    getColorDepth(env?: Record<string, string | undefined>): number {
        const e = env || (typeof process !== "undefined" ? process.env : undefined);
        if (e && (e.NO_COLOR !== undefined || e.NODE_DISABLE_COLORS !== undefined)) {
            return 1;
        }
        if (e && (e.COLORTERM === "truecolor" || e.COLORTERM === "24bit")) {
            return 24;
        }
        if (e && typeof e.TERM === "string" && e.TERM.indexOf("256") !== -1) {
            return 8;
        }
        return this.isTTY ? 8 : 1;
    }

    hasColors(count?: number, env?: Record<string, string | undefined>): boolean {
        if (count === undefined) count = 16;
        const depth = this.getColorDepth(env);
        return depth >= 4 && (1 << depth) >= count;
    }

    destroy(): this {
        this.writable = false;
        this.emit("close");
        return this;
    }
}

export default {
    isatty,
    ReadStream,
    WriteStream,
};
