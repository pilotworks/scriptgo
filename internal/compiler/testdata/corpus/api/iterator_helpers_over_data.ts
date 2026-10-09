// Iterator helpers compute from the source sequence (no fixed results).

// @expect: 63,77,91
const odd = Iterator.from([7, 8, 9, 10, 11, 12, 13])
  .filter((x: number) => x % 2 === 1)
  .drop(1)
  .map((x: number) => x * 7)
  .take(3);
console.log(odd.toArray().join(","));

// @expect: ab|cd
const words = Iterator.from(["ab", "", "cd"]).filter((s: string) => s.length > 0);
console.log(words.toArray().join("|"));

// @expect: 5 false
// @expect: 6 false
// @expect: true
const it = Iterator.from([5, 6]);
const first = it.next();
console.log(first.value, first.done);
const second = it.next();
console.log(second.value, second.done);
console.log(it.next().done);

// @expect: 0
console.log(Iterator.from([1, 2]).drop(5).toArray().length);
