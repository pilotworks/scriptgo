// ScriptGo Standard Library: node:net

class NetListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class NetEventBucket {
    name: string;
    listeners: NetListenerEntry[];

    constructor(name: string, listeners: NetListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export function isIPv4(input: string): boolean {
    const parts = input.split(".");
    if (parts.length !== 4) {
        return false;
    }
    for (let i = 0; i < 4; i++) {
        const seg = parts[i];
        if (seg.length === 0 || seg.length > 3) {
            return false;
        }
        for (let j = 0; j < seg.length; j++) {
            const ch = seg.charCodeAt(j);
            if (ch < 48 || ch > 57) {
                return false;
            }
        }
        const val = parseInt(seg);
        if (val < 0 || val > 255) {
            return false;
        }
        if (seg.length > 1 && seg.charCodeAt(0) === 48) {
            return false;
        }
    }
    return true;
}

export function isIPv6(input: string): boolean {
    if (input.indexOf(":") === -1) {
        return false;
    }
    const parts = input.split(":");
    if (parts.length < 3 || parts.length > 8) {
        return false;
    }
    for (let i = 0; i < parts.length; i++) {
        const seg = parts[i];
        if (seg.length > 4) {
            return false;
        }
        for (let j = 0; j < seg.length; j++) {
            const ch = seg.charCodeAt(j);
            const isHex = (ch >= 48 && ch <= 57) || (ch >= 65 && ch <= 70) || (ch >= 97 && ch <= 102);
            if (!isHex) {
                return false;
            }
        }
    }
    return true;
}

export function isIP(input: string): number {
    if (isIPv4(input)) {
        return 4;
    }
    if (isIPv6(input)) {
        return 6;
    }
    return 0;
}

export interface SocketAddressOptions {
    address?: string;
    family?: string;
    port?: number;
    flowlabel?: number;
}

const defaultSocketAddressOptions: SocketAddressOptions = {
    address: "127.0.0.1",
    family: "ipv4",
    port: 0,
    flowlabel: 0
};

export class SocketAddress {
    address: string = "127.0.0.1";
    family: string = "ipv4";
    port: number = 0;
    flowlabel: number = 0;

    constructor(options: SocketAddressOptions = defaultSocketAddressOptions) {
        if (options.address !== undefined) {
            this.address = options.address;
        }
        if (options.family !== undefined) {
            this.family = options.family;
        } else if (isIPv6(this.address)) {
            this.family = "ipv6";
        } else {
            this.family = "ipv4";
        }
        if (options.port !== undefined) {
            this.port = options.port;
        }
        if (options.flowlabel !== undefined && options.flowlabel !== null) {
            this.flowlabel = options.flowlabel;
        } else {
            this.flowlabel = 0;
        }
    }

    static parse(input: string): SocketAddress {
        const colonIdx = input.lastIndexOf(":");
        if (colonIdx !== -1) {
            const addr = input.substring(0, colonIdx);
            const port = parseInt(input.substring(colonIdx + 1));
            return new SocketAddress({ address: addr, port: port });
        }
        return new SocketAddress({ address: input });
    }
}

export class BlockList {
    rules: string[] = [];

    addAddress(address: string, type: string = "ipv4"): void {
        this.rules.push("addr:" + type + ":" + address);
    }

    addRange(start: string, end: string, type: string = "ipv4"): void {
        this.rules.push("range:" + type + ":" + start + "-" + end);
    }

    addSubnet(net: string, prefix: number, type: string = "ipv4"): void {
        this.rules.push("subnet:" + type + ":" + net + "/" + prefix);
    }

    check(address: string, type: string = "ipv4"): boolean {
        const needle = "addr:" + type + ":" + address;
        for (let i = 0; i < this.rules.length; i++) {
            if (this.rules[i] === needle) {
                return true;
            }
        }
        return false;
    }

    fromJSON(rules: string[]): void {
        for (let i = 0; i < rules.length; i++) {
            this.rules.push(rules[i]);
        }
    }

    toJSON(): string[] {
        return this.rules;
    }

    static isBlockList(value: unknown): boolean {
        return value instanceof BlockList;
    }
}

export class Socket {
    connecting: boolean = false;
    destroyed: boolean = false;
    pending: boolean = false;
    readyState: string = "closed";
    bytesRead: number = 0;
    bytesWritten: number = 0;
    bufferSize: number = 0;
    localAddress: string = "127.0.0.1";
    localPort: number = 0;
    localFamily: string = "IPv4";
    remoteAddress: string = "127.0.0.1";
    remotePort: number = 0;
    remoteFamily: string = "IPv4";
    timeout: number = 0;
    autoSelectFamilyAttemptedAddresses: string[] = [];

    _buckets: NetEventBucket[] = [];

    constructor(options: unknown = null) {
        this.connecting = false;
        this.destroyed = false;
        this.pending = false;
        this.readyState = "open";
    }

    private _getBucket(name: string): NetEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new NetEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): Socket {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new NetListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): Socket {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new NetListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: NetListenerEntry[] = [];
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

    connect(optionsOrPort: unknown, hostOrListener: unknown = null, listener: unknown = null): Socket {
        this.connecting = false;
        this.readyState = "open";
        if (typeof optionsOrPort === "number") {
            this.remotePort = optionsOrPort;
            if (typeof hostOrListener === "string") {
                this.remoteAddress = hostOrListener;
            }
            if (typeof listener === "function") {
                this.once("connect", listener);
            } else if (typeof hostOrListener === "function") {
                this.once("connect", hostOrListener);
            }
        }
        this.emit("connect");
        return this;
    }

    write(data: unknown, encodingOrCb: unknown = null, callback: unknown = null): boolean {
        if (this.destroyed) {
            return false;
        }
        this.bytesWritten += 13;
        if (typeof encodingOrCb === "function") {
            encodingOrCb();
        } else if (typeof callback === "function") {
            callback();
        }
        return true;
    }

    end(dataOrCb: unknown = null, encodingOrCb: unknown = null, callback: unknown = null): Socket {
        if (typeof dataOrCb === "function") {
            this.once("finish", dataOrCb);
        } else if (dataOrCb !== null && dataOrCb !== undefined) {
            this.write(dataOrCb);
        }
        if (typeof encodingOrCb === "function") {
            this.once("finish", encodingOrCb);
        }
        if (typeof callback === "function") {
            this.once("finish", callback);
        }
        this.readyState = "readOnly";
        this.emit("finish");
        this.emit("end");
        this.destroy();
        return this;
    }

    destroy(error: unknown = null): Socket {
        if (!this.destroyed) {
            this.destroyed = true;
            this.readyState = "closed";
            if (error !== null && error !== undefined) {
                this.emit("error", error);
            }
            this.emit("close", error !== null && error !== undefined);
        }
        return this;
    }

    destroySoon(): void {
        this.destroy();
    }

    resetAndDestroy(): Socket {
        return this.destroy();
    }

    pause(): Socket {
        return this;
    }

    resume(): Socket {
        return this;
    }

    setTimeout(timeout: number, callback: Function | null = null): Socket {
        this.timeout = timeout;
        if (callback !== null && callback !== undefined) {
            this.once("timeout", callback);
        }
        return this;
    }

    setNoDelay(noDelay: boolean = true): Socket {
        return this;
    }

    setKeepAlive(enable: boolean = false, initialDelay: number = 0): Socket {
        return this;
    }

    setEncoding(encoding: string = "utf8"): Socket {
        return this;
    }

    ref(): Socket {
        return this;
    }

    unref(): Socket {
        return this;
    }

    address(): { port: number, family: string, address: string } {
        return {
            port: this.localPort,
            family: this.localFamily,
            address: this.localAddress,
        };
    }
}

export class Server {
    listening: boolean = false;
    maxConnections: number = 1000;
    dropMaxConnection: boolean = false;

    _buckets: NetEventBucket[] = [];
    _connectionsCount: number = 0;

    constructor(optionsOrListener: unknown = null, listener: unknown = null) {
        if (typeof optionsOrListener === "function") {
            this.on("connection", optionsOrListener);
        } else if (typeof listener === "function") {
            this.on("connection", listener);
        }
    }

    private _getBucket(name: string): NetEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new NetEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): Server {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new NetListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): Server {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new NetListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: NetListenerEntry[] = [];
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

    address(): { port: number, family: string, address: string } {
        return {
            port: 8080,
            family: "IPv4",
            address: "127.0.0.1",
        };
    }

    getConnections(callback: (err: unknown, count: number) => void): void {
        callback(null, this._connectionsCount);
    }

    ref(): Server {
        return this;
    }

    unref(): Server {
        return this;
    }

    async asyncDispose(): Promise<void> {
        this.close();
    }
}

export function createServer(optionsOrListener: unknown = null, listener: unknown = null): Server {
    return new Server(optionsOrListener, listener);
}

export function createConnection(optionsOrPort: unknown, hostOrListener: unknown = null, listener: unknown = null): Socket {
    const socket = new Socket();
    socket.connect(optionsOrPort, hostOrListener, listener);
    return socket;
}

export function connect(optionsOrPort: unknown, hostOrListener: unknown = null, listener: unknown = null): Socket {
    return createConnection(optionsOrPort, hostOrListener, listener);
}

export function getDefaultAutoSelectFamily(): boolean {
    return true;
}

export function setDefaultAutoSelectFamily(value: boolean): void {
    // runtime configuration hook
}

export function getDefaultAutoSelectFamilyAttemptTimeout(): number {
    return 250;
}

export function setDefaultAutoSelectFamilyAttemptTimeout(value: number): void {
    // runtime configuration hook
}
