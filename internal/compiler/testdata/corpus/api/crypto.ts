import crypto, {
    Certificate,
    Cipher,
    Decipher,
    DiffieHellman,
    DiffieHellmanGroup,
    ECDH,
    Hash,
    Hmac,
    KeyObject,
    Sign,
    Verify,
    X509Certificate,
    constants,
    subtle,
    webcrypto,
    checkPrime,
    checkPrimeSync,
    createCipheriv,
    createDecipheriv,
    createDiffieHellman,
    createDiffieHellmanGroup,
    createECDH,
    createHash,
    createHmac,
    createPrivateKey,
    createPublicKey,
    createSecretKey,
    createSign,
    createVerify,
    generateKey,
    generateKeyPair,
    generateKeyPairSync,
    generateKeySync,
    generatePrime,
    generatePrimeSync,
    getCipherInfo,
    getCiphers,
    getCurves,
    getDiffieHellman,
    getFips,
    getHashes,
    getRandomValues,
    hkdf,
    hkdfSync,
    pbkdf2,
    pbkdf2Sync,
    privateDecrypt,
    privateEncrypt,
    publicDecrypt,
    publicEncrypt,
    randomBytes,
    randomFill,
    randomFillSync,
    randomInt,
    randomUUID,
    scrypt,
    scryptSync,
    secureHeapUsed,
    setEngine,
    setFips,
    timingSafeEqual,
} from "node:crypto";

function onRandomFill(err: Error | null, buf: Buffer): void {}

// @api: crypto.crypto.Hash
// @api: crypto.Hash
// @api: new crypto.Hash
// @api: crypto.createHash
// @api: Hash.update
// @api: Hash.digest
// @api: Hash.copy
// @expect: cr_hash: 32 32
const h = createHash("sha256");
h.update("hello");
const hCopy = h.copy();
console.log("cr_hash: " + h.digest().length + " " + hCopy.digest().length);

// @api: crypto.crypto.Hmac
// @api: crypto.Hmac
// @api: new crypto.Hmac
// @api: crypto.createHmac
// @api: Hmac.update
// @api: Hmac.digest
// @expect: cr_hmac: 32
const hmac = createHmac("sha256", "secret");
hmac.update("hello");
console.log("cr_hmac: " + hmac.digest().length);

// @api: Hash.digest
// @api: Hmac.digest
// @expect: cr_hash_bytes: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad 9c196e32dc0175f86f4b1cb89289d6619de6bee699e4c378e68309ed97a1a6ab
const hashBytes = createHash("sha256").update("abc").digest();
const hmacBytes = createHmac("sha256", "key").update("abc").digest();
console.log("cr_hash_bytes: " + hashBytes.toString("hex") + " " + hmacBytes.toString("hex"));

// @expect: cr_hash_binary: 47ffa3ea45a70b8a41c2c0825df323c00a8b7a01c1ea06083cc41dddcc001123 963d16d355f11798a5434eaadf01feab4e09e8b31ddbdbc85a4c9a05f8dfb0b5
const binaryHashInput = Buffer.from([0, 255, 1]);
const binaryHash = createHash("sha256").update(binaryHashInput).digest("hex");
const binaryHmac = createHmac("sha256", Buffer.from([255, 0])).update(binaryHashInput).digest("hex");
console.log("cr_hash_binary: " + binaryHash + " " + binaryHmac);

