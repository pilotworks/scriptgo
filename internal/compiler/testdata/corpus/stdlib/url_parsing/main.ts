import { URL } from "node:url";

const u: URL = new URL("https://user:pass@example.com:8080/path/to/page?query=123#myhash");
console.log(u.protocol);
console.log(u.username);
console.log(u.password);
console.log(u.host);
console.log(u.hostname);
console.log(u.port);
console.log(u.pathname);
console.log(u.search);
console.log(u.hash);
console.log(u.origin);
console.log(u.href);
console.log(u.toString());
console.log(URL.canParse("https://example.com"));
console.log(URL.canParse("invalid-no-scheme"));
