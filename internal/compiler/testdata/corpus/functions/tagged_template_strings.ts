// @expect: Player [Alice] scored [95] points!
// @expect: one|two|three
function tag(strings: TemplateStringsArray, ...values: (string | number)[]): string {
    let result = "";
    for (let i = 0; i < strings.length; i++) {
        result += strings[i];
        if (i < values.length) {
            result += `[${values[i]}]`;
        }
    }
    return result;
}

const name = "Alice";
const score = 95;
const output = tag`Player ${name} scored ${score} points!`;
console.log(output);

function rawTag(strings: TemplateStringsArray, ..._vals: number[]): string {
    return strings.join("|");
}

console.log(rawTag`one${1}two${2}three`);
