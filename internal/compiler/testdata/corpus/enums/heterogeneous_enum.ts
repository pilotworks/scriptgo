// @expect: 0
// @expect: YES
// @expect: 2
// @expect: UNKNOWN
enum MixedResult {
    No = 0,
    Yes = "YES",
    Maybe = 2,
    Unknown = "UNKNOWN"
}

console.log(MixedResult.No);
console.log(MixedResult.Yes);
console.log(MixedResult.Maybe);
console.log(MixedResult.Unknown);
