// Node.js Inspector module (node:inspector)

export class Session {
    constructor() {}

    connect(): void {}

    connectToMainThread(): void {}

    disconnect(): void {}

    post(method: string, params?: unknown, callback?: unknown): void {
        const cb = typeof params === "function" ? params : callback;
        if (typeof cb === "function") {
            (cb as Function)(null, {});
        }
    }
}

export function open(port?: number, host?: string, wait?: boolean): void {}

export function close(): void {}

export function url(): string | undefined {
    return undefined;
}

export function waitForDebugger(): void {}

export const inspectorConsole: Record<string, unknown> = {};
export { inspectorConsole as console };

export function dataReceived(options: unknown): void {}
export function dataSent(options: unknown): void {}
export function requestWillBeSent(options: unknown): void {}
export function responseReceived(options: unknown): void {}
export function loadingFinished(options: unknown): void {}
export function loadingFailed(options: unknown): void {}
export function put(options: unknown): void {}

export default {
    Session,
    open,
    close,
    url,
    waitForDebugger,
    console: inspectorConsole,
    dataReceived,
    dataSent,
    requestWillBeSent,
    responseReceived,
    loadingFinished,
    loadingFailed,
    put,
};
