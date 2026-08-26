class HttpListenerEntry {
    fn: Function;
    once: boolean;

    constructor(fn: Function, once: boolean) {
        this.fn = fn;
        this.once = once;
    }
}

class HttpEventBucket {
    name: string;
    listeners: HttpListenerEntry[];

    constructor(name: string, listeners: HttpListenerEntry[]) {
        this.name = name;
        this.listeners = listeners;
    }
}

export class Headers {
    _keys: string[] = [];
    _values: string[] = [];

    constructor(init: Headers | null = null) {
        this._keys = [];
        this._values = [];
        if (init !== null) {
            const entries = init.entries();
            for (let i = 0; i < entries.length; i++) {
                this.append(entries[i][0], entries[i][1]);
            }
        }
    }

    append(name: string, value: string): void {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === lower) {
                this._values[i] = this._values[i] + ", " + value;
                return;
            }
        }
        this._keys.push(lower);
        this._values.push(value);
    }

    delete(name: string): void {
        const lower = name.toLowerCase();
        const nextKeys: string[] = [];
        const nextValues: string[] = [];
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] !== lower) {
                nextKeys.push(this._keys[i]);
                nextValues.push(this._values[i]);
            }
        }
        this._keys = nextKeys;
        this._values = nextValues;
    }

    get(name: string): string | null {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === lower) {
                return this._values[i];
            }
        }
        return null;
    }

    has(name: string): boolean {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === lower) {
                return true;
            }
        }
        return false;
    }

    set(name: string, value: string): void {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._keys.length; i++) {
            if (this._keys[i] === lower) {
                this._values[i] = value;
                return;
            }
        }
        this._keys.push(lower);
        this._values.push(value);
    }

    forEach(callback: (value: string, name: string, parent: Headers) => void): void {
        for (let i = 0; i < this._keys.length; i++) {
            callback(this._values[i], this._keys[i], this);
        }
    }

    entries(): string[][] {
        const res: string[][] = [];
        for (let i = 0; i < this._keys.length; i++) {
            const pair: string[] = [this._keys[i], this._values[i]];
            res.push(pair);
        }
        return res;
    }

    keys(): string[] {
        return this._keys;
    }

    values(): string[] {
        return this._values;
    }
}

export interface RequestInit {
    method?: string;
    headers?: unknown;
    body?: string | null;
}

const defaultRequestInit: RequestInit = { method: "", headers: null, body: "" };

export class Request {
    url: string = "";
    method: string = "GET";
    headers: Headers = new Headers();
    body: string = "";

    constructor(input: unknown, init: RequestInit = defaultRequestInit) {
        if (typeof input === "string") {
            this.url = input;
            this.method = "GET";
            this.headers = new Headers();
            this.body = "";
        } else if (input instanceof Request) {
            this.url = input.url;
            this.method = input.method;
            this.headers = new Headers(input.headers);
            this.body = input.body;
        } else {
            this.url = "";
            this.method = "GET";
            this.headers = new Headers();
            this.body = "";
        }

        if (init.method !== undefined && init.method.length > 0) {
            this.method = init.method.toUpperCase();
        }
        if (init.headers instanceof Headers) {
            this.headers = init.headers as Headers;
        }
        if (init.body !== undefined && init.body !== null && init.body.length > 0) {
            this.body = init.body;
        }
    }
}

export interface ResponseInit {
    status?: number;
    statusText?: string;
    headers?: unknown;
}

const defaultResponseInit: ResponseInit = { status: 200, statusText: "OK", headers: null };

export class Response {
    ok: boolean = true;
    status: number = 200;
    statusText: string = "OK";
    headers: Headers = new Headers();
    url: string = "";
    _body: string = "";

    constructor(body: string = "", init: ResponseInit = defaultResponseInit) {
        this._body = body;
        let s = 200;
        let st = "OK";
        let h = new Headers();
        if (init.status !== undefined && init.status > 0) {
            s = init.status;
        }
        if (init.statusText !== undefined && init.statusText.length > 0) {
            st = init.statusText;
        }
        if (init.headers instanceof Headers) {
            h = new Headers(init.headers as Headers);
        }
        this.status = s;
        this.statusText = st;
        this.ok = (s >= 200 && s < 300);
        this.headers = h;
        this.url = "";
    }

