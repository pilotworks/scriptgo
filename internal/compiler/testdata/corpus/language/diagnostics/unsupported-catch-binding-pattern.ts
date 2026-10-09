// @ts-nocheck
// ScriptGo Corpus: Language Negative (catch clause binding pattern)
// @check.err: SG2005
// The caught value is unknown/any, so destructuring it has no static layout.
try {
  throw new Error("boom");
} catch ({ message }) {
  console.log(message);
}
