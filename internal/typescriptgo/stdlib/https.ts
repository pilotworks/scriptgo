// ScriptGo Standard Library: node:https

class HttpsListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class HttpsEventBucket {
    name: string;
    listeners: HttpsListenerEntry[];

    constructor(name: string, listeners: HttpsListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export interface HttpsAgentOptions {
    maxSockets?: number;
    maxFreeSockets?: number;
    maxTotalSockets?: number;
    maxCachedSessions?: number;
    keepAlive?: boolean;
    keepAliveMsecs?: number;
}

const defaultHttpsAgentOptions: HttpsAgentOptions = {
    maxSockets: 999999,
    maxFreeSockets: 256,
    maxTotalSockets: 999999,
    maxCachedSessions: 100,
    keepAlive: false,
    keepAliveMsecs: 1000
};

export class Agent {
    maxSockets: number = 999999;
    maxFreeSockets: number = 256;
    maxTotalSockets: number = 999999;
    maxCachedSessions: number = 100;
    keepAlive: boolean = false;
    keepAliveMsecs: number = 1000;
    freeSockets: string[] = [];
    sockets: string[] = [];
    requests: string[] = [];

    constructor(options: HttpsAgentOptions = defaultHttpsAgentOptions) {
        if (options.maxSockets !== undefined) {
            this.maxSockets = options.maxSockets;
        }
        if (options.maxFreeSockets !== undefined) {
            this.maxFreeSockets = options.maxFreeSockets;
        }
        if (options.maxTotalSockets !== undefined) {
            this.maxTotalSockets = options.maxTotalSockets;
        }
        if (options.keepAlive !== undefined) {
            this.keepAlive = options.keepAlive;
        }
        if (options.keepAliveMsecs !== undefined) {
            this.keepAliveMsecs = options.keepAliveMsecs;
        }
    }

    destroy(): void {
        this.freeSockets = [];
        this.sockets = [];
        this.requests = [];
    }
}

export const globalAgent: Agent = new Agent();

export class ClientRequest {
    url: string = "";
    method: string = "GET";
    headersKeys: string[] = [];
    headersValues: string[] = [];
    finished: boolean = false;

    _buckets: HttpsEventBucket[] = [];

    constructor(urlOrOptions: unknown, cb: unknown = null) {
        if (typeof urlOrOptions === "string") {
            this.url = urlOrOptions;
        }
        if (typeof cb === "function") {
            this.once("response", cb);
        }
    }

    private _getBucket(name: string): HttpsEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new HttpsEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): ClientRequest {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new HttpsListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): ClientRequest {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new HttpsListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: HttpsListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) {
                remaining.push(entry);
            }
        }
        bucket.listeners = remaining;
        return true;
    }

    setHeader(name: string, value: string): void {
        const lower = name.toLowerCase();
        for (let i = 0; i < this.headersKeys.length; i++) {
            if (this.headersKeys[i] === lower) {
                this.headersValues[i] = value;
                return;
            }
        }
        this.headersKeys.push(lower);
        this.headersValues.push(value);
    }

    getHeader(name: string): string | null {
        const lower = name.toLowerCase();
        for (let i = 0; i < this.headersKeys.length; i++) {
            if (this.headersKeys[i] === lower) {
                return this.headersValues[i];
            }
        }
        return null;
    }

    write(chunk: unknown, encodingOrCb: unknown = null, callback: unknown = null): boolean {
        if (typeof encodingOrCb === "function") {
            encodingOrCb();
        } else if (typeof callback === "function") {
            callback();
        }
        return true;
    }

    end(dataOrCb: unknown = null, encodingOrCb: unknown = null, callback: unknown = null): ClientRequest {
        this.finished = true;
        if (typeof dataOrCb === "function") {
            dataOrCb();
        }
        if (typeof encodingOrCb === "function") {
            encodingOrCb();
        }
        if (typeof callback === "function") {
            callback();
        }
        this.emit("finish");
        return this;
    }

    destroy(error: unknown = null): ClientRequest {
        this.finished = true;
        if (error !== null && error !== undefined) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }

    setTimeout(timeout: number, callback: Function | null = null): ClientRequest {
        if (callback !== null && callback !== undefined) {
            this.once("timeout", callback);
        }
        return this;
    }
}

export interface HttpsServerOptions {
    headersTimeout?: number;
    keepAliveTimeout?: number;
    maxHeadersCount?: number;
    requestTimeout?: number;
    timeout?: number;
}

const defaultHttpsServerOptions: HttpsServerOptions = {
    headersTimeout: 60000,
    keepAliveTimeout: 5000,
    maxHeadersCount: 2000,
    requestTimeout: 300000,
    timeout: 0
};

export class Server {
    listening: boolean = false;
    headersTimeout: number = 60000;
    keepAliveTimeout: number = 5000;
    maxHeadersCount: number = 2000;
    requestTimeout: number = 300000;
    timeout: number = 0;

    _buckets: HttpsEventBucket[] = [];

    constructor(optionsOrListener: unknown = null, requestListener: unknown = null) {
        if (typeof optionsOrListener === "function") {
            this.on("request", optionsOrListener);
        } else if (typeof requestListener === "function") {
            this.on("request", requestListener);
        }
    }

    private _getBucket(name: string): HttpsEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new HttpsEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): Server {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new HttpsListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): Server {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new HttpsListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: HttpsListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            const entry = bucket.listeners[i];
            entry.fn(arg1, arg2);
            if (!entry.once) {
                remaining.push(entry);
            }
        }
        bucket.listeners = remaining;
        return true;
    }

    listen(portOrOptions: unknown = 0, hostOrCb: unknown = null, callback: unknown = null): Server {
        this.listening = true;
        if (typeof hostOrCb === "function") {
            this.once("listening", hostOrCb);
        } else if (typeof callback === "function") {
            this.once("listening", callback);
        }
        this.emit("listening");
        return this;
    }

    close(callback: Function | null = null): Server {
        this.listening = false;
        if (callback !== null && callback !== undefined) {
            this.once("close", callback);
        }
        this.emit("close");
        return this;
    }

    closeAllConnections(): void {
        const dummy = true;
    }

    closeIdleConnections(): void {
        const dummy = true;
    }

    setTimeout(msecs: number = 120000, callback: Function | null = null): Server {
        this.timeout = msecs;
        if (callback !== null && callback !== undefined) {
            this.once("timeout", callback);
        }
        return this;
    }

    async asyncDispose(): Promise<void> {
        this.close();
    }
}

export function createServer(optionsOrListener: unknown = null, requestListener: unknown = null): Server {
    return new Server(optionsOrListener, requestListener);
}

export function request(urlOrOptions: unknown, optionsOrCb: unknown = null, cb: unknown = null): ClientRequest {
    if (typeof optionsOrCb === "function") {
        return new ClientRequest(urlOrOptions, optionsOrCb);
    }
    if (typeof cb === "function") {
        return new ClientRequest(urlOrOptions, cb);
    }
    return new ClientRequest(urlOrOptions, null);
}

export function get(urlOrOptions: unknown, optionsOrCb: unknown = null, cb: unknown = null): ClientRequest {
    const req = request(urlOrOptions, optionsOrCb, cb);
    req.end();
    return req;
}

export default {
    Agent,
    globalAgent,
    Server,
    ClientRequest,
    createServer,
    request,
    get,
};
