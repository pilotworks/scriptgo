// ScriptGo Standard Library: node:child_process

import { EventEmitter } from "node:events";
import { Readable, Writable, StreamChunk } from "node:stream";

export interface ExecSyncOptions {
    cwd?: string;
    input?: string;
    encoding?: string;
    timeout?: number;
}

export interface SpawnSyncOptions {
    cwd?: string;
    input?: string;
    encoding?: string;
    timeout?: number;
}

export interface SpawnOptions {
    cwd?: string;
    env?: Record<string, string>;
    argv0?: string;
    stdio?: string | unknown[];
    detached?: boolean;
    uid?: number;
    gid?: number;
    shell?: boolean | string;
    timeout?: number;
    killSignal?: string | number;
}

export interface ExecOptions {
    cwd?: string;
    env?: Record<string, string>;
    encoding?: string;
    shell?: string;
    timeout?: number;
    maxBuffer?: number;
    killSignal?: string | number;
}

export interface ExecFileOptions {
    cwd?: string;
    env?: Record<string, string>;
    encoding?: string;
    timeout?: number;
    maxBuffer?: number;
    killSignal?: string | number;
}

export interface ForkOptions {
    cwd?: string;
    env?: Record<string, string>;
    execPath?: string;
    execArgv?: string[];
    silent?: boolean;
    stdio?: string | unknown[];
    detached?: boolean;
    timeout?: number;
}

export class SpawnSyncReturns {
    stdout: string;
    stderr: string;
    status: number;

    constructor(stdout: string, stderr: string, status: number) {
        this.stdout = stdout;
        this.stderr = stderr;
        this.status = status;
    }
}

export class ChildProcessSpawnResult {
    pid: number;
    stdinFd: number;
    stdoutFd: number;
    stderrFd: number;

    constructor(pid: number, stdinFd: number, stdoutFd: number, stderrFd: number) {
        this.pid = pid;
        this.stdinFd = stdinFd;
        this.stdoutFd = stdoutFd;
        this.stderrFd = stderrFd;
    }
}

export class ChildProcessPollResult {
    status: number;
    exited: boolean;

    constructor(status: number, exited: boolean) {
        this.status = status;
        this.exited = exited;
    }
}

export class ChildProcessReadResult {
    data: string;
    bytesRead: number;
    eof: boolean;

    constructor(data: string, bytesRead: number, eof: boolean) {
        this.data = data;
        this.bytesRead = bytesRead;
        this.eof = eof;
    }
}

const defaultExecOptions: ExecSyncOptions = { cwd: "", input: "" };
const defaultSpawnOptions: SpawnOptions = { cwd: "" };
const defaultSpawnArgs: string[] = [];

declare namespace __scriptgo {
    function execSync(command: string, cwd?: string, input?: string): string;
    function spawnSync(command: string, args?: string[], cwd?: string, input?: string): SpawnSyncReturns;
    function childProcessSpawnAsync(command: string, args?: string[], cwd?: string): ChildProcessSpawnResult;
    function childProcessPollStatus(pid: number): ChildProcessPollResult;
    function childProcessPipeRead(fd: number, maxLen: number): ChildProcessReadResult;
    function childProcessPipeWrite(fd: number, data: string, len: number): number;
    function childProcessPipeClose(fd: number): void;
    function childProcessKill(pid: number, sig: number): boolean;
}

export class ChildProcessPipeWritable extends Writable {
    _fd: number;

    constructor(fd: number) {
        super();
        this._fd = fd;
    }

    _write(chunk: StreamChunk, encoding: string, callback: (err?: Error | null) => void): void {
        if (this._fd < 0) {
            callback(new Error("pipe is closed"));
            return;
        }
        let str = typeof chunk === "string" ? chunk : chunk.toString();
        try {
            __scriptgo.childProcessPipeWrite(this._fd, str, str.length);
            callback(null);
        } catch (e: unknown) {
            callback(e instanceof Error ? e : new Error(String(e)));
        }
    }

    _destroy(error: Error | null, callback: (err: Error | null) => void): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.childProcessPipeClose(this._fd);
            } catch {}
            this._fd = -1;
        }
        callback(error);
    }
}

export class ChildProcessPipeReadable extends Readable {
    _fd: number;
    _readEnded: boolean = false;

    constructor(fd: number) {
        super();
        this._fd = fd;
    }

    _read(size: number): void {
        if (this._fd < 0 || this._readEnded) {
            return;
        }
        try {
            const res = __scriptgo.childProcessPipeRead(this._fd, size > 0 ? size : 65536);
            if (res.bytesRead > 0) {
                this.push(res.data);
            }
            if (res.eof) {
                this._readEnded = true;
                this.push(null);
            }
        } catch {}
    }

    _destroy(error: Error | null, callback: (err: Error | null) => void): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.childProcessPipeClose(this._fd);
            } catch {}
            this._fd = -1;
        }
        callback(error);
    }
}