// @api: crypto.crypto.X509Certificate
// @api: crypto.X509Certificate
// @api: new crypto.X509Certificate
// @api: X509Certificate.ca
// @api: X509Certificate.fingerprint
// @api: X509Certificate.fingerprint256
// @api: X509Certificate.fingerprint512
// @api: X509Certificate.infoAccess
// @api: X509Certificate.issuer
// @api: X509Certificate.issuerCertificate
// @api: X509Certificate.keyUsage
// @api: X509Certificate.publicKey
// @api: X509Certificate.raw
// @api: X509Certificate.serialNumber
// @api: X509Certificate.subject
// @api: X509Certificate.subjectAltName
// @api: X509Certificate.validFrom
// @api: X509Certificate.validFromDate
// @api: X509Certificate.validTo
// @api: X509Certificate.validToDate
// @api: X509Certificate.checkEmail
// @api: X509Certificate.checkHost
// @api: X509Certificate.checkIP
// @api: X509Certificate.checkIssued
// @api: X509Certificate.toJSON
// @api: X509Certificate.toLegacyObject
// @api: X509Certificate.toString
const testCertPem = "-----BEGIN CERTIFICATE-----\n" +
    "MIICujCCAaKgAwIBAgIBATANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDDARUZXN0\n" +
    "MB4XDTI2MDkwMzA0MzY1OVoXDTI3MDkwMzA0MzY1OVowDzENMAsGA1UEAwwEVGVz\n" +
    "dDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKV9foCbRd8dT3Va7Bp9\n" +
    "HQkuUN2RMpHG9hz/32dqjh0WaUL3i7uTnX8ux983AtXTtWdC8P+d7JnTOL5nWSHo\n" +
    "F8xYrxMBWmtb0UKNP/Z3BG34dMdQ+2dWCfqE0NspSOiZE6j67Bwduyy9pUSP2SI3\n" +
    "FrRF30nMreTCvE09n+GuGG5JH73fpLqaTCRmLBnONOzs7seIwW5b22yb59kAXc6S\n" +
    "v5XfHXintpCDsbaWsHUBzQVGU7uAVv1Lp6HSmCrcHVaEI/vpoOHnwkXu747X5A7i\n" +
    "k5YGn2u2k7jqaU7Dk9Eb6Fhr6nnZDgS18Qe8PhErygt1hqSTo9sVEtiEZbVXLSh4\n" +
    "rtsCAwEAAaMhMB8wHQYDVR0OBBYEFBRw4E7K1B1Gxg/Xlq3946bUXHRFMA0GCSqG\n" +
    "SIb3DQEBCwUAA4IBAQAtA7WoFof8XvpgemLyeicPr2zxOYTr25TMRpFN4p2H353b\n" +
    "enf2bxkms/vZQL3hnGFdJJ2SC8/utRdcKaT1WCZtHiiMWN3TEyqOoeBz6TswcKjR\n" +
    "teU1gvE0bIkUupI7gBWLQYunb46zOJl+YdyeJ7ZRVdV0XLGKutFPiMnRCg3V2fxM\n" +
    "ZorBqCGPOZm2+ZRE9j8rMOWuWqTQ10lA9pncmG+p2TPsiCR5LsqM0kaMh5XA7ijA\n" +
    "4pkaiYY/1ZoB6thc7T/xALi/KgVsiYNiDaVbwFnZDW6T8Ik0Xc95Obt6IFsiEciQ\n" +
    "HYZMLWwiBFaGrDi3sgabJEskUFpjcOMuGjWzhgpH\n" +
    "-----END CERTIFICATE-----";
const cert = new X509Certificate(testCertPem);
cert.checkEmail("test@example.com");
cert.checkHost("example.com");
cert.checkIP("127.0.0.1");
cert.checkIssued(cert);
cert.toJSON();
cert.toLegacyObject();
cert.toString();
// @expect: cr_x509: false 01 CN=Test true
console.log("cr_x509: " + cert.ca + " " + cert.serialNumber + " " + cert.subject + " " + (typeof cert.publicKey === "object"));

// @api: crypto.constants
// @api: crypto.subtle
// @api: crypto.webcrypto
// @expect: cr_props: 1 true true
console.log("cr_props: " + constants.RSA_PKCS1_PADDING + " " + (typeof subtle === "object") + " " + (typeof webcrypto === "object"));