    async text(): Promise<string> {
        return this._body;
    }

    async json<T = unknown>(): Promise<T> {
        return JSON.parse(this._body) as T;
    }

    async arrayBuffer(): Promise<ArrayBuffer> {
        const buf = new Uint8Array(this._body.length);
        for (let i = 0; i < this._body.length; i++) {
            buf[i] = this._body.charCodeAt(i);
        }
        return buf.buffer as ArrayBuffer;
    }

    static json(data: string, init: ResponseInit = defaultResponseInit): Response {
        let headers = new Headers();
        if (init.headers instanceof Headers) {
            headers = new Headers(init.headers as Headers);
        }
        if (!headers.has("content-type")) {
            headers.set("content-type", "application/json");
        }
        let s = 200;
        let st = "OK";
        if (init.status !== undefined && init.status > 0) {
            s = init.status;
        }
        if (init.statusText !== undefined && init.statusText.length > 0) {
            st = init.statusText;
        }
        const respInit: ResponseInit = {
            status: s,
            statusText: st,
            headers: headers
        };
        return new Response(data, respInit);
    }

    static error(): Response {
        const respInit: ResponseInit = {
            status: 0,
            statusText: "",
            headers: null
        };
        return new Response("", respInit);
    }

    static redirect(url: string, status: number = 302): Response {
        const headers = new Headers();
        headers.set("location", url);
        const respInit: ResponseInit = {
            status: status,
            statusText: "Found",
            headers: headers
        };
        return new Response("", respInit);
    }
}

export class FetchResponseData {
    status: number;
    statusText: string;
    headers: string[];
    body: string;

    constructor(status: number, statusText: string, headers: string[], body: string) {
        this.status = status;
        this.statusText = statusText;
        this.headers = headers;
        this.body = body;
    }
}

declare namespace __scriptgo {
    function fetchSync(url: string, method?: string, headers?: string[], body?: string): FetchResponseData;
}

export async function fetch(input: unknown, init: RequestInit = defaultRequestInit): Promise<Response> {
    let url = "";
    let method = "GET";
    let body = "";
    let headers: Headers = new Headers();
    if (typeof input === "string") {
        url = input;
    } else if (input instanceof Request) {
        url = input.url;
        method = input.method;
        body = input.body;
        headers = input.headers;
    }
    if (init.method !== undefined && init.method.length > 0) {
        method = init.method;
    }
    if (init.body !== undefined && init.body !== null && init.body.length > 0) {
        body = init.body;
    }
    if (init.headers instanceof Headers) {
        headers = init.headers as Headers;
    }
    const headerEntries = headers.entries();
    const flatHeaders: string[] = [];
    for (let i = 0; i < headerEntries.length; i++) {
        flatHeaders.push(headerEntries[i][0]);
        flatHeaders.push(headerEntries[i][1]);
    }
    const raw = __scriptgo.fetchSync(url, method, flatHeaders, body);
    const respHeaders = new Headers();
    for (let i = 0; i < raw.headers.length; i += 2) {
        if (i + 1 < raw.headers.length) {
            respHeaders.append(raw.headers[i], raw.headers[i + 1]);
        }
    }
    const respInit: ResponseInit = {
        status: raw.status,
        statusText: raw.statusText,
        headers: respHeaders
    };
    return new Response(raw.body, respInit);
}

export const METHODS: string[] = [
    "ACL", "BIND", "CHECKOUT", "CONNECT", "COPY", "DELETE", "GET", "HEAD", "LINK",
    "LOCK", "M-SEARCH", "MERGE", "MKACTIVITY", "MKCALENDAR", "MKCOL", "MOVE", "NOTIFY",
    "OPTIONS", "PATCH", "POST", "PROPFIND", "PROPPATCH", "PURGE", "PUT", "REBIND",
    "REPORT", "SEARCH", "SOURCE", "SUBSCRIBE", "TRACE", "UNBIND", "UNLINK", "UNLOCK",
    "UNSUBSCRIBE"
];

