import {
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
    fips,
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
    generateKeySync,
    generateKeyPair,
    generateKeyPairSync,
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

// @api: crypto.crypto.Certificate
// @api: crypto.Certificate
// @api: new crypto.Certificate
// @expect: cr_cert: true
const spkOk = Certificate.verifySpkac(Buffer.alloc(0));
Certificate.exportChallenge(Buffer.alloc(0));
Certificate.exportPublicKey(Buffer.alloc(0));
console.log("cr_cert: " + spkOk);

// @api: crypto.crypto.Cipher
// @api: crypto.Cipher
// @api: new crypto.Cipher
// @api: crypto.createCipheriv
// @api: Cipher.update
// @api: Cipher.final
// @api: Cipher.setAutoPadding
// @api: Cipher.getAuthTag
// @api: Cipher.setAAD
// @expect: cr_cipher: 0 0 16
const c = createCipheriv("aes-256-gcm", Buffer.alloc(32), Buffer.alloc(12));
c.setAutoPadding(true);
c.setAAD(Buffer.alloc(0));
const u1 = c.update(Buffer.alloc(0));
const f1 = c.final();
const tag = c.getAuthTag();
console.log("cr_cipher: " + u1.length + " " + f1.length + " " + tag.length);

// @api: crypto.crypto.Decipher
// @api: crypto.Decipher
// @api: new crypto.Decipher
// @api: crypto.createDecipheriv
// @api: Decipher.update
// @api: Decipher.final
// @api: Decipher.setAutoPadding
// @api: Decipher.setAuthTag
// @api: Decipher.setAAD
// @expect: cr_decipher: 0 0
const d = createDecipheriv("aes-256-gcm", Buffer.alloc(32), Buffer.alloc(12));
d.setAutoPadding(true);
d.setAuthTag(tag);
d.setAAD(Buffer.alloc(0));
const u2 = d.update(Buffer.alloc(0));
const f2 = d.final();
console.log("cr_decipher: " + u2.length + " " + f2.length);

// @api: crypto.crypto.DiffieHellman
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
// @expect: cr_dh: 32 32 32 1 32 32 0
const dh = createDiffieHellman(256);
const dhKeys = dh.generateKeys();
const secret = dh.computeSecret(dhKeys);
const prime = dh.getPrime();
const gen = dh.getGenerator();
const dhPub = dh.getPublicKey();
const dhPriv = dh.getPrivateKey();
dh.setPublicKey(dhPub);
dh.setPrivateKey(dhPriv);
console.log("cr_dh: " + dhKeys.length + " " + secret.length + " " + prime.length + " " + gen.length + " " + dhPub.length + " " + dhPriv.length + " " + dh.verifyError);

// @api: crypto.crypto.DiffieHellmanGroup
// @api: crypto.DiffieHellmanGroup
// @api: new crypto.DiffieHellmanGroup
// @api: crypto.createDiffieHellmanGroup
// @api: crypto.getDiffieHellman
// @expect: cr_dhg: 32 32
const dhg1 = createDiffieHellmanGroup("modp14");
const dhg2 = getDiffieHellman("modp14");
console.log("cr_dhg: " + dhg1.generateKeys().length + " " + dhg2.generateKeys().length);

// @api: crypto.crypto.ECDH
// @api: crypto.ECDH
// @api: new crypto.ECDH
// @api: crypto.createECDH
// @api: ECDH.generateKeys
// @api: ECDH.computeSecret
// @api: ECDH.getPublicKey
// @api: ECDH.getPrivateKey
// @api: ECDH.setPublicKey
// @api: ECDH.setPrivateKey
// @expect: cr_ecdh: 32 32 32 32
const ecdh = createECDH("prime256v1");
const ecdhPub = ecdh.generateKeys();
const ecdhPriv = ecdh.getPrivateKey();
ecdh.setPublicKey(ecdhPub);
ecdh.setPrivateKey(ecdhPriv);
const ecdhSecret = ecdh.computeSecret(ecdhPub);
console.log("cr_ecdh: " + ecdhPub.length + " " + ecdhPriv.length + " " + ecdh.getPublicKey().length + " " + ecdhSecret.length);

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

// @api: crypto.crypto.KeyObject
// @api: crypto.KeyObject
// @api: new crypto.KeyObject
// @api: crypto.createPrivateKey
// @api: crypto.createPublicKey
// @api: crypto.createSecretKey
// @api: KeyObject.type
// @api: KeyObject.asymmetricKeyType
// @api: KeyObject.symmetricKeySize
// @api: KeyObject.asymmetricKeyDetails
// @api: KeyObject.equals
// @api: KeyObject.export
// @api: KeyObject.toCryptoKey
// @expect: cr_keyobj: secret private public true
const secKey = createSecretKey(Buffer.alloc(32));
const privKey = createPrivateKey("dummy");
const pubKey = createPublicKey("dummy");
secKey.asymmetricKeyDetails();
secKey.export();
secKey.toCryptoKey("AES-GCM", true, ["encrypt"]);
console.log("cr_keyobj: " + secKey.type + " " + privKey.type + " " + pubKey.type + " " + secKey.equals(secKey));

// @api: crypto.crypto.Sign
// @api: crypto.Sign
// @api: new crypto.Sign
// @api: crypto.createSign
// @api: Sign.update
// @api: Sign.sign
// @expect: cr_sign: 64
const s = createSign("SHA256");
s.update("hello");
const sig = s.sign(privKey);
console.log("cr_sign: " + sig.length);

// @api: crypto.crypto.Verify
// @api: crypto.Verify
// @api: new crypto.Verify
// @api: crypto.createVerify
// @api: Verify.update
// @api: Verify.verify
// @expect: cr_verify: true
const v = createVerify("SHA256");
v.update("hello");
console.log("cr_verify: " + v.verify(pubKey, sig));

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
// @api: X509Certificate.checkPrivateKey
// @api: X509Certificate.toJSON
// @api: X509Certificate.toLegacyObject
// @api: X509Certificate.toString
// @api: X509Certificate.verify
// @expect: cr_x509: false 01 CN=Test true
const cert = new X509Certificate("dummy");
cert.checkEmail("test@example.com");
cert.checkHost("example.com");
cert.checkIP("127.0.0.1");
cert.checkIssued(cert);
cert.checkPrivateKey(privKey);
cert.toJSON();
cert.toLegacyObject();
cert.toString();
cert.verify(pubKey);
console.log("cr_x509: " + cert.ca + " " + cert.serialNumber + " " + cert.subject + " " + (typeof cert.publicKey === "object"));

// @api: crypto.constants
// @api: crypto.fips
// @api: crypto.subtle
// @api: crypto.webcrypto
// @expect: cr_props: 1 false true true
console.log("cr_props: " + constants.RSA_PKCS1_PADDING + " " + fips + " " + (typeof subtle === "object") + " " + (typeof webcrypto === "object"));

// @api: crypto.checkPrimeSync
// @api: crypto.generateKeySync
// @api: crypto.generateKeyPairSync
// @api: crypto.generatePrimeSync
// @api: crypto.getCipherInfo
// @api: crypto.getCiphers
// @api: crypto.getCurves
// @api: crypto.getFips
// @api: crypto.getHashes
// @api: crypto.getRandomValues
// @api: crypto.hkdfSync
// @api: crypto.pbkdf2Sync
// @api: crypto.privateDecrypt
// @api: crypto.privateEncrypt
// @api: crypto.publicDecrypt
// @api: crypto.publicEncrypt
// @api: crypto.randomBytes
// @api: crypto.randomFillSync
// @api: crypto.randomInt
// @api: crypto.randomUUID
// @api: crypto.scryptSync
// @api: crypto.secureHeapUsed
// @api: crypto.setEngine
// @api: crypto.setFips
// @api: crypto.timingSafeEqual
// @expect: cr_sync: true secret public 3 32 32 16 true
const primeOk = checkPrimeSync(3);
const genKey = generateKeySync("hmac", {});
const keyPair = generateKeyPairSync("rsa", {});
const pSync = generatePrimeSync(32);
getCipherInfo("aes-256-gcm");
getCiphers();
getCurves();
getFips();
getHashes();
getRandomValues(new Uint8Array(4));
const hkdfRes = hkdfSync("sha256", "ikm", "salt", "info", 32);
const pbkdf2Res = pbkdf2Sync("pass", "salt", 100, 32, "sha256");
privateDecrypt(privKey, Buffer.alloc(0));
privateEncrypt(privKey, Buffer.alloc(0));
publicDecrypt(pubKey, Buffer.alloc(0));
publicEncrypt(pubKey, Buffer.alloc(0));
const rb = randomBytes(16);
randomFillSync(Buffer.alloc(8));
randomInt(0, 10);
randomUUID();
scryptSync("pass", "salt", 32);
secureHeapUsed();
setEngine("engine");
setFips(false);
const eq = timingSafeEqual(Buffer.alloc(4), Buffer.alloc(4));
console.log("cr_sync: " + primeOk + " " + genKey.type + " " + keyPair.publicKey.type + " " + pSync + " " + hkdfRes.byteLength + " " + pbkdf2Res.length + " " + rb.length + " " + eq);

// @api: crypto.checkPrime
// @api: crypto.generateKey
// @api: crypto.generateKeyPair
// @api: crypto.generatePrime
// @api: crypto.hkdf
// @api: crypto.pbkdf2
// @api: crypto.randomFill
// @api: crypto.scrypt
// @expect: cr_async: true
checkPrime(3, {}, (err, res) => {});
generateKey("hmac", {}, (err, k) => {});
generateKeyPair("rsa", {}, (err, pub, priv) => {});
generatePrime(32, {}, (err, prime) => {});
hkdf("sha256", "ikm", "salt", "info", 32, (err, dk) => {});
pbkdf2("pass", "salt", 100, 32, "sha256", (err, dk) => {});
randomFill(Buffer.alloc(8), 0, 8);
scrypt("pass", "salt", 32, (err, dk) => {});
console.log("cr_async: true");
