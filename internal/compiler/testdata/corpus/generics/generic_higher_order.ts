// @expect: 4
// @expect: 16
// @expect: 36
function mapList<T, U>(list: T[], fn: (item: T) => U): U[] {
    const result: U[] = [];
    for (const item of list) {
        result.push(fn(item));
    }
    return result;
}

function filterList<T>(list: T[], predicate: (item: T) => boolean): T[] {
    const result: T[] = [];
    for (const item of list) {
        if (predicate(item)) {
            result.push(item);
        }
    }
    return result;
}

const nums = [1, 2, 3, 4, 5, 6];
const evens = filterList(nums, n => n % 2 === 0);
const squares = mapList(evens, n => n * n);

for (const s of squares) {
    console.log(s);
}
