import { URL } from "node:url";

const base: URL = new URL("https://example.com/api/v1/users");
const rel1: URL = new URL("posts", base.href);
console.log(rel1.href);

const rel2: URL = new URL("/root/path", base.href);
console.log(rel2.href);

const rel3: URL = new URL("?page=2", "https://example.com/items");
console.log(rel3.href);
console.log(rel3.searchParams.get("page"));
