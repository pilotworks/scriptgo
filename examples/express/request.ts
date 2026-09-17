import { IncomingMessage } from "node:http";
import { RequestLike } from "./types";

export class Request implements RequestLike {
    raw: IncomingMessage;
    method: string;
    url: string;
    path: string;
    params: Record<string, string>;
    query: Record<string, string>;
    headers: Record<string, string>;
    body: unknown;

    constructor(raw: IncomingMessage) {
        this.raw = raw;
        this.method = raw.method ? raw.method.toUpperCase() : "GET";
        this.url = raw.url || "/";
        this.params = {};
        this.query = {};
        this.headers = raw.headers || {};
        this.body = null;

        // Parse path and query
        const qIdx = this.url.indexOf("?");
        if (qIdx !== -1) {
            this.path = this.url.slice(0, qIdx);
            const qStr = this.url.slice(qIdx + 1);
            if (qStr.length > 0) {
                const pairs = qStr.split("&");
                for (let i = 0; i < pairs.length; i++) {
                    const eqIdx = pairs[i].indexOf("=");
                    if (eqIdx !== -1) {
                        const k = pairs[i].slice(0, eqIdx);
                        const v = pairs[i].slice(eqIdx + 1);
                        this.query[k] = v;
                    } else {
                        this.query[pairs[i]] = "";
                    }
                }
            }
        } else {
            this.path = this.url;
        }
    }

    get(name: string): string | undefined {
        const lower = name.toLowerCase();
        return this.headers[lower];
    }
}
