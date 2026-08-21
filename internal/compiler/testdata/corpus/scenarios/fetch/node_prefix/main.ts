import { Headers, Request, Response } from "node:http";

const h = new Headers();
h.set("x-api-key", "secret123");

const req = new Request("https://api.example.com/items", {
    method: "POST",
    headers: h,
    body: "item=1"
});

console.log(req.url);
console.log(req.method);
console.log(req.headers.get("x-api-key"));
console.log(req.body);

const res = new Response("done", {
    status: 201,
    statusText: "Created"
});

console.log(res.status);
console.log(res.statusText);
console.log(res.ok);
