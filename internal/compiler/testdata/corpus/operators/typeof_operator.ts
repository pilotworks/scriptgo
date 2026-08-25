// @expect: number
// @expect: string
// @expect: boolean
// @expect: undefined
// @expect: object
// @expect: object
// @expect: function
// @expect: bigint
// @expect: symbol
console.log(typeof 123);
console.log(typeof "hello");
console.log(typeof true);
console.log(typeof undefined);
console.log(typeof { a: 1 });
console.log(typeof [1, 2, 3]);
console.log(typeof (() => {}));
console.log(typeof 100n);
console.log(typeof Symbol("test"));
