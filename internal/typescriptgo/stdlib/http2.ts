// Node.js HTTP/2 module (node:http2)
import { EventEmitter } from "node:events";
import { Readable, Writable, Duplex } from "node:stream";

export const constants: Record<string, number> = {
    NGHTTP2_SESSION_SERVER: 0,
    NGHTTP2_SESSION_CLIENT: 1,
    NGHTTP2_STREAM_STATE_IDLE: 1,
    NGHTTP2_STREAM_STATE_OPEN: 2,
    NGHTTP2_STREAM_STATE_RESERVED_LOCAL: 3,
    NGHTTP2_STREAM_STATE_RESERVED_REMOTE: 4,
    NGHTTP2_STREAM_STATE_HALF_CLOSED_LOCAL: 5,
    NGHTTP2_STREAM_STATE_HALF_CLOSED_REMOTE: 6,
    NGHTTP2_STREAM_STATE_CLOSED: 7,
    NGHTTP2_NO_ERROR: 0,
    NGHTTP2_PROTOCOL_ERROR: 1,
    NGHTTP2_INTERNAL_ERROR: 2,
    NGHTTP2_FLOW_CONTROL_ERROR: 3,
    NGHTTP2_SETTINGS_TIMEOUT: 4,
    NGHTTP2_STREAM_CLOSED: 5,
    NGHTTP2_FRAME_SIZE_ERROR: 6,
    NGHTTP2_REFUSED_STREAM: 7,
    NGHTTP2_CANCEL: 8,
    NGHTTP2_COMPRESSION_ERROR: 9,
    NGHTTP2_CONNECT_ERROR: 10,
    NGHTTP2_ENHANCE_YOUR_CALM: 11,
    NGHTTP2_INADEQUATE_SECURITY: 12,
    NGHTTP2_HTTP_1_1_REQUIRED: 13,
    HTTP2_HEADER_STATUS: 0,
    HTTP2_HEADER_METHOD: 1,
    HTTP2_HEADER_AUTHORITY: 2,
    HTTP2_HEADER_SCHEME: 3,
    HTTP2_HEADER_PATH: 4,
    HTTP_STATUS_OK: 200,
};

export const sensitiveHeaders: symbol = Symbol("sensitiveHeaders");

export class Http2Session extends EventEmitter {
    alpnProtocol: string = "h2";
    closed: boolean = false;
    connecting: boolean = false;
    destroyed: boolean = false;
    encrypted: boolean = true;
    localSettings: Record<string, number> = {};
    originSet: string[] = [];
    pendingSettingsAck: boolean = false;
    remoteSettings: Record<string, number> = {};
    socket: Duplex = new Duplex();
    state: Record<string, number> = {};
    type: number = 0;

    _localWindowSize: number = 65535;
    _timeoutMs: number = 0;

    close(callback?: () => void): void {
        this.closed = true;
        if (callback) {
            callback();
        }
        this.emit("close");
    }

    destroy(error?: Error, code?: number): this {
        this.destroyed = true;
        this.emit("close");
        return this;
    }

    goaway(code?: number, lastStreamID?: number, opaqueData?: ArrayBufferView): void {
        this.emit("goaway", code || 0, lastStreamID || 0, opaqueData);
    }

    ping(payload?: ArrayBufferView, callback?: (err: Error | null, duration: number, payload: Buffer) => void): boolean {
        if (callback) {
            callback(null, 1, Buffer.alloc(8));
        }
        return true;
    }

    ref(): void {}
    unref(): void {}

    setLocalWindowSize(windowSize: number): void {
        this._localWindowSize = windowSize;
    }

    setTimeout(msecs: number, callback?: () => void): void {
        this._timeoutMs = msecs;
        if (callback) {
            this.on("timeout", callback);
        }
    }

    settings(settings: Record<string, number>, callback?: (err: Error | null, settings: Record<string, number>, duration: number) => void): void {
        for (const k in settings) {
            this.localSettings[k] = settings[k];
        }
        if (callback) {
            callback(null, settings, 1);
        }
    }
}

export class ServerHttp2Session extends Http2Session {
    _altsvcList: { alt: string; originOrStream: unknown }[] = [];

    altsvc(alt: string, originOrStream: unknown): void {
        this._altsvcList.push({ alt, originOrStream });
        this.emit("altsvc", alt, originOrStream);
    }

    origin(...origins: string[]): void {
        for (let i = 0; i < origins.length; i++) {
            this.originSet.push(origins[i]);
        }
        this.emit("origin", origins);
    }
}

export class ClientHttp2Session extends Http2Session {
    request(headers?: Record<string, unknown>, options?: unknown): ClientHttp2Stream {
        const stream = new ClientHttp2Stream();
        if (headers) {
            stream.sentHeaders = headers;
        }
        return stream;
    }
}

