// ScriptGo Corpus: Crypto Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as crypto from "node:crypto";
import { createHash } from "node:crypto";

// @api: crypto.createHash
// @expect: true
const hash_crypto_createHash_0 = crypto.createHash("sha256").update("hello").digest("hex");
console.log(hash_crypto_createHash_0.length === 64);

// @api: crypto.digest
// @expect: true
const h_crypto_digest_1 = createHash("sha256");
h_crypto_digest_1.update("test");
const d_crypto_digest_1: string = h_crypto_digest_1.digest("hex");
console.log(d_crypto_digest_1.length > 0);

// @api: crypto.randomUUID
// @expect: true
const uuid_crypto_randomUUID_2: string = crypto.randomUUID();
console.log(uuid_crypto_randomUUID_2.length > 0);

// @api: crypto.update
// @expect: true
const h_crypto_update_3 = createHash("sha256");
h_crypto_update_3.update("hello");
console.log(h_crypto_update_3.digest("hex").length > 0);

// @api: crypto.createHash (md5, sha1, sha512, base64)
// @expect: true
// @expect: true
// @expect: true
// @expect: true
const md5Hash = crypto.createHash("md5").update("hello").digest("hex");
console.log(md5Hash.length === 32);
const sha1Hash = crypto.createHash("sha1").update("hello").digest("hex");
console.log(sha1Hash.length === 40);
const sha512Hash = crypto.createHash("sha512").update("hello").digest("hex");
console.log(sha512Hash.length === 128);
const b64Hash = crypto.createHash("sha256").update("hello").digest("base64");
console.log(b64Hash.length > 0);

// @api: crypto.createHmac
// @expect: true
const hmac = crypto.createHmac("sha256", "secret-key").update("message").digest("hex");
console.log(hmac.length === 64);

// @api: crypto.randomBytes
// @expect: 16
const rBytes = crypto.randomBytes(16);
console.log(rBytes.length);

// @api: crypto.randomInt
// @expect: true
const rInt = crypto.randomInt(10, 20);
console.log(rInt >= 10 && rInt < 20);

// @api: crypto.randomFillSync
// @expect: 8
const fillBuf = Buffer.alloc(8);
crypto.randomFillSync(fillBuf);
console.log(fillBuf.length);

// @api: crypto.timingSafeEqual
// @expect: true
// @expect: false
const bufA = Buffer.from("hello");
const bufB = Buffer.from("hello");
const bufC = Buffer.from("world");
console.log(crypto.timingSafeEqual(bufA, bufB));
console.log(crypto.timingSafeEqual(bufA, bufC));

// @api: crypto.pbkdf2Sync
// @expect: 32
const key = crypto.pbkdf2Sync("password", "salt", 100, 32, "sha256");
console.log(key.length);

// @api: crypto.getHashes
// @expect: true
const hashes = crypto.getHashes();
console.log(hashes.length >= 4);

// @api: crypto.constants
// @expect: 1
console.log(crypto.constants.RSA_PKCS1_PADDING);

