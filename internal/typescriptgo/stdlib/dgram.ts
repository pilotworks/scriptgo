// ScriptGo Standard Library: node:dgram

declare namespace __scriptgo {
    function dgramSocketCreate(family?: number): number;
    function dgramBind(fd: number, address: string, port: number): void;
    function dgramSend(fd: number, data: string, len: number, port: number, address: string): number;
    function dgramRecv(fd: number, maxLen: number): { data: string; bytes: number; address: string; port: number; family: number };
    function dgramSetBroadcast(fd: number, flag: number): void;
    function dgramSetMulticastTTL(fd: number, ttl: number): void;
    function dgramSetMulticastLoopback(fd: number, flag: number): void;
    function dgramSetRecvBufferSize(fd: number, size: number): void;
    function dgramSetSendBufferSize(fd: number, size: number): void;
    function dgramGetRecvBufferSize(fd: number): number;
    function dgramGetSendBufferSize(fd: number): number;
    function dgramSetTTL(fd: number, ttl: number): void;
    function dgramSetMulticastInterface(fd: number, iface: string): void;
    function dgramAddMembership(fd: number, mcast: string, iface: string): void;
    function dgramDropMembership(fd: number, mcast: string, iface: string): void;
    function dgramAddSourceSpecificMembership(fd: number, src: string, group: string, iface: string): void;
    function dgramDropSourceSpecificMembership(fd: number, src: string, group: string, iface: string): void;
    function dgramConnect(fd: number, address: string, port: number): void;
    function dgramDisconnect(fd: number): void;
    function dgramClose(fd: number): void;
}

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
    _fd: number = -1;
    _localPort: number = 0;
    _localAddress: string = "0.0.0.0";

    constructor(typeOrOptions: string | SocketOptions = "udp4", listener?: Function) {
        if (typeof typeOrOptions === "string") {
            this.type = typeOrOptions;
        } else if (typeOrOptions && typeof typeOrOptions === "object") {
            if (typeOrOptions.type) this.type = typeOrOptions.type;
            if (typeOrOptions.recvBufferSize) this._recvBufferSize = typeOrOptions.recvBufferSize;
            if (typeOrOptions.sendBufferSize) this._sendBufferSize = typeOrOptions.sendBufferSize;
        }
        try {
            this._fd = __scriptgo.dgramSocketCreate(this.type === "udp6" ? 6 : 4);
        } catch {
            this._fd = -1;
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
        let port = 0;
        let address = "0.0.0.0";
        let callback: Function | null = null;

        if (typeof portOrOptions === "number") {
            port = portOrOptions;
            if (typeof addressOrCb === "string") address = addressOrCb;
            else if (typeof addressOrCb === "function") callback = addressOrCb;
            if (typeof cb === "function") callback = cb;
        } else if (typeof portOrOptions === "object" && portOrOptions !== null) {
            const opt = portOrOptions as { port?: number; address?: string };
            if (opt.port !== undefined) port = opt.port;
            if (opt.address !== undefined) address = opt.address;
            if (typeof addressOrCb === "function") callback = addressOrCb;
        } else if (typeof portOrOptions === "function") {
            callback = portOrOptions;
        }

        this._localPort = port;
        this._localAddress = address;

        if (this._fd >= 0) {
            try {
                __scriptgo.dgramBind(this._fd, address, port);
            } catch (err) {
                this.emit("error", err);
            }
        }
        if (callback) {
            this.once("listening", callback);
        }
        this.emit("listening");
        return this;
    }

    send(msg: unknown, offsetOrPort?: unknown, lengthOrAddress?: unknown, portOrCb?: unknown, addressOrCb?: unknown, cb?: Function): void {
        let port = 0;
        let address = "127.0.0.1";
        let callback: Function | null = null;

        if (typeof portOrCb === "number") {
            port = portOrCb;
            if (typeof addressOrCb === "string") address = addressOrCb;
            if (typeof cb === "function") callback = cb;
        } else if (typeof offsetOrPort === "number") {
            port = offsetOrPort;
            if (typeof lengthOrAddress === "string") address = lengthOrAddress;
            if (typeof portOrCb === "function") callback = portOrCb as Function;
        }

        const strData = typeof msg === "string" ? msg : (msg ? String(msg) : "");
        let sentBytes = 0;
        if (this._fd >= 0) {
            try {
                sentBytes = __scriptgo.dgramSend(this._fd, strData, strData.length, port, address);
            } catch (err) {
                if (callback) callback(err, 0);
                return;
            }
        }
        if (callback) {
            callback(null, sentBytes);
        }
    }

    close(callback?: Function): Socket {
        this._closed = true;
        this._bound = false;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramClose(this._fd);
            } catch {}
            this._fd = -1;
        }
        if (callback) {
            this.once("close", callback);
        }
        this.emit("close");
        return this;
    }

    address(): { address: string; family: string; port: number } {
        return {
            address: this._localAddress.length > 0 && this._localAddress !== "0.0.0.0" ? this._localAddress : "127.0.0.1",
            family: this.type === "udp6" ? "IPv6" : "IPv4",
            port: this._localPort > 0 ? this._localPort : 41234,
        };
    }

    connect(port: number, addressOrCb?: unknown, cb?: Function): void {
        this._connected = true;
        this._remotePort = port;
        let addr = "127.0.0.1";
        if (typeof addressOrCb === "string") {
            this._remoteAddress = addressOrCb;
            addr = addressOrCb;
        } else if (typeof addressOrCb === "function") {
            this.once("connect", addressOrCb);
        }
        if (typeof cb === "function") {
            this.once("connect", cb);
        }
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramConnect(this._fd, addr, port);
            } catch {}
        }
        this.emit("connect");
    }

    disconnect(): void {
        this._connected = false;
        this._remotePort = 0;
        this._remoteAddress = "";
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramDisconnect(this._fd);
            } catch {}
        }
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
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetBroadcast(this._fd, flag ? 1 : 0);
            } catch {}
        }
    }

    setTTL(ttl: number): number {
        this._ttl = ttl;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetTTL(this._fd, ttl);
            } catch {}
        }
        return this._ttl;
    }

    setMulticastTTL(ttl: number): number {
        this._multicastTTL = ttl;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetMulticastTTL(this._fd, ttl);
            } catch {}
        }
        return this._multicastTTL;
    }

    setMulticastLoopback(flag: boolean = true): boolean {
        this._multicastLoopback = flag;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetMulticastLoopback(this._fd, flag ? 1 : 0);
            } catch {}
        }
        return this._multicastLoopback;
    }

    setMulticastInterface(multicastInterface: string): void {
        this._multicastInterface = multicastInterface;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetMulticastInterface(this._fd, multicastInterface);
            } catch {}
        }
    }

    addMembership(multicastAddress: string, multicastInterface?: string): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramAddMembership(this._fd, multicastAddress, multicastInterface || "");
            } catch {}
        }
    }

    dropMembership(multicastAddress: string, multicastInterface?: string): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramDropMembership(this._fd, multicastAddress, multicastInterface || "");
            } catch {}
        }
    }

    addSourceSpecificMembership(sourceAddress: string, groupAddress: string, multicastInterface?: string): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramAddSourceSpecificMembership(this._fd, sourceAddress, groupAddress, multicastInterface || "");
            } catch {}
        }
    }

    dropSourceSpecificMembership(sourceAddress: string, groupAddress: string, multicastInterface?: string): void {
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramDropSourceSpecificMembership(this._fd, sourceAddress, groupAddress, multicastInterface || "");
            } catch {}
        }
    }

    setRecvBufferSize(size: number): void {
        this._recvBufferSize = size;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetRecvBufferSize(this._fd, size);
            } catch {}
        }
    }

    setSendBufferSize(size: number): void {
        this._sendBufferSize = size;
        if (this._fd >= 0) {
            try {
                __scriptgo.dgramSetSendBufferSize(this._fd, size);
            } catch {}
        }
    }

    getRecvBufferSize(): number {
        if (this._fd >= 0) {
            try {
                return __scriptgo.dgramGetRecvBufferSize(this._fd);
            } catch {}
        }
        return this._recvBufferSize;
    }

    getSendBufferSize(): number {
        if (this._fd >= 0) {
            try {
                return __scriptgo.dgramGetSendBufferSize(this._fd);
            } catch {}
        }
        return this._sendBufferSize;
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
