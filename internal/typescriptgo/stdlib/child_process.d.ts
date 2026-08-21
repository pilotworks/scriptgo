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
    constructor(stdout: string, stderr: string, status: number);
}
export function execSync(command: string, options?: ExecSyncOptions): string;
export function spawnSync(command: string, args?: string[], options?: SpawnSyncOptions): SpawnSyncReturns;
