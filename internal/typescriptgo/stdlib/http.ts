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

class WebSocketListenerEntry {
    type: string;
    callback: Function;
    once: boolean;

    constructor(type: string, callback: Function, once: boolean) {
        this.type = type;
        this.callback = callback;
        this.once = once;
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


export default {
    METHODS,
    STATUS_CODES,
    getStatusText,
    maxHeaderSize,
    validateHeaderName,
    validateHeaderValue,
    Headers,
    Request,
    Response,
    fetch,
};
