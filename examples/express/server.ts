import express, { Request, Response, NextFunction } from "./index";
import { IncomingMessage, ServerResponse } from "node:http";
import { Socket } from "node:net";

const app = express();
const port = 8080;

// Global middleware: logging
app.use((req, res, next) => {
    console.log(`[Middleware] ${req.method} ${req.path}`);
    next();
});

// Built-in body parser
app.use(express.json());

// 1. Root route
app.get("/", (req, res) => {
    res.send("Hello from ScriptGo Express Native!\n");
});

// 2. Param route
app.get("/api/users/:id", (req, res) => {
    res.json({
        userId: req.params["id"],
        role: "admin"
    });
});

// 3. POST JSON echo route
app.post("/api/echo", (req, res) => {
    res.status(201).json({
        status: "created",
        received: req.body
    });
});

console.log("=== Testing Express Pipeline In-Memory ===");
// 1. Test GET /
const rawReq1 = new IncomingMessage(new Socket());
rawReq1.method = "GET";
rawReq1.url = "/";
const rawRes1 = new ServerResponse(new Socket());
app.handle(rawReq1, rawRes1);
console.log(`GET / -> status ${rawRes1.statusCode}`);

// 2. Test GET /api/users/42
const rawReq2 = new IncomingMessage(new Socket());
rawReq2.method = "GET";
rawReq2.url = "/api/users/42";
const rawRes2 = new ServerResponse(new Socket());
app.handle(rawReq2, rawRes2);
console.log(`GET /api/users/42 -> status ${rawRes2.statusCode}`);

// 3. Test 404 Route
const rawReq3 = new IncomingMessage(new Socket());
rawReq3.method = "GET";
rawReq3.url = "/not-found";
const rawRes3 = new ServerResponse(new Socket());
app.handle(rawReq3, rawRes3);
console.log(`GET /not-found -> status ${rawRes3.statusCode}`);

console.log("=== Testing Network Server Listen & Close ===");
const server = app.listen(port, () => {
    console.log(`Express server listening on port ${port}`);
    server.close();
    console.log("Server closed successfully.");
});

