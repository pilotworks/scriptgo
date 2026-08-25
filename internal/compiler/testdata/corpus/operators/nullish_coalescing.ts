// @expect: default1
// @expect: default2
// @expect: 0
// @expect: 
// @expect: false
const n1: string | null = null;
const val1 = n1 ?? "default1";
const n2: string | undefined = undefined;
const val2 = n2 ?? "default2";
const n3: number | null = 0;
const val3 = n3 ?? 10;
const n4: string | null = "";
const val4 = n4 ?? "fallback";
const n5: boolean | null = false;
const val5 = n5 ?? true;

console.log(val1);
console.log(val2);
console.log(val3);
console.log(val4);
console.log(val5);
