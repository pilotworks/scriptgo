import { Response } from "node:http";
const r = new Response("hello response", { status: 200 });
console.log(r.status);
console.log(r.ok);
