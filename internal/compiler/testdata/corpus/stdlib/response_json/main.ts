import { Response } from "node:http";

async function testResponse(): Promise<void> {
    const data = { message: "hello", count: 42 };
    const resp = Response.json(JSON.stringify(data));

    console.log(resp.status);
    console.log(resp.statusText);
    console.log(resp.ok);
    console.log(resp.headers.get("content-type"));

    const text = await resp.text();
    console.log(text);

    const redirect = Response.redirect("https://example.com/login", 301);
    console.log(redirect.status);
    console.log(redirect.headers.get("location"));
}

testResponse();