export const STATUS_CODES: Record<string, string> = {
    "100": "Continue",
    "101": "Switching Protocols",
    "102": "Processing",
    "103": "Early Hints",
    "200": "OK",
    "201": "Created",
    "202": "Accepted",
    "203": "Non-Authoritative Information",
    "204": "No Content",
    "205": "Reset Content",
    "206": "Partial Content",
    "300": "Multiple Choices",
    "301": "Moved Permanently",
    "302": "Found",
    "303": "See Other",
    "304": "Not Modified",
    "307": "Temporary Redirect",
    "308": "Permanent Redirect",
    "400": "Bad Request",
    "401": "Unauthorized",
    "402": "Payment Required",
    "403": "Forbidden",
    "404": "Not Found",
    "405": "Method Not Allowed",
    "408": "Request Timeout",
    "409": "Conflict",
    "410": "Gone",
    "418": "I'm a Teapot",
    "429": "Too Many Requests",
    "500": "Internal Server Error",
    "501": "Not Implemented",
    "502": "Bad Gateway",
    "503": "Service Unavailable",
    "504": "Gateway Timeout"
};

export function getStatusText(code: number): string {
    if (code === 100) return "Continue";
    if (code === 101) return "Switching Protocols";
    if (code === 102) return "Processing";
    if (code === 103) return "Early Hints";
    if (code === 200) return "OK";
    if (code === 201) return "Created";
    if (code === 202) return "Accepted";
    if (code === 203) return "Non-Authoritative Information";
    if (code === 204) return "No Content";
    if (code === 205) return "Reset Content";
    if (code === 206) return "Partial Content";
    if (code === 300) return "Multiple Choices";
    if (code === 301) return "Moved Permanently";
    if (code === 302) return "Found";
    if (code === 303) return "See Other";
    if (code === 304) return "Not Modified";
    if (code === 307) return "Temporary Redirect";
    if (code === 308) return "Permanent Redirect";
    if (code === 400) return "Bad Request";
    if (code === 401) return "Unauthorized";
    if (code === 402) return "Payment Required";
    if (code === 403) return "Forbidden";
    if (code === 404) return "Not Found";
    if (code === 405) return "Method Not Allowed";
    if (code === 408) return "Request Timeout";
    if (code === 409) return "Conflict";
    if (code === 410) return "Gone";
    if (code === 418) return "I'm a Teapot";
    if (code === 429) return "Too Many Requests";
    if (code === 500) return "Internal Server Error";
    if (code === 501) return "Not Implemented";
    if (code === 502) return "Bad Gateway";
    if (code === 503) return "Service Unavailable";
    if (code === 504) return "Gateway Timeout";
    return "";
}

export const maxHeaderSize: number = 16384;

export function validateHeaderName(name: string, label: string = ""): void {
    if (typeof name !== "string" || name.length === 0) {
        throw new TypeError((label.length > 0 ? label : "Header name") + " must be a non-empty string");
    }
    for (let i = 0; i < name.length; i++) {
        const code = name.charCodeAt(i);
        if (code <= 32 || code >= 127) {
            throw new TypeError("Invalid character in header name [" + name + "]");
        }
    }
}

export function validateHeaderValue(name: string, value: unknown): void {
    if (value === undefined) {
        throw new TypeError("Invalid value \"undefined\" for header \"" + name + "\"");
    }
}

export function setMaxIdleHTTPParsers(max: number): void {
    // runtime configuration hook
}

export function getMaxIdleHTTPParsers(): number {
    return 1000;
}

export interface AgentOptions {
    maxFreeSockets?: number;
    maxSockets?: number;
    maxTotalSockets?: number;
    maxRequestsPerSocket?: number;
}

const defaultAgentOptions: AgentOptions = {
    maxFreeSockets: 256,
    maxSockets: 999999,
    maxTotalSockets: 999999,
    maxRequestsPerSocket: 0
};

