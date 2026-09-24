import { EventEmitter } from "node:events";
import {
    Socket,
    Server as NetServer,
    ServerOptions as NetServerOptions,
    connect as netConnect
} from "node:net";
import { URL, URLSearchParams } from "node:url";
import { FormData } from "node:formdata";
import { Blob, File, resolveObjectURL } from "node:buffer";
import { WebSocket } from "node:ws";

export { WebSocket };

let _maxIdleHTTPParsers: number = 1000;

export function setMaxIdleHTTPParsers(max: number): void {
    if (typeof max !== "number" || max < 0 || isNaN(max)) {
        throw new TypeError('The "max" argument must be a non-negative number');
    }
    _maxIdleHTTPParsers = max;
}

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

function encodeFormDataBody(formData: FormData): { body: string; contentType: string } {
    const boundary = "----ScriptGoFormBoundary" + String(Date.now()) + String(Math.floor(Math.random() * 1000000));
    const entries = formData.entries();
    let body = "";
    for (let it = entries.next(); !it.done; it = entries.next()) {
        const entry = it.value;
        if (!entry) continue;
        const name = entry[0];
        const val = entry[1];
        body += "--" + boundary + "\r\n";
        if (val instanceof File) {
            const file = val as File;
            const filename = file.name && file.name.length > 0 ? file.name : "blob";
            const fileType = file.type && file.type.length > 0 ? file.type : "application/octet-stream";
            body += "Content-Disposition: form-data; name=\"" + name + "\"; filename=\"" + filename + "\"\r\n";
            body += "Content-Type: " + fileType + "\r\n\r\n";
            body += file._bytes.toString() + "\r\n";
        } else if (val instanceof Blob) {
            const blob = val as Blob;
            const fileType = blob.type && blob.type.length > 0 ? blob.type : "application/octet-stream";
            body += "Content-Disposition: form-data; name=\"" + name + "\"; filename=\"blob\"\r\n";
            body += "Content-Type: " + fileType + "\r\n\r\n";
            body += blob._bytes.toString() + "\r\n";
        } else {
            body += "Content-Disposition: form-data; name=\"" + name + "\"\r\n\r\n";
            body += String(val) + "\r\n";
        }
    }
    body += "--" + boundary + "--\r\n";
    return {
        body: body,
        contentType: "multipart/form-data; boundary=" + boundary
    };
}