export class ChildProcess extends EventEmitter {
    pid: number;
    stdin: ChildProcessPipeWritable | null = null;
    stdout: ChildProcessPipeReadable | null = null;
    stderr: ChildProcessPipeReadable | null = null;
    stdio: [ChildProcessPipeWritable | null, ChildProcessPipeReadable | null, ChildProcessPipeReadable | null];
    killed: boolean = false;
    exitCode: number = -1;
    signalCode: string | null = null;
    spawnfile: string;
    spawnargs: string[];
    connected: boolean = false;
    private _intervalId: number | null = null;
    private _closed: boolean = false;

    constructor(command: string, args: string[], spawnRes: ChildProcessSpawnResult) {
        super();
        this.spawnfile = command;
        this.spawnargs = args;
        this.pid = spawnRes.pid;

        if (spawnRes.stdinFd >= 0) {
            this.stdin = new ChildProcessPipeWritable(spawnRes.stdinFd);
        }
        if (spawnRes.stdoutFd >= 0) {
            this.stdout = new ChildProcessPipeReadable(spawnRes.stdoutFd);
        }
        if (spawnRes.stderrFd >= 0) {
            this.stderr = new ChildProcessPipeReadable(spawnRes.stderrFd);
        }
        this.stdio = [this.stdin, this.stdout, this.stderr];

        this._intervalId = setInterval(() => {
            this._poll();
        }, 10);
    }

    kill(signal: string | number = "SIGTERM"): boolean {
        let sigNum = 15;
        if (typeof signal === "number") {
            sigNum = signal;
        } else if (signal === "SIGKILL" || signal === "9") {
            sigNum = 9;
        } else if (signal === "SIGINT" || signal === "2") {
            sigNum = 2;
        }
        this.killed = true;
        return __scriptgo.childProcessKill(this.pid, sigNum);
    }

    ref(): this { return this; }
    unref(): this { return this; }
    disconnect(): void { this.connected = false; }

    _poll(): void {
        if (this._closed) return;

        // Drain stdout
        if (this.stdout !== null && !this.stdout._readEnded && this.stdout._fd >= 0) {
            try {
                const r = __scriptgo.childProcessPipeRead(this.stdout._fd, 65536);
                if (r.bytesRead > 0) {
                    this.stdout.push(r.data);
                }
                if (r.eof) {
                    this.stdout._readEnded = true;
                    this.stdout.push(null);
                }
            } catch {}
        }

        // Drain stderr
        if (this.stderr !== null && !this.stderr._readEnded && this.stderr._fd >= 0) {
            try {
                const r = __scriptgo.childProcessPipeRead(this.stderr._fd, 65536);
                if (r.bytesRead > 0) {
                    this.stderr.push(r.data);
                }
                if (r.eof) {
                    this.stderr._readEnded = true;
                    this.stderr.push(null);
                }
            } catch {}
        }

        // Check process exit status
        try {
            const st = __scriptgo.childProcessPollStatus(this.pid);
            if (st.exited) {
                this._finish(st.status);
            }
        } catch {}
    }

    private _finish(status: number): void {
        if (this._closed) return;
        this._closed = true;

        if (this._intervalId !== null) {
            clearInterval(this._intervalId);
            this._intervalId = null;
        }

        // Final drain of stdout
        if (this.stdout !== null && !this.stdout._readEnded && this.stdout._fd >= 0) {
            while (true) {
                const r = __scriptgo.childProcessPipeRead(this.stdout._fd, 65536);
                if (r.bytesRead > 0) {
                    this.stdout.push(r.data);
                }
                if (r.eof || r.bytesRead === 0) {
                    break;
                }
            }
            this.stdout._readEnded = true;
            this.stdout.push(null);
            this.stdout.destroy();
        }

        // Final drain of stderr
        if (this.stderr !== null && !this.stderr._readEnded && this.stderr._fd >= 0) {
            while (true) {
                const r = __scriptgo.childProcessPipeRead(this.stderr._fd, 65536);
                if (r.bytesRead > 0) {
                    this.stderr.push(r.data);
                }
                if (r.eof || r.bytesRead === 0) {
                    break;
                }
            }
            this.stderr._readEnded = true;
            this.stderr.push(null);
            this.stderr.destroy();
        }

        if (this.stdin !== null) {
            this.stdin.destroy();
        }

        this.exitCode = status;
        this.signalCode = status > 128 ? "SIGTERM" : null;

        this.emit("exit", status, this.signalCode);
        this.emit("close", status, this.signalCode);
    }
}

export function execSync(command: string, options: ExecSyncOptions = defaultExecOptions): string {
    let cwd = "";
    let input = "";
    if (options.cwd !== undefined) {
        cwd = options.cwd;
    }
    if (options.input !== undefined) {
        input = options.input;
    }
    return __scriptgo.execSync(command, cwd, input);
}

export function spawnSync(command: string, args: string[] = defaultSpawnArgs, options: SpawnSyncOptions = defaultSpawnOptions): SpawnSyncReturns {
    let cwd = "";
    let input = "";
    if (options.cwd !== undefined) {
        cwd = options.cwd;
    }
    if (options.input !== undefined) {
        input = options.input;
    }
    const raw = __scriptgo.spawnSync(command, args, cwd, input);
    return new SpawnSyncReturns(raw.stdout, raw.stderr, raw.status);
}

