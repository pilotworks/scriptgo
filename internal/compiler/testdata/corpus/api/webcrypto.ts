import { webcrypto } from "node:crypto";

// @api: crypto.webcrypto
// @api: Crypto.subtle
// @api: Crypto.randomUUID
// @api: Crypto.getRandomValues
// @expect: wc_crypto: true string 4
const cr = webcrypto;
const uuid = cr.randomUUID();
const arr = cr.getRandomValues(new Uint8Array(4));
console.log("wc_crypto: " + (typeof cr.subtle === "object") + " " + typeof uuid + " " + arr.length);

// @api: CryptoKey
// @api: SubtleCrypto.importKey
// @api: CryptoKey.type
// @api: CryptoKey.extractable
// @api: CryptoKey.algorithm
// @api: CryptoKey.usages
// @expect: wc_key: secret true AES-GCM 2
const key = await cr.subtle.importKey("raw", new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]), "AES-GCM", true, ["encrypt", "decrypt"]);
console.log("wc_key: " + key.type + " " + key.extractable + " " + key.algorithm.name + " " + key.usages.length);
const hmacKey = await cr.subtle.importKey("raw", new Uint8Array([1, 2, 3, 4]), { name: "HMAC", hash: "SHA-256" }, true, ["sign", "verify"]);

// @api: crypto.webcrypto.subtle
// @api: SubtleCrypto.deriveBits
// @api: SubtleCrypto.digest
// @api: SubtleCrypto.exportKey
// @api: SubtleCrypto.importKey
// @api: SubtleCrypto.sign
// @api: SubtleCrypto.verify
// @expect: wc_async: true
const runWebcryptoAsync = async () => {
    const subtle = cr.subtle;
    const k = hmacKey;
    const data = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8]);
    const sig = await subtle.sign("HMAC", k, data);
    const valid = await subtle.verify("HMAC", k, sig, data);
    const digest = await subtle.digest("SHA-256", data);
    const imported = await subtle.importKey("raw", data, { name: "HMAC", hash: "SHA-256" }, true, ["sign"]);
    const exported = await subtle.exportKey("raw", imported);
    const pbkdf2Params = { name: "PBKDF2", hash: "SHA-256", salt: new Uint8Array([1, 2, 3, 4]), iterations: 1000 };
    const pbkdfKey = await subtle.importKey("raw", data, { name: "PBKDF2" }, false, ["deriveBits"]);
    const bits = await subtle.deriveBits(pbkdf2Params, pbkdfKey, 128);
    const allOk = valid && sig.byteLength === 32 && digest.byteLength === 32 && (exported as ArrayBuffer).byteLength === 8 && bits.byteLength === 16;
    console.log("wc_async: " + allOk);
};
runWebcryptoAsync();
