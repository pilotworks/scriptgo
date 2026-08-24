# Crypto Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:crypto`  
> **Specification Reference**: [Node.js 22 LTS Crypto Documentation](https://nodejs.org/docs/latest-v22.x/api/crypto.html)  
> **Type Definition Source**: [@types/node/crypto.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-crypto-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:crypto`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `cipher.update(data[, inputEncoding][, outputEncoding])` | `(...) => any` | `__crypto.cipher.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `crypto.createHash(algorithm[, options])` | `(...) => any` | `__crypto.crypto.createHash` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `crypto.randomUUID([options])` | `(...) => any` | `__crypto.crypto.randomUUID` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `decipher.update(data[, inputEncoding][, outputEncoding])` | `(...) => any` | `__crypto.decipher.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `hash.digest([encoding])` | `(...) => any` | `__crypto.hash.digest` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `hash.update(data[, inputEncoding])` | `(...) => any` | `__crypto.hash.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `hmac.digest([encoding])` | `(...) => any` | `__crypto.hmac.digest` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `hmac.update(data[, inputEncoding])` | `(...) => any` | `__crypto.hmac.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `sign.update(data[, inputEncoding])` | `(...) => any` | `__crypto.sign.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `verify.update(data[, inputEncoding])` | `(...) => any` | `__crypto.verify.update` | ✅ Done | `internal/compiler/testdata/corpus/api/crypto.ts` |
| `Certificate` | `(...) => any` | `__crypto.Certificate` | 📋 Planned | - |
| `Cipher` | `(...) => any` | `__crypto.Cipher` | 📋 Planned | - |
| `Decipher` | `(...) => any` | `__crypto.Decipher` | 📋 Planned | - |
| `DiffieHellman` | `(...) => any` | `__crypto.DiffieHellman` | 📋 Planned | - |
| `DiffieHellmanGroup` | `(...) => any` | `__crypto.DiffieHellmanGroup` | 📋 Planned | - |
| `ECDH` | `(...) => any` | `__crypto.ECDH` | 📋 Planned | - |
| `Hash` | `(...) => any` | `__crypto.Hash` | 📋 Planned | - |
| `Hmac` | `(...) => any` | `__crypto.Hmac` | 📋 Planned | - |
| `KeyObject` | `(...) => any` | `__crypto.KeyObject` | 📋 Planned | - |
| `Sign` | `(...) => any` | `__crypto.Sign` | 📋 Planned | - |
| `Verify` | `(...) => any` | `__crypto.Verify` | 📋 Planned | - |
| `X509Certificate` | `(...) => any` | `__crypto.X509Certificate` | 📋 Planned | - |
| `asymmetricKeyDetails` | `any` | `__crypto.asymmetricKeyDetails` | 📋 Planned | - |
| `asymmetricKeyType` | `any` | `__crypto.asymmetricKeyType` | 📋 Planned | - |
| `ca` | `any` | `__crypto.ca` | 📋 Planned | - |
| `cipher.final([outputEncoding])` | `(...) => any` | `__crypto.cipher.final` | 📋 Planned | - |
| `cipher.getAuthTag()` | `(...) => any` | `__crypto.cipher.getAuthTag` | 📋 Planned | - |
| `cipher.setAAD(buffer[, options])` | `(...) => any` | `__crypto.cipher.setAAD` | 📋 Planned | - |
| `cipher.setAutoPadding([autoPadding])` | `(...) => any` | `__crypto.cipher.setAutoPadding` | 📋 Planned | - |
| `constants` | `any` | `__crypto.constants` | 📋 Planned | - |
| `crypto.checkPrime(candidate[, options], callback)` | `(...) => any` | `__crypto.crypto.checkPrime` | 📋 Planned | - |
| `crypto.checkPrimeSync(candidate[, options])` | `(...) => any` | `__crypto.crypto.checkPrimeSync` | 📋 Planned | - |
| `crypto.createCipheriv(algorithm, key, iv[, options])` | `(...) => any` | `__crypto.crypto.createCipheriv` | 📋 Planned | - |
| `crypto.createDecipheriv(algorithm, key, iv[, options])` | `(...) => any` | `__crypto.crypto.createDecipheriv` | 📋 Planned | - |
| `crypto.createDiffieHellman(primeLength[, generator])` | `(...) => any` | `__crypto.crypto.createDiffieHellman` | 📋 Planned | - |
| `crypto.createDiffieHellman(prime[, primeEncoding][, generator][, generatorEncoding])` | `(...) => any` | `__crypto.crypto.createDiffieHellman` | 📋 Planned | - |
| `crypto.createDiffieHellmanGroup(name)` | `(...) => any` | `__crypto.crypto.createDiffieHellmanGroup` | 📋 Planned | - |
| `crypto.createECDH(curveName)` | `(...) => any` | `__crypto.crypto.createECDH` | 📋 Planned | - |
| `crypto.createHmac(algorithm, key[, options])` | `(...) => any` | `__crypto.crypto.createHmac` | 📋 Planned | - |
| `crypto.createPrivateKey(key)` | `(...) => any` | `__crypto.crypto.createPrivateKey` | 📋 Planned | - |
| `crypto.createPublicKey(key)` | `(...) => any` | `__crypto.crypto.createPublicKey` | 📋 Planned | - |
| `crypto.createSecretKey(key[, encoding])` | `(...) => any` | `__crypto.crypto.createSecretKey` | 📋 Planned | - |
| `crypto.createSign(algorithm[, options])` | `(...) => any` | `__crypto.crypto.createSign` | 📋 Planned | - |
| `crypto.createVerify(algorithm[, options])` | `(...) => any` | `__crypto.crypto.createVerify` | 📋 Planned | - |
| `crypto.diffieHellman(options)` | `(...) => any` | `__crypto.crypto.diffieHellman` | 📋 Planned | - |
| `crypto.fips` | `any` | `__crypto.crypto.fips` | 📋 Planned | - |
| `crypto.generateKey(type, options, callback)` | `(...) => any` | `__crypto.crypto.generateKey` | 📋 Planned | - |
| `crypto.generateKeyPair(type, options, callback)` | `(...) => any` | `__crypto.crypto.generateKeyPair` | 📋 Planned | - |
| `crypto.generateKeyPairSync(type, options)` | `(...) => any` | `__crypto.crypto.generateKeyPairSync` | 📋 Planned | - |
| `crypto.generateKeySync(type, options)` | `(...) => any` | `__crypto.crypto.generateKeySync` | 📋 Planned | - |
| `crypto.generatePrime(size[, options], callback)` | `(...) => any` | `__crypto.crypto.generatePrime` | 📋 Planned | - |
| `crypto.generatePrimeSync(size[, options])` | `(...) => any` | `__crypto.crypto.generatePrimeSync` | 📋 Planned | - |
| `crypto.getCipherInfo(nameOrNid[, options])` | `(...) => any` | `__crypto.crypto.getCipherInfo` | 📋 Planned | - |
| `crypto.getCiphers()` | `(...) => any` | `__crypto.crypto.getCiphers` | 📋 Planned | - |
| `crypto.getCurves()` | `(...) => any` | `__crypto.crypto.getCurves` | 📋 Planned | - |
| `crypto.getDiffieHellman(groupName)` | `(...) => any` | `__crypto.crypto.getDiffieHellman` | 📋 Planned | - |
| `crypto.getFips()` | `(...) => any` | `__crypto.crypto.getFips` | 📋 Planned | - |
| `crypto.getHashes()` | `(...) => any` | `__crypto.crypto.getHashes` | 📋 Planned | - |
| `crypto.getRandomValues(typedArray)` | `(...) => any` | `__crypto.crypto.getRandomValues` | 📋 Planned | - |
| `crypto.hash(algorithm, data[, outputEncoding])` | `(...) => any` | `__crypto.crypto.hash` | 📋 Planned | - |
| `crypto.hkdf(digest, ikm, salt, info, keylen, callback)` | `(...) => any` | `__crypto.crypto.hkdf` | 📋 Planned | - |
| `crypto.hkdfSync(digest, ikm, salt, info, keylen)` | `(...) => any` | `__crypto.crypto.hkdfSync` | 📋 Planned | - |
| `crypto.pbkdf2(password, salt, iterations, keylen, digest, callback)` | `(...) => any` | `__crypto.crypto.pbkdf2` | 📋 Planned | - |
| `crypto.pbkdf2Sync(password, salt, iterations, keylen, digest)` | `(...) => any` | `__crypto.crypto.pbkdf2Sync` | 📋 Planned | - |
| `crypto.privateDecrypt(privateKey, buffer)` | `(...) => any` | `__crypto.crypto.privateDecrypt` | 📋 Planned | - |
| `crypto.privateEncrypt(privateKey, buffer)` | `(...) => any` | `__crypto.crypto.privateEncrypt` | 📋 Planned | - |
| `crypto.publicDecrypt(key, buffer)` | `(...) => any` | `__crypto.crypto.publicDecrypt` | 📋 Planned | - |
| `crypto.publicEncrypt(key, buffer)` | `(...) => any` | `__crypto.crypto.publicEncrypt` | 📋 Planned | - |
| `crypto.randomBytes(size[, callback])` | `(...) => any` | `__crypto.crypto.randomBytes` | 📋 Planned | - |
| `crypto.randomFill(buffer[, offset][, size], callback)` | `(...) => any` | `__crypto.crypto.randomFill` | 📋 Planned | - |
| `crypto.randomFillSync(buffer[, offset][, size])` | `(...) => any` | `__crypto.crypto.randomFillSync` | 📋 Planned | - |
| `crypto.randomInt([min, ]max[, callback])` | `(...) => any` | `__crypto.crypto.randomInt` | 📋 Planned | - |
| `crypto.scrypt(password, salt, keylen[, options], callback)` | `(...) => any` | `__crypto.crypto.scrypt` | 📋 Planned | - |
| `crypto.scryptSync(password, salt, keylen[, options])` | `(...) => any` | `__crypto.crypto.scryptSync` | 📋 Planned | - |
| `crypto.secureHeapUsed()` | `(...) => any` | `__crypto.crypto.secureHeapUsed` | 📋 Planned | - |
| `crypto.setEngine(engine[, flags])` | `(...) => any` | `__crypto.crypto.setEngine` | 📋 Planned | - |
| `crypto.setFips(bool)` | `(...) => any` | `__crypto.crypto.setFips` | 📋 Planned | - |
| `crypto.sign(algorithm, data, key[, callback])` | `(...) => any` | `__crypto.crypto.sign` | 📋 Planned | - |
| `crypto.timingSafeEqual(a, b)` | `(...) => any` | `__crypto.crypto.timingSafeEqual` | 📋 Planned | - |
| `crypto.verify(algorithm, data, key, signature[, callback])` | `(...) => any` | `__crypto.crypto.verify` | 📋 Planned | - |
| `crypto.webcrypto` | `any` | `__crypto.crypto.webcrypto` | 📋 Planned | - |
| `decipher.final([outputEncoding])` | `(...) => any` | `__crypto.decipher.final` | 📋 Planned | - |
| `decipher.setAAD(buffer[, options])` | `(...) => any` | `__crypto.decipher.setAAD` | 📋 Planned | - |
| `decipher.setAuthTag(buffer[, encoding])` | `(...) => any` | `__crypto.decipher.setAuthTag` | 📋 Planned | - |
| `decipher.setAutoPadding([autoPadding])` | `(...) => any` | `__crypto.decipher.setAutoPadding` | 📋 Planned | - |
| `diffieHellman.computeSecret(otherPublicKey[, inputEncoding][, outputEncoding])` | `(...) => any` | `__crypto.diffieHellman.computeSecret` | 📋 Planned | - |
| `diffieHellman.generateKeys([encoding])` | `(...) => any` | `__crypto.diffieHellman.generateKeys` | 📋 Planned | - |
| `diffieHellman.getGenerator([encoding])` | `(...) => any` | `__crypto.diffieHellman.getGenerator` | 📋 Planned | - |
| `diffieHellman.getPrime([encoding])` | `(...) => any` | `__crypto.diffieHellman.getPrime` | 📋 Planned | - |
| `diffieHellman.getPrivateKey([encoding])` | `(...) => any` | `__crypto.diffieHellman.getPrivateKey` | 📋 Planned | - |
| `diffieHellman.getPublicKey([encoding])` | `(...) => any` | `__crypto.diffieHellman.getPublicKey` | 📋 Planned | - |
| `diffieHellman.setPrivateKey(privateKey[, encoding])` | `(...) => any` | `__crypto.diffieHellman.setPrivateKey` | 📋 Planned | - |
| `diffieHellman.setPublicKey(publicKey[, encoding])` | `(...) => any` | `__crypto.diffieHellman.setPublicKey` | 📋 Planned | - |
| `diffieHellman.verifyError` | `any` | `__crypto.diffieHellman.verifyError` | 📋 Planned | - |
| `ecdh.computeSecret(otherPublicKey[, inputEncoding][, outputEncoding])` | `(...) => any` | `__crypto.ecdh.computeSecret` | 📋 Planned | - |
| `ecdh.generateKeys([encoding[, format]])` | `(...) => any` | `__crypto.ecdh.generateKeys` | 📋 Planned | - |
| `ecdh.getPrivateKey([encoding])` | `(...) => any` | `__crypto.ecdh.getPrivateKey` | 📋 Planned | - |
| `ecdh.getPublicKey([encoding][, format])` | `(...) => any` | `__crypto.ecdh.getPublicKey` | 📋 Planned | - |
| `ecdh.setPrivateKey(privateKey[, encoding])` | `(...) => any` | `__crypto.ecdh.setPrivateKey` | 📋 Planned | - |
| `ecdh.setPublicKey(publicKey[, encoding])` | `(...) => any` | `__crypto.ecdh.setPublicKey` | 📋 Planned | - |
| `fingerprint` | `any` | `__crypto.fingerprint` | 📋 Planned | - |
| `fingerprint256` | `any` | `__crypto.fingerprint256` | 📋 Planned | - |
| `fingerprint512` | `any` | `__crypto.fingerprint512` | 📋 Planned | - |
| `hash.copy([options])` | `(...) => any` | `__crypto.hash.copy` | 📋 Planned | - |
| `infoAccess` | `any` | `__crypto.infoAccess` | 📋 Planned | - |
| `issuer` | `any` | `__crypto.issuer` | 📋 Planned | - |
| `issuerCertificate` | `any` | `__crypto.issuerCertificate` | 📋 Planned | - |
| `keyObject.equals(otherKeyObject)` | `(...) => any` | `__crypto.keyObject.equals` | 📋 Planned | - |
| `keyObject.export([options])` | `(...) => any` | `__crypto.keyObject.export` | 📋 Planned | - |
| `keyObject.toCryptoKey(algorithm, extractable, keyUsages)` | `(...) => any` | `__crypto.keyObject.toCryptoKey` | 📋 Planned | - |
| `keyUsage` | `any` | `__crypto.keyUsage` | 📋 Planned | - |
| `publicKey` | `any` | `__crypto.publicKey` | 📋 Planned | - |
| `raw` | `any` | `__crypto.raw` | 📋 Planned | - |
| `serialNumber` | `any` | `__crypto.serialNumber` | 📋 Planned | - |
| `sign.sign(privateKey[, outputEncoding])` | `(...) => any` | `__crypto.sign.sign` | 📋 Planned | - |
| `subject` | `any` | `__crypto.subject` | 📋 Planned | - |
| `subjectAltName` | `any` | `__crypto.subjectAltName` | 📋 Planned | - |
| `subtle` | `any` | `__crypto.subtle` | 📋 Planned | - |
| `symmetricKeySize` | `any` | `__crypto.symmetricKeySize` | 📋 Planned | - |
| `type` | `any` | `__crypto.type` | 📋 Planned | - |
| `validFrom` | `any` | `__crypto.validFrom` | 📋 Planned | - |
| `validFromDate` | `any` | `__crypto.validFromDate` | 📋 Planned | - |
| `validTo` | `any` | `__crypto.validTo` | 📋 Planned | - |
| `validToDate` | `any` | `__crypto.validToDate` | 📋 Planned | - |
| `verify.verify(object, signature[, signatureEncoding])` | `(...) => any` | `__crypto.verify.verify` | 📋 Planned | - |
| `x509.checkEmail(email[, options])` | `(...) => any` | `__crypto.x509.checkEmail` | 📋 Planned | - |
| `x509.checkHost(name[, options])` | `(...) => any` | `__crypto.x509.checkHost` | 📋 Planned | - |
| `x509.checkIP(ip)` | `(...) => any` | `__crypto.x509.checkIP` | 📋 Planned | - |
| `x509.checkIssued(otherCert)` | `(...) => any` | `__crypto.x509.checkIssued` | 📋 Planned | - |
| `x509.checkPrivateKey(privateKey)` | `(...) => any` | `__crypto.x509.checkPrivateKey` | 📋 Planned | - |
| `x509.toJSON()` | `(...) => any` | `__crypto.x509.toJSON` | 📋 Planned | - |
| `x509.toLegacyObject()` | `(...) => any` | `__crypto.x509.toLegacyObject` | 📋 Planned | - |
| `x509.toString()` | `(...) => any` | `__crypto.x509.toString` | 📋 Planned | - |
| `x509.verify(publicKey)` | `(...) => any` | `__crypto.x509.verify` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `crypto` are organized per API under `internal/compiler/testdata/corpus/crypto/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/crypto/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