// @api: crypto.checkPrimeSync
// @api: crypto.generatePrimeSync
// @api: crypto.getCiphers
// @api: crypto.getCurves
// @api: crypto.getHashes
// @api: crypto.getRandomValues
// @api: crypto.hkdfSync
// @api: crypto.pbkdf2Sync
// @api: crypto.randomBytes
// @api: crypto.randomFillSync
// @api: crypto.randomInt
// @api: crypto.randomUUID
// @api: crypto.scryptSync
// @api: crypto.timingSafeEqual
// @expect: cr_sync: true true 32 32 16 true
const primeOk = checkPrimeSync(3n);
const pSync = generatePrimeSync(16);
getCiphers();
getCurves();
getHashes();
getRandomValues(new Uint8Array(4));
const hkdfRes = hkdfSync("sha256", "ikm", "salt", "info", 32);
const pbkdf2Res = pbkdf2Sync("pass", "salt", 100, 32, "sha256");
const scryptRes = scryptSync("pass", "salt", 32);
const rb = randomBytes(16);
randomFillSync(Buffer.alloc(8));
randomInt(0, 10);
randomUUID();
scryptSync("pass", "salt", 32);
const eq = timingSafeEqual(Buffer.alloc(4), Buffer.alloc(4));
console.log("cr_sync: " + primeOk + " " + checkPrimeSync(pSync) + " " + hkdfRes.byteLength + " " + pbkdf2Res.length + " " + rb.length + " " + eq);
// @expect: cr_scrypt: 4cac4540992d51feeaefe4668bbfed7222f02b445aaffbbe60cfec110fb2735c
console.log("cr_scrypt: " + scryptRes.toString("hex"));
// @expect: cr_hkdf: fe8f9615d2374c0d17f77d1aeaf408c2e75fe0466073d0def23c733e2f862dfd
console.log("cr_hkdf: " + Buffer.from(hkdfRes).toString("hex"));

// @api: crypto.checkPrime
// @api: crypto.generatePrime
// @api: crypto.hkdf
// @api: crypto.pbkdf2
// @api: crypto.randomFill
// @api: crypto.scrypt
// @expect: cr_async: true
checkPrime(3n, {}, (err, res) => {});
generatePrime(32, {}, (err, prime) => {});
hkdf("sha256", "ikm", "salt", "info", 32, (err, dk) => {});
pbkdf2("pass", "salt", 100, 32, "sha256", (err, dk) => {});
randomFill(Buffer.alloc(8), onRandomFill);
scrypt("pass", "salt", 32, (err, dk) => {});
console.log("cr_async: true");
// @api: crypto.KeyObject
// @api: new crypto.KeyObject
// @api: KeyObject.type
// @api: KeyObject.symmetricKeySize
// @api: KeyObject.export
// @api: KeyObject.equals
// @api: KeyObject.toCryptoKey
// @api: crypto.createSecretKey
// @api: crypto.generateKeySync
// @api: crypto.generateKey
const secKey = createSecretKey(Buffer.from("12345678901234567890123456789012"));
const secExp = secKey.export();
const secEq = secKey.equals(secKey);
const cryptoKey = secKey.toCryptoKey("AES-GCM", true, ["encrypt"]);
const genKey = generateKeySync("aes", { length: 256 });
generateKey("hmac", { length: 256 }, (err, k) => {});
// @expect: cr_keyobject: secret 32 true true 32
console.log("cr_keyobject: " + secKey.type + " " + secKey.symmetricKeySize + " " + (Buffer.isBuffer(secExp)) + " " + secEq + " " + (genKey.symmetricKeySize || 0));

// @api: crypto.generateKeyPairSync
// @api: crypto.generateKeyPair
// @api: crypto.createPublicKey
// @api: crypto.createPrivateKey
// @api: KeyObject.asymmetricKeyType
// @api: KeyObject.asymmetricKeyDetails
const rsaKp = generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
    privateKeyEncoding: { type: "pkcs8", format: "pem" }
});
generateKeyPair("rsa", { modulusLength: 2048 }, (err, pub, priv) => {});
const pubObj = createPublicKey(rsaKp.publicKey);
const privObj = createPrivateKey(rsaKp.privateKey);
// @expect: cr_rsa_keygen: public private rsa 2048
console.log("cr_rsa_keygen: " + pubObj.type + " " + privObj.type + " " + pubObj.asymmetricKeyType + " " + (pubObj.asymmetricKeyDetails ? pubObj.asymmetricKeyDetails.modulusLength : 0));

