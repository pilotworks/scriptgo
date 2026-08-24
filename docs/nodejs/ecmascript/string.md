# String Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 String Specification](https://tc39.es/ecma262/#sec-string-objects)  
> **Type Definition Source**: [microsoft/TypeScript lib.es2024.d.ts](https://github.com/microsoft/TypeScript/tree/main/src/lib)  
> **Gate Oracle**: TC39 Test262 Test Suite & TypeScript baselines

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
| `String.anchor(name: string): string` | `anchor(name: string): string` | `__string.anchor` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.at(index: number): string \| undefined` | `at(index: number): string \| undefined` | `__string.at` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.big(): string` | `big(): string` | `__string.big` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.blink(): string` | `blink(): string` | `__string.blink` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.bold(): string` | `bold(): string` | `__string.bold` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.charAt(pos: number): string` | `charAt(pos: number): string` | `__string.charAt` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.charCodeAt(index: number): number` | `charCodeAt(index: number): number` | `__string.charCodeAt` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.codePointAt(pos: number): number \| undefined` | `codePointAt(pos: number): number \| undefined` | `__string.codePointAt` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.concat(...strings: string[]): string` | `concat(...strings: string[]): string` | `__string.concat` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.endsWith(searchString: string, endPosition?: number): boolean` | `endsWith(searchString: string, endPosition?: number): boolean` | `__string.endsWith` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.fixed(): string` | `fixed(): string` | `__string.fixed` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.fontcolor(color: string): string` | `fontcolor(color: string): string` | `__string.fontcolor` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.fontsize(size: number): string` | `fontsize(size: number): string` | `__string.fontsize` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.fromCharCode(...codes: number[]): string` | `fromCharCode(...codes: number[]): string` | `__string.fromCharCode` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.fromCodePoint(...codePoints: number[]): string` | `fromCodePoint(...codePoints: number[]): string` | `__string.fromCodePoint` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.includes(searchString: string, position?: number): boolean` | `includes(searchString: string, position?: number): boolean` | `__string.includes` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.indexOf(searchString: string, position?: number): number` | `indexOf(searchString: string, position?: number): number` | `__string.indexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.isWellFormed(): boolean` | `isWellFormed(): boolean` | `__string.isWellFormed` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.italics(): string` | `italics(): string` | `__string.italics` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.lastIndexOf(searchString: string, position?: number): number` | `lastIndexOf(searchString: string, position?: number): number` | `__string.lastIndexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.link(url: string): string` | `link(url: string): string` | `__string.link` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.localeCompare(that: string, locales?: Intl.LocalesArgument, options?: Intl.CollatorOptions): number` | `localeCompare(that: string, locales?: Intl.LocalesArgument, options?: Intl.CollatorOptions): number` | `__string.localeCompare` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.match(matcher: { [Symbol.match](string: string): RegExpMatchArray \| null; }): RegExpMatchArray \| null` | `match(matcher: { [Symbol.match](string: string): RegExpMatchArray \| null; }): RegExpMatchArray \| null` | `__string.match` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.matchAll(regexp: RegExp): RegExpStringIterator<RegExpExecArray>` | `matchAll(regexp: RegExp): RegExpStringIterator<RegExpExecArray>` | `__string.matchAll` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.normalize(form: "NFC" \| "NFD" \| "NFKC" \| "NFKD"): string` | `normalize(form: "NFC" \| "NFD" \| "NFKC" \| "NFKD"): string` | `__string.normalize` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.padEnd(targetLength: number, padString?: string): string` | `padEnd(targetLength: number, padString?: string): string` | `__string.padEnd` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.padStart(targetLength: number, padString?: string): string` | `padStart(targetLength: number, padString?: string): string` | `__string.padStart` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.raw(template: { raw: readonly string[] \| ArrayLike<string>; }, ...substitutions: any[]): string` | `raw(template: { raw: readonly string[] \| ArrayLike<string>; }, ...substitutions: any[]): string` | `__string.raw` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.readonly length: number` | `readonly length: number` | `__string.length` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.repeat(count: number): string` | `repeat(count: number): string` | `__string.repeat` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.replace(searchValue: { [Symbol.replace](string: string, replaceValue: string): string; }, replaceValue: string): string` | `replace(searchValue: { [Symbol.replace](string: string, replaceValue: string): string; }, replaceValue: string): string` | `__string.replace` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.replaceAll(searchValue: string \| RegExp, replaceValue: string): string` | `replaceAll(searchValue: string \| RegExp, replaceValue: string): string` | `__string.replaceAll` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.search(searcher: { [Symbol.search](string: string): number; }): number` | `search(searcher: { [Symbol.search](string: string): number; }): number` | `__string.search` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.slice(start?: number, end?: number): string` | `slice(start?: number, end?: number): string` | `__string.slice` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.small(): string` | `small(): string` | `__string.small` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.split(splitter: { [Symbol.split](string: string, limit?: number): string[]; }, limit?: number): string[]` | `split(splitter: { [Symbol.split](string: string, limit?: number): string[]; }, limit?: number): string[]` | `__string.split` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.startsWith(searchString: string, position?: number): boolean` | `startsWith(searchString: string, position?: number): boolean` | `__string.startsWith` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.strike(): string` | `strike(): string` | `__string.strike` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.sub(): string` | `sub(): string` | `__string.sub` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.substr(from: number, length?: number): string` | `substr(from: number, length?: number): string` | `__string.substr` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.substring(start: number, end?: number): string` | `substring(start: number, end?: number): string` | `__string.substring` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.sup(): string` | `sup(): string` | `__string.sup` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toLocaleLowerCase(locales?: Intl.LocalesArgument): string` | `toLocaleLowerCase(locales?: Intl.LocalesArgument): string` | `__string.toLocaleLowerCase` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toLocaleUpperCase(locales?: Intl.LocalesArgument): string` | `toLocaleUpperCase(locales?: Intl.LocalesArgument): string` | `__string.toLocaleUpperCase` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toLowerCase(): string` | `toLowerCase(): string` | `__string.toLowerCase` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toString(): string` | `toString(): string` | `__string.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toUpperCase(): string` | `toUpperCase(): string` | `__string.toUpperCase` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.toWellFormed(): string` | `toWellFormed(): string` | `__string.toWellFormed` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.trim(): string` | `trim(): string` | `__string.trim` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.trimEnd(): string` | `trimEnd(): string` | `__string.trimEnd` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.trimLeft(): string` | `trimLeft(): string` | `__string.trimLeft` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.trimRight(): string` | `trimRight(): string` | `__string.trimRight` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.trimStart(): string` | `trimStart(): string` | `__string.trimStart` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `String.valueOf(): string` | `valueOf(): string` | `__string.valueOf` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |
| `new String(value?: any): String` | `new (value?: any): String` | `__string.new` | ✅ Done | `internal/compiler/testdata/corpus/api/string.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `string` are organized per API under `internal/compiler/testdata/corpus/string/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/string/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
