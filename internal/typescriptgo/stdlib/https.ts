// ScriptGo Standard Library: node:https

import { EventEmitter } from "node:events";
import {
    TLSSocket,
    Server as TLSServer,
    TLSServerOptions,
    TLSConnectionOptions,
    connect as tlsConnect,
    createServer as tlsCreateServer,
    SecureContext
} from "node:tls";
import { URL } from "node:url";

export interface AgentOptions {
    keepAlive?: boolean;
    keepAliveMsecs?: number;
    maxSockets?: number;
    maxFreeSockets?: number;
    timeout?: number;
    rejectUnauthorized?: boolean;
    ca?: string | string[];
    cert?: string;
    key?: string;
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
    servername?: string;
}

export class Agent extends EventEmitter {
    maxSockets: number = 100;
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

class IncomingMessage extends EventEmitter {
    statusCode: number = 200;
    statusMessage: string = "OK";
    headers: Record<string, string> = {};
    rawHeaders: string[] = [];
    httpVersion: string = "1.1";
    method: string = "";
    url: string = "";
    socket: TLSSocket;
    complete: boolean = false;
    private _encoding: string = "utf8";

    constructor(socket: TLSSocket) {
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

class ServerResponse extends EventEmitter {
    statusCode: number = 200;
    statusMessage: string = "OK";
    headersSent: boolean = false;
    socket: TLSSocket;
    finished: boolean = false;
    private _headers: Record<string, string> = {};

    constructor(socket: TLSSocket) {
        super();
        this.socket = socket;
    }

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

    hasHeader(name: string): boolean {
        return this._headers[name.toLowerCase()] !== undefined;
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
        } else if (typeof statusMessageOrHeaders === "object" && statusMessageOrHeaders !== null) {
            hdrs = statusMessageOrHeaders as Record<string, string>;
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
        this.socket.write(head);
    }

    write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        if (!this.headersSent) {
            this._sendHeaders();
        }
        const str = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        this.socket.write(str);
        if (typeof callback === "function") {
            callback();
        }
        return true;
    }

    end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this.finished) return this;
        if (chunk !== undefined && chunk !== null) {
            this.write(chunk);
        } else if (!this.headersSent) {
            this._sendHeaders();
        }
        this.finished = true;
        this.emit("finish");
        if (typeof callback === "function") {
            callback();
        }
        this.socket.end();
        return this;
    }
}

class ClientRequest extends EventEmitter {
    method: string = "GET";
    path: string = "/";
    host: string = "localhost";
    port: number = 443;
    headers: Record<string, string> = {};
    socket: TLSSocket | null = null;
    aborted: boolean = false;
    reusedSocket: boolean = false;
    private _bodyChunks: string[] = [];
    private _ended: boolean = false;
    private _options: RequestOptions;
    private _res: IncomingMessage | null = null;

