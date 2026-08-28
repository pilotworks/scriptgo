// @expect: 2
// @expect: 128640
// @expect: 55357
// @expect: 56960
// @expect: Path:\project\build_2026\bin
// @expect: alpha-beta-gamma-delta-epsilon
// @expect: alpha-beta
// @expect: 0
// @expect: 2026|-|08|-|28
// @expect: hello [world]
// @expect: bonono
// @expect: foo $
// @expect: ba 2
// @expect: ba 4
// @expect: ba 6
// 1. ECMAScript UTF-16 length and codePointAt on surrogate pairs
const rocket: string = "🚀";
console.log(rocket.length);
console.log(rocket.codePointAt(0));
console.log(rocket.charCodeAt(0));
console.log(rocket.charCodeAt(1));

// 2. String.raw tagged template with interpolation and raw escapes
const year: number = 2026;
const rawOutput: string = String.raw`Path:\project\build_${year}\bin`;
console.log(rawOutput);

// 3. String.prototype.split with various limits and capture groups
const sentence: string = "alpha,beta,gamma,delta,epsilon";
console.log(sentence.split(",").join("-"));
console.log(sentence.split(",", 2).join("-"));
console.log(sentence.split(",", 0).length);

const rawDate: string = "2026-08-28";
const splitTokens: string[] = rawDate.split(/(-)/);
console.log(splitTokens.join("|"));

// 4. String.prototype.replace / replaceAll patterns
const greeting: string = "hello world";
console.log(greeting.replace("world", "[$&]"));
console.log("banana".replaceAll("a", "o"));
console.log("foo bar".replace("bar", "$$"));

// 5. Stateful RegExp execution with global /g flag
const regex: RegExp = /ba/g;
const sourceText: string = "babababa";

let matchResult: RegExpExecArray | null = regex.exec(sourceText);
console.log(matchResult ? matchResult[0] : "null", regex.lastIndex);

matchResult = regex.exec(sourceText);
console.log(matchResult ? matchResult[0] : "null", regex.lastIndex);

matchResult = regex.exec(sourceText);
console.log(matchResult ? matchResult[0] : "null", regex.lastIndex);
