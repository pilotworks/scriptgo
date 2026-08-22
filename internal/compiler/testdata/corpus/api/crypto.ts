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
