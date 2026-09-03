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

export function execFileSync(file: string, args: string[] = defaultSpawnArgs, options?: ExecSyncOptions): string {
    return execSync(file, options);
}

export default {
    SpawnSyncReturns,
    execFileSync,
    execSync,
    spawnSync,
};
