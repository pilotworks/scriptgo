// @expect: apple:10
// @expect: banana:20
// @expect: orange:30
const obj: Record<string, number> = {
    apple: 10,
    banana: 20,
    orange: 30
};

for (const k in obj) {
    console.log(k + ":" + obj[k]);
}
