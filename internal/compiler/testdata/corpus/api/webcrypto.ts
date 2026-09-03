import {
    webcrypto,
    Crypto,
    CryptoKey,
    CryptoKeyPair,
    SubtleCrypto,
    Algorithm,
    KeyAlgorithm,
    AesDerivedKeyParams,
    AesCbcParams,
    AesCtrParams,
    AesGcmParams,
    AesKeyAlgorithm,
    AesKeyGenParams,
    EcdhKeyDeriveParams,
    EcdsaParams,
    EcKeyAlgorithm,
    EcKeyGenParams,
    EcKeyImportParams,
    Ed448Params,
    HkdfParams,
    HmacImportParams,
    HmacKeyAlgorithm,
    HmacKeyGenParams,
    Pbkdf2Params,
    RsaHashedImportParams,
    RsaHashedKeyAlgorithm,
    RsaHashedKeyGenParams,
    RsaOaepParams,
    RsaPssParams,
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

// @api: webcrypto.webcrypto.Algorithm
// @api: webcrypto.Algorithm
// @api: new webcrypto.Algorithm
// @api: Algorithm.name
// @expect: wc_algo: test
const algo = new Algorithm("test");
console.log("wc_algo: " + algo.name);

// @api: webcrypto.webcrypto.KeyAlgorithm
// @api: webcrypto.KeyAlgorithm
// @api: new webcrypto.KeyAlgorithm
// @api: KeyAlgorithm.name
// @expect: wc_keyalgo: AES-GCM
const keyAlgo = new KeyAlgorithm();
console.log("wc_keyalgo: " + keyAlgo.name);

// @api: webcrypto.webcrypto.AesDerivedKeyParams
// @api: webcrypto.AesDerivedKeyParams
// @api: new webcrypto.AesDerivedKeyParams
// @api: AesDerivedKeyParams.name
// @api: AesDerivedKeyParams.length
// @expect: wc_aesDerived: AES-CTR 256
const aesDerived = new AesDerivedKeyParams();
console.log("wc_aesDerived: " + aesDerived.name + " " + aesDerived.length);

// @api: webcrypto.webcrypto.AesCbcParams
// @api: webcrypto.AesCbcParams
// @api: new webcrypto.AesCbcParams
// @api: AesCbcParams.name
// @api: AesCbcParams.iv
// @expect: wc_aesCbc: AES-CBC true
const aesCbc = new AesCbcParams();
console.log("wc_aesCbc: " + aesCbc.name + " " + (typeof aesCbc.iv === "object"));

// @api: webcrypto.webcrypto.AesCtrParams
// @api: webcrypto.AesCtrParams
// @api: new webcrypto.AesCtrParams
// @api: AesCtrParams.name
// @api: AesCtrParams.counter
// @api: AesCtrParams.length
// @expect: wc_aesCtr: AES-CTR true 64
const aesCtr = new AesCtrParams();
console.log("wc_aesCtr: " + aesCtr.name + " " + (typeof aesCtr.counter === "object") + " " + aesCtr.length);

// @api: webcrypto.webcrypto.AesGcmParams
// @api: webcrypto.AesGcmParams
// @api: new webcrypto.AesGcmParams
// @api: AesGcmParams.name
// @api: AesGcmParams.iv
// @api: AesGcmParams.additionalData
// @api: AesGcmParams.tagLength
// @expect: wc_aesGcm: AES-GCM 128
const aesGcm = new AesGcmParams();
console.log("wc_aesGcm: " + aesGcm.name + " " + aesGcm.tagLength);

// @api: webcrypto.webcrypto.AesKeyAlgorithm
// @api: webcrypto.AesKeyAlgorithm
// @api: new webcrypto.AesKeyAlgorithm
// @api: AesKeyAlgorithm.name
// @api: AesKeyAlgorithm.length
// @expect: wc_aesKeyAlgo: AES-GCM 256
const aesKeyAlgo = new AesKeyAlgorithm();
console.log("wc_aesKeyAlgo: " + aesKeyAlgo.name + " " + aesKeyAlgo.length);

// @api: webcrypto.webcrypto.AesKeyGenParams
// @api: webcrypto.AesKeyGenParams
// @api: new webcrypto.AesKeyGenParams
// @api: AesKeyGenParams.name
// @api: AesKeyGenParams.length
// @expect: wc_aesKeyGen: AES-GCM 256
const aesKeyGen = new AesKeyGenParams();
console.log("wc_aesKeyGen: " + aesKeyGen.name + " " + aesKeyGen.length);

// @api: webcrypto.webcrypto.EcdhKeyDeriveParams
// @api: webcrypto.EcdhKeyDeriveParams
// @api: new webcrypto.EcdhKeyDeriveParams
// @api: EcdhKeyDeriveParams.name
// @api: EcdhKeyDeriveParams.public
// @expect: wc_ecdhKeyDerive: ECDH public
const ecdhKeyDerive = new EcdhKeyDeriveParams();
console.log("wc_ecdhKeyDerive: " + ecdhKeyDerive.name + " " + ecdhKeyDerive.public.type);

// @api: webcrypto.webcrypto.EcdsaParams
// @api: webcrypto.EcdsaParams
// @api: new webcrypto.EcdsaParams
// @api: EcdsaParams.name
// @api: EcdsaParams.hash
// @expect: wc_ecdsa: ECDSA SHA-256
const ecdsa = new EcdsaParams();
console.log("wc_ecdsa: " + ecdsa.name + " " + ecdsa.hash);

// @api: webcrypto.webcrypto.EcKeyAlgorithm
// @api: webcrypto.EcKeyAlgorithm
// @api: new webcrypto.EcKeyAlgorithm
// @api: EcKeyAlgorithm.name
// @api: EcKeyAlgorithm.namedCurve
// @expect: wc_ecKeyAlgo: ECDSA P-256
const ecKeyAlgo = new EcKeyAlgorithm();
console.log("wc_ecKeyAlgo: " + ecKeyAlgo.name + " " + ecKeyAlgo.namedCurve);

// @api: webcrypto.webcrypto.EcKeyGenParams
// @api: webcrypto.EcKeyGenParams
// @api: new webcrypto.EcKeyGenParams
// @api: EcKeyGenParams.name
// @api: EcKeyGenParams.namedCurve
// @expect: wc_ecKeyGen: ECDSA P-256
const ecKeyGen = new EcKeyGenParams();
console.log("wc_ecKeyGen: " + ecKeyGen.name + " " + ecKeyGen.namedCurve);

// @api: webcrypto.webcrypto.EcKeyImportParams
// @api: webcrypto.EcKeyImportParams
// @api: new webcrypto.EcKeyImportParams
// @api: EcKeyImportParams.name
// @api: EcKeyImportParams.namedCurve
// @expect: wc_ecKeyImport: ECDSA P-256
const ecKeyImport = new EcKeyImportParams();
console.log("wc_ecKeyImport: " + ecKeyImport.name + " " + ecKeyImport.namedCurve);

// @api: webcrypto.webcrypto.Ed448Params
// @api: webcrypto.Ed448Params
// @api: new webcrypto.Ed448Params
// @api: Ed448Params.name
// @api: Ed448Params.context
// @expect: wc_ed448: Ed448 true
const ed448 = new Ed448Params();
console.log("wc_ed448: " + ed448.name + " " + (typeof ed448.context === "object"));

// @api: webcrypto.webcrypto.HkdfParams
// @api: webcrypto.HkdfParams
// @api: new webcrypto.HkdfParams
// @api: HkdfParams.name
// @api: HkdfParams.hash
// @api: HkdfParams.salt
// @api: HkdfParams.info
// @expect: wc_hkdf: HKDF SHA-256
const hkdfParams = new HkdfParams();
console.log("wc_hkdf: " + hkdfParams.name + " " + hkdfParams.hash);

// @api: webcrypto.webcrypto.HmacImportParams
// @api: webcrypto.HmacImportParams
// @api: new webcrypto.HmacImportParams
// @api: HmacImportParams.name
// @api: HmacImportParams.hash
// @api: HmacImportParams.length
// @expect: wc_hmacImport: HMAC SHA-256 256
const hmacImport = new HmacImportParams();
console.log("wc_hmacImport: " + hmacImport.name + " " + hmacImport.hash + " " + hmacImport.length);

// @api: webcrypto.webcrypto.HmacKeyAlgorithm
// @api: webcrypto.HmacKeyAlgorithm
// @api: new webcrypto.HmacKeyAlgorithm
// @api: HmacKeyAlgorithm.name
// @api: HmacKeyAlgorithm.hash
// @api: HmacKeyAlgorithm.length
// @expect: wc_hmacKeyAlgo: HMAC SHA-256 256
const hmacKeyAlgo = new HmacKeyAlgorithm();
console.log("wc_hmacKeyAlgo: " + hmacKeyAlgo.name + " " + hmacKeyAlgo.hash.name + " " + hmacKeyAlgo.length);

// @api: webcrypto.webcrypto.HmacKeyGenParams
// @api: webcrypto.HmacKeyGenParams
// @api: new webcrypto.HmacKeyGenParams
// @api: HmacKeyGenParams.name
// @api: HmacKeyGenParams.hash
// @api: HmacKeyGenParams.length
// @expect: wc_hmacKeyGen: HMAC SHA-256 256
const hmacKeyGen = new HmacKeyGenParams();
console.log("wc_hmacKeyGen: " + hmacKeyGen.name + " " + hmacKeyGen.hash + " " + hmacKeyGen.length);

// @api: webcrypto.webcrypto.Pbkdf2Params
// @api: webcrypto.Pbkdf2Params
// @api: new webcrypto.Pbkdf2Params
// @api: Pbkdf2Params.name
// @api: Pbkdf2Params.hash
// @api: Pbkdf2Params.salt
// @api: Pbkdf2Params.iterations
// @expect: wc_pbkdf2: PBKDF2 SHA-256 100000
const pbkdf2Params = new Pbkdf2Params();
console.log("wc_pbkdf2: " + pbkdf2Params.name + " " + pbkdf2Params.hash + " " + pbkdf2Params.iterations);

// @api: webcrypto.webcrypto.RsaHashedImportParams
// @api: webcrypto.RsaHashedImportParams
// @api: new webcrypto.RsaHashedImportParams
// @api: RsaHashedImportParams.name
// @api: RsaHashedImportParams.hash
// @expect: wc_rsaHashedImport: RSA-OAEP SHA-256
const rsaHashedImport = new RsaHashedImportParams();
console.log("wc_rsaHashedImport: " + rsaHashedImport.name + " " + rsaHashedImport.hash);

// @api: webcrypto.webcrypto.RsaHashedKeyAlgorithm
// @api: webcrypto.RsaHashedKeyAlgorithm
// @api: new webcrypto.RsaHashedKeyAlgorithm
// @api: RsaHashedKeyAlgorithm.name
// @api: RsaHashedKeyAlgorithm.modulusLength
// @api: RsaHashedKeyAlgorithm.publicExponent
// @api: RsaHashedKeyAlgorithm.hash
// @expect: wc_rsaHashedKeyAlgo: RSA-OAEP 2048 SHA-256
const rsaHashedKeyAlgo = new RsaHashedKeyAlgorithm();
console.log("wc_rsaHashedKeyAlgo: " + rsaHashedKeyAlgo.name + " " + rsaHashedKeyAlgo.modulusLength + " " + rsaHashedKeyAlgo.hash.name);

// @api: webcrypto.webcrypto.RsaHashedKeyGenParams
// @api: webcrypto.RsaHashedKeyGenParams
// @api: new webcrypto.RsaHashedKeyGenParams
// @api: RsaHashedKeyGenParams.name
// @api: RsaHashedKeyGenParams.modulusLength
// @api: RsaHashedKeyGenParams.publicExponent
// @api: RsaHashedKeyGenParams.hash
// @expect: wc_rsaHashedKeyGen: RSA-OAEP 2048 SHA-256
const rsaHashedKeyGen = new RsaHashedKeyGenParams();
console.log("wc_rsaHashedKeyGen: " + rsaHashedKeyGen.name + " " + rsaHashedKeyGen.modulusLength + " " + rsaHashedKeyGen.hash);

// @api: webcrypto.webcrypto.RsaOaepParams
// @api: webcrypto.RsaOaepParams
// @api: new webcrypto.RsaOaepParams
// @api: RsaOaepParams.name
// @api: RsaOaepParams.label
// @expect: wc_rsaOaep: RSA-OAEP true
const rsaOaep = new RsaOaepParams();
console.log("wc_rsaOaep: " + rsaOaep.name + " " + (typeof rsaOaep.label === "object"));

// @api: webcrypto.webcrypto.RsaPssParams
// @api: webcrypto.RsaPssParams
// @api: new webcrypto.RsaPssParams
// @api: RsaPssParams.name
// @api: RsaPssParams.saltLength
// @expect: wc_rsaPss: RSA-PSS 32
const rsaPss = new RsaPssParams();
console.log("wc_rsaPss: " + rsaPss.name + " " + rsaPss.saltLength);

// @api: webcrypto.webcrypto.SubtleCrypto
// @api: webcrypto.SubtleCrypto
// @api: new webcrypto.SubtleCrypto
// @api: SubtleCrypto.deriveBits
// @api: SubtleCrypto.deriveKey
// @api: SubtleCrypto.digest
// @api: SubtleCrypto.exportKey
// @api: SubtleCrypto.generateKey
// @api: SubtleCrypto.importKey
// @api: SubtleCrypto.sign
// @api: SubtleCrypto.verify
// @expect: wc_async: true
const runWebcryptoAsync = async () => {
    const subtle = new SubtleCrypto();
    const k = new CryptoKey();
    const data = new Uint8Array(16);
    await subtle.sign("HMAC", k, data);
    await subtle.verify("HMAC", k, data, data);
    await subtle.digest("SHA-256", data);
    await subtle.generateKey("AES-GCM", true, ["encrypt"]);
    await subtle.deriveKey("HKDF", k, "AES-GCM", true, ["encrypt"]);
    await subtle.deriveBits("HKDF", k, 256);
    await subtle.importKey("raw", data, "AES-GCM", true, ["encrypt"]);
    await subtle.exportKey("raw", k);
    console.log("wc_async: true");
};
runWebcryptoAsync();
