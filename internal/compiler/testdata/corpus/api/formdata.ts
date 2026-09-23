// ScriptGo Corpus: WHATWG FormData Standard APIs
// Consolidated test suite with inline assertions.

import { Blob, File } from "node:buffer";
import "node:http";

// --- Global FormData instantiation & basic CRUD ---
// @api: formdata.constructor
// @api: formdata.append
// @api: formdata.get
// @api: formdata.has
// @expect: Alice
// @expect: true
// @expect: false
const fd = new FormData();
fd.append("name", "Alice");
console.log(fd.get("name"));
console.log(fd.has("name"));
console.log(fd.has("nonexistent"));

// --- Multiple values and getAll ---
// @api: formdata.getAll
// @expect: Alice
// @expect: Alice,Bob
fd.append("name", "Bob");
console.log(fd.get("name"));
const allNames = fd.getAll("name");
console.log(allNames.join(","));

// --- set replaces all existing entries ---
// @api: formdata.set
// @expect: Charlie
// @expect: 1
fd.set("name", "Charlie");
console.log(fd.get("name"));
console.log(fd.getAll("name").length);

// --- delete removes all matching entries ---
// @api: formdata.delete
// @expect: false
// @expect: null
fd.delete("name");
console.log(fd.has("name"));
console.log(fd.get("name"));

// --- Blob and File handling ---
// @expect: blob
// @expect: image/png
// @expect: avatar.png
// @expect: text/plain
const blob = new Blob(["blob content"], { type: "image/png" });
fd.append("avatar", blob);
const avatarEntry = fd.get("avatar");
if (avatarEntry instanceof File) {
    console.log(avatarEntry.name);
    console.log(avatarEntry.type);
}

const file = new File(["file content"], "avatar.png", { type: "text/plain" });
fd.set("avatar", file);
const avatarFile = fd.get("avatar");
if (avatarFile instanceof File) {
    console.log(avatarFile.name);
    console.log(avatarFile.type);
}

// --- Iteration: forEach, keys, values, entries, Symbol.iterator ---
// @api: formdata.keys
// @api: formdata.values
// @api: formdata.entries
// @api: formdata.forEach
// @api: formdata.iterator
// @expect: k1,k2
// @expect: v1,v2
// @expect: k1:v1
// @expect: k2:v2
// @expect: k1=v1
// @expect: k2=v2
const fd2 = new FormData();
fd2.append("k1", "v1");
fd2.append("k2", "v2");
console.log(Array.from(fd2.keys()).join(","));
console.log(Array.from(fd2.values()).join(","));
for (const [k, v] of fd2.entries()) {
    console.log(k + ":" + v);
}
fd2.forEach((v, k) => {
    console.log(k + "=" + v);
});

// --- Additional FormData instance check ---
// @expect: bar
const nfd = new FormData();
nfd.append("foo", "bar");
console.log(nfd.get("foo"));

// --- Response.prototype.formData() ---
// @api: response.formData
// @expect: urlencoded: 1, world
// @expect: multipart: John Doe, bio.txt, Engineer, isFile=true
async function testResponseFormData() {
    const urlEncodedBody = "id=1&greeting=world";
    const res1 = new Response(urlEncodedBody, {
        headers: { "content-type": "application/x-www-form-urlencoded" }
    });
    const fd1 = await res1.formData();
    console.log("urlencoded: " + fd1.get("id") + ", " + fd1.get("greeting"));

    const boundary = "----TestBoundary123";
    const multipartBody = "--" + boundary + "\r\n" +
        "Content-Disposition: form-data; name=\"username\"\r\n\r\n" +
        "John Doe\r\n" +
        "--" + boundary + "\r\n" +
        "Content-Disposition: form-data; name=\"bio\"; filename=\"bio.txt\"\r\n" +
        "Content-Type: text/plain\r\n\r\n" +
        "Engineer\r\n" +
        "--" + boundary + "--\r\n";
    const res2 = new Response(multipartBody, {
        headers: { "content-type": "multipart/form-data; boundary=" + boundary }
    });
    const fd2Parsed = await res2.formData();
    const bioEntry = fd2Parsed.get("bio");
    let isFile = false;
    if (bioEntry instanceof File) {
        isFile = true;
    }
    const bioFile = bioEntry as File;
    const bioName = bioFile.name;
    const bioContent = await bioFile.text();
    console.log("multipart: " + fd2Parsed.get("username") + ", " + bioName + ", " + bioContent + ", isFile=" + isFile);
}
testResponseFormData();
