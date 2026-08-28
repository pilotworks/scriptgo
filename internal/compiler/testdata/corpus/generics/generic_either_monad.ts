// @expect: Success: 50
// @expect: Error: Division by zero
// @expect: Error: Invalid number: invalid
abstract class Either<L, R> {
    abstract isLeft(): boolean;
    abstract isRight(): boolean;
    abstract map<U>(fn: (val: R) => U): Either<L, U>;
    abstract flatMap<U>(fn: (val: R) => Either<L, U>): Either<L, U>;
    abstract fold<T>(onLeft: (left: L) => T, onRight: (right: R) => T): T;
}

class Left<L, R> extends Either<L, R> {
    private value: L;

    constructor(value: L) {
        super();
        this.value = value;
    }

    isLeft(): boolean { return true; }
    isRight(): boolean { return false; }

    map<U>(fn: (val: R) => U): Either<L, U> {
        return new Left<L, U>(this.value);
    }

    flatMap<U>(fn: (val: R) => Either<L, U>): Either<L, U> {
        return new Left<L, U>(this.value);
    }

    fold<T>(onLeft: (left: L) => T, onRight: (right: R) => T): T {
        return onLeft(this.value);
    }
}

class Right<L, R> extends Either<L, R> {
    private value: R;

    constructor(value: R) {
        super();
        this.value = value;
    }

    isLeft(): boolean { return false; }
    isRight(): boolean { return true; }

    map<U>(fn: (val: R) => U): Either<L, U> {
        return new Right<L, U>(fn(this.value));
    }

    flatMap<U>(fn: (val: R) => Either<L, U>): Either<L, U> {
        return fn(this.value);
    }

    fold<T>(onLeft: (left: L) => T, onRight: (right: R) => T): T {
        return onRight(this.value);
    }
}

function safeDivide(a: number, b: number): Either<string, number> {
    if (b === 0) {
        return new Left<string, number>("Division by zero");
    }
    return new Right<string, number>(a / b);
}

function parseNumber(str: string): Either<string, number> {
    const val = Number(str);
    if (isNaN(val)) {
        return new Left<string, number>(`Invalid number: ${str}`);
    }
    return new Right<string, number>(val);
}

const res1 = parseNumber("100")
    .flatMap(num => safeDivide(num, 4))
    .map(val => val * 2);

console.log(res1.fold(err => `Error: ${err}`, val => `Success: ${val}`));

const res2 = parseNumber("100")
    .flatMap(num => safeDivide(num, 0))
    .map(val => val * 2);

console.log(res2.fold(err => `Error: ${err}`, val => `Success: ${val}`));

const res3 = parseNumber("invalid")
    .flatMap(num => safeDivide(num, 2))
    .map(val => val * 2);

console.log(res3.fold(err => `Error: ${err}`, val => `Success: ${val}`));