export interface AgentNameOptions {
    host?: string;
    port?: number | string;
    localAddress?: string;
}

const defaultAgentNameOptions: AgentNameOptions = { host: "localhost", port: 80, localAddress: "" };

export class HttpConnection {
    connected: boolean = true;
    options: string = "";
    constructor(connected: boolean = true, options: string = "") {
        this.connected = connected;
        this.options = options;
    }
    on(event: string, listener: unknown = null): this {
        return this;
    }
}

export class Agent {
    options: AgentOptions | null = null;
    freeSockets: string[] = [];
    sockets: string[] = [];
    requests: string[] = [];
    maxFreeSockets: number = 256;
    maxSockets: number = 999999;
    maxTotalSockets: number = 999999;
    maxRequestsPerSocket: number = 0;

    constructor(options: AgentOptions = defaultAgentOptions) {
        this.options = options;
        if (options.maxFreeSockets !== undefined) {
            this.maxFreeSockets = options.maxFreeSockets;
        }
        if (options.maxSockets !== undefined) {
            this.maxSockets = options.maxSockets;
        }
        if (options.maxTotalSockets !== undefined) {
            this.maxTotalSockets = options.maxTotalSockets;
        }
    }

    createConnection(options: unknown = null, callback: unknown = null): HttpConnection {
        const conn = new HttpConnection(true, String(options));
        if (typeof callback === "function") {
            const fn = callback as Function;
            fn(null, conn);
        }
        return conn;
    }

    destroy(): void {
        this.freeSockets = [];
        this.sockets = [];
        this.requests = [];
    }

    getName(options: AgentNameOptions = defaultAgentNameOptions): string {
        let host = "localhost";
        let port = "80";
        let localAddress = "";
        if (options.host !== undefined && options.host.length > 0) {
            host = options.host;
        }
        if (options.port !== undefined) {
            port = String(options.port);
        }
        if (options.localAddress !== undefined && options.localAddress.length > 0) {
            localAddress = options.localAddress;
        }
        return host + ":" + port + ":" + localAddress;
    }

    keepSocketAlive(socket: unknown): boolean {
        return true;
    }

    reuseSocket(socket: unknown, request: unknown): void {
        const dummy = socket;
    }
}

export const globalAgent: Agent = new Agent();

export class OutgoingMessage {
    connection: string | null = null;
    socket: string | null = null;
    headersSent: boolean = false;
    writableCorked: number = 0;
    writableEnded: boolean = false;
    writableFinished: boolean = false;
    writableHighWaterMark: number = 16384;
    writableLength: number = 0;
    writableObjectMode: boolean = false;
    strictContentLength: boolean = false;
    sendDate: boolean = true;
    req: IncomingMessage | null = null;

    private _events: HttpEventBucket[] = [];
    private _headerKeys: string[] = [];
    private _headerValues: string[] = [];
    private _rawHeaderNames: string[] = [];
    private _bodyChunks: string[] = [];