function parseFormDataFromBody(rawBody: string, headers: Headers): FormData {
    const fd = new FormData();
    const ct = headers.get("content-type") || "";
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
            const parts = rawBody.split(delimiter);
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

    if (rawBody.length > 0) {
        const pairs = rawBody.split("&");
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

export interface RequestInit {
    method?: string;
    headers?: unknown;
    body?: unknown;
}

const defaultRequestInit: RequestInit = { method: "", headers: null, body: "" };

export class Request {
    url: string = "";
    method: string = "GET";
    headers: Headers = new Headers();
    body: string = "";

    constructor(input: unknown, init: RequestInit = defaultRequestInit) {
        if (typeof input === "string") {
            this.url = input as string;
            this.method = "GET";
            this.headers = new Headers();
            this.body = "";
        } else if (input instanceof Request) {
            const other = input as Request;
            this.url = other.url;
            this.method = other.method;
            this.headers = new Headers(other.headers);
            this.body = other.body;
        } else {
            this.url = "";
            this.method = "GET";
            this.headers = new Headers();
            this.body = "";
        }

        if (init.method !== undefined && init.method.length > 0) {
            this.method = init.method.toUpperCase();
        }
        if (init.headers !== undefined && init.headers !== null) {
            if (init.headers instanceof Headers) {
                this.headers = init.headers as Headers;
            } else {
                this.headers = new Headers(init.headers as any);
            }
        }
        if (init.body !== undefined && init.body !== null) {
            const b = init.body as any;
            if (typeof b === "string") {
                this.body = b as string;
            } else if (b instanceof FormData) {
                const enc = encodeFormDataBody(b as FormData);
                this.body = enc.body;
                if (!this.headers.has("content-type")) {
                    this.headers.set("content-type", enc.contentType);
                }
            } else if (b instanceof URLSearchParams) {
                this.body = (b as URLSearchParams).toString();
                if (!this.headers.has("content-type")) {
                    this.headers.set("content-type", "application/x-www-form-urlencoded;charset=UTF-8");
                }
            } else if (b instanceof Blob) {
                this.body = (b as Blob)._bytes.toString();
                if (!this.headers.has("content-type") && (b as Blob).type.length > 0) {
                    this.headers.set("content-type", (b as Blob).type);
                }
            } else {
                this.body = String(b);
            }
        }
    }

    async text(): Promise<string> {
        return this.body;
    }

    async json<T = unknown>(): Promise<T> {
        return JSON.parse(this.body) as T;
    }

    async arrayBuffer(): Promise<ArrayBuffer> {
        const buf = new Uint8Array(this.body.length);
        for (let i = 0; i < this.body.length; i++) {
            buf[i] = this.body.charCodeAt(i);
        }
        return buf.buffer as ArrayBuffer;
    }

    async blob(): Promise<Blob> {
        const ct = this.headers.get("content-type") || "";
        return new Blob([this.body], { type: ct });
    }

    async bytes(): Promise<Uint8Array> {
        const ab = await this.arrayBuffer();
        return new Uint8Array(ab);
    }

    async formData(): Promise<FormData> {
        return parseFormDataFromBody(this.body, this.headers);
    }

    clone(): Request {
        const init: RequestInit = {
            method: this.method,
            headers: new Headers(this.headers),
            body: this.body
        };
        return new Request(this.url, init);
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

    constructor(body: unknown = "", init: ResponseInit = defaultResponseInit) {
        this.body = null;
        let s = 200;
        let st = "OK";
        let h = new Headers();
        if (init.status !== undefined && init.status !== null) {
            s = init.status;
        }
        if (init.statusText !== undefined && init.statusText !== null) {
            st = init.statusText;
        }
        if (init.headers !== undefined && init.headers !== null) {
            if (init.headers instanceof Headers) {
                h = new Headers(init.headers as Headers);
            } else {
                h = new Headers(init.headers as any);
            }
        }

        let bodyStr = "";
        if (body !== undefined && body !== null) {
            const b = body as any;
            if (typeof b === "string") {
                bodyStr = b as string;
            } else if (b instanceof FormData) {
                const enc = encodeFormDataBody(b as FormData);
                bodyStr = enc.body;
                if (!h.has("content-type")) {
                    h.set("content-type", enc.contentType);
                }
            } else if (b instanceof URLSearchParams) {
                bodyStr = (b as URLSearchParams).toString();
                if (!h.has("content-type")) {
                    h.set("content-type", "application/x-www-form-urlencoded;charset=UTF-8");
                }
            } else if (b instanceof Blob) {
                bodyStr = (b as Blob)._bytes.toString();
                if (!h.has("content-type") && (b as Blob).type.length > 0) {
                    h.set("content-type", (b as Blob).type);
                }
            } else {
                bodyStr = String(b);
            }
        }
        this._body = bodyStr;
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

    async blob(): Promise<Blob> {
        const ct = this.headers.get("content-type") || "";
        return new Blob([this._body], { type: ct });
    }

    async bytes(): Promise<Uint8Array> {
        const ab = await this.arrayBuffer();
        return new Uint8Array(ab);
    }

    async formData(): Promise<FormData> {
        return parseFormDataFromBody(this._body, this.headers);
    }

    clone(): Response {
        const init: ResponseInit = {
            status: this.status,
            statusText: this.statusText,
            headers: new Headers(this.headers)
        };
        const cloned = new Response(this._body, init);
        cloned.url = this.url;
        return cloned;
    }

    static json(data: unknown, init: ResponseInit = defaultResponseInit): Response {
        let headers = new Headers();
        if (init.headers instanceof Headers) {
            headers = new Headers(init.headers as Headers);
        } else if (init.headers !== null && init.headers !== undefined) {
            headers = new Headers(init.headers as any);
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
        const bodyStr = typeof data === "string" ? data : JSON.stringify(data);
        return new Response(bodyStr, respInit);
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
    const req = (input instanceof Request && init === defaultRequestInit)
        ? (input as Request)
        : new Request(input, init);

    if (typeof req.url === "string" && req.url.startsWith("blob:")) {
        const blob = resolveObjectURL(req.url);
        if (!blob) {
            throw new TypeError("Failed to fetch: blob URL not found");
        }
        const text = await blob.text();
        const headers = new Headers();
        if (blob.type.length > 0) {
            headers.set("content-type", blob.type);
        }
        const resp = new Response(text, { status: 200, statusText: "OK", headers: headers });
        resp.url = req.url;
        return resp;
    }

    const flatHeaders: string[] = [];
    for (let i = 0; i < req.headers._keys.length; i++) {
        flatHeaders.push(req.headers._keys[i]);
        flatHeaders.push(req.headers._values[i]);
    }
    const raw = __scriptgo.fetchSync(req.url, req.method, flatHeaders, req.body);
    const respHeaders = new Headers();
    for (let i = 0; i < raw.headers.length; i += 2) {
        if (i + 1 < raw.headers.length) {
            respHeaders.append(raw.headers[i], raw.headers[i + 1]);
        }
    }
    const resp = new Response(raw.body, {
        status: raw.status,
        statusText: raw.statusText,
        headers: respHeaders
    });
    resp.url = req.url;
    return resp;
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
    maxTotalSockets?: number;
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
    maxTotalSockets: number = Infinity;
    keepAlive: boolean = false;
    keepAliveMsecs: number = 1000;
    freeSockets: Record<string, Socket[]> = {};
    sockets: Record<string, Socket[]> = {};
    requests: Record<string, ClientRequest[]> = {};
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
        if (this.options.maxFreeSockets !== undefined && typeof this.options.maxFreeSockets === "number") {
            this.maxFreeSockets = this.options.maxFreeSockets;
        }
        if (this.options.maxTotalSockets !== undefined && typeof this.options.maxTotalSockets === "number") {
            this.maxTotalSockets = this.options.maxTotalSockets;
        }
    }

    createConnection(options: any, callback?: (err: Error | null, stream: Socket) => void): Socket {
        const s = netConnect(options);
        if (callback) {
            s.once("connect", () => callback(null, s));
            s.once("error", (err: Error) => callback(err, s));
        }
        return s;
    }

    keepSocketAlive(socket: Socket): boolean {
        if (socket && typeof socket.setKeepAlive === "function") {
            socket.setKeepAlive(true, this.keepAliveMsecs);
            return true;
        }
        return false;
    }

    reuseSocket(socket: Socket, request: ClientRequest): void {
        if (socket && request) {
            request.socket = socket;
            request.reusedSocket = true;
        }
    }

    destroy(): void {
        for (const host of Object.keys(this.sockets)) {
            const list = this.sockets[host];
            if (Array.isArray(list)) {
                for (let i = 0; i < list.length; i++) {
                    list[i].destroy();
                }
            }
        }
        for (const host of Object.keys(this.freeSockets)) {
            const list = this.freeSockets[host];
            if (Array.isArray(list)) {
                for (let i = 0; i < list.length; i++) {
                    list[i].destroy();
                }
            }
        }
        this.sockets = {};
        this.freeSockets = {};
        this.requests = {};
        this.emit("free");
    }

    getName(options?: { host?: string; port?: number; localAddress?: string; family?: number }): string {
        const opts = options || {};
        let name = opts.host || "localhost";
        name += ":";
        if (opts.port !== undefined) name += String(opts.port);
        name += ":";
        if (opts.localAddress !== undefined) name += opts.localAddress;
        if (opts.family !== undefined) name += ":" + String(opts.family);
        return name;
    }
}

export const globalAgent = new Agent();

export class OutgoingMessage extends EventEmitter {
    headersSent: boolean = false;
    finished: boolean = false;
    socket: Socket | null = null;
    writableCorked: number = 0;
    writableEnded: boolean = false;
    writableFinished: boolean = false;
    writableHighWaterMark: number = 16384;
    writableLength: number = 0;
    readonly writableObjectMode: boolean = false;
    protected _headers: Record<string, string> = {};
    protected _rawHeaderNames: Record<string, string> = {};
    protected _trailers: Record<string, string> = {};
    protected _destroyed: boolean = false;

    get connection(): Socket | null {
        return this.socket;
    }

    setHeader(name: string, value: string | number | readonly string[]): this {
        if (this.headersSent) {
            throw new Error("Cannot set headers after they are sent to the client");
        }
        const key = name.toLowerCase();
        const valStr = Array.isArray(value) ? value.join(", ") : String(value);
        this._headers[key] = valStr;
        this._rawHeaderNames[key] = name;
        return this;
    }

    setHeaders(headers: Map<string, any> | Headers | Record<string, any>): this {
        if (headers instanceof Map) {
            (headers as Map<string, any>).forEach((v: any, k: string) => {
                this.setHeader(k, v);
            });
        } else if (headers instanceof Headers) {
            const h = headers as Headers;
            for (let i = 0; i < h._keys.length; i++) {
                this.setHeader(h._keys[i], h._values[i]);
            }
        } else if (typeof headers === "object" && headers !== null) {
            const keys = Object.keys(headers);
            for (let i = 0; i < keys.length; i++) {
                const k = keys[i];
                this.setHeader(k, (headers as any)[k]);
            }
        }
        return this;
    }

    appendHeader(name: string, value: string | readonly string[]): this {
        if (this.headersSent) {
            throw new Error("Cannot append header after they are sent");
        }
        const key = name.toLowerCase();
        const valStr = Array.isArray(value) ? value.join(", ") : String(value);
        if (this._headers[key] !== undefined) {
            this._headers[key] = this._headers[key] + ", " + valStr;
        } else {
            this._headers[key] = valStr;
            this._rawHeaderNames[key] = name;
        }
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
        const nextRaw: Record<string, string> = {};
        for (const k of Object.keys(this._rawHeaderNames)) {
            if (k !== key) {
                nextRaw[k] = this._rawHeaderNames[k];
            }
        }
        this._rawHeaderNames = nextRaw;
        return this;
    }

    getHeaderNames(): string[] {
        return Object.keys(this._headers);
    }

    getRawHeaderNames(): string[] {
        const res: string[] = [];
        for (const k of Object.keys(this._headers)) {
            res.push(this._rawHeaderNames[k] || k);
        }
        return res;
    }

    getHeaders(): Record<string, string> {
        const copy: Record<string, string> = {};
        for (const k of Object.keys(this._headers)) {
            copy[k] = this._headers[k];
        }
        return copy;
    }

    addTrailers(headers: Record<string, string>): void {
        for (const k of Object.keys(headers)) {
            this._trailers[k.toLowerCase()] = String(headers[k]);
        }
    }

    flushHeaders(): void {}

    cork(): void {
        this.writableCorked = this.writableCorked + 1;
        if (this.socket && typeof (this.socket as any).cork === "function") {
            (this.socket as any).cork();
        }
    }

    uncork(): void {
        if (this.writableCorked > 0) {
            this.writableCorked = this.writableCorked - 1;
        }
        if (this.socket && typeof (this.socket as any).uncork === "function") {
            (this.socket as any).uncork();
        }
    }

    pipe(destination?: unknown, options?: unknown): never {
        throw new Error("ERR_STREAM_CANNOT_PIPE: Cannot pipe, not readable");
    }

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.once("timeout", callback);
        }
        if (this.socket && typeof this.socket.setTimeout === "function") {
            this.socket.setTimeout(msecs);
        }
        return this;
    }

    destroy(error?: Error): this {
        if (this._destroyed) return this;
        this._destroyed = true;
        if (this.socket) {
            this.socket.destroy(error);
        }
        if (error) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }

    write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        const data = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        this.writableLength = this.writableLength + data.length;
        if (this.socket) {
            this.socket.write(data);
        }
        if (callback) {
            callback();
        }
        return true;
    }

    end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this.writableEnded) return this;
        this.writableEnded = true;
        if (chunk !== undefined && chunk !== null) {
            this.write(chunk, encoding);
        }
        this.finished = true;
        this.writableFinished = true;
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

export class ServerResponse extends OutgoingMessage {
    statusCode: number = 200;
    statusMessage: string = "OK";
    sendDate: boolean = true;
    strictContentLength: boolean = false;
    req: IncomingMessage | null = null;

    constructor(reqOrSocket?: IncomingMessage | Socket) {
        super();
        if (reqOrSocket instanceof Socket) {
            this.socket = reqOrSocket;
        } else if (reqOrSocket) {
            this.req = reqOrSocket as IncomingMessage;
            if (this.req.socket) {
                this.socket = this.req.socket;
            }
        }
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

    writeContinue(): void {
        if (this.socket) {
            this.socket.write("HTTP/1.1 100 Continue\r\n\r\n");
        }
    }

    writeProcessing(): void {
        if (this.socket) {
            this.socket.write("HTTP/1.1 102 Processing\r\n\r\n");
        }
    }

    writeEarlyHints(hints: Record<string, any>, callback?: () => void): void {
        if (this.socket) {
            let msg = "HTTP/1.1 103 Early Hints\r\n";
            for (const k of Object.keys(hints)) {
                const val: any = (hints as any)[k];
                if (Array.isArray(val)) {
                    for (let i = 0; i < val.length; i++) {
                        msg += `${k}: ${val[i]}\r\n`;
                    }
                } else if (val !== null && val !== undefined) {
                    msg += `${k}: ${String(val)}\r\n`;
                }
            }
            msg += "\r\n";
            this.socket.write(msg);
        }
        if (callback) {
            callback();
        }
    }

    private _sendHeaders(): void {
        if (this.headersSent) return;
        this.headersSent = true;
        let head = `HTTP/1.1 ${this.statusCode} ${this.statusMessage}\r\n`;
        if (this.sendDate && !this.hasHeader("date")) {
            head += `date: ${new Date().toUTCString()}\r\n`;
        }
        for (const key of Object.keys(this._headers)) {
            head += `${key}: ${this._headers[key]}\r\n`;
        }
        head += "\r\n";
        if (this.socket) {
            this.socket.write(head);
        }
    }

    override write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        if (!this.headersSent) {
            this._sendHeaders();
        }
        return super.write(chunk, encoding, callback);
    }

    override end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this.finished) return this;
        if (!this.headersSent) {
            this._sendHeaders();
        }
        return super.end(chunk, encoding, callback);
    }
}

