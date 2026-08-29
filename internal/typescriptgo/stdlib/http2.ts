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

    goaway(code?: number, lastStreamID?: number, opaqueData?: ArrayBufferView): void {}

    ping(payload?: ArrayBufferView, callback?: (err: Error | null, duration: number, payload: Buffer) => void): boolean {
        if (callback) {
            callback(null, 1, Buffer.alloc(8));
        }
        return true;
    }

    ref(): void {}
    unref(): void {}

    setLocalWindowSize(windowSize: number): void {}
    setTimeout(msecs: number, callback?: () => void): void {}

    settings(settings: Record<string, number>, callback?: (err: Error | null, settings: Record<string, number>, duration: number) => void): void {
        if (callback) {
            callback(null, settings, 1);
        }
    }
}

export class ServerHttp2Session extends Http2Session {
    altsvc(alt: string, originOrStream: unknown): void {}
    origin(...origins: string[]): void {}
}

export class ClientHttp2Session extends Http2Session {
    request(headers?: Record<string, unknown>, options?: unknown): ClientHttp2Stream {
        return new ClientHttp2Stream();
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

    close(code?: number, callback?: () => void): void {
        this.closed = true;
        if (callback) {
            callback();
        }
        this.emit("close");
    }

    priority(options: unknown): void {}
    setTimeout(msecs: number, callback?: () => void): void {}
    sendTrailers(headers: Record<string, unknown>): void {}
}

export class ClientHttp2Stream extends Http2Stream {}

export class ServerHttp2Stream extends Http2Stream {
    headersSent: boolean = false;
    pushAllowed: boolean = true;

    additionalHeaders(headers: Record<string, unknown>): void {}
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

    setTimeout(msecs: number, callback?: () => void): this {
        this.timeout = msecs;
        return this;
    }

    updateSettings(settings: Record<string, number>): void {}

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

    setTimeout(msecs: number, callback?: () => void): this {
        this.timeout = msecs;
        return this;
    }

    updateSettings(settings: Record<string, number>): void {}

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

    addTrailers(headers: Record<string, unknown>): void {}
    appendHeader(name: string, value: unknown): this {
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
        return {};
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
        return this;
    }
    write(chunk: unknown, encoding?: unknown, callback?: unknown): boolean {
        return true;
    }
    writeContinue(): void {}
    writeEarlyHints(hints: Record<string, unknown>, callback?: () => void): void {
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

export function getPackedSettings(settings: unknown): Buffer {
    return Buffer.alloc(0);
}

export function getUnpackedSettings(buf: ArrayBufferView): Settings {
    return new Settings();
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
