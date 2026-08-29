// Node.js Worker Threads module (node:worker_threads)
import { EventEmitter } from "node:events";
import { Readable, Writable } from "node:stream";

export class ResourceLimits {
    maxYoungGenerationSizeMb: number = 0;
    maxOldGenerationSizeMb: number = 0;
    codeRangeSizeMb: number = 0;
    stackSizeMb: number = 4;
}

export const isMainThread: boolean = true;
export const isInternalThread: boolean = false;
export const parentPort: MessagePort | null = null;
export const threadId: number = 0;
export const threadName: string = "";
export const workerData: unknown = undefined;
export const SHARE_ENV: symbol = Symbol("SHARE_ENV");
export const resourceLimits: ResourceLimits = new ResourceLimits();

let _envKeys: string[] = [];
let _envVals: unknown[] = [];

export function getEnvironmentData(key: string): unknown {
    const idx = _envKeys.indexOf(key);
    if (idx !== -1) {
        return _envVals[idx];
    }
    return undefined;
}

export function setEnvironmentData(key: string, value: unknown): void {
    const idx = _envKeys.indexOf(key);
    if (idx !== -1) {
        _envVals[idx] = value;
    } else {
        _envKeys.push(key);
        _envVals.push(value);
    }
}

export function markAsUntransferable(object: object): void {}
export function isMarkedAsUntransferable(object: object): boolean {
    return false;
}
export function markAsUncloneable(object: object): void {}
export function moveMessagePortToContext(port: MessagePort, context: unknown): MessagePort {
    return port;
}
export function postMessageToThread(threadId: number, value: unknown, transferList?: unknown[]): void {}
export function receiveMessageOnPort(port: MessagePort): { message: unknown } | undefined {
    return undefined;
}

export class MessagePort extends EventEmitter {
    private _refed: boolean = true;

    constructor() {
        super();
        this._refed = true;
    }

    postMessage(value: unknown, transferList?: unknown[]): void {
        this.emit("message", value);
    }

    start(): void {}

    close(): void {
        this.emit("close");
    }

    ref(): this {
        this._refed = true;
        return this;
    }

    unref(): this {
        this._refed = false;
        return this;
    }

    hasRef(): boolean {
        return this._refed;
    }
}

export class MessageChannel {
    port1: MessagePort;
    port2: MessagePort;

    constructor() {
        this.port1 = new MessagePort();
        this.port2 = new MessagePort();
    }
}

export class BroadcastChannel {
    name: string = "";
    onmessage?: (event: unknown) => void;
    onmessageerror?: (event: unknown) => void;
    private _refed: boolean = true;

    constructor(name: string) {
        this.name = name;
        this._refed = true;
    }

    postMessage(message: unknown): void {
        if (this.onmessage) {
            this.onmessage({ data: message });
        }
    }

    close(): void {}

    ref(): this {
        this._refed = true;
        return this;
    }

    unref(): this {
        this._refed = false;
        return this;
    }
}

export class WorkerPerformance {
    eventLoopUtilization(): { idle: number; active: number; utilization: number } {
        return { idle: 0, active: 0, utilization: 0 };
    }
}

export class WorkerCpuUsage {
    user: number = 0;
    system: number = 0;
}

export class Worker extends EventEmitter {
    threadId: number = 1;
    threadName: string = "";
    stdin: Writable | null = null;
    stdout: Readable | null = null;
    stderr: Readable | null = null;
    performance: WorkerPerformance;
    resourceLimits: ResourceLimits;
    private _refed: boolean = true;

    constructor(filename: string | URL, options?: unknown) {
        super();
        this.threadId = 1;
        this.threadName = "";
        this.stdin = null;
        this.stdout = null;
        this.stderr = null;
        this.performance = new WorkerPerformance();
        this.resourceLimits = new ResourceLimits();
        this._refed = true;
    }

    postMessage(value: unknown, transferList?: unknown[]): void {
        this.emit("message", value);
    }

    ref(): this {
        this._refed = true;
        return this;
    }

    unref(): this {
        this._refed = false;
        return this;
    }

    async terminate(): Promise<number> {
        this.emit("exit", 0);
        return 0;
    }

    getHeapSnapshot(options?: unknown): Promise<Readable> {
        return Promise.resolve(new Readable());
    }

    getHeapStatistics(): Promise<unknown> {
        return Promise.resolve({});
    }

    cpuUsage(previousValue?: WorkerCpuUsage): WorkerCpuUsage {
        const usage = new WorkerCpuUsage();
        usage.user = 0;
        usage.system = 0;
        return usage;
    }

    startCpuProfile(): unknown {
        return undefined;
    }

    async [Symbol.asyncDispose](): Promise<void> {
        await this.terminate();
    }
}

export default {
    isMainThread,
    isInternalThread,
    parentPort,
    threadId,
    threadName,
    workerData,
    SHARE_ENV,
    resourceLimits,
    ResourceLimits,
    getEnvironmentData,
    setEnvironmentData,
    markAsUntransferable,
    isMarkedAsUntransferable,
    markAsUncloneable,
    moveMessagePortToContext,
    postMessageToThread,
    receiveMessageOnPort,
    MessagePort,
    MessageChannel,
    BroadcastChannel,
    Worker,
    WorkerPerformance,
    WorkerCpuUsage,
};
