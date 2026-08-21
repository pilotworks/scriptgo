function tag(strings: TemplateStringsArray, name: string, age: number): string {
    return strings[0] + name + strings[1] + age + strings[2];
}

const user = "Alice";
const years = 30;
const formatted = tag`User: ${user}, Age: ${years}.`;
console.log(formatted);

function simpleTag(strings: TemplateStringsArray): string {
    return strings[0] + " (tagged)";
}

console.log(simpleTag`Hello world`);
