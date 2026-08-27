// @expect: [[1,"a"],[2,"b"],[3,"c"]]
// @expect: [[1,2,3],["a","b","c"]]
// @expect: ["hello",42]
function zip<T, U>(arr1: T[], arr2: U[]): [T, U][] {
    const len = Math.min(arr1.length, arr2.length);
    const result: [T, U][] = [];
    for (let i = 0; i < len; i++) {
        result.push([arr1[i], arr2[i]]);
    }
    return result;
}

function unzip<T, U>(pairs: [T, U][]): [T[], U[]] {
    const list1: T[] = [];
    const list2: U[] = [];
    for (let i = 0; i < pairs.length; i++) {
        const [first, second] = pairs[i];
        list1.push(first);
        list2.push(second);
    }
    return [list1, list2];
}

function swapTuple<T, U>(tuple: [T, U]): [U, T] {
    return [tuple[1], tuple[0]];
}

const numbers = [1, 2, 3];
const letters = ["a", "b", "c"];

const zipped = zip(numbers, letters);
console.log(JSON.stringify(zipped));

const [unzippedNums, unzippedLetters] = unzip(zipped);
console.log(JSON.stringify([unzippedNums, unzippedLetters]));

const swapped = swapTuple([42, "hello"]);
console.log(JSON.stringify(swapped));