export class IncomingMessage extends EventEmitter {
    statusCode: number = 200;
    statusMessage: string = "OK";
    headers: Record<string, string> = {};
    headersDistinct: Record<string, string[]> = {};
    rawHeaders: string[] = [];
    trailers: Record<string, string> = {};
    trailersDistinct: Record<string, string[]> = {};
    rawTrailers: string[] = [];
    httpVersion: string = "1.1";
    method: string = "";
    url: string = "";
    socket: Socket;
    complete: boolean = false;
    aborted: boolean = false;
    private _encoding: string = "utf8";

    constructor(socket: Socket) {
        super();
        this.socket = socket;
    }

    get connection(): Socket {
        return this.socket;
    }

    setEncoding(encoding: string): this {
        this._encoding = encoding;
        return this;
    }

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.once("timeout", callback);
        }
        if (this.socket && typeof this.socket.setTimeout === "function") {
            this.socket.setTimeout(msecs);
        }
        return this;
    }

    destroy(error?: Error): this {
        this.complete = true;
        if (this.socket) {
            this.socket.destroy(error);
        }
        if (error) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }
}

export class Server extends NetServer {
    headersTimeout: number = 60000;
    requestTimeout: number = 300000;
    maxHeadersCount: number | null = null;
    maxRequestsPerSocket: number = 0;
    timeout: number = 0;
    keepAliveTimeout: number = 5000;
    keepAliveTimeoutBuffer: number = 1000;
    private _connections: Socket[] = [];

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
            this._connections.push(socket);
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
                                const rawName = lines[i].slice(0, colonIdx).trim();
                                const hName = rawName.toLowerCase();
                                const hVal = lines[i].slice(colonIdx + 1).trim();
                                req.headers[hName] = hVal;
                                req.rawHeaders.push(rawName);
                                req.rawHeaders.push(hVal);
                                if (req.headersDistinct[hName] === undefined) {
                                    req.headersDistinct[hName] = [];
                                }
                                req.headersDistinct[hName].push(hVal);
                            }
                        }

                        res = new ServerResponse(req);
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
                const idx = this._connections.indexOf(socket);
                if (idx !== -1) {
                    this._connections.splice(idx, 1);
                }
                if (res) {
                    res.emit("close");
                }
            });
        });
    }

    closeAllConnections(): void {
        for (let i = 0; i < this._connections.length; i++) {
            this._connections[i].destroy();
        }
        this._connections = [];
    }

    closeIdleConnections(): void {
        for (let i = 0; i < this._connections.length; i++) {
            this._connections[i].destroy();
        }
        this._connections = [];
    }

    setTimeout(msecs?: number, callback?: () => void): this {
        this.timeout = msecs || 0;
        if (callback) {
            this.on("timeout", callback);
        }
        return this;
    }

    [Symbol.asyncDispose](): Promise<void> {
        return new Promise<void>((resolve, reject) => {
            this.close((err) => {
                if (err) reject(err);
                else resolve();
            });
        });
    }
}

