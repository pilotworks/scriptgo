// Node.js TTY module (node:tty)

export function isatty(fd: number): boolean {
    if (fd === 0 || fd === 1 || fd === 2) {
        return true;
    }
    return false;
}

export class ReadStream {
    fd: number = 0;
    isRaw: boolean = false;
    isTTY: boolean = true;

    constructor(fd: number = 0, options: unknown = null) {
        this.fd = fd;
        this.isRaw = false;
        this.isTTY = true;
    }

    setRawMode(mode: boolean): ReadStream {
        this.isRaw = mode;
        return this;
    }
}

export class WriteStream {
    fd: number = 1;
    isTTY: boolean = true;
    columns: number = 80;
    rows: number = 24;

    constructor(fd: number = 1, options: unknown = null) {
        this.fd = fd;
        this.isTTY = true;
        this.columns = 80;
        this.rows = 24;
    }

    clearLine(dir: number, callback: unknown = null): boolean {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return true;
    }

    clearScreenDown(callback: unknown = null): boolean {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return true;
    }

    cursorTo(x: number, y: number = 0, callback: unknown = null): boolean {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return true;
    }

    moveCursor(dx: number, dy: number, callback: unknown = null): boolean {
        if (typeof callback === "function") {
            (callback as Function)();
        }
        return true;
    }

    getColorDepth(env: unknown = null): number {
        return 24;
    }

    hasColors(count: unknown = 16, env: unknown = null): boolean {
        if (typeof count === "number") {
            return (count as number) <= 16777216;
        }
        return true;
    }

    getWindowSize(): number[] {
        return [this.columns, this.rows];
    }
}

export default {
    isatty,
    ReadStream,
    WriteStream,
};
