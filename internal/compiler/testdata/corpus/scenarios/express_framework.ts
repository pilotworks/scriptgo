// @expect: [Middleware] GET /
// @expect: GET / -> status 200, body: Hello Express
// @expect: [Middleware] GET /users/42
// @expect: GET /users/42 -> status 200, userId: 42
// @expect: [Middleware] POST /echo
// @expect: POST /echo -> status 201, payload: {"msg":"ping"}
// @expect: [Middleware] GET /not-found
// @expect: GET /not-found -> status 404
// @expect: Server listening on 8089
// @expect: Server closed successfully

import { createServer, IncomingMessage, ServerResponse } from "node:http";
import { Socket } from "node:net";

type NextFunction = (err?: Error | null) => void;

interface RequestLike {
    method: string;
    url: string;
    path: string;
    params: Record<string, string>;
    body: string;
}

interface ResponseLike {
    statusCode: number;
    body: string;
    status(code: number): ResponseLike;
    send(data: string): void;
    json(data: unknown): void;
}

type Handler = (req: RequestLike, res: ResponseLike, next: NextFunction) => void;

class ExpressRequest implements RequestLike {
    method: string;
    url: string;
    path: string;
    params: Record<string, string> = {};
    body: string;

    constructor(rawReq: IncomingMessage, bodyStr: string = "") {
        this.method = rawReq.method || "GET";
        this.url = rawReq.url || "/";
        this.path = (rawReq.url || "/").split("?")[0];
        this.params = {};
        this.body = bodyStr;
    }
}

class ExpressResponse implements ResponseLike {
    statusCode: number = 200;
    body: string = "";
    rawRes: ServerResponse;

    constructor(rawRes: ServerResponse) {
        this.rawRes = rawRes;
    }

    status(code: number): ResponseLike {
        this.statusCode = code;
        this.rawRes.statusCode = code;
        return this;
    }

    send(data: string): void {
        this.body = data;
        this.rawRes.writeHead(this.statusCode);
        this.rawRes.end(data);
    }

    json(data: unknown): void {
        const str = JSON.stringify(data);
        this.body = str;
        this.rawRes.setHeader("content-type", "application/json");
        this.rawRes.writeHead(this.statusCode);
        this.rawRes.end(str);
    }
}

class ExpressApp {
    private middlewares: Handler[] = [];
    private routes: { method: string; path: string; handler: Handler }[] = [];

    use(fn: Handler): void {
        this.middlewares.push(fn);
    }

    get(path: string, fn: Handler): void {
        this.routes.push({ method: "GET", path: path, handler: fn });
    }

    post(path: string, fn: Handler): void {
        this.routes.push({ method: "POST", path: path, handler: fn });
    }

    handle(rawReq: IncomingMessage, rawRes: ServerResponse, bodyStr: string = ""): ExpressResponse {
        const req = new ExpressRequest(rawReq, bodyStr);
        const res = new ExpressResponse(rawRes);

        let mIdx = 0;
        const next: NextFunction = (err?: Error | null): void => {
            if (err) {
                res.status(500).send("Error: " + err.message);
                return;
            }
            if (mIdx < this.middlewares.length) {
                const mw = this.middlewares[mIdx++];
                mw(req, res, next);
                return;
            }

            for (let i = 0; i < this.routes.length; i++) {
                const route = this.routes[i];
                if (route.method !== req.method) continue;

                // Match param route e.g. /users/:id
                if (route.path.indexOf(":") !== -1) {
                    const rParts = route.path.split("/");
                    const uParts = req.path.split("/");
                    if (rParts.length === uParts.length) {
                        let matched = true;
                        for (let j = 0; j < rParts.length; j++) {
                            if (rParts[j].startsWith(":")) {
                                req.params[rParts[j].slice(1)] = uParts[j];
                            } else if (rParts[j] !== uParts[j]) {
                                matched = false;
                                break;
                            }
                        }
                        if (matched) {
                            route.handler(req, res, next);
                            return;
                        }
                    }
                } else if (route.path === req.path) {
                    route.handler(req, res, next);
                    return;
                }
            }

            res.status(404).send("Cannot " + req.method + " " + req.path);
        };

        next();
        return res;
    }
}

const app = new ExpressApp();

app.use((req, res, next) => {
    console.log("[Middleware] " + req.method + " " + req.path);
    next();
});

app.get("/", (req, res) => {
    res.send("Hello Express");
});

app.get("/users/:id", (req, res) => {
    res.json({ userId: req.params["id"] });
});

app.post("/echo", (req, res) => {
    res.status(201).send(req.body);
});

// Test 1: GET /
const rawReq1 = new IncomingMessage(new Socket());
rawReq1.method = "GET";
rawReq1.url = "/";
const rawRes1 = new ServerResponse(new Socket());
const res1 = app.handle(rawReq1, rawRes1);
console.log("GET / -> status " + res1.statusCode + ", body: " + res1.body);

// Test 2: GET /users/42
const rawReq2 = new IncomingMessage(new Socket());
rawReq2.method = "GET";
rawReq2.url = "/users/42";
const rawRes2 = new ServerResponse(new Socket());
const res2 = app.handle(rawReq2, rawRes2);
const userObj = JSON.parse(res2.body) as { userId: string };
console.log("GET /users/42 -> status " + res2.statusCode + ", userId: " + userObj.userId);

// Test 3: POST /echo
const rawReq3 = new IncomingMessage(new Socket());
rawReq3.method = "POST";
rawReq3.url = "/echo";
const rawRes3 = new ServerResponse(new Socket());
const res3 = app.handle(rawReq3, rawRes3, '{"msg":"ping"}');
console.log("POST /echo -> status " + res3.statusCode + ", payload: " + res3.body);

// Test 4: GET /not-found
const rawReq4 = new IncomingMessage(new Socket());
rawReq4.method = "GET";
rawReq4.url = "/not-found";
const rawRes4 = new ServerResponse(new Socket());
const res4 = app.handle(rawReq4, rawRes4);
console.log("GET /not-found -> status " + res4.statusCode);

// Test 5: Server listen and close
const server = createServer();
server.listen(8089, () => {
    console.log("Server listening on 8089");
    server.close();
    console.log("Server closed successfully");
});
