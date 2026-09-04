// @api: Intl.NumberFormat.prototype.format
// @api: Intl.DateTimeFormat.prototype.format
// @api: Intl.Collator.prototype.compare
// @api: Intl.Segmenter.prototype.segment
// @api: Intl.getCanonicalLocales
// @api: Intl.DisplayNames.prototype.of
// @api: Intl.ListFormat.prototype.format
// @api: Intl.RelativeTimeFormat.prototype.format
// @api: Intl.PluralRules.prototype.select
// @expect: 1,234,567
// @expect: -1
// @expect: {}
// @expect: Hello
// @expect: 0
// @expect: 3
// @expect: [ 'en-US' ]
// @expect: English
// @expect: Apple, Banana, and Cherry
// @expect: in 3 days
// @expect: one
// @expect: other

const nf = new Intl.NumberFormat("en-US");
console.log(nf.format(1234567));

const collator = new Intl.Collator("en");
console.log(collator.compare("a", "b"));

const segmenter = new Intl.Segmenter("en", { granularity: "word" });
console.log(segmenter.segment("Hello World"));
const segments = segmenter.segment("Hello World");
console.log(segments.containing(1)!.segment);
console.log(segments.containing(1)!.index);
let segmentCount = 0;
for (const part of segments) {
    segmentCount++;
}
console.log(segmentCount);

const locales = Intl.getCanonicalLocales("EN-us");
console.log(locales);

const dn = new Intl.DisplayNames("en-US", { type: "language" });
console.log(dn.of("en"));

const lf = new Intl.ListFormat("en");
console.log(lf.format(["Apple", "Banana", "Cherry"]));

const rtf = new Intl.RelativeTimeFormat("en");
console.log(rtf.format(3, "day"));

const pr = new Intl.PluralRules("en");
console.log(pr.select(1));
console.log(pr.select(5));
