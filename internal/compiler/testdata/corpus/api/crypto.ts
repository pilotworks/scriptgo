import {
    Hash,
    Hmac,
    X509Certificate,
    constants,
    subtle,
    webcrypto,
    checkPrime,
    checkPrimeSync,
    createHash,
    createHmac,
    generatePrime,
    generatePrimeSync,
    getCiphers,
    getCurves,
    getHashes,
    getRandomValues,
    hkdf,
    hkdfSync,
    pbkdf2,
    pbkdf2Sync,
    randomBytes,
    randomFill,
    randomFillSync,
    randomInt,
    randomUUID,
    scrypt,
    scryptSync,
    timingSafeEqual,
} from "node:crypto";

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
const primeOk = checkPrimeSync(3);
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
checkPrime(3, {}, (err, res) => {});
generatePrime(32, {}, (err, prime) => {});
hkdf("sha256", "ikm", "salt", "info", 32, (err, dk) => {});
pbkdf2("pass", "salt", 100, 32, "sha256", (err, dk) => {});
randomFill(Buffer.alloc(8), 0, 8);
scrypt("pass", "salt", 32, (err, dk) => {});
console.log("cr_async: true");
