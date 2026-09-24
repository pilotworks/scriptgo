import { EventEmitter } from "node:events";
import {
    Socket,
    Server as NetServer,
    ServerOptions as NetServerOptions,
    connect as netConnect
} from "node:net";
import { URL } from "node:url";
import { FormData } from "node:formdata";
import { File } from "node:buffer";

export class HeadersIterator<T = unknown> {
    private _values: T[];
    private _index: number = 0;

    constructor(values: T[]) {
        this._values = values;
        this._index = 0;
    }

    next(): { value: T | undefined; done: boolean } {
        if (this._index < this._values.length) {
            const value = this._values[this._index];
            this._index = this._index + 1;
            return { value: value, done: false };
        }
        return { value: undefined, done: true };
    }

    [Symbol.iterator](): HeadersIterator<T> {
        return this;
    }
}

export class Headers {
    _keys: string[] = [];
    _values: string[] = [];

    constructor(init: Headers | Record<string, string> | [string, string][] | null = null) {
        this._keys = [];
        this._values = [];
        if (init !== null && init !== undefined) {
            if (init instanceof Headers) {
                const other = init as Headers;
                for (let i = 0; i < other._keys.length; i++) {
                    this.append(other._keys[i], other._values[i]);
                }
            } else if (Array.isArray(init)) {
                for (let i = 0; i < init.length; i++) {
                    const item = init[i];
                    if (Array.isArray(item) && item.length >= 2) {
                        this.append(item[0], item[1]);
                    }
                }
            } else if (typeof init === "object") {
                const keys = Object.keys(init);
                for (let i = 0; i < keys.length; i++) {
                    const k = keys[i];
                    this.append(k, String((init as any)[k]));
                }
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

    entries(): HeadersIterator<[string, string]> {
        const res: [string, string][] = [];
        for (let i = 0; i < this._keys.length; i++) {
            const pair: [string, string] = [this._keys[i], this._values[i]];
            res.push(pair);
        }
        return new HeadersIterator<[string, string]>(res);
    }

    keys(): HeadersIterator<string> {
        return new HeadersIterator<string>(this._keys);
    }

    values(): HeadersIterator<string> {
        return new HeadersIterator<string>(this._values);
    }

    [Symbol.iterator](): HeadersIterator<[string, string]> {
        return this.entries();
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
    body: unknown = null;
    _body: string = "";

    constructor(body: string = "", init: ResponseInit = defaultResponseInit) {
        this._body = body;
        this.body = null;
        let s = 200;
        let st = "OK";
        let h = new Headers();
        if (init.status !== undefined && init.status > 0) {
            s = init.status;
        }
        if (init.statusText !== undefined && init.statusText.length > 0) {
            st = init.statusText;
        }
        if (init.headers !== undefined && init.headers !== null) {
            h = new Headers(init.headers as any);
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

    async formData(): Promise<FormData> {
        const fd = new FormData();
        const ct = this.headers.get("content-type") || "";
        if (ct.includes("multipart/form-data")) {
            let boundary = "";
            const bIdx = ct.indexOf("boundary=");
            if (bIdx !== -1) {
                boundary = ct.substring(bIdx + 9).trim();
                if (boundary.startsWith('"') && boundary.endsWith('"')) {
                    boundary = boundary.substring(1, boundary.length - 1);
                }
            }
            if (boundary.length > 0) {
                const delimiter = "--" + boundary;
                const parts = this._body.split(delimiter);
                for (let i = 0; i < parts.length; i++) {
                    const part = parts[i];
                    if (part.length === 0 || part === "--" || part === "--\r\n" || part.startsWith("--")) {
                        continue;
                    }
                    const headerEnd = part.indexOf("\r\n\r\n");
                    const lfHeaderEnd = part.indexOf("\n\n");
                    let headerBlock = "";
                    let bodyBlock = "";
                    if (headerEnd !== -1) {
                        headerBlock = part.substring(0, headerEnd);
                        bodyBlock = part.substring(headerEnd + 4);
                        if (bodyBlock.endsWith("\r\n")) {
                            bodyBlock = bodyBlock.substring(0, bodyBlock.length - 2);
                        }
                    } else if (lfHeaderEnd !== -1) {
                        headerBlock = part.substring(0, lfHeaderEnd);
                        bodyBlock = part.substring(lfHeaderEnd + 2);
                        if (bodyBlock.endsWith("\n")) {
                            bodyBlock = bodyBlock.substring(0, bodyBlock.length - 1);
                        }
                    } else {
                        continue;
                    }
                    let name = "";
                    let filename = "";
                    let partContentType = "text/plain";
                    const lines = headerBlock.split("\n");
                    for (let j = 0; j < lines.length; j++) {
                        const line = lines[j].trim();
                        const lower = line.toLowerCase();
                        if (lower.startsWith("content-disposition:")) {
                            const nameMatch = line.indexOf("name=\"");
                            if (nameMatch !== -1) {
                                const endQuote = line.indexOf("\"", nameMatch + 6);
                                if (endQuote !== -1) {
                                    name = line.substring(nameMatch + 6, endQuote);
                                }
                            }
                            const fnMatch = line.indexOf("filename=\"");
                            if (fnMatch !== -1) {
                                const fnEnd = line.indexOf("\"", fnMatch + 10);
                                if (fnEnd !== -1) {
                                    filename = line.substring(fnMatch + 10, fnEnd);
                                }
                            }
                        } else if (lower.startsWith("content-type:")) {
                            partContentType = line.substring(13).trim();
                        }
                    }
                    if (name.length > 0) {
                        if (filename.length > 0) {
                            fd.append(name, new File([bodyBlock], filename, { type: partContentType }));
                        } else {
                            fd.append(name, bodyBlock);
                        }
                    }
                }
            }
            return fd;
        }

        if (this._body.length > 0) {
            const pairs = this._body.split("&");
            for (let i = 0; i < pairs.length; i++) {
                const pair = pairs[i];
                if (pair.length === 0) continue;
                const eq = pair.indexOf("=");
                if (eq !== -1) {
                    const rawKey = pair.substring(0, eq).replaceAll("+", " ");
                    const rawVal = pair.substring(eq + 1).replaceAll("+", " ");
                    try {
                        fd.append(decodeURIComponent(rawKey), decodeURIComponent(rawVal));
                    } catch (e) {
                        fd.append(rawKey, rawVal);
                    }
                } else {
                    const rawKey = pair.replaceAll("+", " ");
                    try {
                        fd.append(decodeURIComponent(rawKey), "");
                    } catch (e) {
                        fd.append(rawKey, "");
                    }
                }
            }
        }
        return fd;
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
    const flatHeaders: string[] = [];
    for (let i = 0; i < headers._keys.length; i++) {
        flatHeaders.push(headers._keys[i]);
        flatHeaders.push(headers._values[i]);
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


export interface AgentOptions {
    keepAlive?: boolean;
    keepAliveMsecs?: number;
    maxSockets?: number;
    maxFreeSockets?: number;
    timeout?: number;
}

export interface RequestOptions extends AgentOptions {
    protocol?: string;
    host?: string;
    hostname?: string;
    port?: number;
    method?: string;
    path?: string;
    headers?: Record<string, string>;
    auth?: string;
    agent?: Agent | boolean;
}

export class Agent extends EventEmitter {
    maxSockets: number = Infinity;
    maxFreeSockets: number = 256;
    keepAlive: boolean = false;
    keepAliveMsecs: number = 1000;
    options: AgentOptions;

    constructor(options?: AgentOptions) {
        super();
        this.options = options || {};
        if (this.options.keepAlive !== undefined) {
            this.keepAlive = !!this.options.keepAlive;
        }
        if (this.options.maxSockets !== undefined && typeof this.options.maxSockets === "number") {
            this.maxSockets = this.options.maxSockets;
        }
    }

    destroy(): void {
        this.emit("free");
    }
}

export const globalAgent = new Agent();

export class OutgoingMessage extends EventEmitter {
    headersSent: boolean = false;
    finished: boolean = false;
    socket: Socket | null = null;
    protected _headers: Record<string, string> = {};

    setHeader(name: string, value: string): this {
        if (this.headersSent) {
            throw new Error("Cannot set headers after they are sent to the client");
        }
        this._headers[name.toLowerCase()] = String(value);
        return this;
    }

    getHeader(name: string): string | undefined {
        return this._headers[name.toLowerCase()];
    }

    hasHeader(name: string): boolean {
        return this._headers[name.toLowerCase()] !== undefined;
    }

    removeHeader(name: string): this {
        if (this.headersSent) {
            throw new Error("Cannot remove headers after they are sent to the client");
        }
        const key = name.toLowerCase();
        const nextHeaders: Record<string, string> = {};
        for (const k of Object.keys(this._headers)) {
            if (k !== key) {
                nextHeaders[k] = this._headers[k];
            }
        }
        this._headers = nextHeaders;
        return this;
    }

    getHeaderNames(): string[] {
        return Object.keys(this._headers);
    }

    getHeaders(): Record<string, string> {
        const copy: Record<string, string> = {};
        for (const k of Object.keys(this._headers)) {
            copy[k] = this._headers[k];
        }
        return copy;
    }

    flushHeaders(): void {}
}

export class ServerResponse extends OutgoingMessage {
    statusCode: number = 200;
    statusMessage: string = "OK";
    sendDate: boolean = true;

    constructor(socket: Socket) {
        super();
        this.socket = socket;
    }

    writeHead(
        statusCode: number,
        statusMessageOrHeaders?: string | Record<string, string>,
        headers?: Record<string, string>
    ): this {
        if (this.headersSent) {
            throw new Error("Cannot write head after headers are sent");
        }
        this.statusCode = statusCode;
        let hdrs: Record<string, string> | undefined = undefined;
        if (typeof statusMessageOrHeaders === "string") {
            this.statusMessage = statusMessageOrHeaders;
            hdrs = headers;
        } else {
            const codeStr = String(statusCode);
            if (STATUS_CODES[codeStr] !== undefined) {
                this.statusMessage = STATUS_CODES[codeStr];
            } else {
                this.statusMessage = "OK";
            }
            if (typeof statusMessageOrHeaders === "object" && statusMessageOrHeaders !== null) {
                hdrs = statusMessageOrHeaders as Record<string, string>;
            }
        }
        if (hdrs) {
            for (const key of Object.keys(hdrs)) {
                this.setHeader(key, hdrs[key]);
            }
        }
        return this;
    }

    private _sendHeaders(): void {
        if (this.headersSent) return;
        this.headersSent = true;
        let head = `HTTP/1.1 ${this.statusCode} ${this.statusMessage}\r\n`;
        for (const key of Object.keys(this._headers)) {
            head += `${key}: ${this._headers[key]}\r\n`;
        }
        head += "\r\n";
        if (this.socket) {
            this.socket.write(head);
        }
    }

    write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        if (!this.headersSent) {
            this._sendHeaders();
        }
        const data = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        if (this.socket) {
            this.socket.write(data);
        }
        if (callback) {
            callback();
        }
        return true;
    }

    end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this.finished) return this;
        if (!this.headersSent) {
            this._sendHeaders();
        }
        if (chunk !== undefined && chunk !== null) {
            const data = typeof chunk === "string" ? chunk : String(chunk);
            if (this.socket) {
                this.socket.write(data);
            }
        }
        this.finished = true;
        if (callback) {
            callback();
        }
        this.emit("finish");
        if (this.socket) {
            this.socket.end();
        }
        return this;
    }
}

export class IncomingMessage extends EventEmitter {
    statusCode: number = 200;
    statusMessage: string = "OK";
    headers: Record<string, string> = {};
    rawHeaders: string[] = [];
    httpVersion: string = "1.1";
    method: string = "";
    url: string = "";
    socket: Socket;
    complete: boolean = false;
    private _encoding: string = "utf8";

    constructor(socket: Socket) {
        super();
        this.socket = socket;
    }

    setEncoding(encoding: string): this {
        this._encoding = encoding;
        return this;
    }

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.once("timeout", callback);
        }
        return this;
    }

    destroy(error?: Error): this {
        this.complete = true;
        if (error) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }
}

export class Server extends NetServer {
    constructor(
        optionsOrListener?: NetServerOptions | ((req: IncomingMessage, res: ServerResponse) => void),
        requestListener?: (req: IncomingMessage, res: ServerResponse) => void
    ) {
        let opts: NetServerOptions | null = null;
        let listener: ((req: IncomingMessage, res: ServerResponse) => void) | undefined = undefined;

        if (typeof optionsOrListener === "function") {
            listener = optionsOrListener;
        } else if (optionsOrListener) {
            opts = optionsOrListener as NetServerOptions;
            listener = requestListener;
        }

        super(opts);

        if (listener) {
            this.on("request", listener);
        }

        this.on("connection", (socket: Socket) => {
            let buffer = "";
            let req: IncomingMessage | null = null;
            let res: ServerResponse | null = null;
            let headersParsed = false;

            socket.on("data", (data: unknown) => {
                const str = typeof data === "string" ? data : (data !== null && data !== undefined ? String(data) : "");
                buffer += str;

                if (!headersParsed) {
                    const headerEndIdx = buffer.indexOf("\r\n\r\n");
                    if (headerEndIdx !== -1) {
                        headersParsed = true;
                        const headerPart = buffer.slice(0, headerEndIdx);
                        const bodyPart = buffer.slice(headerEndIdx + 4);

                        const lines = headerPart.split("\r\n");
                        const reqLine = lines[0] || "";
                        const reqParts = reqLine.split(" ");
                        const method = reqParts[0] || "GET";
                        const url = reqParts.length > 1 ? reqParts[1] : "/";

                        req = new IncomingMessage(socket);
                        req.method = method;
                        req.url = url;

                        for (let i = 1; i < lines.length; i++) {
                            const colonIdx = lines[i].indexOf(":");
                            if (colonIdx !== -1) {
                                const hName = lines[i].slice(0, colonIdx).trim().toLowerCase();
                                const hVal = lines[i].slice(colonIdx + 1).trim();
                                req.headers[hName] = hVal;
                                req.rawHeaders.push(lines[i].slice(0, colonIdx).trim());
                                req.rawHeaders.push(hVal);
                            }
                        }

                        res = new ServerResponse(socket);
                        this.emit("request", req, res);

                        if (bodyPart.length > 0) {
                            req.emit("data", bodyPart);
                        }
                    }
                } else if (req) {
                    req.emit("data", str);
                }
            });

            socket.on("end", () => {
                if (req) {
                    req.complete = true;
                    req.emit("end");
                }
            });

            socket.on("close", () => {
                if (res) {
                    res.emit("close");
                }
            });
        });
    }
}

export class ClientRequest extends OutgoingMessage {
    method: string = "GET";
    path: string = "/";
    host: string = "localhost";
    port: number = 80;
    aborted: boolean = false;
    reusedSocket: boolean = false;
    private _bodyChunks: string[] = [];
    private _ended: boolean = false;
    private _options: RequestOptions;
    private _res: IncomingMessage | null = null;

    constructor(options: RequestOptions, callback?: (res: IncomingMessage) => void) {
        super();
        this._options = options;
        if (options.method) {
            const m = options.method.toUpperCase();
            if (METHODS.indexOf(m) !== -1) {
                this.method = m;
            } else {
                this.method = options.method;
            }
        }
        if (options.path) this.path = options.path;
        if (options.host) this.host = options.host;
        else if (options.hostname) this.host = options.hostname;

        if (typeof options.port === "number") {
            this.port = options.port;
        } else {
            this.port = 80;
        }

        if (options.headers) {
            for (const k of Object.keys(options.headers)) {
                this.setHeader(k, options.headers[k]);
            }
        }

        if (callback) {
            this.once("response", callback);
        }
    }

    write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        if (this._ended) {
            throw new Error("write after end");
        }
        const str = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        this._bodyChunks.push(str);
        if (this.socket) {
            this.socket.write(str);
        }
        if (callback) {
            callback();
        }
        return true;
    }

    end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this._ended) return this;
        this._ended = true;

        if (chunk !== undefined && chunk !== null) {
            const str = typeof chunk === "string" ? chunk : String(chunk);
            this._bodyChunks.push(str);
        }

        if (callback) {
            callback();
        }

        this._connectAndSend();
        return this;
    }

    private _connectAndSend(): void {
        const socket = netConnect({ host: this.host, port: this.port }, () => {
            this.socket = socket;
            this.emit("socket", socket);

            let reqStr = `${this.method} ${this.path} HTTP/1.1\r\n`;
            if (!this.getHeader("host")) {
                reqStr += `host: ${this.host}:${this.port}\r\n`;
            }
            if (!this.getHeader("connection")) {
                reqStr += `connection: close\r\n`;
            }

            const body = this._bodyChunks.join("");
            if (body.length > 0 && !this.getHeader("content-length")) {
                reqStr += `content-length: ${body.length}\r\n`;
            }

            for (const name of Object.keys(this._headers)) {
                reqStr += `${name}: ${this._headers[name]}\r\n`;
            }
            reqStr += "\r\n";
            reqStr += body;

            socket.write(reqStr);
        });

        this.socket = socket;

        let buffer = "";
        let res: IncomingMessage | null = null;
        let headersParsed = false;

        socket.on("data", (data: unknown) => {
            const str = typeof data === "string" ? data : (data !== null && data !== undefined ? String(data) : "");
            buffer += str;

            if (!headersParsed) {
                const headerEndIdx = buffer.indexOf("\r\n\r\n");
                if (headerEndIdx !== -1) {
                    headersParsed = true;
                    const headerPart = buffer.slice(0, headerEndIdx);
                    const bodyPart = buffer.slice(headerEndIdx + 4);

                    const lines = headerPart.split("\r\n");
                    const statusLine = lines[0] || "";
                    const statusParts = statusLine.split(" ");
                    const statusCode = statusParts.length > 1 ? parseInt(statusParts[1]) : 200;
                    const statusMessage = statusParts.length > 2 ? statusParts.slice(2).join(" ") : "OK";

                    res = new IncomingMessage(socket);
                    res.statusCode = statusCode;
                    res.statusMessage = statusMessage;
                    this._res = res;

                    for (let i = 1; i < lines.length; i++) {
                        const colonIdx = lines[i].indexOf(":");
                        if (colonIdx !== -1) {
                            const hName = lines[i].slice(0, colonIdx).trim().toLowerCase();
                            const hVal = lines[i].slice(colonIdx + 1).trim();
                            res.headers[hName] = hVal;
                            res.rawHeaders.push(lines[i].slice(0, colonIdx).trim());
                            res.rawHeaders.push(hVal);
                        }
                    }

                    this.emit("response", res);

                    if (bodyPart.length > 0) {
                        res.emit("data", bodyPart);
                    }
                }
            } else if (res) {
                res.emit("data", str);
            }
        });

        socket.on("end", () => {
            if (res) {
                res.complete = true;
                res.emit("end");
            }
            this.emit("close");
        });

        socket.on("error", (err: Error) => {
            this.emit("error", err);
        });

        socket.on("close", () => {
            if (res) {
                res.emit("close");
            }
        });
    }

    abort(): void {
        this.aborted = true;
        this.destroy(new Error("Request aborted"));
    }

    destroy(error?: Error): this {
        if (this.socket) {
            this.socket.destroy(error);
        }
        if (error) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.once("timeout", callback);
        }
        return this;
    }
}

