// @expect: 11
// @expect: 5
// @expect: 42
interface HasLength {
    length: number;
}

function getLength<T extends HasLength>(item: T): number {
    return item.length;
}

console.log(getLength("hello world"));
console.log(getLength([1, 2, 3, 4, 5]));
console.log(getLength({ length: 42 }));
