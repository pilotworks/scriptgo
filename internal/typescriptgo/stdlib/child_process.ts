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

export class ChildProcess {
    channel: unknown = null;
    connected: boolean = false;
    exitCode: number = 0;
    killed: boolean = false;
    pid: number = 0;
    signalCode: string | null = null;
    spawnargs: string[] = [];
    spawnfile: string = "";
    stdin: unknown = null;
    stdout: unknown = null;
    stderr: unknown = null;
    stdio: unknown[] = [];

    constructor(command: string = "", args: string[] = []) {
        this.spawnfile = command;
        this.spawnargs = args;
        this.pid = 1234;
        this.exitCode = 0;
        this.connected = false;
        this.killed = false;
    }

    kill(signal?: string | number): boolean {
        this.killed = true;
        return true;
    }

    disconnect(): void {
        this.connected = false;
    }

    ref(): ChildProcess {
        return this;
    }

    unref(): ChildProcess {
        return this;
    }

    send(message: unknown, sendHandle?: unknown, options?: unknown, callback?: (err: unknown) => void): boolean {
        if (callback) callback(null);
        return true;
    }

    [Symbol.dispose](): void {
        this.kill();
    }
}

const defaultExecOptions: ExecSyncOptions = { cwd: "", input: "" };
const defaultSpawnOptions: SpawnSyncOptions = { cwd: "", input: "" };
const defaultSpawnArgs: string[] = [];

declare namespace __scriptgo {
    function execSync(command: string, cwd?: string, input?: string): string;
    function spawnSync(command: string, args?: string[], cwd?: string, input?: string): SpawnSyncReturns;
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

export function spawn(command: string, args: string[] = defaultSpawnArgs, options?: unknown): ChildProcess {
    return new ChildProcess(command, args);
}

export function exec(command: string, optionsOrCb?: unknown, cb?: unknown): ChildProcess {
    const cp = new ChildProcess(command);
    if (typeof optionsOrCb === "function") {
        (optionsOrCb as Function)(null, "", "");
    } else if (typeof cb === "function") {
        (cb as Function)(null, "", "");
    }
    return cp;
}

export function execFile(file: string, argsOrOptions?: unknown, optionsOrCb?: unknown, cb?: unknown): ChildProcess {
    const cp = new ChildProcess(file);
    if (typeof optionsOrCb === "function") {
        (optionsOrCb as Function)(null, "", "");
    } else if (typeof cb === "function") {
        (cb as Function)(null, "", "");
    } else if (typeof argsOrOptions === "function") {
        (argsOrOptions as Function)(null, "", "");
    }
    return cp;
}

export function fork(modulePath: string, args: string[] = defaultSpawnArgs, options?: unknown): ChildProcess {
    return new ChildProcess(modulePath, args);
}

export function execFileSync(file: string, args: string[] = defaultSpawnArgs, options?: ExecSyncOptions): string {
    return execSync(file, options);
}

export default {
    ChildProcess,
    SpawnSyncReturns,
    exec,
    execFile,
    execFileSync,
    execSync,
    fork,
    spawn,
    spawnSync,
};
