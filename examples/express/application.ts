import { createServer, Server, IncomingMessage, ServerResponse } from "node:http";
import { Handler, RequestLike, ResponseLike, NextFunction } from "./types";
import { Request } from "./request";
import { Response } from "./response";
import { Router } from "./router";

export class Application {
    private _router: Router = new Router();

    use(handlerOrPath: string | Handler, maybeHandler?: Handler): this {
        this._router.use(handlerOrPath, maybeHandler);
        return this;
    }

    get(path: string, handler: Handler): this {
        this._router.get(path, handler);
        return this;
    }

    post(path: string, handler: Handler): this {
        this._router.post(path, handler);
        return this;
    }

    put(path: string, handler: Handler): this {
        this._router.put(path, handler);
        return this;
    }

    delete(path: string, handler: Handler): this {
        this._router.delete(path, handler);
        return this;
    }

    patch(path: string, handler: Handler): this {
        this._router.patch(path, handler);
        return this;
    }

    all(path: string, handler: Handler): this {
        this._router.all(path, handler);
        return this;
    }

    handle(rawReq: IncomingMessage, rawRes: ServerResponse): void {
        const req = new Request(rawReq);
        const res = new Response(rawRes);

        const mws = this._router.getMatchedMiddlewares(req.path);
        const match = this._router.match(req.method, req.path);

        let idx = 0;

        const next: NextFunction = (err?: Error | null): void => {
            if (err) {
                res.status(500).json({ error: err.message || "Internal Server Error" });
                return;
            }

            if (idx < mws.length) {
                const currentMw = mws[idx];
                idx++;
                currentMw(req, res, next);
                return;
            }

            if (match) {
                req.params = match.params;
                match.handler(req, res, (routeErr?: Error | null) => {
                    if (routeErr) {
                        res.status(500).json({ error: routeErr.message || "Internal Server Error" });
                    }
                });
                return;
            }

            res.status(404).send(`Cannot ${req.method} ${req.path}\n`);
        };

        next();
    }

    listen(port: number, callback?: () => void): Server {
        const server = createServer((rawReq: IncomingMessage, rawRes: ServerResponse) => {
            this.handle(rawReq, rawRes);
        });
        server.listen(port, callback);
        return server;
    }
}
