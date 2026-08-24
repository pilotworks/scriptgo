export class Headers {
    _keys: string[];
    _values: string[];
    constructor(init?: Headers | null);
    append(name: string, value: string): void;
    delete(name: string): void;
    get(name: string): string | null;
    has(name: string): boolean;
    set(name: string, value: string): void;
    forEach(callback?: (value: string, name: string, parent: Headers)): void;
    entries(): string[][];
    keys(): string[];
    values(): string[];
}
export interface RequestInit {
    method?: string;
    headers?: unknown;
    body?: string | null;
}
export class Request {
    url: string;
    method: string;
    headers: Headers;
    body: string;
    constructor(input: unknown, init?: RequestInit);
}
export interface ResponseInit {
    status?: number;
    statusText?: string;
    headers?: unknown;
}
export class Response {
    ok: boolean;
    status: number;
    statusText: string;
    headers: Headers;
    url: string;
    _body: string;
    constructor(body?: string, init?: ResponseInit);
    async text(): Promise<string>;
    async json<T = unknown>(): Promise<T>;
    async arrayBuffer(): Promise<ArrayBuffer>;
    static json(data: string, init?: ResponseInit): Response;
    static error(): Response;
    static redirect(url: string, status?: number): Response;
}
export class FetchResponseData {
    status: number;
    statusText: string;
    headers: string[];
    body: string;
    constructor(status: number, statusText: string, headers: string[], body: string);
}
export async function fetch(input: unknown, init?: RequestInit): Promise<Response>;
export const METHODS: string[];
export const STATUS_CODES: Record<string, string>;
export function getStatusText(code: number): string;
export const maxHeaderSize: number;
export function validateHeaderName(name: string, label?: string): void;
export function validateHeaderValue(name: string, value: unknown): void;
export function setMaxIdleHTTPParsers(max: number): void;
export interface AgentOptions {
    maxFreeSockets?: number;
    maxSockets?: number;
    maxTotalSockets?: number;
    maxRequestsPerSocket?: number;
}
export interface AgentNameOptions {
    host?: string;
    port?: number | string;
    localAddress?: string;
}
export class HttpConnection {
    connected: boolean;
    options: string;
    constructor(connected?: boolean, options?: string);
    on(event: string, listener?: unknown): this;
}
export class Agent {
    options: AgentOptions | null;
    freeSockets: string[];
    sockets: string[];
    requests: string[];
    maxFreeSockets: number;
    maxSockets: number;
    maxTotalSockets: number;
    maxRequestsPerSocket: number;
    constructor(options?: AgentOptions);
    createConnection(options?: unknown, callback?: unknown): HttpConnection;
    destroy(): void;
    getName(options?: AgentNameOptions): string;
    keepSocketAlive(socket: unknown): boolean;
    reuseSocket(socket: unknown, request: unknown): void;
}
export const globalAgent: Agent;
export class OutgoingMessage {
    connection: string | null;
    socket: string | null;
    headersSent: boolean;
    writableCorked: number;
    writableEnded: boolean;
    writableFinished: boolean;
    writableHighWaterMark: number;
    writableLength: number;
    writableObjectMode: boolean;
    strictContentLength: boolean;
    sendDate: boolean;
    req: IncomingMessage | null;
    on(event: string, listener: Function): OutgoingMessage;
    once(event: string, listener: Function): OutgoingMessage;
    emit(event: string, arg1?: unknown, arg2?: unknown): boolean;
    setHeader(name: string, value: unknown): OutgoingMessage;
    getHeader(name: string): string | null;
    getHeaders(): Headers;
    getHeaderNames(): string[];
    getRawHeaderNames(): string[];
    hasHeader(name: string): boolean;
    removeHeader(name: string): void;
    appendHeader(name: string, value: unknown): OutgoingMessage;
    setHeaders(headers: any): OutgoingMessage;
    flushHeaders(): void;
    write(chunk: unknown, encoding?: string, callback?: unknown): boolean;
    end(chunk?: unknown, encoding?: string, callback?: unknown): OutgoingMessage;
    destroy(error?: unknown): OutgoingMessage;
    cork(): void;
    uncork(): void;
    setTimeout(msecs?: number, callback?: unknown): OutgoingMessage;
    addTrailers(headers?: unknown): void;
    pipe(dest: unknown): unknown;
}
export interface ClientRequestOptions {
    path?: string;
    pathname?: string;
    method?: string;
    host?: string;
    hostname?: string;
    protocol?: string;
    headers?: Headers;
}
export class ClientRequest extends OutgoingMessage {
    path: string;
    method: string;
    host: string;
    protocol: string;
    reusedSocket: boolean;
    maxHeadersCount: number;
    constructor(urlOrOptions: unknown, cb?: unknown);
    abort(): void;
    setNoDelay(noDelay?: boolean): void;
    setSocketKeepAlive(enable?: boolean, initialDelay?: number): void;
}
export class IncomingMessage {
    aborted: boolean;
    complete: boolean;
    connection: string | null;
    socket: string | null;
    headers: string[];
    headersDistinct: string[];
    httpVersion: string;
    method: string;
    rawHeaders: string[];
    rawTrailers: string[];
    statusCode: number;
    statusMessage: string;
    trailers: string[];
    trailersDistinct: string[];
    url: string;
    constructor(socket?: any);
    on(event: string, listener: Function): IncomingMessage;
    once(event: string, listener: Function): IncomingMessage;
    emit(event: string, arg1?: unknown, arg2?: unknown): boolean;
    destroy(error?: unknown): IncomingMessage;
    setTimeout(msecs?: number, callback?: unknown): IncomingMessage;
    pipe<T = any>(destination: T): T;
}
export class ServerResponse extends OutgoingMessage {
    statusCode: number;
    statusMessage: string;
    constructor(req?: IncomingMessage | null);
    writeHead(statusCode: number, statusMessageOrHeaders?: unknown, headers?: Headers | null): ServerResponse;
    writeContinue(): void;
    writeEarlyHints(hints?: unknown, callback?: unknown): void;
    writeProcessing(): void;
}
export class Server {
    listening: boolean;
    maxHeadersCount: number;
    maxRequestsPerSocket: number;
    timeout: number;
    headersTimeout: number;
    keepAliveTimeout: number;
    keepAliveTimeoutBuffer: number;
    requestTimeout: number;
    constructor(optionsOrListener?: unknown, listener?: unknown);
    on(event: string, listener: Function): Server;
    once(event: string, listener: Function): Server;
    emit(event: string, arg1?: unknown, arg2?: unknown): boolean;
    listen(port?: unknown, hostOrCb?: unknown, cb?: unknown): Server;
    close(callback?: unknown): Server;
    closeAllConnections(): void;
    closeIdleConnections(): void;
    setTimeout(msecs?: number, callback?: unknown): Server;
    async asyncDispose(): Promise<void>;
}
export function createServer(optionsOrListener?: unknown, listener?: unknown): Server;
export function request(urlOrOptions: unknown, optionsOrCb?: unknown, cb?: unknown): ClientRequest;
export function get(urlOrOptions: unknown, optionsOrCb?: unknown, cb?: unknown): ClientRequest;
export class WebSocket {
    url: string;
    readyState: number;
    constructor(url: string);
    close(): void;
    send(data: unknown): void;
}
