import { ServerResponse } from "node:http";
import { ResponseLike } from "./types";

export class Response implements ResponseLike {
    raw: ServerResponse;
    statusCode: number = 200;
    private _headersSent: boolean = false;

    constructor(raw: ServerResponse) {
        this.raw = raw;
    }

    status(code: number): this {
        this.statusCode = code;
        this.raw.statusCode = code;
        return this;
    }

    setHeader(name: string, value: string): this {
        this.raw.setHeader(name, value);
        return this;
    }

    getHeader(name: string): string | undefined {
        return this.raw.getHeader(name);
    }

    json(data: unknown): void {
        const payload = JSON.stringify(data);
        if (!this.raw.hasHeader("content-type")) {
            this.raw.setHeader("content-type", "application/json; charset=utf-8");
        }
        if (!this.raw.hasHeader("content-length")) {
            this.raw.setHeader("content-length", String(payload.length));
        }
        this.raw.writeHead(this.statusCode);
        this.raw.end(payload);
        this._headersSent = true;
    }

    send(body: string): void {
        if (!this.raw.hasHeader("content-type")) {
            this.raw.setHeader("content-type", "text/html; charset=utf-8");
        }
        if (!this.raw.hasHeader("content-length")) {
            this.raw.setHeader("content-length", String(body.length));
        }
        this.raw.writeHead(this.statusCode);
        this.raw.end(body);
        this._headersSent = true;
    }

    redirect(url: string, status: number = 302): void {
        this.raw.setHeader("location", url);
        this.raw.writeHead(status);
        this.raw.end();
        this._headersSent = true;
    }

    end(): void {
        if (!this._headersSent) {
            this.raw.writeHead(this.statusCode);
        }
        this.raw.end();
        this._headersSent = true;
    }
}