    private _findBucketIndex(event: string): number {
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].name === event) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateBucketIndex(event: string): number {
        let idx = this._findBucketIndex(event);
        if (idx < 0) {
            this._events.push(new HttpEventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    on(event: string, listener: Function): OutgoingMessage {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): OutgoingMessage {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return false;
        const bucket = this._events[idx];
        const snapshot: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            snapshot.push(bucket.listeners[i]);
        }
        const next: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                next.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = next;
        for (let i = 0; i < snapshot.length; i++) {
            const fn = snapshot[i].fn;
            if (arg1 === undefined) {
                fn();
            } else if (arg2 === undefined) {
                fn(arg1);
            } else {
                fn(arg1, arg2);
            }
        }
        return true;
    }

    setHeader(name: string, value: unknown): OutgoingMessage {
        validateHeaderName(name);
        validateHeaderValue(name, value);
        const lower = name.toLowerCase();
        const valStr = String(value);
        for (let i = 0; i < this._headerKeys.length; i++) {
            if (this._headerKeys[i] === lower) {
                this._headerValues[i] = valStr;
                this._rawHeaderNames[i] = name;
                return this;
            }
        }
        this._headerKeys.push(lower);
        this._headerValues.push(valStr);
        this._rawHeaderNames.push(name);
        return this;
    }

    getHeader(name: string): string | null {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._headerKeys.length; i++) {
            if (this._headerKeys[i] === lower) {
                return this._headerValues[i];
            }
        }
        return null;
    }

    getHeaders(): Headers {
        const res = new Headers();
        for (let i = 0; i < this._headerKeys.length; i++) {
            res.set(this._headerKeys[i], this._headerValues[i]);
        }
        return res;
    }

    getHeaderNames(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._headerKeys.length; i++) {
            res.push(this._headerKeys[i]);
        }
        return res;
    }

    getRawHeaderNames(): string[] {
        const res: string[] = [];
        for (let i = 0; i < this._rawHeaderNames.length; i++) {
            res.push(this._rawHeaderNames[i]);
        }
        return res;
    }

    hasHeader(name: string): boolean {
        const lower = name.toLowerCase();
        for (let i = 0; i < this._headerKeys.length; i++) {
            if (this._headerKeys[i] === lower) {
                return true;
            }
        }
        return false;
    }

    removeHeader(name: string): void {
        const lower = name.toLowerCase();
        const nextKeys: string[] = [];
        const nextVals: string[] = [];
        const nextRaw: string[] = [];
        for (let i = 0; i < this._headerKeys.length; i++) {
            if (this._headerKeys[i] !== lower) {
                nextKeys.push(this._headerKeys[i]);
                nextVals.push(this._headerValues[i]);
                nextRaw.push(this._rawHeaderNames[i]);
            }
        }
        this._headerKeys = nextKeys;
        this._headerValues = nextVals;
        this._rawHeaderNames = nextRaw;
    }

    appendHeader(name: string, value: unknown): OutgoingMessage {
        validateHeaderName(name);
        const lower = name.toLowerCase();
        const valStr = String(value);
        for (let i = 0; i < this._headerKeys.length; i++) {
            if (this._headerKeys[i] === lower) {
                this._headerValues[i] = this._headerValues[i] + ", " + valStr;
                return this;
            }
        }
        this._headerKeys.push(lower);
        this._headerValues.push(valStr);
        this._rawHeaderNames.push(name);
        return this;
    }

    setHeaders(headers: unknown): OutgoingMessage {
        if (headers instanceof Headers) {
            const h = headers as Headers;
            const entries = h.entries();
            for (let i = 0; i < entries.length; i++) {
                this.setHeader(entries[i][0], entries[i][1]);
            }
        }
        return this;
    }

    flushHeaders(): void {
        this.headersSent = true;
    }

    write(chunk: unknown, encoding: string = "", callback: unknown = null): boolean {
        const str = String(chunk);
        this._bodyChunks.push(str);
        this.writableLength += str.length;
        if (typeof callback === "function") {
            const fn = callback as Function;
            fn(null);
        }
        return true;
    }

    end(chunk: unknown = null, encoding: string = "", callback: unknown = null): OutgoingMessage {
        if (chunk !== null && chunk !== undefined) {
            this.write(chunk, encoding);
        }
        this.writableEnded = true;
        this.writableFinished = true;
        this.headersSent = true;
        if (typeof callback === "function") {
            const fn = callback as Function;
            fn();
        }
        this.emit("finish");
        return this;
    }

    destroy(error: unknown = null): OutgoingMessage {
        this.writableEnded = true;
        this.emit("close");
        return this;
    }

    cork(): void {
        this.writableCorked++;
    }

    uncork(): void {
        if (this.writableCorked > 0) {
            this.writableCorked--;
        }
    }

    setTimeout(msecs: number = 0, callback: unknown = null): OutgoingMessage {
        if (typeof callback === "function") {
            this.once("timeout", callback as Function);
        }
        return this;
    }

    addTrailers(headers: unknown = null): void {
        const dummy = headers;
    }

    pipe(dest: unknown): unknown {
        return dest;
    }
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
    path: string = "/";
    method: string = "GET";
    host: string = "localhost";
    protocol: string = "http:";
    reusedSocket: boolean = false;
    maxHeadersCount: number = 2000;

    constructor(urlOrOptions: unknown, cb: unknown = null) {
        super();
        if (typeof urlOrOptions === "string") {
            const s = urlOrOptions as string;
            const slash = s.indexOf("/", 8);
            this.path = slash >= 0 ? s.substring(slash) : s;
        } else if (urlOrOptions !== null && typeof urlOrOptions === "object") {
            const opts = urlOrOptions as ClientRequestOptions;
            if (opts.path !== undefined) this.path = opts.path;
            if (opts.pathname !== undefined) this.path = opts.pathname;
            if (opts.method !== undefined) this.method = opts.method.toUpperCase();
            if (opts.host !== undefined) this.host = opts.host;
            if (opts.hostname !== undefined) this.host = opts.hostname;
            if (opts.protocol !== undefined) this.protocol = opts.protocol;
            if (opts.headers !== undefined) {
                this.setHeaders(opts.headers);
            }
        }
        if (typeof cb === "function") {
            const fn = cb as Function;
            this.once("response", fn);
        }
    }

    abort(): void {
        this.destroy();
        this.emit("abort");
    }

    setNoDelay(noDelay: boolean = true): void {
        const dummy = noDelay;
    }

    setSocketKeepAlive(enable: boolean = false, initialDelay: number = 0): void {
        const dummy = enable;
    }
}

