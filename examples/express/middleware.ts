import { Handler, RequestLike, ResponseLike, NextFunction } from "./types";
import { Request } from "./request";

export function json(): Handler {
    return (req: RequestLike, res: ResponseLike, next: NextFunction): void => {
        const cType = req.headers["content-type"] || "";
        if (cType.indexOf("application/json") === -1 && req.method !== "POST" && req.method !== "PUT" && req.method !== "PATCH") {
            next();
            return;
        }

        const concreteReq = req as Request;
        let bodyStr = "";

        concreteReq.raw.on("data", (chunk: unknown) => {
            const s = typeof chunk === "string" ? chunk : String(chunk);
            bodyStr += s;
        });

        concreteReq.raw.on("end", () => {
            if (bodyStr.length > 0) {
                try {
                    req.body = JSON.parse(bodyStr);
                } catch (e) {
                    req.body = null;
                }
            } else {
                req.body = {};
            }
            next();
        });
    };
}

export function urlencoded(): Handler {
    return (req: RequestLike, res: ResponseLike, next: NextFunction): void => {
        const cType = req.headers["content-type"] || "";
        if (cType.indexOf("application/x-www-form-urlencoded") === -1) {
            next();
            return;
        }

        const concreteReq = req as Request;
        let bodyStr = "";

        concreteReq.raw.on("data", (chunk: unknown) => {
            const s = typeof chunk === "string" ? chunk : String(chunk);
            bodyStr += s;
        });

        concreteReq.raw.on("end", () => {
            const parsed: Record<string, string> = {};
            if (bodyStr.length > 0) {
                const pairs = bodyStr.split("&");
                for (let i = 0; i < pairs.length; i++) {
                    const eqIdx = pairs[i].indexOf("=");
                    if (eqIdx !== -1) {
                        const k = pairs[i].slice(0, eqIdx);
                        const v = pairs[i].slice(eqIdx + 1);
                        parsed[k] = v;
                    } else {
                        parsed[pairs[i]] = "";
                    }
                }
            }
            req.body = parsed;
            next();
        });
    };
}
