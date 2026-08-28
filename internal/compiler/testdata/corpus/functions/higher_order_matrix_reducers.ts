// @expect: Zipped: 10, 40, 90, 160, 250
// @expect: Evens: 2, 4
// @expect: Odds: 1, 3, 5
// @expect: Sum: 15
// @expect: Reversed concat: dcba
function zipWith<T, U, R>(arr1: T[], arr2: U[], fn: (a: T, b: U) => R): R[] {
    const result: R[] = [];
    const len = arr1.length < arr2.length ? arr1.length : arr2.length;
    for (let i = 0; i < len; i++) {
        result.push(fn(arr1[i], arr2[i]));
    }
    return result;
}

function partition<T>(arr: T[], predicate: (item: T) => boolean): [T[], T[]] {
    const pass: T[] = [];
    const fail: T[] = [];
    for (let i = 0; i < arr.length; i++) {
        if (predicate(arr[i])) {
            pass.push(arr[i]);
        } else {
            fail.push(arr[i]);
        }
    }
    return [pass, fail];
}

function foldl<T, R>(arr: T[], initial: R, fn: (acc: R, item: T) => R): R {
    let acc = initial;
    for (let i = 0; i < arr.length; i++) {
        acc = fn(acc, arr[i]);
    }
    return acc;
}

function foldr<T, R>(arr: T[], initial: R, fn: (acc: R, item: T) => R): R {
    let acc = initial;
    for (let i = arr.length - 1; i >= 0; i--) {
        acc = fn(acc, arr[i]);
    }
    return acc;
}

const a = [1, 2, 3, 4, 5];
const b = [10, 20, 30, 40, 50];

const zipped = zipWith(a, b, (x, y) => x * y);
console.log("Zipped: " + zipped.join(", "));

const [evens, odds] = partition(a, (x) => x % 2 === 0);
console.log("Evens: " + evens.join(", "));
console.log("Odds: " + odds.join(", "));

const sum = foldl(a, 0, (acc, x) => acc + x);
console.log("Sum: " + sum);

const reversedStr = foldr(["a", "b", "c", "d"], "", (acc, x) => acc + x);
console.log("Reversed concat: " + reversedStr);