export class Http2Stream extends Duplex {
    aborted: boolean = false;
    bufferSize: number = 0;
    closed: boolean = false;
    destroyed: boolean = false;
    endAfterHeaders: boolean = false;
    id: number = 1;
    pending: boolean = false;
    rstCode: number = 0;
    sentHeaders: Record<string, unknown> = {};
    sentInfoHeaders: Record<string, unknown>[] = [];
    sentTrailers: Record<string, unknown> = {};
    session: Http2Session = new Http2Session();
    state: Record<string, number> = {};
    _priorityOptions: unknown = null;

    close(code?: number, callback?: () => void): void {
        this.closed = true;
        if (callback) {
            callback();
        }
        this.emit("close");
    }

    priority(options: unknown): void {
        this._priorityOptions = options;
        this.emit("priority", options);
    }

    setTimeout(msecs: number, callback?: () => void): void {
        if (callback) {
            this.on("timeout", callback);
        }
    }

    sendTrailers(headers: Record<string, unknown>): void {
        this.sentTrailers = headers;
        this.emit("trailers", headers);
    }
}

export class ClientHttp2Stream extends Http2Stream {}

export class ServerHttp2Stream extends Http2Stream {
    headersSent: boolean = false;
    pushAllowed: boolean = true;

    additionalHeaders(headers: Record<string, unknown>): void {
        this.sentInfoHeaders.push(headers);
        this.emit("headers", headers);
    }

    pushStream(headers: Record<string, unknown>, options?: unknown, callback?: (err: Error | null, pushStream: ServerHttp2Stream, headers: Record<string, unknown>) => void): void {
        if (callback) {
            callback(null, new ServerHttp2Stream(), headers);
        }
    }

    respond(headers?: Record<string, unknown>, options?: unknown): void {
        this.headersSent = true;
    }

    respondWithFD(fd: number, headers?: Record<string, unknown>, options?: unknown): void {
        this.headersSent = true;
    }

    respondWithFile(path: string, headers?: Record<string, unknown>, options?: unknown): void {
        this.headersSent = true;
    }
}

export class Http2Server extends EventEmitter {
    timeout: number = 0;
    localSettings: Record<string, number> = {};

    setTimeout(msecs: number, callback?: () => void): this {
        this.timeout = msecs;
        return this;
    }

    updateSettings(settings: Record<string, number>): void {
        for (const k in settings) {
            this.localSettings[k] = settings[k];
        }
    }

    close(callback?: (err?: Error) => void): this {
        if (callback) {
            callback();
        }
        this.emit("close");
        return this;
    }

    async [Symbol.asyncDispose](): Promise<void> {
        this.close();
    }
}

export class Http2SecureServer extends EventEmitter {
    timeout: number = 0;
    localSettings: Record<string, number> = {};

    setTimeout(msecs: number, callback?: () => void): this {
        this.timeout = msecs;
        return this;
    }

    updateSettings(settings: Record<string, number>): void {
        for (const k in settings) {
            this.localSettings[k] = settings[k];
        }
    }

    close(callback?: (err?: Error) => void): this {
        if (callback) {
            callback();
        }
        this.emit("close");
        return this;
    }
}

export class Http2ServerRequest extends Readable {
    aborted: boolean = false;
    authority: string = "";
    complete: boolean = true;
    connection: Duplex = new Duplex();
    headers: Record<string, string> = {};
    httpVersion: string = "2.0";
    method: string = "GET";
    rawHeaders: string[] = [];
    rawTrailers: string[] = [];
    scheme: string = "https";
    socket: Duplex = new Duplex();
    stream: ServerHttp2Stream = new ServerHttp2Stream();
    trailers: Record<string, string> = {};
    url: string = "/";

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.on("timeout", callback);
        }
        return this;
    }

    destroy(error?: Error): this {
        this.emit("close");
        return this;
    }
}

export class Http2ServerResponse extends Writable {
    connection: Duplex = new Duplex();
    finished: boolean = false;
    headersSent: boolean = false;
    req: Http2ServerRequest = new Http2ServerRequest();
    sendDate: boolean = true;
    socket: Duplex = new Duplex();
    statusCode: number = 200;
    statusMessage: string = "OK";
    stream: ServerHttp2Stream = new ServerHttp2Stream();
    writableEnded: boolean = false;
    private _headers: Map<string, unknown> = new Map();

    addTrailers(headers: Record<string, unknown>): void {
        for (const k in headers) {
            this._headers.set(k, headers[k]);
        }
    }

    appendHeader(name: string, value: unknown): this {
        const prev = this._headers.get(name);
        if (prev) {
            this._headers.set(name, Array.isArray(prev) ? [...(prev as unknown[]), value] : [prev, value]);
        } else {
            this._headers.set(name, value);
        }
        return this;
    }

    createPushResponse(headers: Record<string, unknown>, callback: (err: Error | null, response: Http2ServerResponse) => void): void {
        callback(null, new Http2ServerResponse());
    }

    end(data?: unknown, encoding?: unknown, callback?: unknown): this {
        this.finished = true;
        this.writableEnded = true;
        return this;
    }

    getHeader(name: string): unknown {
        return this._headers.get(name);
    }

    getHeaderNames(): string[] {
        return Array.from(this._headers.keys());
    }

