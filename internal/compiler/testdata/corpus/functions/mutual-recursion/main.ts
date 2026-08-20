function isEven(n: number): boolean {
    if (n === 0) {
        return true;
    }
    if (n < 0) {
        return isEven(-n);
    }
    return isOdd(n - 1);
}

function isOdd(n: number): boolean {
    if (n === 0) {
        return false;
    }
    if (n < 0) {
        return isOdd(-n);
    }
    return isEven(n - 1);
}

console.log(isEven(0));
console.log(isEven(4));
console.log(isEven(7));
console.log(isOdd(0));
console.log(isOdd(3));
console.log(isOdd(8));
console.log(isEven(-6));
console.log(isOdd(-5));
