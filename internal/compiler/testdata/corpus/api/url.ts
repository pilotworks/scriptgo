// ScriptGo Corpus: Node.js URL Module (Strict 1:1 Parity Tests)
import {
    URL,
    URLSearchParams,
    Url,
    fileURLToPath,
    format,
    parse,
    pathToFileURL,
    resolve,
    urlToHttpOptions
} from "node:url";
import * as url from "node:url";

// @api: url.URL
// @expect: https://example.com/
const u1 = new URL("https://example.com/");
console.log(u1.href);

// @api: url.constructor
// @expect: https://example.com/sub
const uCtor = new URL("/sub", "https://example.com");
console.log(uCtor.href);

// @api: url.toJSON
// @expect: https://example.com/
console.log(u1.toJSON());

// @api: url.toString
// @expect: https://example.com/
console.log(u1.toString());

// @api: url.canParse
// @expect: true
console.log(URL.canParse("https://example.com"));

// @api: url.parse
// @expect: https://example.com/test
const parsedU = URL.parse("https://example.com/test");
if (parsedU !== null) {
    console.log(parsedU.href);
}

// @api: url.URLSearchParams
// @expect: 1
const sp1 = new URLSearchParams("a=1&b=2");
console.log(sp1.get("a"));

// @api: url.hash
// @expect: #frag
const uHash = new URL("https://example.com/path#frag");
console.log(uHash.hash);

// @api: url.host
// @expect: example.com:8080
const uHost = new URL("https://example.com:8080/path");
console.log(uHost.host);

// @api: url.hostname
// @expect: example.com
const uHn = new URL("https://example.com:8080/path");
console.log(uHn.hostname);

// @api: url.href
// @expect: https://example.com/hello
const uHref = new URL("https://example.com/hello");
console.log(uHref.href);

// @api: url.origin
// @expect: https://example.com:8080
const uOrigin = new URL("https://example.com:8080/path");
console.log(uOrigin.origin);

// @api: url.password
// @expect: pass123
const uPass = new URL("https://user:pass123@example.com/path");
console.log(uPass.password);

// @api: url.pathname
// @expect: /api/v1
const uPathname = new URL("https://example.com/api/v1?query=1");
console.log(uPathname.pathname);

// @api: url.port
// @expect: 8080
const uPort = new URL("https://example.com:8080/path");
console.log(uPort.port);

// @api: url.protocol
// @expect: https:
const uProto = new URL("https://example.com/");
console.log(uProto.protocol);

// @api: url.search
// @expect: ?a=1&b=2
const uSearch = new URL("https://example.com/path?a=1&b=2");
console.log(uSearch.search);

// @api: url.searchParams
// @expect: 1
const uSP = new URL("https://example.com/path?a=1&b=2");
console.log(uSP.searchParams.get("a"));

// @api: url.username
// @expect: user123
const uUser = new URL("https://user123:pass@example.com/");
console.log(uUser.username);

// @api: URLSearchParams.size
// @expect: 2
const spSize = new URLSearchParams("a=1&b=2");
console.log(spSize.size);

// @api: URLSearchParams.append
// @expect: 1
const spApp = new URLSearchParams();
spApp.append("x", "1");
console.log(spApp.get("x"));

// @api: URLSearchParams.delete
// @expect: true
const spDel = new URLSearchParams("x=1&y=2");
spDel.delete("x");
console.log(spDel.get("x") === null);

// @api: URLSearchParams.entries
// @expect: true
const spEnt = new URLSearchParams("x=1&y=2");
console.log(spEnt.entries() !== null);

// @api: URLSearchParams.forEach
// @expect: true
let forEachCount = 0;
spEnt.forEach(() => { forEachCount++; });
console.log(forEachCount === 2);

// @api: URLSearchParams.get
// @expect: hello
const spGet = new URLSearchParams("msg=hello");
console.log(spGet.get("msg"));

// @api: URLSearchParams.getAll
// @expect: 2
const spGetAll = new URLSearchParams("item=1&item=2");
console.log(spGetAll.getAll("item").length);

// @api: URLSearchParams.has
// @expect: true
console.log(spGetAll.has("item"));

// @api: URLSearchParams.keys
// @expect: true
console.log(spGetAll.keys() !== null);

// @api: URLSearchParams.set
// @expect: 99
const spSet = new URLSearchParams("x=1");
spSet.set("x", "99");
console.log(spSet.get("x"));

// @api: URLSearchParams.sort
// @expect: a=1&b=2
const spSort = new URLSearchParams("b=2&a=1");
spSort.sort();
console.log(spSort.toString());

// @api: URLSearchParams.toString
// @expect: a=1&b=2
console.log(spSort.toString());

// @api: URLSearchParams.values
// @expect: true
console.log(spSort.values() !== null);

// @api: URLSearchParams.[Symbol.iterator]
// @expect: true
console.log(spSort.entries() !== null);

// @api: url.fileURLToPath
// @expect: /path/to/file.txt
console.log(fileURLToPath("file:///path/to/file.txt"));

// @api: url.format
// @expect: https://example.com/
console.log(format(new URL("https://example.com/")));

// @api: url.pathToFileURL
// @expect: file:///path/file.txt
console.log(pathToFileURL("/path/file.txt").href);

// @api: url.urlToHttpOptions
// @expect: example.com
console.log(urlToHttpOptions(new URL("https://user:pass@example.com:8080/path")).hostname);

// @api: url.Url
// @expect: true
const legacyUrl = parse("http://user:pass@example.com:8080/path/sub?query=1#hash");
console.log(typeof legacyUrl === "object");

// @api: url.auth
// @expect: user:pass
console.log(legacyUrl.auth);

// @api: url.path
// @expect: /path/sub?query=1
console.log(legacyUrl.path);

// @api: url.query
// @expect: query=1
console.log(legacyUrl.query);

// @api: url.slashes
// @expect: true
console.log(legacyUrl.slashes);

// @api: url.resolve
// @expect: https://example.com/sub/file.txt
console.log(resolve("https://example.com/a/b", "/sub/file.txt"));