export function execFileSync(file: string, args: string[] = defaultSpawnArgs, options: ExecSyncOptions = defaultExecOptions): string {
    const ret = spawnSync(file, args, options);
    if (ret.status !== 0) {
        throw new Error("Command failed: " + file + (ret.stderr.length > 0 ? "\n" + ret.stderr : ""));
    }
    return ret.stdout;
}

export function spawn(command: string, args: string[] = defaultSpawnArgs, options: SpawnOptions = defaultSpawnOptions): ChildProcess {
    let cwd = "";
    if (options.cwd !== undefined) {
        cwd = options.cwd;
    }
    const raw = __scriptgo.childProcessSpawnAsync(command, args, cwd);
    const cp = new ChildProcess(command, args, raw);
    queueMicrotask(() => {
        cp.emit("spawn");
    });
    return cp;
}

export function exec(
    command: string,
    optionsOrCallback?: ExecOptions | ((error: Error | null, stdout: string, stderr: string) => void),
    callback?: (error: Error | null, stdout: string, stderr: string) => void
): ChildProcess {
    let opts: ExecOptions = defaultExecOptions;
    let cb: ((error: Error | null, stdout: string, stderr: string) => void) | undefined = callback;

    if (typeof optionsOrCallback === "function") {
        cb = optionsOrCallback as (error: Error | null, stdout: string, stderr: string) => void;
    } else if (optionsOrCallback !== undefined && optionsOrCallback !== null) {
        opts = optionsOrCallback;
    }

    const cp = spawn("/bin/sh", ["-c", command], { cwd: opts.cwd });
    let stdoutData = "";
    let stderrData = "";

    if (cp.stdout !== null) {
        cp.stdout.on("data", (chunk: StreamChunk) => {
            stdoutData += String(chunk);
        });
    }
    if (cp.stderr !== null) {
        cp.stderr.on("data", (chunk: StreamChunk) => {
            stderrData += String(chunk);
        });
    }

    cp.on("close", (code: number) => {
        if (cb !== undefined) {
            if (code !== 0) {
                const err = new Error("Command failed: " + command + (stderrData.length > 0 ? "\n" + stderrData : ""));
                cb(err, stdoutData, stderrData);
            } else {
                cb(null, stdoutData, stderrData);
            }
        }
    });

    return cp;
}

export function execFile(
    file: string,
    argsOrOptionsOrCallback?: string[] | ExecFileOptions | ((error: Error | null, stdout: string, stderr: string) => void),
    optionsOrCallback?: ExecFileOptions | ((error: Error | null, stdout: string, stderr: string) => void),
    callback?: (error: Error | null, stdout: string, stderr: string) => void
): ChildProcess {
    let args: string[] = defaultSpawnArgs;
    let opts: ExecFileOptions = defaultExecOptions;
    let cb: ((error: Error | null, stdout: string, stderr: string) => void) | undefined = callback;

    if (Array.isArray(argsOrOptionsOrCallback)) {
        args = argsOrOptionsOrCallback;
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback as (error: Error | null, stdout: string, stderr: string) => void;
        } else if (optionsOrCallback !== undefined && optionsOrCallback !== null) {
            opts = optionsOrCallback;
        }
    } else if (typeof argsOrOptionsOrCallback === "function") {
        cb = argsOrOptionsOrCallback as (error: Error | null, stdout: string, stderr: string) => void;
    } else if (argsOrOptionsOrCallback !== undefined && argsOrOptionsOrCallback !== null) {
        opts = argsOrOptionsOrCallback as ExecFileOptions;
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback as (error: Error | null, stdout: string, stderr: string) => void;
        }
    }

    const cp = spawn(file, args, { cwd: opts.cwd });
    let stdoutData = "";
    let stderrData = "";

    if (cp.stdout !== null) {
        cp.stdout.on("data", (chunk: StreamChunk) => {
            stdoutData += String(chunk);
        });
    }
    if (cp.stderr !== null) {
        cp.stderr.on("data", (chunk: StreamChunk) => {
            stderrData += String(chunk);
        });
    }

    cp.on("close", (code: number) => {
        if (cb !== undefined) {
            if (code !== 0) {
                const err = new Error("Command failed: " + file + (stderrData.length > 0 ? "\n" + stderrData : ""));
                cb(err, stdoutData, stderrData);
            } else {
                cb(null, stdoutData, stderrData);
            }
        }
    });

    return cp;
}

export function fork(modulePath: string, args: string[] = defaultSpawnArgs, options: ForkOptions = defaultSpawnOptions): ChildProcess {
    const execPath = "scriptgo";
    const runArgs = ["run", modulePath];
    for (let i = 0; i < args.length; i++) {
        runArgs.push(args[i]);
    }
    return spawn(execPath, runArgs, options);
}

export default {
    ChildProcess,
    ChildProcessPipeWritable,
    ChildProcessPipeReadable,
    ChildProcessSpawnResult,
    ChildProcessPollResult,
    ChildProcessReadResult,
    SpawnSyncReturns,
    execFileSync,
    execSync,
    spawnSync,
    spawn,
    exec,
    execFile,
    fork,
};
