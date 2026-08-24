// Type definitions for node:https

export interface AgentOptions {
    ca?: string | Buffer | Array<string | Buffer>;
    cert?: string | Buffer | Array<string | Buffer>;
    ciphers?: string;
    clientCertEngine?: string;
    crl?: string | Buffer | Array<string | Buffer>;
    dhparam?: string | Buffer;
    ecdhCurve?: string;
    honorCipherOrder?: boolean;
    key?: string | Buffer | Array<any>;
    passphrase?: string;
    pfx?: string | Buffer | Array<string | Buffer | any>;
    rejectUnauthorized?: boolean;
    secureOptions?: number;
    secureProtocol?: string;
    servername?: string;
    sessionIdContext?: string;
    maxCachedSessions?: number;
    maxSockets?: number;
    maxFreeSockets?: number;
    maxTotalSockets?: number;
    keepAlive?: boolean;
    keepAliveMsecs?: number;
}

export declare class Agent {
    maxSockets: number;
    maxFreeSockets: number;
    maxTotalSockets: number;
    maxCachedSessions: number;
    options: AgentOptions;
    constructor(options?: AgentOptions);
    destroy(): void;
}

export declare const globalAgent: Agent;

export interface ServerOptions {
    ca?: string | Buffer | Array<string | Buffer>;
    cert?: string | Buffer | Array<string | Buffer>;
    key?: string | Buffer | Array<any>;
    passphrase?: string;
    rejectUnauthorized?: boolean;
    headersTimeout?: number;
    keepAliveTimeout?: number;
    maxHeadersCount?: number;
    requestTimeout?: number;
    timeout?: number;
}

export declare class ClientRequest {
    readonly url: string;
    readonly method: string;
    readonly finished: boolean;

    constructor(urlOrOptions: unknown, cb?: (...args: any[]) => void);

    setHeader(name: string, value: string): void;
    getHeader(name: string): string | null;
    write(chunk: any, cb?: () => void): boolean;
    write(chunk: any, encoding?: string, cb?: () => void): boolean;
    end(cb?: () => void): this;
    end(data: any, cb?: () => void): this;
    end(data: any, encoding?: string, cb?: () => void): this;
    destroy(error?: Error): this;
    setTimeout(timeout: number, callback?: () => void): this;

    on(event: string, listener: (...args: any[]) => void): this;
    once(event: string, listener: (...args: any[]) => void): this;
    emit(event: string, ...args: any[]): boolean;
}

export declare class Server {
    headersTimeout: number;
    keepAliveTimeout: number;
    maxHeadersCount: number;
    requestTimeout: number;
    timeout: number;
    readonly listening: boolean;

    constructor(requestListener?: (...args: any[]) => void);
    constructor(options: ServerOptions, requestListener?: (...args: any[]) => void);

    listen(port?: number, hostname?: string, backlog?: number, listeningListener?: () => void): this;
    listen(port?: number, hostname?: string, listeningListener?: () => void): this;
    listen(port?: number, backlog?: number, listeningListener?: () => void): this;
    listen(port?: number, listeningListener?: () => void): this;
    listen(options: unknown, listeningListener?: () => void): this;

    close(callback?: (err?: Error) => void): this;
    closeAllConnections(): void;
    closeIdleConnections(): void;
    setTimeout(msecs?: number, callback?: () => void): this;

    on(event: string, listener: (...args: any[]) => void): this;
    once(event: string, listener: (...args: any[]) => void): this;
    emit(event: string, ...args: any[]): boolean;

    asyncDispose(): Promise<void>;
}

export declare function createServer(requestListener?: (...args: any[]) => void): Server;
export declare function createServer(options: ServerOptions, requestListener?: (...args: any[]) => void): Server;

export declare function request(options: unknown, callback?: (...args: any[]) => void): ClientRequest;
export declare function request(url: string | URL, options?: unknown, callback?: (...args: any[]) => void): ClientRequest;

export declare function get(options: unknown, callback?: (...args: any[]) => void): ClientRequest;
export declare function get(url: string | URL, options?: unknown, callback?: (...args: any[]) => void): ClientRequest;