    getHeaders(): Record<string, unknown> {
        return Object.fromEntries(this._headers.entries());
    }

    hasHeader(name: string): boolean {
        return this._headers.has(name);
    }

    removeHeader(name: string): void {
        this._headers.delete(name);
    }

    setHeader(name: string, value: unknown): this {
        this._headers.set(name, value);
        return this;
    }

    setTimeout(msecs: number, callback?: () => void): this {
        if (callback) {
            this.on("timeout", callback);
        }
        return this;
    }

    write(chunk: unknown, encoding?: unknown, callback?: unknown): boolean {
        return true;
    }

    writeContinue(): void {
        this.emit("continue");
    }

    writeEarlyHints(hints: Record<string, unknown>, callback?: () => void): void {
        this.emit("earlyHints", hints);
        if (callback) {
            callback();
        }
    }

    writeHead(statusCode: number, statusMessage?: unknown, headers?: unknown): this {
        this.statusCode = statusCode;
        this.headersSent = true;
        return this;
    }
}

export function createServer(options?: unknown, onRequestHandler?: unknown): Http2Server {
    return new Http2Server();
}

export function createSecureServer(options?: unknown, onRequestHandler?: unknown): Http2SecureServer {
    return new Http2SecureServer();
}

export function connect(authority: string | URL, options?: unknown, listener?: unknown): ClientHttp2Session {
    return new ClientHttp2Session();
}

export class Settings {
    headerTableSize: number = 4096;
    enablePush: number = 1;
    initialWindowSize: number = 65535;
    maxFrameSize: number = 16384;
    maxConcurrentStreams: number = 100;
    maxHeaderListSize: number = 65535;
    enableConnectProtocol: number = 0;
}

export function getDefaultSettings(): Settings {
    return new Settings();
}

export function getPackedSettings(settings?: unknown): Uint8Array {
    if (!settings || typeof settings !== "object") {
        return new Uint8Array(0);
    }
    const s = settings as Settings;
    const entries: { id: number; val: number }[] = [];
    if (s.headerTableSize !== undefined) entries.push({ id: 1, val: s.headerTableSize });
    if (s.enablePush !== undefined) entries.push({ id: 2, val: s.enablePush ? 1 : 0 });
    if (s.maxConcurrentStreams !== undefined) entries.push({ id: 3, val: s.maxConcurrentStreams });
    if (s.initialWindowSize !== undefined) entries.push({ id: 4, val: s.initialWindowSize });
    if (s.maxFrameSize !== undefined) entries.push({ id: 5, val: s.maxFrameSize });
    if (s.maxHeaderListSize !== undefined) entries.push({ id: 6, val: s.maxHeaderListSize });
    if (s.enableConnectProtocol !== undefined) entries.push({ id: 8, val: s.enableConnectProtocol ? 1 : 0 });

    const buf = new Uint8Array(entries.length * 6);
    for (let i = 0; i < entries.length; i++) {
        const id = entries[i].id;
        const val = entries[i].val;
        const offset = i * 6;
        buf[offset] = (id >> 8) & 0xff;
        buf[offset + 1] = id & 0xff;
        buf[offset + 2] = (val >>> 24) & 0xff;
        buf[offset + 3] = (val >>> 16) & 0xff;
        buf[offset + 4] = (val >>> 8) & 0xff;
        buf[offset + 5] = val & 0xff;
    }
    return buf;
}

export function getUnpackedSettings(buf?: unknown): Settings {
    const s = new Settings();
    if (!buf || !(buf instanceof Uint8Array)) return s;
    const view = buf as Uint8Array;
    const count = Math.floor(view.length / 6);
    for (let i = 0; i < count; i++) {
        const offset = i * 6;
        const id = (view[offset] << 8) | view[offset + 1];
        const val = ((view[offset + 2] << 24) >>> 0) | (view[offset + 3] << 16) | (view[offset + 4] << 8) | view[offset + 5];
        if (id === 1) s.headerTableSize = val;
        else if (id === 2) s.enablePush = val ? 1 : 0;
        else if (id === 3) s.maxConcurrentStreams = val;
        else if (id === 4) s.initialWindowSize = val;
        else if (id === 5) s.maxFrameSize = val;
        else if (id === 6) s.maxHeaderListSize = val;
        else if (id === 8) s.enableConnectProtocol = val ? 1 : 0;
    }
    return s;
}

export function performServerHandshake(socket: Duplex, options?: unknown): ServerHttp2Session {
    return new ServerHttp2Session();
}

export default {
    constants,
    sensitiveHeaders,
    Http2Session,
    ServerHttp2Session,
    ClientHttp2Session,
    Http2Stream,
    ClientHttp2Stream,
    ServerHttp2Stream,
    Http2Server,
    Http2SecureServer,
    Http2ServerRequest,
    Http2ServerResponse,
    createServer,
    createSecureServer,
    connect,
    getDefaultSettings,
    getPackedSettings,
    getUnpackedSettings,
    performServerHandshake,
};
