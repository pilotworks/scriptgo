// ScriptGo Standard Library: node:net

import { EventEmitter } from "node:events";

declare namespace __scriptgo {
    function netSocketCreate(family?: number, sockType?: number): number;
    function netSocketConnect(fd: number, host: string, port: number): void;
    function netSocketWrite(fd: number, data: string, len: number): number;
    function netSocketRead(fd: number, maxLen: number): string;
    function netSocketClose(fd: number): void;
    function netSocketSetNoDelay(fd: number, noDelay: number): void;
    function netSocketSetKeepAlive(fd: number, enable: number, initialDelay: number): void;
    function netServerListen(host: string, port: number, backlog: number): number;
    function netServerAccept(serverFd: number): { fd: number; ip: string; port: number };
}

let _defaultAutoSelectFamily: boolean = true;
let _defaultAutoSelectFamilyAttemptTimeout: number = 250;

export function getDefaultAutoSelectFamily(): boolean {
    return _defaultAutoSelectFamily;
}

export function setDefaultAutoSelectFamily(value: boolean): void {
    _defaultAutoSelectFamily = value ? true : false;
}

export function getDefaultAutoSelectFamilyAttemptTimeout(): number {
    return _defaultAutoSelectFamilyAttemptTimeout;
}

export function setDefaultAutoSelectFamilyAttemptTimeout(value: number): void {
    if (typeof value === "number" && value >= 1) {
        _defaultAutoSelectFamilyAttemptTimeout = value;
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

function parseIPv4ToNum(input: string): number {
    const parts = input.split(".");
    if (parts.length !== 4) return -1;
    let res = 0;
    for (let i = 0; i < 4; i++) {
        const seg = parts[i];
        if (seg.length === 0 || seg.length > 3) return -1;
        const val = parseInt(seg);
        if (isNaN(val) || val < 0 || val > 255) return -1;
        res = (res * 256) + val;
    }
    return res >>> 0;
}

function parseIPv6ToParts(input: string): number[] | null {
    if (!isIPv6(input)) return null;
    const doubleColon = input.indexOf("::");
    let leftParts: string[] = [];
    let rightParts: string[] = [];
    if (doubleColon !== -1) {
        const left = input.substring(0, doubleColon);
        const right = input.substring(doubleColon + 2);
        if (left.length > 0) leftParts = left.split(":");
        if (right.length > 0) rightParts = right.split(":");
    } else {
        leftParts = input.split(":");
    }
    const missing = 8 - (leftParts.length + rightParts.length);
    if (missing < 0) return null;
    const res: number[] = [];
    for (let i = 0; i < leftParts.length; i++) {
        res.push(parseInt(leftParts[i], 16));
    }
    for (let i = 0; i < missing; i++) {
        res.push(0);
    }
    for (let i = 0; i < rightParts.length; i++) {
        res.push(parseInt(rightParts[i], 16));
    }
    return res;
}

function compareIPv6Parts(a: number[], b: number[]): number {
    for (let i = 0; i < 8; i++) {
        if (a[i] < b[i]) return -1;
        if (a[i] > b[i]) return 1;
    }
    return 0;
}

function computeIPv6Subnet(parts: number[], prefix: number): [number[], number[]] {
    const start: number[] = [];
    const end: number[] = [];
    let remPrefix = prefix;
    for (let i = 0; i < 8; i++) {
        if (remPrefix >= 16) {
            start.push(parts[i]);
            end.push(parts[i]);
            remPrefix -= 16;
        } else if (remPrefix > 0) {
            const mask = ((0xFFFF << (16 - remPrefix)) & 0xFFFF);
            start.push(parts[i] & mask);
            end.push(parts[i] | (~mask & 0xFFFF));
            remPrefix = 0;
        } else {
            start.push(0);
            end.push(0xFFFF);
        }
    }
    return [start, end];
}

export interface SocketAddressOptions {
    address?: string;
    family?: "ipv4" | "ipv6";
    port?: number;
    flowlabel?: number;
}

export interface SocketOptions {
    fd?: number;
    allowHalfOpen?: boolean;
    readable?: boolean;
    writable?: boolean;
}

export interface SocketConnectOptions {
    port?: number;
    host?: string;
    localAddress?: string;
    localPort?: number;
    family?: number;
    path?: string;
}

export interface ServerOptions {
    allowHalfOpen?: boolean;
    pauseOnConnect?: boolean;
    noDelay?: boolean;
    keepAlive?: boolean;
    keepAliveInitialDelay?: number;
}

export interface ListenOptions {
    port?: number;
    host?: string;
    backlog?: number;
    path?: string;
}

const defaultSocketAddressOptions: SocketAddressOptions = {
    address: "127.0.0.1",
    family: "ipv4",
    port: 0,
    flowlabel: 0
};

export class SocketAddress {
    address: string = "127.0.0.1";
    family: "ipv4" | "ipv6" = "ipv4";
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
    private _rules: string[] = [];
    private _ipv4Ranges: [number, number][] = [];
    private _ipv6Ranges: [number[], number[]][] = [];

    constructor() {}

    static isBlockList(value: unknown): boolean {
        return value !== null && typeof value === "object" && (value instanceof BlockList || (value as any)._isBlockList === true);
    }

    readonly _isBlockList: boolean = true;

    addAddress(address: string | SocketAddress, type: "ipv4" | "ipv6" = "ipv4"): void {
        let addrStr = typeof address === "string" ? address : address.address;
        let fam = typeof address === "string" ? type : address.family;
        if (fam === "ipv4") {
            const num = parseIPv4ToNum(addrStr);
            if (num < 0) throw new TypeError("Invalid IPv4 address: " + addrStr);
            this._ipv4Ranges.push([num, num]);
            this._rules.push("Address: IPv4 " + addrStr);
        } else {
            const parts = parseIPv6ToParts(addrStr);
            if (!parts) throw new TypeError("Invalid IPv6 address: " + addrStr);
            this._ipv6Ranges.push([parts, parts]);
            this._rules.push("Address: IPv6 " + addrStr);
        }
    }

    addRange(start: string | SocketAddress, end: string | SocketAddress, type: "ipv4" | "ipv6" = "ipv4"): void {
        let startStr = typeof start === "string" ? start : start.address;
        let endStr = typeof end === "string" ? end : end.address;
        let fam = typeof start === "string" ? type : start.family;
        if (fam === "ipv4") {
            const sNum = parseIPv4ToNum(startStr);
            const eNum = parseIPv4ToNum(endStr);
            if (sNum < 0 || eNum < 0 || sNum > eNum) throw new RangeError("Invalid IPv4 range");
            this._ipv4Ranges.push([sNum, eNum]);
            this._rules.push("Range: IPv4 " + startStr + "-" + endStr);
        } else {
            const sParts = parseIPv6ToParts(startStr);
            const eParts = parseIPv6ToParts(endStr);
            if (!sParts || !eParts || compareIPv6Parts(sParts, eParts) > 0) throw new RangeError("Invalid IPv6 range");
            this._ipv6Ranges.push([sParts, eParts]);
            this._rules.push("Range: IPv6 " + startStr + "-" + endStr);
        }
    }

    addSubnet(net: string | SocketAddress, prefix: number, type: "ipv4" | "ipv6" = "ipv4"): void {
        let netStr = typeof net === "string" ? net : net.address;
        let fam = typeof net === "string" ? type : net.family;
        if (fam === "ipv4") {
            if (prefix < 0 || prefix > 32) throw new RangeError("IPv4 prefix must be between 0 and 32");
            const netNum = parseIPv4ToNum(netStr);
            if (netNum < 0) throw new TypeError("Invalid IPv4 address: " + netStr);
            const mask = prefix === 0 ? 0 : ((0xFFFFFFFF << (32 - prefix)) >>> 0);
            const start = (netNum & mask) >>> 0;
            const end = (start | (~mask >>> 0)) >>> 0;
            this._ipv4Ranges.push([start, end]);
            this._rules.push("Subnet: IPv4 " + netStr + "/" + prefix);
        } else {
            if (prefix < 0 || prefix > 128) throw new RangeError("IPv6 prefix must be between 0 and 128");
            const netParts = parseIPv6ToParts(netStr);
            if (!netParts) throw new TypeError("Invalid IPv6 address: " + netStr);
            const [startParts, endParts] = computeIPv6Subnet(netParts, prefix);
            this._ipv6Ranges.push([startParts, endParts]);
            this._rules.push("Subnet: IPv6 " + netStr + "/" + prefix);
        }
    }

    check(address: string | SocketAddress, type: "ipv4" | "ipv6" = "ipv4"): boolean {
        let addrStr = typeof address === "string" ? address : address.address;
        let fam = typeof address === "string" ? (isIPv6(addrStr) ? "ipv6" : type) : address.family;
        if (fam === "ipv4") {
            const num = parseIPv4ToNum(addrStr);
            if (num < 0) return false;
            for (let i = 0; i < this._ipv4Ranges.length; i++) {
                const [start, end] = this._ipv4Ranges[i];
                if (num >= start && num <= end) return true;
            }
            return false;
        } else {
            const parts = parseIPv6ToParts(addrStr);
            if (!parts) return false;
            for (let i = 0; i < this._ipv6Ranges.length; i++) {
                const [start, end] = this._ipv6Ranges[i];
                if (compareIPv6Parts(parts, start) >= 0 && compareIPv6Parts(parts, end) <= 0) return true;
            }
            return false;
        }
    }

    fromJSON(rules: string[]): void {
        if (!Array.isArray(rules)) return;
        for (let i = 0; i < rules.length; i++) {
            const r = rules[i];
            if (typeof r !== "string") continue;
            if (r.startsWith("Address: IPv4 ")) {
                this.addAddress(r.substring(14), "ipv4");
            } else if (r.startsWith("Address: IPv6 ")) {
                this.addAddress(r.substring(14), "ipv6");
            } else if (r.startsWith("Range: IPv4 ")) {
                const dash = r.indexOf("-", 12);
                if (dash !== -1) {
                    this.addRange(r.substring(12, dash), r.substring(dash + 1), "ipv4");
                }
            } else if (r.startsWith("Range: IPv6 ")) {
                const dash = r.indexOf("-", 12);
                if (dash !== -1) {
                    this.addRange(r.substring(12, dash), r.substring(dash + 1), "ipv6");
                }
            } else if (r.startsWith("Subnet: IPv4 ")) {
                const slash = r.indexOf("/", 13);
                if (slash !== -1) {
                    this.addSubnet(r.substring(13, slash), parseInt(r.substring(slash + 1)), "ipv4");
                }
            } else if (r.startsWith("Subnet: IPv6 ")) {
                const slash = r.indexOf("/", 13);
                if (slash !== -1) {
                    this.addSubnet(r.substring(13, slash), parseInt(r.substring(slash + 1)), "ipv6");
                }
            }
        }
    }

    toJSON(): string[] {
        return this.rules;
    }

    get rules(): string[] {
        const copy: string[] = [];
        for (let i = 0; i < this._rules.length; i++) {
            copy.push(this._rules[i]);
        }
        return copy;
    }
}

export class Socket extends EventEmitter {
    connecting: boolean = false;
    destroyed: boolean = false;
    pending: boolean = true;
    readyState: string = "closed";
    bytesRead: number = 0;
    bytesWritten: number = 0;
    bufferSize: number | undefined = undefined;
    localAddress: string | undefined = undefined;
    localPort: number | undefined = undefined;
    localFamily: string | undefined = undefined;
    remoteAddress: string | undefined = undefined;
    remotePort: number | undefined = undefined;
    remoteFamily: string | undefined = undefined;
    timeout: number = 0;
    autoSelectFamilyAttemptedAddresses: string[] = [];
    _fd: number = -1;
    _noDelay: boolean = false;
    _keepAlive: boolean = false;
    _keepAliveInitialDelay: number = 0;

    constructor(options: SocketOptions | null = null) {
        super();
        this.connecting = false;
        this.destroyed = false;
        this.pending = true;
        this.readyState = "open";
        this._fd = -1;
    }

    pause(): this {
        return this;
    }

    resume(): this {
        return this;
    }

    ref(): this {
        return this;
    }

    unref(): this {
        return this;
    }

    setEncoding(encoding?: string): this {
        return this;
    }

    destroySoon(): void {
        this.end();
    }

    resetAndDestroy(): this {
        if (this.destroyed || this._fd < 0) {
            const err = new Error("Socket is closed");
            this.emit("error", err);
            return this;
        }
        this.destroy();
        return this;
    }

    connect(optionsOrPort: number | string | SocketConnectOptions, hostOrListener: string | (() => void) | null = null, listener: (() => void) | null = null): Socket {
        this.connecting = false;
        this.pending = false;
        this.readyState = "open";
        let isIPC = false;
        let rPort: number = 0;
        let rAddr: string = "127.0.0.1";
        if (typeof optionsOrPort === "number") {
            rPort = optionsOrPort;
            rAddr = typeof hostOrListener === "string" ? hostOrListener : "127.0.0.1";
            this.remotePort = rPort;
            this.remoteAddress = rAddr;
            this.remoteFamily = "IPv4";
            this.localAddress = "127.0.0.1";
            this.localPort = 0;
            this.localFamily = "IPv4";
        } else if (typeof optionsOrPort === "string") {
            rAddr = optionsOrPort;
            rPort = 0;
            this.remoteAddress = rAddr;
            this.remotePort = rPort;
            this.remoteFamily = "IPC";
            this.localAddress = rAddr;
            this.localPort = 0;
            this.localFamily = "IPC";
            isIPC = true;
        } else {
            if (optionsOrPort.port !== undefined) rPort = optionsOrPort.port;
            if (optionsOrPort.host !== undefined) rAddr = optionsOrPort.host;
            this.remotePort = rPort;
            this.remoteAddress = rAddr;
            this.remoteFamily = "IPv4";
            this.localAddress = "127.0.0.1";
            this.localPort = 0;
            this.localFamily = "IPv4";
            if (optionsOrPort.path !== undefined) {
                rAddr = optionsOrPort.path;
                rPort = 0;
                this.remoteAddress = rAddr;
                this.remotePort = rPort;
                this.remoteFamily = "IPC";
                this.localAddress = rAddr;
                this.localPort = 0;
                this.localFamily = "IPC";
                isIPC = true;
            }
        }
        if (listener !== null) {
            this.once("connect", listener);
        } else if (typeof hostOrListener === "function") {
            this.once("connect", hostOrListener);
        }
        try {
            if (this._fd < 0) {
                this._fd = __scriptgo.netSocketCreate(isIPC ? 1 : 4, 1);
            }
            if (this._noDelay) {
                try { __scriptgo.netSocketSetNoDelay(this._fd, 1); } catch {}
            }
            if (this._keepAlive) {
                try { __scriptgo.netSocketSetKeepAlive(this._fd, 1, this._keepAliveInitialDelay); } catch {}
            }
            __scriptgo.netSocketConnect(this._fd, rAddr, rPort);
            this.emit("connect");
        } catch (err) {
            this.emit("error", err);
            this.destroy();
        }
        return this;
    }

    write(data: string | Uint8Array, encodingOrCb: string | (() => void) | null = null, callback: (() => void) | null = null): boolean {
        if (this.destroyed) {
            return false;
        }
        let byteCount = 0;
        let strData = "";
        if (typeof data === "string") {
            strData = data;
            byteCount = data.length;
        } else if (data && typeof data === "object") {
            const arr = data as Uint8Array;
            byteCount = arr.byteLength !== undefined ? arr.byteLength : (arr.length !== undefined ? arr.length : 0);
            strData = String(data);
        }
        this.bytesWritten += byteCount;
        if (this._fd >= 0) {
            try {
                __scriptgo.netSocketWrite(this._fd, strData, byteCount);
            } catch (err) {
                this.emit("error", err);
            }
        }
        if (typeof encodingOrCb === "function") {
            encodingOrCb();
        } else if (callback !== null) {
            callback();
        }
        return true;
    }

    read(size: number = 65536): string {
        if (this._fd >= 0 && !this.destroyed) {
            try {
                const data = __scriptgo.netSocketRead(this._fd, size);
                this.bytesRead += data.length;
                if (data.length > 0) {
                    this.emit("data", data);
                }
                return data;
            } catch (err) {
                this.emit("error", err);
            }
        }
        return "";
    }

    end(dataOrCb: string | Uint8Array | (() => void) | null = null, encodingOrCb: string | (() => void) | null = null, callback: (() => void) | null = null): Socket {
        if (typeof dataOrCb === "function") {
            this.once("finish", dataOrCb);
        } else if (dataOrCb !== null) {
            this.write(dataOrCb);
        }
        if (typeof encodingOrCb === "function") {
            this.once("finish", encodingOrCb);
        }
        if (callback !== null) {
            this.once("finish", callback);
        }
        this.readyState = "readOnly";
        this.emit("finish");
        this.emit("end");
        this.destroy();
        return this;
    }

    destroy(error: Error | null = null): Socket {
        if (!this.destroyed) {
            if (this._fd >= 0) {
                try {
                    __scriptgo.netSocketClose(this._fd);
                } catch {}
                this._fd = -1;
            }
            this.destroyed = true;
            this.pending = false;
            this.readyState = "closed";
            if (error !== null && error !== undefined) {
                this.emit("error", error);
            }
            this.emit("close", error !== null && error !== undefined);
        }
        return this;
    }

    setTimeout(timeout: number, callback: (() => void) | null = null): Socket {
        this.timeout = timeout;
        if (callback !== null && callback !== undefined) {
            this.once("timeout", callback);
        }
        return this;
    }

    setKeepAlive(enable: boolean = false, initialDelay: number = 0): Socket {
        this._keepAlive = enable;
        this._keepAliveInitialDelay = initialDelay;
        if (this._fd >= 0) {
            try {
                __scriptgo.netSocketSetKeepAlive(this._fd, enable ? 1 : 0, initialDelay);
            } catch {}
        }
        return this;
    }

    setNoDelay(noDelay: boolean = true): Socket {
        this._noDelay = noDelay;
        if (this._fd >= 0) {
            try {
                __scriptgo.netSocketSetNoDelay(this._fd, noDelay ? 1 : 0);
            } catch {}
        }
        return this;
    }

    address(): { port?: number, family?: string, address?: string } {
        if (this.localPort === undefined) {
            return {};
        }
        return {
            port: this.localPort,
            family: this.localFamily,
            address: this.localAddress,
        };
    }

    [Symbol.asyncDispose](): Promise<void> {
        this.destroy();
        return Promise.resolve(undefined);
    }
}

export class Server extends EventEmitter {
    listening: boolean = false;
    maxConnections: number | undefined = undefined;
    maxHeadersCount: number = 2000;
    timeout: number = 0;
    keepAliveTimeout: number = 5000;
    dropMaxConnection: boolean | undefined = undefined;

    _connectionsCount: number = 0;
    _addressPort: number = 0;
    _addressPath: string = "";
    _serverFd: number = -1;

    constructor(optionsOrListener: ServerOptions | ((socket: Socket) => void) | null = null, listener: ((socket: Socket) => void) | null = null) {
        super();
        this._serverFd = -1;
        if (typeof optionsOrListener === "function") {
            this.on("connection", optionsOrListener);
        } else if (typeof listener === "function") {
            this.on("connection", listener);
        }
    }

    ref(): this {
        return this;
    }

    unref(): this {
        return this;
    }

    listen(portOrOptions: number | string | ListenOptions = 0, hostOrCb: string | (() => void) | null = null, callback: (() => void) | null = null): Server {
        this.listening = true;
        let port = 0;
        let host = "0.0.0.0";
        let backlog = 511;

        if (typeof portOrOptions === "number") {
            port = portOrOptions;
            this._addressPort = port;
            if (typeof hostOrCb === "string") host = hostOrCb;
        } else if (typeof portOrOptions === "string") {
            host = portOrOptions;
            port = 0;
            this._addressPath = portOrOptions;
        } else if (typeof portOrOptions === "object" && portOrOptions !== null) {
            if (portOrOptions.port !== undefined) {
                port = portOrOptions.port;
                this._addressPort = port;
            }
            if (portOrOptions.host !== undefined) host = portOrOptions.host;
            if (portOrOptions.backlog !== undefined) backlog = portOrOptions.backlog;
            if (portOrOptions.path !== undefined) {
                host = portOrOptions.path;
                port = 0;
                this._addressPath = host;
            }
        }
        if (typeof hostOrCb === "function") {
            this.once("listening", hostOrCb);
        } else if (callback !== null) {
            this.once("listening", callback);
        }
        try {
            this._serverFd = __scriptgo.netServerListen(host, port, backlog);
            queueMicrotask(() => this.emit("listening"));
        } catch (err) {
            queueMicrotask(() => this.emit("error", err));
        }
        return this;
    }

    _acceptConnection(): Socket | null {
        if (this._serverFd < 0) return null;
        try {
            const client = __scriptgo.netServerAccept(this._serverFd);
            if (client && client.fd >= 0) {
                const sock = new Socket();
                sock._fd = client.fd;
                sock.remoteAddress = client.ip;
                sock.remotePort = client.port;
                sock.readyState = "open";
                sock.pending = false;
                sock.connecting = false;
                this._connectionsCount++;
                sock.on("close", () => {
                    this._connectionsCount = Math.max(0, this._connectionsCount - 1);
                });
                this.emit("connection", sock);
                return sock;
            }
        } catch {}
        return null;
    }

    close(callback: Function | null = null): Server {
        this.listening = false;
        if (this._serverFd >= 0) {
            try {
                __scriptgo.netSocketClose(this._serverFd);
            } catch {}
            this._serverFd = -1;
        }
        if (callback !== null && callback !== undefined) {
            this.once("close", callback);
        }
        this.emit("close");
        return this;
    }

    address(): { port: number, family: string, address: string } | string {
        if (this._addressPath.length > 0) {
            return this._addressPath;
        }
        return {
            port: this._addressPort,
            family: "IPv4",
            address: "127.0.0.1",
        };
    }

    getConnections(callback: (err: Error | null, count: number) => void): void {
        callback(null, this._connectionsCount);
    }

    [Symbol.asyncDispose](): Promise<void> {
        this.close();
        return Promise.resolve(undefined);
    }
}

export function createServer(optionsOrListener: ServerOptions | ((socket: Socket) => void) | null = null, listener: ((socket: Socket) => void) | null = null): Server {
    return new Server(optionsOrListener, listener);
}

export function createConnection(optionsOrPort: number | string | SocketConnectOptions, hostOrListener: string | (() => void) | null = null, listener: (() => void) | null = null): Socket {
    const socket = new Socket();
    socket.connect(optionsOrPort, hostOrListener, listener);
    return socket;
}

export function connect(optionsOrPort: number | string | SocketConnectOptions, hostOrListener: string | (() => void) | null = null, listener: (() => void) | null = null): Socket {
    return createConnection(optionsOrPort, hostOrListener, listener);
}

export default {
    BlockList,
    SocketAddress,
    Socket,
    Server,
    isIP,
    isIPv4,
    isIPv6,
    createServer,
    createConnection,
    connect,
    getDefaultAutoSelectFamily,
    setDefaultAutoSelectFamily,
    getDefaultAutoSelectFamilyAttemptTimeout,
    setDefaultAutoSelectFamilyAttemptTimeout,
};
