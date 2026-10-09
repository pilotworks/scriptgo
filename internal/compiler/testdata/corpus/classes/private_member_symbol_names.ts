// @expect: dollar,underscore,unicode
// Private member names that are not bare LLVM identifiers are emitted as
// quoted symbols.
class Registry {
  static #$(): string { return "dollar"; }
  static #_(): string { return "underscore"; }
  static #℘(): string { return "unicode"; }
  static names(): string { return [Registry.#$(), Registry.#_(), Registry.#℘()].join(","); }
}
console.log(Registry.names());