export class ClientRequest extends OutgoingMessage {
    method: string = "GET";
    path: string = "/";
    host: string = "localhost";
    port: number = 80;
    protocol: string = "http:";
    aborted: boolean = false;
    reusedSocket: boolean = false;
    maxHeadersCount: number | null = null;
    private _bodyChunks: string[] = [];
    private _ended: boolean = false;
    private _options: RequestOptions;
    private _res: IncomingMessage | null = null;

    constructor(options?: RequestOptions | string | URL, callback?: (res: IncomingMessage) => void) {
        super();
        let opts: RequestOptions = {};
        if (typeof options === "string") {
            const parsed = new URL(options);
            opts.host = parsed.hostname;
            opts.port = parsed.port ? parseInt(parsed.port) : 80;
            opts.path = parsed.pathname + parsed.search;
            opts.protocol = parsed.protocol;
        } else if (options instanceof URL) {
            opts.host = options.hostname;
            opts.port = options.port ? parseInt(options.port) : 80;
            opts.path = options.pathname + options.search;
            opts.protocol = options.protocol;
        } else if (options) {
            opts = options;
        }
        this._options = opts;
        if (opts.protocol) this.protocol = opts.protocol;
        if (opts.method) {
            const m = opts.method.toUpperCase();
            if (METHODS.indexOf(m) !== -1) {
                this.method = m;
            } else {
                this.method = opts.method;
            }
        }
        if (opts.path) this.path = opts.path;
        if (opts.host) this.host = opts.host;
        else if (opts.hostname) this.host = opts.hostname;

        if (typeof opts.port === "number") {
            this.port = opts.port;
        } else {
            this.port = 80;
        }

        if (opts.headers) {
            for (const k of Object.keys(opts.headers)) {
                this.setHeader(k, opts.headers[k]);
            }
        }

        if (callback) {
            this.once("response", callback);
        }
    }

