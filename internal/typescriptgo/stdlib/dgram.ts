// ScriptGo Standard Library: node:dgram

class DgramListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class DgramEventBucket {
    name: string;
    listeners: DgramListenerEntry[];

    constructor(name: string, listeners: DgramListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export interface SocketOptions {
    type: string;
    reuseAddr?: boolean;
    ipv6Only?: boolean;
    recvBufferSize?: number;
    sendBufferSize?: number;
    signal?: unknown;
}

export class Socket {
    type: string = "udp4";
    private _bound: boolean = false;
    private _closed: boolean = false;
    private _connected: boolean = false;
    private _broadcast: boolean = false;
    private _multicastLoopback: boolean = true;
    private _ttl: number = 64;
    private _multicastTTL: number = 1;
    private _multicastInterface: string = "";
    private _recvBufferSize: number = 65536;
    private _sendBufferSize: number = 65536;
    private _remotePort: number = 0;
    private _remoteAddress: string = "";
    private _remoteFamily: string = "IPv4";

    _buckets: DgramEventBucket[] = [];

    constructor(typeOrOptions: string | SocketOptions = "udp4", listener?: Function) {
        if (typeof typeOrOptions === "string") {
            this.type = typeOrOptions;
        } else if (typeOrOptions && typeof typeOrOptions === "object") {
            if (typeOrOptions.type) this.type = typeOrOptions.type;
            if (typeOrOptions.recvBufferSize) this._recvBufferSize = typeOrOptions.recvBufferSize;
            if (typeOrOptions.sendBufferSize) this._sendBufferSize = typeOrOptions.sendBufferSize;
        }
        if (typeof listener === "function") {
            this.on("message", listener);
        }
    }

    private _getBucket(name: string): DgramEventBucket {
        for (let i = 0; i < this._buckets.length; i++) {
            if (this._buckets[i].name === name) {
                return this._buckets[i];
            }
        }
        const created = new DgramEventBucket(name, []);
        this._buckets.push(created);
        return created;
    }

    on(event: string, listener: Function): Socket {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new DgramListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): Socket {
        const bucket = this._getBucket(event);
        bucket.listeners.push(new DgramListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const bucket = this._getBucket(event);
        if (bucket.listeners.length === 0) {
            return false;
        }
        const remaining: DgramListenerEntry[] = [];
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

    bind(portOrOptions?: unknown, addressOrCb?: unknown, cb?: Function): Socket {
        this._bound = true;
        let callback: Function | null = null;
        if (typeof addressOrCb === "function") {
            callback = addressOrCb;
        } else if (typeof cb === "function") {
            callback = cb;
        }
        if (callback) {
            this.once("listening", callback);
        }
        this.emit("listening");
        return this;
    }

    send(msg: unknown, offsetOrPort?: unknown, lengthOrAddress?: unknown, portOrCb?: unknown, addressOrCb?: unknown, cb?: Function): void {
        let callback: Function | null = null;
        if (typeof cb === "function") {
            callback = cb;
        } else if (typeof addressOrCb === "function") {
            callback = addressOrCb;
        } else if (typeof portOrCb === "function") {
            callback = portOrCb;
        }
        if (callback) {
            callback(null, 0);
        }
    }

    close(callback?: Function): Socket {
        this._closed = true;
        this._bound = false;
        if (callback) {
            this.once("close", callback);
        }
        this.emit("close");
        return this;
    }

    address(): { address: string; family: string; port: number } {
        return {
            address: "127.0.0.1",
            family: "IPv4",
            port: 41234,
        };
    }

    connect(port: number, addressOrCb?: unknown, cb?: Function): void {
        this._connected = true;
        this._remotePort = port;
        if (typeof addressOrCb === "string") {
            this._remoteAddress = addressOrCb;
        } else if (typeof addressOrCb === "function") {
            this.once("connect", addressOrCb);
        }
        if (typeof cb === "function") {
            this.once("connect", cb);
        }
        this.emit("connect");
    }

    disconnect(): void {
        this._connected = false;
        this._remotePort = 0;
        this._remoteAddress = "";
    }

    remoteAddress(): { address: string; family: string; port: number } {
        return {
            address: this._remoteAddress.length > 0 ? this._remoteAddress : "127.0.0.1",
            family: this._remoteFamily,
            port: this._remotePort,
        };
    }

    setBroadcast(flag: boolean = true): void {
        this._broadcast = flag;
    }

    setTTL(ttl: number): number {
        this._ttl = ttl;
        return this._ttl;
    }

    setMulticastTTL(ttl: number): number {
        this._multicastTTL = ttl;
        return this._multicastTTL;
    }

    setMulticastLoopback(flag: boolean = true): boolean {
        this._multicastLoopback = flag;
        return this._multicastLoopback;
    }

    setMulticastInterface(multicastInterface: string): void {
        this._multicastInterface = multicastInterface;
    }

    addMembership(multicastAddress: string, multicastInterface?: string): void {
    }

    dropMembership(multicastAddress: string, multicastInterface?: string): void {
    }

    addSourceSpecificMembership(sourceAddress: string, groupAddress: string, multicastInterface?: string): void {
    }

    dropSourceSpecificMembership(sourceAddress: string, groupAddress: string, multicastInterface?: string): void {
    }

    setRecvBufferSize(size: number): void {
        this._recvBufferSize = size;
    }

    setSendBufferSize(size: number): void {
        this._sendBufferSize = size;
    }

    getRecvBufferSize(): number {
        return this._recvBufferSize;
    }

    getSendBufferSize(): number {
        return this._sendBufferSize;
    }

    getSendQueueSize(): number {
        return 0;
    }

    getSendQueueCount(): number {
        return 0;
    }

    ref(): Socket {
        return this;
    }

    unref(): Socket {
        return this;
    }

    [Symbol.asyncDispose](): Promise<void> {
        this.close();
        return Promise.resolve(undefined);
    }
}

export function createSocket(typeOrOptions: string | SocketOptions = "udp4", callback?: Function): Socket {
    return new Socket(typeOrOptions, callback);
}

export default {
    Socket,
    createSocket,
};
