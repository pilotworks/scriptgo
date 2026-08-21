const re = /foo/;
const m = re.exec("foobar") as string[];
console.log(m[0]);