export function createServer(
    optionsOrListener?: NetServerOptions | ((req: IncomingMessage, res: ServerResponse) => void),
    requestListener?: (req: IncomingMessage, res: ServerResponse) => void
): Server {
    return new Server(optionsOrListener, requestListener);
}

export function request(
    urlOrOptions: string | URL | RequestOptions,
    optionsOrCallback?: RequestOptions | ((res: IncomingMessage) => void),
    callback?: (res: IncomingMessage) => void
): ClientRequest {
    let opts: RequestOptions = {};
    let cb: ((res: IncomingMessage) => void) | undefined = callback;

    if (typeof urlOrOptions === "string") {
        const parsed = new URL(urlOrOptions);
        opts.host = parsed.hostname;
        opts.port = parsed.port ? parseInt(parsed.port) : 80;
        opts.path = parsed.pathname + parsed.search;
        opts.method = "GET";
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback;
        } else if (optionsOrCallback) {
            for (const k of Object.keys(optionsOrCallback)) {
                (opts as Record<string, unknown>)[k] = (optionsOrCallback as Record<string, unknown>)[k];
            }
        }
    } else if (urlOrOptions instanceof URL) {
        opts.host = urlOrOptions.hostname;
        opts.port = urlOrOptions.port ? parseInt(urlOrOptions.port) : 80;
        opts.path = urlOrOptions.pathname + urlOrOptions.search;
        opts.method = "GET";
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback;
        } else if (optionsOrCallback) {
            for (const k of Object.keys(optionsOrCallback)) {
                (opts as Record<string, unknown>)[k] = (optionsOrCallback as Record<string, unknown>)[k];
            }
        }
    } else {
        opts = urlOrOptions as RequestOptions;
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback;
        }
    }

    return new ClientRequest(opts, cb);
}

export function get(
    urlOrOptions: string | URL | RequestOptions,
    optionsOrCallback?: RequestOptions | ((res: IncomingMessage) => void),
    callback?: (res: IncomingMessage) => void
): ClientRequest {
    const req = request(urlOrOptions, optionsOrCallback, callback);
    req.end();
    return req;
}

export default {
    Agent,
    globalAgent,
    IncomingMessage,
    OutgoingMessage,
    ServerResponse,
    ClientRequest,
    Server,
    createServer,
    request,
    get,
    METHODS,
    STATUS_CODES,
    maxHeaderSize,
    validateHeaderName,
    validateHeaderValue,
    // WHATWG Fetch exports
    Headers,
    Request,
    Response,
    fetch,
    getStatusText
};
