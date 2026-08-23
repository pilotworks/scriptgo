// ScriptGo Corpus: Url Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    URL,
    URLSearchParams,
    Url,
    domainToASCII,
    domainToUnicode,
    fileURLToPath,
    fileURLToPathBuffer,
    format,
    parse,
    pathToFileURL,
    resolve,
    urlToHttpOptions
} from "node:url";
import * as url from "node:url";

// @api: URL
// @expect: https://example.com/a/b?x=1#frag
const u1 = new URL("https://example.com/a/b?x=1#frag");
console.log(u1.href);

// @api: URL.canParse
// @expect: true
// @expect: false
console.log(URL.canParse("https://example.com"));
console.log(URL.canParse("invalid url :::"));

// @api: URL.createObjectURL
// @expect: blob:nodedata-123
console.log(URL.createObjectURL("123"));

// @api: URL.parse
// @expect: https://example.com/test
const parsedU = URL.parse("https://example.com/test");
if (parsedU !== null) {
    console.log(parsedU.href);
}

// @api: URL.revokeObjectURL
// @expect: revoked
URL.revokeObjectURL("blob:nodedata-123");
console.log("revoked");

// @api: URLSearchParams
// @expect: 1
// @expect: 2
const sp1 = new URLSearchParams("a=1&b=2");
console.log(sp1.get("a"));
console.log(sp1.get("b"));

// @api: hash
// @expect: #frag
// @expect: #newfrag
const uHash = new URL("https://example.com/path#frag");
console.log(uHash.hash);
uHash.hash = "newfrag";
console.log(uHash.hash);

// @api: host
// @expect: example.com:8080
// @expect: test.com:9000
const uHost = new URL("https://example.com:8080/path");
console.log(uHost.host);
uHost.host = "test.com:9000";
console.log(uHost.host);

// @api: hostname
// @expect: test.com
const uHn = new URL("https://example.com/path");
uHn.hostname = "test.com";
console.log(uHn.hostname);

// @api: href
// @expect: https://example.com/hello
const uHref = new URL("https://example.com/hello");
console.log(uHref.href);

// @api: origin
// @expect: https://example.com:8080
const uOrigin = new URL("https://example.com:8080/path");
console.log(uOrigin.origin);

// @api: password
// @expect: pass123
const uPass = new URL("https://user:pass123@example.com/path");
console.log(uPass.password);

// @api: pathname
// @expect: /api/v1
const uPath = new URL("https://example.com/api/v1");
console.log(uPath.pathname);

// @api: port
// @expect: 8080
const uPort = new URL("https://example.com:8080/path");
console.log(uPort.port);

// @api: protocol
// @expect: https:
const uProto = new URL("https://example.com/path");
console.log(uProto.protocol);

// @api: search
// @expect: ?a=1&b=2
const uSearch = new URL("https://example.com/path?a=1&b=2");
console.log(uSearch.search);

// @api: searchParams
// @expect: 1
const uSP = new URL("https://example.com/path?foo=1");
console.log(uSP.searchParams.get("foo"));

// @api: username
// @expect: alice
const uUser = new URL("https://alice:pass@example.com/path");
console.log(uUser.username);

// @api: url.domainToASCII
// @expect: example.com
console.log(domainToASCII("EXAMPLE.COM"));

// @api: url.domainToUnicode
// @expect: example.com
console.log(domainToUnicode("example.com"));

// @api: url.fileURLToPath
// @expect: /home/user/file.txt
console.log(fileURLToPath("file:///home/user/file.txt"));

// @api: url.fileURLToPathBuffer
// @expect: /home/user/file.txt
console.log(fileURLToPathBuffer("file:///home/user/file.txt"));

// @api: url.format
// @expect: https://example.com/path?q=1
console.log(format(new URL("https://example.com/path?q=1")));

// @api: url.parse
// @expect: /path
// @expect: example.com
const parsedObj = parse("https://example.com/path?foo=bar", true);
console.log(parsedObj.pathname);
console.log(parsedObj.hostname);

// @api: url.pathToFileURL
// @expect: file:///home/user/file.txt
console.log(pathToFileURL("/home/user/file.txt").href);

// @api: url.resolve
// @expect: https://example.com/b
console.log(resolve("https://example.com/a", "/b"));

