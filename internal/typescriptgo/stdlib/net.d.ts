// Type definitions for node:net

export interface AddressInfo {
    address: string;
    family: string;
    port: number;
}

export interface SocketConstructorOpts {
    fd?: number;
    allowHalfOpen?: boolean;
    readable?: boolean;
    writable?: boolean;
    signal?: unknown;
}

export interface SocketConnectOpts {
    port: number;
    host?: string;
    localAddress?: string;
    localPort?: number;
    family?: number;
    hints?: number;
    lookup?: Function;
    noDelay?: boolean;
    keepAlive?: boolean;
    keepAliveInitialDelay?: number;
    autoSelectFamily?: boolean;
    autoSelectFamilyAttemptTimeout?: number;
}

export declare function isIP(input: string): number;
export declare function isIPv4(input: string): boolean;
export declare function isIPv6(input: string): boolean;

export declare class SocketAddress {
    readonly address: string;
    readonly family: string;
    readonly port: number;
    readonly flowlabel: number;
    constructor(options?: { address?: string; family?: string; port?: number; flowlabel?: number });
    static parse(input: string): SocketAddress;
}

export declare class BlockList {
    readonly rules: string[];
    addAddress(address: string, type?: string): void;
    addRange(start: string, end: string, type?: string): void;
    addSubnet(net: string, prefix: number, type?: string): void;
    check(address: string, type?: string): boolean;
    fromJSON(value: unknown): void;
    toJSON(): string[];
    static isBlockList(value: unknown): boolean;
}

export declare class Socket {
    readonly bufferSize: number;
    readonly bytesRead: number;
    readonly bytesWritten: number;
    readonly connecting: boolean;
    readonly destroyed: boolean;
    readonly localAddress?: string;
    readonly localPort?: number;
    readonly localFamily?: string;
    readonly pending: boolean;
    readonly readyState: string;
    readonly remoteAddress?: string;
    readonly remoteFamily?: string;
    readonly remotePort?: number;
    readonly timeout?: number;
    readonly autoSelectFamilyAttemptedAddresses: string[];

    constructor(options?: SocketConstructorOpts);

    connect(options: SocketConnectOpts, connectionListener?: () => void): this;
    connect(port: number, host: string, connectionListener?: () => void): this;
    connect(port: number, connectionListener?: () => void): this;

    setEncoding(encoding?: string): this;
    pause(): this;
    resume(): this;
    setTimeout(timeout: number, callback?: () => void): this;
    setNoDelay(noDelay?: boolean): this;
    setKeepAlive(enable?: boolean, initialDelay?: number): this;
    address(): AddressInfo | {};
    unref(): this;
    ref(): this;
    destroy(error?: Error): this;
    destroySoon(): void;
    resetAndDestroy(): this;

    write(buffer: Uint8Array | string, cb?: (err?: Error) => void): boolean;
    write(str: Uint8Array | string, encoding?: string, cb?: (err?: Error) => void): boolean;
    end(cb?: () => void): this;
    end(data: string | Uint8Array, cb?: () => void): this;
    end(str: string | Uint8Array, encoding?: string, cb?: () => void): this;

    on(event: string, listener: (...args: any[]) => void): this;
    once(event: string, listener: (...args: any[]) => void): this;
    emit(event: string, ...args: any[]): boolean;
}

export interface ServerOpts {
    allowHalfOpen?: boolean;
    pauseOnConnect?: boolean;
    noDelay?: boolean;
    keepAlive?: boolean;
    keepAliveInitialDelay?: number;
    highWaterMark?: number;
}

export declare class Server {
    readonly maxConnections: number;
    readonly connections: number;
    readonly listening: boolean;
    dropMaxConnection: boolean;

    constructor(connectionListener?: (socket: Socket) => void);
    constructor(options?: ServerOpts, connectionListener?: (socket: Socket) => void);

    listen(port?: number, hostname?: string, backlog?: number, listeningListener?: () => void): this;
    listen(port?: number, hostname?: string, listeningListener?: () => void): this;
    listen(port?: number, backlog?: number, listeningListener?: () => void): this;
    listen(port?: number, listeningListener?: () => void): this;
    listen(options: unknown, listeningListener?: () => void): this;

    close(callback?: (err?: Error) => void): this;
    address(): AddressInfo | string | null;
    getConnections(cb: (error: Error | null, count: number) => void): void;
    ref(): this;
    unref(): this;

    on(event: string, listener: (...args: any[]) => void): this;
    once(event: string, listener: (...args: any[]) => void): this;
    emit(event: string, ...args: any[]): boolean;

    asyncDispose(): Promise<void>;
}

export declare function createServer(connectionListener?: (socket: Socket) => void): Server;
export declare function createServer(options?: ServerOpts, connectionListener?: (socket: Socket) => void): Server;

export declare function connect(options: SocketConnectOpts, connectionListener?: () => void): Socket;
export declare function connect(port: number, host?: string, connectionListener?: () => void): Socket;

export declare function createConnection(options: SocketConnectOpts, connectionListener?: () => void): Socket;
export declare function createConnection(port: number, host?: string, connectionListener?: () => void): Socket;

export declare function getDefaultAutoSelectFamily(): boolean;
export declare function setDefaultAutoSelectFamily(value: boolean): void;
export declare function getDefaultAutoSelectFamilyAttemptTimeout(): number;
export declare function setDefaultAutoSelectFamilyAttemptTimeout(value: number): void;