    constructor(options: RequestOptions, callback?: (res: IncomingMessage) => void) {
        super();
        this._options = options;
        if (options.method) this.method = options.method.toUpperCase();
        if (options.path) this.path = options.path;
        if (options.host) this.host = options.host;
        else if (options.hostname) this.host = options.hostname;

        if (typeof options.port === "number") {
            this.port = options.port;
        } else {
            this.port = 443;
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

    setHeader(name: string, value: string): void {
        this.headers[name.toLowerCase()] = String(value);
    }

    getHeader(name: string): string | undefined {
        return this.headers[name.toLowerCase()];
    }

    removeHeader(name: string): void {
        const key = name.toLowerCase();
        const nextHeaders: Record<string, string> = {};
        for (const k of Object.keys(this.headers)) {
            if (k !== key) {
                nextHeaders[k] = this.headers[k];
            }
        }
        this.headers = nextHeaders;
    }

    getHeaders(): Record<string, string> {
        return { ...this.headers };
    }

    getHeaderNames(): string[] {
        return Object.keys(this.headers);
    }

    hasHeader(name: string): boolean {
        return this.headers[name.toLowerCase()] !== undefined;
    }

    write(chunk: unknown, encoding?: string, callback?: Function): boolean {
        const str = typeof chunk === "string" ? chunk : (chunk !== null && chunk !== undefined ? String(chunk) : "");
        this._bodyChunks.push(str);
        if (typeof callback === "function") {
            callback();
        }
        return true;
    }

    end(chunk?: unknown, encoding?: string, callback?: Function): this {
        if (this._ended) return this;
        this._ended = true;
        if (chunk !== undefined && chunk !== null) {
            this.write(chunk);
        }
        if (typeof callback === "function") {
            callback();
        }

        const body = this._bodyChunks.join("");
        if (!this.hasHeader("host")) {
            this.setHeader("host", this.host);
        }
        if (!this.hasHeader("connection")) {
            this.setHeader("connection", "close");
        }
        if (body.length > 0 && !this.hasHeader("content-length")) {
            this.setHeader("content-length", String(body.length));
        }

        let reqText = `${this.method} ${this.path} HTTP/1.1\r\n`;
        for (const k of Object.keys(this.headers)) {
            reqText += `${k}: ${this.headers[k]}\r\n`;
        }
        reqText += "\r\n";
        reqText += body;

        const connOpts: TLSConnectionOptions = {
            host: this.host,
            port: this.port,
            servername: this._options.servername || this.host,
            rejectUnauthorized: this._options.rejectUnauthorized !== undefined ? !!this._options.rejectUnauthorized : true
        };

        const socket = tlsConnect(connOpts);
        this.socket = socket;
        this.emit("socket", socket);

        let buffer = "";
        let headersParsed = false;
        let res: IncomingMessage | null = null;

        socket.on("secureConnect", () => {
            socket.write(reqText);
        });

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
                    const statusCode = statusParts.length > 1 ? parseInt(statusParts[1], 10) : 200;
                    const statusMessage = statusParts.length > 2 ? statusParts.slice(2).join(" ") : "OK";

                    res = new IncomingMessage(socket);
                    this._res = res;
                    res.statusCode = statusCode;
                    res.statusMessage = statusMessage;

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

        return this;
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

export class Server extends TLSServer {
    constructor(
        options?: TLSServerOptions | ((req: IncomingMessage, res: ServerResponse) => void),
        requestListener?: (req: IncomingMessage, res: ServerResponse) => void
    ) {
        let opts: TLSServerOptions = {};
        let listener: ((req: IncomingMessage, res: ServerResponse) => void) | undefined = undefined;

        if (typeof options === "function") {
            listener = options;
            opts = {};
        } else if (options) {
            opts = options as TLSServerOptions;
            listener = requestListener;
        }

        super(opts);

        this.on("secureConnection", (socket: TLSSocket) => {
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
                        if (listener) {
                            listener(req, res);
                        }

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
        });
    }
}

function parseUrlOptions(input: string | URL | RequestOptions): RequestOptions {
    if (typeof input === "string") {
        try {
            const u = new URL(input);
            return {
                protocol: u.protocol,
                hostname: u.hostname,
                host: u.host,
                port: u.port ? parseInt(u.port, 10) : 443,
                path: (u.pathname || "/") + (u.search || "")
            };
        } catch (_) {
            return { path: input };
        }
    }
    if (input instanceof URL) {
        return {
            protocol: input.protocol,
            hostname: input.hostname,
            host: input.host,
            port: input.port ? parseInt(input.port, 10) : 443,
            path: (input.pathname || "/") + (input.search || "")
        };
    }
    return input;
}

export function request(
    urlOrOptions: string | URL | RequestOptions,
    optionsOrCallback?: RequestOptions | ((res: IncomingMessage) => void),
    callback?: (res: IncomingMessage) => void
): ClientRequest {
    let opts: RequestOptions = {};
    let cb: ((res: IncomingMessage) => void) | undefined = undefined;

    if (typeof urlOrOptions === "string" || urlOrOptions instanceof URL) {
        opts = parseUrlOptions(urlOrOptions);
        if (typeof optionsOrCallback === "function") {
            cb = optionsOrCallback;
        } else if (optionsOrCallback) {
            opts = { ...opts, ...optionsOrCallback };
            cb = callback;
        }
    } else {
        opts = { ...urlOrOptions };
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

export function createServer(
    optionsOrListener?: TLSServerOptions | ((req: IncomingMessage, res: ServerResponse) => void),
    requestListener?: (req: IncomingMessage, res: ServerResponse) => void
): Server {
    return new Server(optionsOrListener, requestListener);
}

const METHODS: string[] = [
    "ACL", "BIND", "CHECKOUT", "CONNECT", "COPY", "DELETE", "GET", "HEAD", "LINK",
    "LOCK", "M-SEARCH", "MERGE", "MKACTIVITY", "MKCALENDAR", "MKCOL", "MOVE", "NOTIFY",
    "OPTIONS", "PATCH", "POST", "PROPFIND", "PROPPATCH", "PURGE", "PUT", "REBIND",
    "REPORT", "SEARCH", "SOURCE", "SUBSCRIBE", "TRACE", "UNBIND", "UNLINK", "UNLOCK",
    "UNSUBSCRIBE"
];

const STATUS_CODES: Record<string, string> = {
    "100": "Continue",
    "101": "Switching Protocols",
    "200": "OK",
    "201": "Created",
    "202": "Accepted",
    "204": "No Content",
    "301": "Moved Permanently",
    "302": "Found",
    "304": "Not Modified",
    "400": "Bad Request",
    "401": "Unauthorized",
    "403": "Forbidden",
    "404": "Not Found",
    "405": "Method Not Allowed",
    "408": "Request Timeout",
    "500": "Internal Server Error",
    "502": "Bad Gateway",
    "503": "Service Unavailable"
};

export default {
    Agent,
    globalAgent,
    Server,
    request,
    get,
    createServer
};