// @api: crypto.Cipher
// @api: new crypto.Cipher
// @api: crypto.createCipheriv
// @api: Cipher.update
// @api: Cipher.final
// @api: Cipher.setAutoPadding
// @api: crypto.Decipher
// @api: new crypto.Decipher
// @api: crypto.createDecipheriv
// @api: Decipher.update
// @api: Decipher.final
// @api: Decipher.setAutoPadding
const aesKey = Buffer.alloc(32, 1);
const aesIv = Buffer.alloc(16, 2);
const cipher = createCipheriv("aes-256-cbc", aesKey, aesIv);
cipher.setAutoPadding(true);
const enc1 = cipher.update(Buffer.from("Hello, ScriptGo Crypto!"));
const enc2 = cipher.final();
const encrypted = Buffer.concat([enc1, enc2]);

const decipher = createDecipheriv("aes-256-cbc", aesKey, aesIv);
decipher.setAutoPadding(true);
const dec1 = decipher.update(encrypted);
const dec2 = decipher.final();
const decrypted = Buffer.concat([dec1, dec2]).toString("utf8");
// @expect: cr_cipher_cbc: Hello, ScriptGo Crypto!
console.log("cr_cipher_cbc: " + decrypted);

// @api: Cipher.setAAD
// @api: Cipher.getAuthTag
// @api: Decipher.setAAD
// @api: Decipher.setAuthTag
// @api: crypto.getCipherInfo
const gcmKey = Buffer.alloc(32, 3);
const gcmIv = Buffer.alloc(12, 4);
const cGcm = createCipheriv("aes-256-gcm", gcmKey, gcmIv);
cGcm.setAAD(Buffer.from("additional data"));
const gcmEnc1 = cGcm.update(Buffer.from("GCM authenticated payload"));
const gcmEnc2 = cGcm.final();
const gcmEnc = Buffer.concat([gcmEnc1, gcmEnc2]);
const gcmTag = cGcm.getAuthTag();

const dGcm = createDecipheriv("aes-256-gcm", gcmKey, gcmIv);
dGcm.setAAD(Buffer.from("additional data"));
dGcm.setAuthTag(gcmTag);
const gcmDec1 = dGcm.update(gcmEnc);
const gcmDec2 = dGcm.final();
const gcmDecrypted = Buffer.concat([gcmDec1, gcmDec2]).toString("utf8");
const cInfo = getCipherInfo("aes-256-gcm");
// @expect: cr_cipher_gcm: GCM authenticated payload 16 gcm
console.log("cr_cipher_gcm: " + gcmDecrypted + " " + gcmTag.length + " " + (cInfo ? cInfo.mode : ""));

// @api: crypto.Sign
// @api: new crypto.Sign
// @api: crypto.createSign
// @api: Sign.update
// @api: Sign.sign
// @api: crypto.Verify
// @api: new crypto.Verify
// @api: crypto.createVerify
// @api: Verify.update
// @api: Verify.verify
const sign = createSign("sha256");
sign.update("message to authenticate");
const sig = sign.sign(rsaKp.privateKey);

const verify = createVerify("sha256");
verify.update("message to authenticate");
const isSigValid = verify.verify(rsaKp.publicKey, sig);
// @expect: cr_sign_verify: true true
console.log("cr_sign_verify: " + (sig.length > 0) + " " + isSigValid);

