// Node.js Cluster module (node:cluster)
import { EventEmitter } from "node:events";

export interface ClusterSettings {
    execArgv?: string[];
    exec?: string;
    args?: string[];
    cwd?: string;
    serialization?: string;
    silent?: boolean;
    stdio?: unknown[];
    uid?: number;
    gid?: number;
    inspectPort?: number;
    windowsHide?: boolean;
}

export class WorkerProcess extends EventEmitter {
    pid: number = 1;
}

export class Worker extends EventEmitter {
    id: number = 1;
    process: WorkerProcess;
    exitedAfterDisconnect: boolean = false;
    private _dead: boolean = false;
    private _connected: boolean = true;

    constructor(id: number = 1) {
        super();
        this.id = id;
        this.process = new WorkerProcess();
        this.exitedAfterDisconnect = false;
        this._dead = false;
        this._connected = true;
    }

    isConnected(): boolean {
        return this._connected;
    }

    isDead(): boolean {
        return this._dead;
    }

    send(message: unknown, sendHandle?: unknown, options?: unknown, callback?: (error: Error | null) => void): boolean {
        if (callback) {
            callback(null);
        }
        return true;
    }

    kill(signal?: string): void {
        this._dead = true;
        this._connected = false;
        this.emit("exit", 0, signal || "SIGTERM");
    }

    disconnect(): this {
        this.exitedAfterDisconnect = true;
        this._connected = false;
        this.emit("disconnect");
        return this;
    }
}

export const isMaster: boolean = true;
export const isPrimary: boolean = true;
export const isWorker: boolean = false;
export const schedulingPolicy: number = 2; // SCHED_RR
export const settings: ClusterSettings = {};
export const worker: Worker | undefined = undefined;
export const workers: Record<string, Worker | undefined> = {};

export function setupMaster(settings?: ClusterSettings): void {
    setupPrimary(settings);
}

export function setupPrimary(settings?: ClusterSettings): void {
    if (settings) {
        Object.assign(settings, settings);
    }
}

export function fork(env?: Record<string, string>): Worker {
    const w = new Worker(1);
    return w;
}

export function disconnect(callback?: () => void): void {
    if (callback) {
        callback();
    }
}

export default {
    isMaster,
    isPrimary,
    isWorker,
    schedulingPolicy,
    settings,
    worker,
    workers,
    setupMaster,
    setupPrimary,
    fork,
    disconnect,
    Worker,
    WorkerProcess,
};
