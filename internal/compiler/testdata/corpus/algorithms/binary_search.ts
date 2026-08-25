// @expect: 5
// @expect: 0
// @expect: 9
// @expect: -1
function binarySearch(arr: number[], target: number): number {
    let low = 0;
    let high = arr.length - 1;

    while (low <= high) {
        const mid = Math.floor((low + high) / 2);
        if (arr[mid] === target) {
            return mid;
        } else if (arr[mid] < target) {
            low = mid + 1;
        } else {
            high = mid - 1;
        }
    }
    return -1;
}

const arr = [2, 5, 8, 12, 16, 23, 38, 56, 72, 91];
console.log(binarySearch(arr, 23));
console.log(binarySearch(arr, 2));
console.log(binarySearch(arr, 91));
console.log(binarySearch(arr, 45));
