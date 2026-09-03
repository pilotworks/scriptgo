import {
    webcrypto,
    Crypto,
    CryptoKey,
    CryptoKeyPair,
    SubtleCrypto,
} from "webcrypto";

// @api: webcrypto.webcrypto.Crypto
// @api: webcrypto.Crypto
// @api: new webcrypto.Crypto
// @api: Crypto.subtle
// @api: Crypto.randomUUID
// @api: Crypto.getRandomValues
// @expect: wc_crypto: true string 4
const cr = new Crypto();
const uuid = cr.randomUUID();
const arr = cr.getRandomValues(new Uint8Array(4));
console.log("wc_crypto: " + (typeof cr.subtle === "object") + " " + typeof uuid + " " + arr.length);

// @api: webcrypto.webcrypto.CryptoKey
// @api: webcrypto.CryptoKey
// @api: new webcrypto.CryptoKey
// @api: CryptoKey.type
// @api: CryptoKey.extractable
// @api: CryptoKey.algorithm
// @api: CryptoKey.usages
// @expect: wc_key: secret true AES-GCM 2
const key = new CryptoKey();
console.log("wc_key: " + key.type + " " + key.extractable + " " + key.algorithm.name + " " + key.usages.length);

// @api: webcrypto.webcrypto.CryptoKeyPair
// @api: webcrypto.CryptoKeyPair
// @api: new webcrypto.CryptoKeyPair
// @api: CryptoKeyPair.privateKey
// @api: CryptoKeyPair.publicKey
// @expect: wc_keypair: private public
const pair = new CryptoKeyPair();
console.log("wc_keypair: " + pair.privateKey.type + " " + pair.publicKey.type);

// @api: webcrypto.webcrypto.SubtleCrypto
// @api: webcrypto.SubtleCrypto
// @api: new webcrypto.SubtleCrypto
// @api: SubtleCrypto.deriveBits
// @api: SubtleCrypto.digest
// @api: SubtleCrypto.exportKey
// @api: SubtleCrypto.importKey
// @api: SubtleCrypto.sign
// @api: SubtleCrypto.verify
// @expect: wc_async: true
const runWebcryptoAsync = async () => {
    const subtle = new SubtleCrypto();
    const k = new CryptoKey();
    const data = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8]);
    const sig = await subtle.sign("HMAC", k, data);
    const valid = await subtle.verify("HMAC", k, sig, data);
    const digest = await subtle.digest("SHA-256", data);
    const imported = await subtle.importKey("raw", data, "AES-GCM", true, ["encrypt"]);
    const exported = await subtle.exportKey("raw", imported);
    const pbkdf2Params = { name: "PBKDF2", hash: "sha256", salt: new Uint8Array([1, 2, 3, 4]), iterations: 1000 };
    const bits = await subtle.deriveBits(pbkdf2Params, imported, 128);
    const allOk = valid && sig.byteLength === 32 && digest.byteLength === 32 && (exported as ArrayBuffer).byteLength === 8 && bits.byteLength === 16;
    console.log("wc_async: " + allOk);
};
runWebcryptoAsync();