export class IncomingMessage {
    aborted: boolean = false;
    complete: boolean = true;
    connection: string = "";
    socket: string = "";
    headers: string[] = [];
    headersDistinct: string[] = [];
    httpVersion: string = "1.1";
    method: string = "GET";
    rawHeaders: string[] = [];
    rawTrailers: string[] = [];
    statusCode: number = 200;
    statusMessage: string = "OK";
    trailers: string[] = [];
    trailersDistinct: string[] = [];
    url: string = "/";

    private _events: HttpEventBucket[] = [];

    constructor(socket: unknown = null) {
        if (socket !== null && socket !== undefined) {
            this.socket = String(socket);
            this.connection = String(socket);
        }
    }

    private _findBucketIndex(event: string): number {
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].name === event) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateBucketIndex(event: string): number {
        let idx = this._findBucketIndex(event);
        if (idx < 0) {
            this._events.push(new HttpEventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    on(event: string, listener: Function): IncomingMessage {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): IncomingMessage {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return false;
        const bucket = this._events[idx];
        const snapshot: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            snapshot.push(bucket.listeners[i]);
        }
        const next: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                next.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = next;
        for (let i = 0; i < snapshot.length; i++) {
            const fn = snapshot[i].fn;
            if (arg1 === undefined) {
                fn();
            } else if (arg2 === undefined) {
                fn(arg1);
            } else {
                fn(arg1, arg2);
            }
        }
        return true;
    }

    destroy(error: unknown = null): IncomingMessage {
        this.aborted = true;
        this.emit("close");
        return this;
    }

    setTimeout(msecs: number = 0, callback: unknown = null): IncomingMessage {
        if (typeof callback === "function") {
            this.once("timeout", callback as Function);
        }
        return this;
    }

    pipe(dest: unknown): unknown {
        return dest;
    }
}

export class ServerResponse extends OutgoingMessage {
    statusCode: number = 200;
    statusMessage: string = "OK";

    constructor(req: IncomingMessage | null = null) {
        super();
        this.req = req;
    }

    writeHead(statusCode: number, statusMessageOrHeaders: unknown = null, headers: Headers | null = null): ServerResponse {
        this.statusCode = statusCode;
        if (typeof statusMessageOrHeaders === "string") {
            this.statusMessage = statusMessageOrHeaders;
            if (headers !== null) {
                this.setHeaders(headers);
            }
        } else if (statusMessageOrHeaders instanceof Headers) {
            this.setHeaders(statusMessageOrHeaders as Headers);
        }
        this.headersSent = true;
        return this;
    }

    writeContinue(): void {
        const dummy = true;
    }

    writeEarlyHints(hints: unknown = null, callback: unknown = null): void {
        if (typeof callback === "function") {
            const fn = callback as Function;
            fn();
        }
    }

    writeProcessing(): void {
        const dummy = true;
    }
}

export class Server {
    listening: boolean = false;
    maxHeadersCount: number = 2000;
    maxRequestsPerSocket: number = 0;
    timeout: number = 0;
    headersTimeout: number = 60000;
    keepAliveTimeout: number = 5000;
    keepAliveTimeoutBuffer: number = 1000;
    requestTimeout: number = 300000;

    private _events: HttpEventBucket[] = [];

    constructor(optionsOrListener: unknown = null, listener: unknown = null) {
        if (typeof optionsOrListener === "function") {
            const fn = optionsOrListener as Function;
            this.on("request", fn);
        } else if (optionsOrListener !== null && typeof listener === "function") {
            const fn = listener as Function;
            this.on("request", fn);
        }
    }

    private _findBucketIndex(event: string): number {
        for (let i = 0; i < this._events.length; i++) {
            if (this._events[i].name === event) {
                return i;
            }
        }
        return -1;
    }

    private _getOrCreateBucketIndex(event: string): number {
        let idx = this._findBucketIndex(event);
        if (idx < 0) {
            this._events.push(new HttpEventBucket(event, []));
            idx = this._events.length - 1;
        }
        return idx;
    }

    on(event: string, listener: Function): Server {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, false));
        return this;
    }

    once(event: string, listener: Function): Server {
        const idx = this._getOrCreateBucketIndex(event);
        this._events[idx].listeners.push(new HttpListenerEntry(listener, true));
        return this;
    }

    emit(event: string, arg1: unknown = undefined, arg2: unknown = undefined): boolean {
        const idx = this._findBucketIndex(event);
        if (idx < 0) return false;
        const bucket = this._events[idx];
        const snapshot: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            snapshot.push(bucket.listeners[i]);
        }
        const next: HttpListenerEntry[] = [];
        for (let i = 0; i < bucket.listeners.length; i++) {
            if (!bucket.listeners[i].once) {
                next.push(bucket.listeners[i]);
            }
        }
        bucket.listeners = next;
        for (let i = 0; i < snapshot.length; i++) {
            const fn = snapshot[i].fn;
            if (arg1 === undefined) {
                fn();
            } else if (arg2 === undefined) {
                fn(arg1);
            } else {
                fn(arg1, arg2);
            }
        }
        return true;
    }

    listen(port: unknown = 0, hostOrCb: unknown = null, cb: unknown = null): Server {
        this.listening = true;
        if (typeof hostOrCb === "function") {
            const fn = hostOrCb as Function;
            fn();
        } else if (typeof cb === "function") {
            const fn = cb as Function;
            fn();
        }
        this.emit("listening");
        return this;
    }

    close(callback: unknown = null): Server {
        this.listening = false;
        if (typeof callback === "function") {
            const fn = callback as Function;
            fn();
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

    setTimeout(msecs: number = 0, callback: unknown = null): Server {
        this.timeout = msecs;
        if (typeof callback === "function") {
            this.on("timeout", callback as Function);
        }
        return this;
    }

    async asyncDispose(): Promise<void> {
        this.close();
    }
}

export function createServer(optionsOrListener: unknown = null, listener: unknown = null): Server {
    if (typeof optionsOrListener === "function") {
        return new Server(optionsOrListener, null);
    }
    if (typeof listener === "function") {
        return new Server(optionsOrListener, listener);
    }
    return new Server(optionsOrListener, null);
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
    METHODS,
    STATUS_CODES,
    Agent,
    globalAgent,
    OutgoingMessage,
    ServerResponse,
    IncomingMessage,
    ClientRequest,
    Server,
    createServer,
    request,
    get,
    validateHeaderName,
    validateHeaderValue,
    setMaxIdleHTTPParsers,
    getMaxIdleHTTPParsers,
};
