import { Handler, RouteEntry } from "./types";

export function splitPath(path: string): string[] {
    const parts = path.split("/");
    const res: string[] = [];
    for (let i = 0; i < parts.length; i++) {
        const seg = parts[i].trim();
        if (seg.length > 0) {
            res.push(seg);
        }
    }
    return res;
}

export interface MatchResult {
    handler: Handler;
    params: Record<string, string>;
}

export class Router {
    routes: RouteEntry[] = [];
    middlewares: { prefix: string[]; handler: Handler }[] = [];

    use(handlerOrPath: string | Handler, maybeHandler?: Handler): this {
        if (typeof handlerOrPath === "string" && maybeHandler) {
            this.middlewares.push({
                prefix: splitPath(handlerOrPath),
                handler: maybeHandler
            });
        } else if (typeof handlerOrPath === "function") {
            this.middlewares.push({
                prefix: [],
                handler: handlerOrPath
            });
        }
        return this;
    }

    get(path: string, handler: Handler): this {
        return this.addRoute("GET", path, handler);
    }

    post(path: string, handler: Handler): this {
        return this.addRoute("POST", path, handler);
    }

    put(path: string, handler: Handler): this {
        return this.addRoute("PUT", path, handler);
    }

    delete(path: string, handler: Handler): this {
        return this.addRoute("DELETE", path, handler);
    }

    patch(path: string, handler: Handler): this {
        return this.addRoute("PATCH", path, handler);
    }

    all(path: string, handler: Handler): this {
        return this.addRoute("ALL", path, handler);
    }

    addRoute(method: string, path: string, handler: Handler): this {
        const segments = splitPath(path);
        this.routes.push({
            method: method.toUpperCase(),
            path: path,
            segments: segments,
            handler: handler
        });
        return this;
    }

    match(method: string, pathname: string): MatchResult | null {
        const upperMethod = method.toUpperCase();
        const reqSegments = splitPath(pathname);

        for (let i = 0; i < this.routes.length; i++) {
            const r = this.routes[i];
            if (r.method !== "ALL" && r.method !== upperMethod) {
                continue;
            }

            const isWildcardTail = r.segments.length > 0 && r.segments[r.segments.length - 1] === "*";
            if (!isWildcardTail && r.segments.length !== reqSegments.length) {
                continue;
            }
            if (isWildcardTail && reqSegments.length < r.segments.length - 1) {
                continue;
            }

            const params: Record<string, string> = {};
            let matched = true;
            const checkLen = isWildcardTail ? r.segments.length - 1 : r.segments.length;

            for (let j = 0; j < checkLen; j++) {
                const rSeg = r.segments[j];
                const qSeg = reqSegments[j];

                if (rSeg.length > 0 && rSeg.charCodeAt(0) === 58) { // ':'
                    const paramName = rSeg.slice(1);
                    params[paramName] = qSeg;
                } else if (rSeg === "*") {
                    // Match single segment wildcard
                } else if (rSeg !== qSeg) {
                    matched = false;
                    break;
                }
            }

            if (matched) {
                if (isWildcardTail) {
                    const tail = reqSegments.slice(checkLen).join("/");
                    params["*"] = tail;
                }
                return {
                    handler: r.handler,
                    params: params
                };
            }
        }

        return null;
    }

    getMatchedMiddlewares(pathname: string): Handler[] {
        const reqSegments = splitPath(pathname);
        const result: Handler[] = [];

        for (let i = 0; i < this.middlewares.length; i++) {
            const mw = this.middlewares[i];
            if (mw.prefix.length === 0) {
                result.push(mw.handler);
                continue;
            }

            if (reqSegments.length >= mw.prefix.length) {
                let match = true;
                for (let j = 0; j < mw.prefix.length; j++) {
                    if (mw.prefix[j] !== reqSegments[j]) {
                        match = false;
                        break;
                    }
                }
                if (match) {
                    result.push(mw.handler);
                }
            }
        }

        return result;
    }
}
