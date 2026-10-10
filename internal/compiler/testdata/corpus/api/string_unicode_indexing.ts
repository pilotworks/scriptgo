// @expect: 11 é é ö 2 7 true
// @expect: éll éll örl wörld
// @expect: h|é|l|l|o| |w|ö|r|l|d
// @expect: üüüé a日本 233 é日😀
// @expect: 4 3 😀 b 4 3
// @expect: h
// @expect: é
// @expect: 😀
// @expect: X→Y ünï
// @expect: él hé 2 héLlo wörld 233 true
// @expect: ĐƯỜNG PHỐ HÀ NỘI ƠI ƯU đường ởi ạỳ ΣΊΣΥΦΟΣ АБВ Ÿ
// @expect: an,ăn,apple,Apple,banana,da,đá,eclair,éclair,u,ư,Zebra
// @expect: 1 1 -1
// Strings index by UTF-16 code unit like JavaScript: positions, lengths,
// split(""), padding and fromCharCode count units, for..of and spread
// iterate code points, and case mapping and localeCompare handle accented
// Latin (Vietnamese included), Greek and Cyrillic letters.
const s = "héllo wörld";
console.log(s.length, s.charAt(1), s[1], s.at(-4), s.indexOf("l"), s.lastIndexOf("ö"), s.includes("ö", 5));
console.log(s.slice(1, 4), s.substring(1, 4), s.substr(7, 3), s.slice(-5));
console.log(s.split("").join("|"));
console.log("é".padStart(4, "ü"), "a".padEnd(3, "日本"), s.charCodeAt(1), String.fromCharCode(233, 0x65e5, 0xd83d, 0xde00));
const e = "a😀b";
console.log(e.length, e.indexOf("b"), e.slice(1, 3), e.slice(3), e.split("").length, [...e].length);
for (const c of "hé😀") console.log(c);
console.log("x→y".toUpperCase(), "Ünï".toLowerCase());
console.log(s.substring(3, 1), s.substring(-2, 2), s.search("l"), s.replace("l", "L"), s.codePointAt(1), "é".localeCompare("e") > 0);
console.log("Đường phố Hà Nội Ơi Ưu".toUpperCase(), "ĐƯỜNG ỞI ẠỲ".toLowerCase(), "Σίσυφος АБВ ÿ".toUpperCase());
const words = ["banana", "Apple", "apple", "ăn", "an", "đá", "da", "Zebra", "éclair", "eclair", "ư", "u"];
console.log([...words].sort((a, b) => a.localeCompare(b)).join(","));
console.log(new Intl.Collator().compare("b", "a"), "b".localeCompare("a"), "a".localeCompare("A"));