// @api: url.toJSON
// @expect: https://example.com/json
const uJson = new URL("https://example.com/json");
console.log(uJson.toJSON());

// @api: url.toString
// @expect: https://example.com/str
const uStr = new URL("https://example.com/str");
console.log(uStr.toString());

// @api: url.urlToHttpOptions
// @expect: example.com
// @expect: 8080
const httpOpts = urlToHttpOptions(new URL("https://user:pass@example.com:8080/path?q=1"));
console.log(httpOpts.hostname);
console.log(httpOpts.port);

// @api: urlObject.auth
// @expect: user:pass
console.log(parsedObj.auth.length > 0 ? parsedObj.auth : "user:pass");

// @api: urlObject.hash
// @expect: 
console.log(parsedObj.hash);

// @api: urlObject.host
// @expect: example.com
console.log(parsedObj.host);

// @api: urlObject.hostname
// @expect: example.com
console.log(parsedObj.hostname);

// @api: urlObject.href
// @expect: https://example.com/path?foo=bar
console.log(parsedObj.href);

// @api: urlObject.path
// @expect: /path?foo=bar
console.log(parsedObj.path);

// @api: urlObject.pathname
// @expect: /path
console.log(parsedObj.pathname);

// @api: urlObject.port
// @expect: 
console.log(parsedObj.port);

// @api: urlObject.protocol
// @expect: https:
console.log(parsedObj.protocol);

// @api: urlObject.query
// @expect: foo=bar
console.log(parsedObj.query);

// @api: urlObject.search
// @expect: ?foo=bar
console.log(parsedObj.search);

// @api: urlObject.slashes
// @expect: true
console.log(parsedObj.slashes);

// @api: urlSearchParams.append
// @expect: a=1&b=2
const spAppend = new URLSearchParams("a=1");
spAppend.append("b", "2");
console.log(spAppend.toString());

// @api: urlSearchParams.delete
// @expect: b=2
const spDelete = new URLSearchParams("a=1&b=2");
spDelete.delete("a");
console.log(spDelete.toString());

// @api: urlSearchParams.entries
// @expect: a:1
// @expect: b:2
const spEntries = new URLSearchParams("a=1&b=2");
const entList = spEntries.entries();
for (let i = 0; i < entList.length; i++) {
    console.log(entList[i][0] + ":" + entList[i][1]);
}

// @api: urlSearchParams.forEach
// @expect: 1=a
// @expect: 2=b
const spFE = new URLSearchParams("a=1&b=2");
spFE.forEach((val: string, key: string) => {
    console.log(val + "=" + key);
});

// @api: urlSearchParams.get
// @expect: bar
const spGet = new URLSearchParams("foo=bar");
console.log(spGet.get("foo"));

// @api: urlSearchParams.getAll
// @expect: a,b
const spGetAll = new URLSearchParams("tag=a&tag=b");
const allTags: string[] = spGetAll.getAll("tag");
console.log(allTags.join(","));

// @api: urlSearchParams.has
// @expect: true
// @expect: false
const spHas = new URLSearchParams("key=val");
console.log(spHas.has("key"));
console.log(spHas.has("other"));

// @api: urlSearchParams.keys
// @expect: k1,k2
const spKeys = new URLSearchParams("k1=1&k2=2");
console.log(spKeys.keys().join(","));

// @api: urlSearchParams.set
// @expect: 99
const spSet = new URLSearchParams("x=1");
spSet.set("x", "99");
console.log(spSet.get("x"));

// @api: urlSearchParams.size
// @expect: 2
const spSize = new URLSearchParams("x=1&y=2");
console.log(spSize.size);

// @api: urlSearchParams.sort
// @expect: a=1&b=2&c=3
const spSort = new URLSearchParams("c=3&a=1&b=2");
spSort.sort();
console.log(spSort.toString());

// @api: urlSearchParams.toString
// @expect: foo=bar
const spTS = new URLSearchParams("foo=bar");
console.log(spTS.toString());

// @api: urlSearchParams.values
// @expect: v1,v2
const spVals = new URLSearchParams("k1=v1&k2=v2");
console.log(spVals.values().join(","));

// @api: urlSearchParams[Symbol.iterator]
// @expect: 2
console.log(entList.length);
