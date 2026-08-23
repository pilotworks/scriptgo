// ScriptGo Corpus: Iterator & IteratorObject Builtin APIs
// Consolidated test suite with inline assertions.

// @api: Iterator.from
// @api: IteratorObject.toArray
// @expect: 1,2,3
const iter_from = Iterator.from([1, 2, 3]);
console.log(iter_from.toArray().join(","));

// @api: IteratorObject.map
// @expect: 2,4,6
const iter_map = Iterator.from([1, 2, 3]).map((x: number) => x * 2);
console.log(iter_map.toArray().join(","));

// @api: IteratorObject.filter
// @expect: 2,4
const iter_filter = Iterator.from([1, 2, 3, 4]).filter((x: number) => x % 2 === 0);
console.log(iter_filter.toArray().join(","));

// @api: IteratorObject.take
// @expect: 1,2
const iter_take = Iterator.from([1, 2, 3, 4, 5]).take(2);
console.log(iter_take.toArray().join(","));

// @api: IteratorObject.drop
// @expect: 3,4,5
const iter_drop = Iterator.from([1, 2, 3, 4, 5]).drop(2);
console.log(iter_drop.toArray().join(","));

// @api: IteratorObject.flatMap
// @expect: 1,10,2,20
const iter_flatmap = Iterator.from([1, 2]).flatMap((x: number) => [x, x * 10]);
console.log(iter_flatmap.toArray().join(","));

// @api: IteratorObject.reduce
// @expect: 10
const iter_sum = Iterator.from([1, 2, 3, 4]).reduce((acc: number, x: number) => acc + x, 0);
console.log(iter_sum);

// @api: IteratorObject.some
// @expect: true
// @expect: false
console.log(Iterator.from([1, 2, 3]).some((x: number) => x === 2));
console.log(Iterator.from([1, 2, 3]).some((x: number) => x === 99));

// @api: IteratorObject.every
// @expect: true
// @expect: false
console.log(Iterator.from([2, 4, 6]).every((x: number) => x % 2 === 0));
console.log(Iterator.from([2, 3, 6]).every((x: number) => x % 2 === 0));

// @api: IteratorObject.find
// @expect: 30
console.log(Iterator.from([10, 20, 30, 40]).find((x: number) => x > 25));

// @api: IteratorObject.forEach
// @expect: 1
// @expect: 2
Iterator.from([1, 2]).forEach((x: number) => {
    console.log(x);
});

// @api: IteratorObject.next
// @expect: 10
// @expect: false
// @expect: 20
// @expect: false
// @expect: true
const it_next = Iterator.from([10, 20]);
const r1 = it_next.next();
console.log(r1.value);
console.log(r1.done);
const r2 = it_next.next();
console.log(r2.value);
console.log(r2.done);
const r3 = it_next.next();
console.log(r3.done);
