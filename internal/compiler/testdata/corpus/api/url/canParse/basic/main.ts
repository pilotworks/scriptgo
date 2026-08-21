import { URL } from "node:url";
console.log(URL.canParse("https://example.com"));
console.log(URL.canParse("not a valid url:::"));
