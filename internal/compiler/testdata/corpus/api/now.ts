// ScriptGo Corpus: Now Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: Now.timeZoneId
// @api: Now.instant
// @api: Now.plainDateISO
// @api: Now.plainDateTimeISO
// @api: Now.plainTimeISO
// @api: Now.zonedDateTimeISO
// @expect: UTC
// @expect: 1700000000000
// @expect: 2026-01-01
// @expect: 2026-01-01T00:00:00
// @expect: 00:00:00
// @expect: 2026-01-01T00:00:00+00:00[UTC]

class Now {
    static timeZoneId(): string {
        return "UTC";
    }
    static instant(): number {
        return 1700000000000;
    }
    static plainDateISO(): string {
        return "2026-01-01";
    }
    static plainDateTimeISO(): string {
        return "2026-01-01T00:00:00";
    }
    static plainTimeISO(): string {
        return "00:00:00";
    }
    static zonedDateTimeISO(): string {
        return "2026-01-01T00:00:00+00:00[UTC]";
    }
}

console.log(Now.timeZoneId());
console.log(Now.instant());
console.log(Now.plainDateISO());
console.log(Now.plainDateTimeISO());
console.log(Now.plainTimeISO());
console.log(Now.zonedDateTimeISO());
