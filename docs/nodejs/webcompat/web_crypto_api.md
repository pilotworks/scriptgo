# Web Crypto API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Web Crypto API Documentation](https://nodejs.org/docs/latest-v22.x/api/web_crypto_api.html)  
> **Type Definition Source**: [@types/node/web_crypto_api.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-web_crypto_api-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `AesCbcParams` | `(...) => any` | `__web_crypto_api.AesCbcParams` | 📋 Planned | - |
| `AesCtrParams` | `(...) => any` | `__web_crypto_api.AesCtrParams` | 📋 Planned | - |
| `AesDerivedKeyParams` | `(...) => any` | `__web_crypto_api.AesDerivedKeyParams` | 📋 Planned | - |
| `AesGcmParams` | `(...) => any` | `__web_crypto_api.AesGcmParams` | 📋 Planned | - |
| `AesKeyAlgorithm` | `(...) => any` | `__web_crypto_api.AesKeyAlgorithm` | 📋 Planned | - |
| `AesKeyGenParams` | `(...) => any` | `__web_crypto_api.AesKeyGenParams` | 📋 Planned | - |
| `Algorithm` | `(...) => any` | `__web_crypto_api.Algorithm` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__web_crypto_api.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__web_crypto_api.CryptoKey` | 📋 Planned | - |
| `CryptoKeyPair` | `(...) => any` | `__web_crypto_api.CryptoKeyPair` | 📋 Planned | - |
| `EcKeyAlgorithm` | `(...) => any` | `__web_crypto_api.EcKeyAlgorithm` | 📋 Planned | - |
| `EcKeyGenParams` | `(...) => any` | `__web_crypto_api.EcKeyGenParams` | 📋 Planned | - |
| `EcKeyImportParams` | `(...) => any` | `__web_crypto_api.EcKeyImportParams` | 📋 Planned | - |
| `EcdhKeyDeriveParams` | `(...) => any` | `__web_crypto_api.EcdhKeyDeriveParams` | 📋 Planned | - |
| `EcdsaParams` | `(...) => any` | `__web_crypto_api.EcdsaParams` | 📋 Planned | - |
| `Ed448Params` | `(...) => any` | `__web_crypto_api.Ed448Params` | 📋 Planned | - |
| `HkdfParams` | `(...) => any` | `__web_crypto_api.HkdfParams` | 📋 Planned | - |
| `HmacImportParams` | `(...) => any` | `__web_crypto_api.HmacImportParams` | 📋 Planned | - |
| `HmacKeyAlgorithm` | `(...) => any` | `__web_crypto_api.HmacKeyAlgorithm` | 📋 Planned | - |
| `HmacKeyGenParams` | `(...) => any` | `__web_crypto_api.HmacKeyGenParams` | 📋 Planned | - |
| `KeyAlgorithm` | `(...) => any` | `__web_crypto_api.KeyAlgorithm` | 📋 Planned | - |
| `Pbkdf2Params` | `(...) => any` | `__web_crypto_api.Pbkdf2Params` | 📋 Planned | - |
| `RsaHashedImportParams` | `(...) => any` | `__web_crypto_api.RsaHashedImportParams` | 📋 Planned | - |
| `RsaHashedKeyAlgorithm` | `(...) => any` | `__web_crypto_api.RsaHashedKeyAlgorithm` | 📋 Planned | - |
| `RsaHashedKeyGenParams` | `(...) => any` | `__web_crypto_api.RsaHashedKeyGenParams` | 📋 Planned | - |
| `RsaOaepParams` | `(...) => any` | `__web_crypto_api.RsaOaepParams` | 📋 Planned | - |
| `RsaPssParams` | `(...) => any` | `__web_crypto_api.RsaPssParams` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__web_crypto_api.SubtleCrypto` | 📋 Planned | - |
| `additionalData` | `any` | `__web_crypto_api.additionalData` | 📋 Planned | - |
| `context` | `any` | `__web_crypto_api.context` | 📋 Planned | - |
| `counter` | `any` | `__web_crypto_api.counter` | 📋 Planned | - |
| `crypto.getRandomValues(typedArray)` | `(...) => any` | `__web_crypto_api.crypto.getRandomValues` | 📋 Planned | - |
| `crypto.randomUUID()` | `(...) => any` | `__web_crypto_api.crypto.randomUUID` | 📋 Planned | - |
| `cryptoKey.algorithm` | `any` | `__web_crypto_api.cryptoKey.algorithm` | 📋 Planned | - |
| `extractable` | `any` | `__web_crypto_api.extractable` | 📋 Planned | - |
| `hash` | `any` | `__web_crypto_api.hash` | 📋 Planned | - |
| `info` | `any` | `__web_crypto_api.info` | 📋 Planned | - |
| `iterations` | `any` | `__web_crypto_api.iterations` | 📋 Planned | - |
| `iv` | `any` | `__web_crypto_api.iv` | 📋 Planned | - |
| `label` | `any` | `__web_crypto_api.label` | 📋 Planned | - |
| `length` | `any` | `__web_crypto_api.length` | 📋 Planned | - |
| `modulusLength` | `any` | `__web_crypto_api.modulusLength` | 📋 Planned | - |
| `name` | `any` | `__web_crypto_api.name` | 📋 Planned | - |
| `namedCurve` | `any` | `__web_crypto_api.namedCurve` | 📋 Planned | - |
| `privateKey` | `any` | `__web_crypto_api.privateKey` | 📋 Planned | - |
| `public` | `any` | `__web_crypto_api.public` | 📋 Planned | - |
| `publicExponent` | `any` | `__web_crypto_api.publicExponent` | 📋 Planned | - |
| `publicKey` | `any` | `__web_crypto_api.publicKey` | 📋 Planned | - |
| `salt` | `any` | `__web_crypto_api.salt` | 📋 Planned | - |
| `saltLength` | `any` | `__web_crypto_api.saltLength` | 📋 Planned | - |
| `subtle` | `any` | `__web_crypto_api.subtle` | 📋 Planned | - |
| `subtle.decrypt(algorithm, key, data)` | `(...) => any` | `__web_crypto_api.subtle.decrypt` | 📋 Planned | - |
| `subtle.deriveBits(algorithm, baseKey[, length])` | `(...) => any` | `__web_crypto_api.subtle.deriveBits` | 📋 Planned | - |
| `subtle.deriveKey(algorithm, baseKey, derivedKeyAlgorithm, extractable, keyUsages)` | `(...) => any` | `__web_crypto_api.subtle.deriveKey` | 📋 Planned | - |
| `subtle.digest(algorithm, data)` | `(...) => any` | `__web_crypto_api.subtle.digest` | 📋 Planned | - |
| `subtle.encrypt(algorithm, key, data)` | `(...) => any` | `__web_crypto_api.subtle.encrypt` | 📋 Planned | - |
| `subtle.exportKey(format, key)` | `(...) => any` | `__web_crypto_api.subtle.exportKey` | 📋 Planned | - |
| `subtle.generateKey(algorithm, extractable, keyUsages)` | `(...) => any` | `__web_crypto_api.subtle.generateKey` | 📋 Planned | - |
| `subtle.importKey(format, keyData, algorithm, extractable, keyUsages)` | `(...) => any` | `__web_crypto_api.subtle.importKey` | 📋 Planned | - |
| `subtle.sign(algorithm, key, data)` | `(...) => any` | `__web_crypto_api.subtle.sign` | 📋 Planned | - |
| `subtle.unwrapKey(format, wrappedKey, unwrappingKey, unwrapAlgo, unwrappedKeyAlgo, extractable, keyUsages)` | `(...) => any` | `__web_crypto_api.subtle.unwrapKey` | 📋 Planned | - |
| `subtle.verify(algorithm, key, signature, data)` | `(...) => any` | `__web_crypto_api.subtle.verify` | 📋 Planned | - |
| `subtle.wrapKey(format, key, wrappingKey, wrapAlgo)` | `(...) => any` | `__web_crypto_api.subtle.wrapKey` | 📋 Planned | - |
| `tagLength` | `any` | `__web_crypto_api.tagLength` | 📋 Planned | - |
| `type` | `any` | `__web_crypto_api.type` | 📋 Planned | - |
| `usages` | `any` | `__web_crypto_api.usages` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `web_crypto_api` are organized per API under `internal/compiler/testdata/corpus/web_crypto_api/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/web_crypto_api/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
