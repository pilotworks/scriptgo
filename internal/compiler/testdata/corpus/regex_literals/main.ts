let re = /^[0-9]+$/;
console.log(re.source);
console.log(re.flags);
console.log(re.test("12345"));
console.log(re.test("123a"));

let reI = /hello/i;
console.log(reI.test("HELLO"));
console.log(reI.test("world"));

let text = "The quick brown fox";
console.log(text.search(/brown/));
console.log(text.search(/cat/));

let replaced = text.replace(/fox/, "dog");
console.log(replaced);

let repGlobal = "a1b2c3".replace(/[0-9]/g, "X");
console.log(repGlobal);

let newRe = new RegExp("abc", "i");
console.log(newRe.test("ABC"));