// @api: crypto.publicEncrypt
// @api: crypto.privateDecrypt
// @api: crypto.privateEncrypt
// @api: crypto.publicDecrypt
const asymPlain = Buffer.from("Secret RSA Message");
const asymEnc = publicEncrypt(rsaKp.publicKey, asymPlain);
const asymDec = privateDecrypt(rsaKp.privateKey, asymEnc);
// @expect: cr_asym_encrypt: Secret RSA Message
console.log("cr_asym_encrypt: " + asymDec.toString("utf8"));

// @api: crypto.DiffieHellman
// @api: new crypto.DiffieHellman
// @api: crypto.createDiffieHellman
// @api: DiffieHellman.generateKeys
// @api: DiffieHellman.computeSecret
// @api: DiffieHellman.getPrime
// @api: DiffieHellman.getGenerator
// @api: DiffieHellman.getPublicKey
// @api: DiffieHellman.getPrivateKey
// @api: DiffieHellman.setPublicKey
// @api: DiffieHellman.setPrivateKey
// @api: DiffieHellman.verifyError
// @api: crypto.DiffieHellmanGroup
// @api: new crypto.DiffieHellmanGroup
// @api: crypto.createDiffieHellmanGroup
// @api: crypto.getDiffieHellman
const dh1 = createDiffieHellman(512);
const dh1Pub = dh1.generateKeys();
const dh2 = createDiffieHellman(dh1.getPrime(), dh1.getGenerator());
const dh2Pub = dh2.generateKeys();
const sec1 = dh1.computeSecret(dh2Pub);
const sec2 = dh2.computeSecret(dh1Pub);
const dhPrime = dh1.getPrime();
const dhGen = dh1.getGenerator();
const dhPriv = dh1.getPrivateKey();
dh1.setPublicKey(dh1Pub);
dh1.setPrivateKey(dhPriv);
const dhGroup = createDiffieHellmanGroup("modp14");
getDiffieHellman("modp14");
// @expect: cr_dh: true 0 true true
console.log("cr_dh: " + (sec1.toString("hex") === sec2.toString("hex")) + " " + dh1.verifyError + " " + (dhPrime.length > 0) + " " + (dhGen.length > 0));

// @api: crypto.ECDH
// @api: new crypto.ECDH
// @api: crypto.createECDH
// @api: ECDH.generateKeys
// @api: ECDH.computeSecret
// @api: ECDH.getPublicKey
// @api: ECDH.getPrivateKey
// @api: ECDH.setPublicKey
// @api: ECDH.setPrivateKey
const ecdh1 = createECDH("prime256v1");
const ecdh1Pub = ecdh1.generateKeys();
const ecdh2 = createECDH("prime256v1");
const ecdh2Pub = ecdh2.generateKeys();
const ecSec1 = ecdh1.computeSecret(ecdh2Pub);
const ecSec2 = ecdh2.computeSecret(ecdh1Pub);
const ecPub = ecdh1.getPublicKey();
const ecPriv = ecdh1.getPrivateKey();
ecdh1.setPublicKey(ecPub);
ecdh1.setPrivateKey(ecPriv);
// @expect: cr_ecdh: true 32
console.log("cr_ecdh: " + (ecSec1.toString("hex") === ecSec2.toString("hex")) + " " + ecSec1.length);

// @api: crypto.Certificate
// @api: new crypto.Certificate
const spkacCert = new Certificate();
// @api: X509Certificate.checkPrivateKey
// @api: X509Certificate.verify
const certKeyMatch = cert.checkPrivateKey(privObj);
const certKeyVerify = cert.verify(pubObj);
// @api: crypto.getFips
// @api: crypto.setFips
// @api: crypto.fips
// @api: crypto.secureHeapUsed
// @api: crypto.setEngine
const fipsEnabled = getFips();
setFips(false);
const heap = secureHeapUsed();
try { setEngine("nonexistent"); } catch (e) {}
// @expect: cr_misc: 0 0 object
console.log("cr_misc: " + crypto.fips + " " + fipsEnabled + " " + (typeof heap));

