# Web Cryptography API (SubtleCrypto) Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG Web Cryptography API (SubtleCrypto) Specification](https://wintercg.org/)  
> **Type Definition Source**: [microsoft/TypeScript lib.dom.d.ts (Server subset)](https://github.com/microsoft/TypeScript/blob/main/src/lib/lib.dom.d.ts)  
> **Gate Oracle**: Web Platform Tests (WPT) & Node.js 22 LTS WPT runner

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `AesCbcParams` | `(...) => any` | `__crypto.AesCbcParams` | 📋 Planned | - |
| `AesCtrParams` | `(...) => any` | `__crypto.AesCtrParams` | 📋 Planned | - |
| `AesDerivedKeyParams` | `(...) => any` | `__crypto.AesDerivedKeyParams` | 📋 Planned | - |
| `AesGcmParams` | `(...) => any` | `__crypto.AesGcmParams` | 📋 Planned | - |
| `AesKeyAlgorithm` | `(...) => any` | `__crypto.AesKeyAlgorithm` | 📋 Planned | - |
| `AesKeyGenParams` | `(...) => any` | `__crypto.AesKeyGenParams` | 📋 Planned | - |
| `Algorithm` | `(...) => any` | `__crypto.Algorithm` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__crypto.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__crypto.CryptoKey` | 📋 Planned | - |
| `CryptoKeyPair` | `(...) => any` | `__crypto.CryptoKeyPair` | 📋 Planned | - |
| `EcKeyAlgorithm` | `(...) => any` | `__crypto.EcKeyAlgorithm` | 📋 Planned | - |
| `EcKeyGenParams` | `(...) => any` | `__crypto.EcKeyGenParams` | 📋 Planned | - |
| `EcKeyImportParams` | `(...) => any` | `__crypto.EcKeyImportParams` | 📋 Planned | - |
| `EcdhKeyDeriveParams` | `(...) => any` | `__crypto.EcdhKeyDeriveParams` | 📋 Planned | - |
| `EcdsaParams` | `(...) => any` | `__crypto.EcdsaParams` | 📋 Planned | - |
| `Ed448Params` | `(...) => any` | `__crypto.Ed448Params` | 📋 Planned | - |
| `HkdfParams` | `(...) => any` | `__crypto.HkdfParams` | 📋 Planned | - |
| `HmacImportParams` | `(...) => any` | `__crypto.HmacImportParams` | 📋 Planned | - |
| `HmacKeyAlgorithm` | `(...) => any` | `__crypto.HmacKeyAlgorithm` | 📋 Planned | - |
| `HmacKeyGenParams` | `(...) => any` | `__crypto.HmacKeyGenParams` | 📋 Planned | - |
| `KeyAlgorithm` | `(...) => any` | `__crypto.KeyAlgorithm` | 📋 Planned | - |
| `Pbkdf2Params` | `(...) => any` | `__crypto.Pbkdf2Params` | 📋 Planned | - |
| `RsaHashedImportParams` | `(...) => any` | `__crypto.RsaHashedImportParams` | 📋 Planned | - |
| `RsaHashedKeyAlgorithm` | `(...) => any` | `__crypto.RsaHashedKeyAlgorithm` | 📋 Planned | - |
| `RsaHashedKeyGenParams` | `(...) => any` | `__crypto.RsaHashedKeyGenParams` | 📋 Planned | - |
| `RsaOaepParams` | `(...) => any` | `__crypto.RsaOaepParams` | 📋 Planned | - |
| `RsaPssParams` | `(...) => any` | `__crypto.RsaPssParams` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__crypto.SubtleCrypto` | 📋 Planned | - |
| `additionalData` | `any` | `__crypto.additionalData` | 📋 Planned | - |
| `context` | `any` | `__crypto.context` | 📋 Planned | - |
| `counter` | `any` | `__crypto.counter` | 📋 Planned | - |
| `crypto.getRandomValues(typedArray)` | `(...) => any` | `__crypto.crypto.getRandomValues` | 📋 Planned | - |
| `crypto.randomUUID()` | `(...) => any` | `__crypto.crypto.randomUUID` | 📋 Planned | - |
| `cryptoKey.algorithm` | `any` | `__crypto.cryptoKey.algorithm` | 📋 Planned | - |
| `extractable` | `any` | `__crypto.extractable` | 📋 Planned | - |
| `hash` | `any` | `__crypto.hash` | 📋 Planned | - |
| `info` | `any` | `__crypto.info` | 📋 Planned | - |
| `iterations` | `any` | `__crypto.iterations` | 📋 Planned | - |
| `iv` | `any` | `__crypto.iv` | 📋 Planned | - |
| `label` | `any` | `__crypto.label` | 📋 Planned | - |
| `length` | `any` | `__crypto.length` | 📋 Planned | - |
| `modulusLength` | `any` | `__crypto.modulusLength` | 📋 Planned | - |
| `name` | `any` | `__crypto.name` | 📋 Planned | - |
| `namedCurve` | `any` | `__crypto.namedCurve` | 📋 Planned | - |
| `privateKey` | `any` | `__crypto.privateKey` | 📋 Planned | - |
| `public` | `any` | `__crypto.public` | 📋 Planned | - |
| `publicExponent` | `any` | `__crypto.publicExponent` | 📋 Planned | - |
| `publicKey` | `any` | `__crypto.publicKey` | 📋 Planned | - |
| `salt` | `any` | `__crypto.salt` | 📋 Planned | - |
| `saltLength` | `any` | `__crypto.saltLength` | 📋 Planned | - |
| `subtle` | `any` | `__crypto.subtle` | 📋 Planned | - |
| `subtle.decrypt(algorithm, key, data)` | `(...) => any` | `__crypto.subtle.decrypt` | 📋 Planned | - |
| `subtle.deriveBits(algorithm, baseKey[, length])` | `(...) => any` | `__crypto.subtle.deriveBits` | 📋 Planned | - |
| `subtle.deriveKey(algorithm, baseKey, derivedKeyAlgorithm, extractable, keyUsages)` | `(...) => any` | `__crypto.subtle.deriveKey` | 📋 Planned | - |
| `subtle.digest(algorithm, data)` | `(...) => any` | `__crypto.subtle.digest` | 📋 Planned | - |
| `subtle.encrypt(algorithm, key, data)` | `(...) => any` | `__crypto.subtle.encrypt` | 📋 Planned | - |
| `subtle.exportKey(format, key)` | `(...) => any` | `__crypto.subtle.exportKey` | 📋 Planned | - |
| `subtle.generateKey(algorithm, extractable, keyUsages)` | `(...) => any` | `__crypto.subtle.generateKey` | 📋 Planned | - |
| `subtle.importKey(format, keyData, algorithm, extractable, keyUsages)` | `(...) => any` | `__crypto.subtle.importKey` | 📋 Planned | - |
| `subtle.sign(algorithm, key, data)` | `(...) => any` | `__crypto.subtle.sign` | 📋 Planned | - |
| `subtle.unwrapKey(format, wrappedKey, unwrappingKey, unwrapAlgo, unwrappedKeyAlgo, extractable, keyUsages)` | `(...) => any` | `__crypto.subtle.unwrapKey` | 📋 Planned | - |
| `subtle.verify(algorithm, key, signature, data)` | `(...) => any` | `__crypto.subtle.verify` | 📋 Planned | - |
| `subtle.wrapKey(format, key, wrappingKey, wrapAlgo)` | `(...) => any` | `__crypto.subtle.wrapKey` | 📋 Planned | - |
| `tagLength` | `any` | `__crypto.tagLength` | 📋 Planned | - |
| `type` | `any` | `__crypto.type` | 📋 Planned | - |
| `usages` | `any` | `__crypto.usages` | 📋 Planned | - |

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
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/crypto/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