    setNoDelay(noDelay: boolean = true): void {
        if (this.socket && typeof (this.socket as any).setNoDelay === "function") {
            (this.socket as any).setNoDelay(noDelay);
        }
    }

    setSocketKeepAlive(enable: boolean = false, initialDelay: number = 0): void {
        if (this.socket && typeof (this.socket as any).setKeepAlive === "function") {
            (this.socket as any).setKeepAlive(enable, initialDelay);
        }
    }

    abort(): void {
        this.aborted = true;
        this.destroy(new Error("Request aborted"));
    }

    override write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        if (this._ended) {
            throw new Error("write after end");
        }
        const str = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        this._bodyChunks.push(str);
        this.writableLength = this.writableLength + str.length;
        if (this.socket) {
            this.socket.write(str);
        }
        if (callback) {
            callback();
        }
        return true;
    }

    override end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this._ended) return this;
        this._ended = true;
        this.writableEnded = true;

        if (chunk !== undefined && chunk !== null) {
            const str = typeof chunk === "string" ? chunk : String(chunk);
            this._bodyChunks.push(str);
            this.writableLength = this.writableLength + str.length;
        }

        if (callback) {
            callback();
        }

        this._connectAndSend();
        return this;
    }

    private _connectAndSend(): void {
        const socket = new Socket();
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
                            const rawName = lines[i].slice(0, colonIdx).trim();
                            const hName = rawName.toLowerCase();
                            const hVal = lines[i].slice(colonIdx + 1).trim();
                            res.headers[hName] = hVal;
                            res.rawHeaders.push(rawName);
                            res.rawHeaders.push(hVal);
                            if (res.headersDistinct[hName] === undefined) {
                                res.headersDistinct[hName] = [];
                            }
                            res.headersDistinct[hName].push(hVal);
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
            this.finished = true;
            this.writableFinished = true;
            if (res) {
                res.complete = true;
                res.emit("end");
            }
            this.emit("finish");
            this.emit("close");
        });

        socket.on("error", (err: Error) => {
            this.emit("error", err);
        });

        socket.on("close", () => {
            this.finished = true;
            this.writableFinished = true;
            if (res) {
                res.emit("close");
            }
        });

        socket.connect({ host: this.host, port: this.port }, () => {
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
    }

    override destroy(error?: Error): this {
        if (this.socket) {
            this.socket.destroy(error);
        }
        if (error) {
            this.emit("error", error);
        }
        this.emit("close");
        return this;
    }

    override setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.once("timeout", callback);
        }
        if (this.socket && typeof this.socket.setTimeout === "function") {
            this.socket.setTimeout(msecs);
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
    setMaxIdleHTTPParsers,
    WebSocket,
    // WHATWG Fetch exports
    Headers,
    Request,
    Response,
    fetch,
    getStatusText
};
