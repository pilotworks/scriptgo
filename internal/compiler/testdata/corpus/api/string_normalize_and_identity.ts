// normalize validates its form and keeps ASCII text; toString/valueOf on a
// string primitive return the string itself.

// @expect: hello abc x
console.log("hello".normalize(), "abc".normalize("NFKD"), "x".normalize(undefined));

// @expect: st
console.log("s".toString() + "t".valueOf());

// @expect: true The normalization form should be one of NFC, NFD, NFKC, NFKD.
try {
  "a".normalize("bad");
} catch (e) {
  console.log(e instanceof RangeError, (e as Error).message);
}
