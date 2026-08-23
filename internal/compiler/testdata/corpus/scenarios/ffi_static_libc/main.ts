// @native.expected: valid pid: true
// @native.expected: fabs(-42): 42
// @native.expected: strlen("Hello"): 5
// @native.expected: Hello from native puts!

declare function getpid(): bigint;
declare function fabs(n: number): number;
declare function strlen(s: string): bigint;
declare function puts(s: string): bigint;

const pid = getpid();
console.log("valid pid:", pid > 0n);
console.log("fabs(-42):", fabs(-42));
console.log("strlen(\"Hello\"):", strlen("Hello"));
puts("Hello from native puts!");
