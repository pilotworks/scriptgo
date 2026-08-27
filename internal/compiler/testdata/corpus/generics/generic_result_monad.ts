// @expect: success: 100
// @expect: error: Division by zero
// @expect: transformed: 200
// @expect: fallback: -1
class Result<T, E> {
    private constructor(
        private readonly _isOk: boolean,
        private readonly _value?: T,
        private readonly _error?: E
    ) {}

    static ok<T, E>(val: T): Result<T, E> {
        return new Result<T, E>(true, val, undefined);
    }

    static err<T, E>(err: E): Result<T, E> {
        return new Result<T, E>(false, undefined, err);
    }

    isOk(): boolean {
        return this._isOk;
    }

    unwrap(): T {
        if (!this._isOk) {
            throw new Error("Called unwrap on Err");
        }
        return this._value!;
    }

    unwrapOr(fallback: T): T {
        return this._isOk ? this._value! : fallback;
    }

    unwrapErr(): E {
        return this._error!;
    }

    map<U>(fn: (val: T) => U): Result<U, E> {
        if (this._isOk) {
            return Result.ok<U, E>(fn(this._value!));
        }
        return Result.err<U, E>(this._error!);
    }
}

function safeDivide(a: number, b: number): Result<number, string> {
    if (b === 0) return Result.err<number, string>("Division by zero");
    return Result.ok<number, string>(a / b);
}

const r1 = safeDivide(200, 2);
if (r1.isOk()) {
    console.log("success: " + r1.unwrap());
}

const r2 = safeDivide(10, 0);
if (!r2.isOk()) {
    console.log("error: " + r2.unwrapErr());
}

const r3 = r1.map(v => v * 2);
console.log("transformed: " + r3.unwrapOr(0));
console.log("fallback: " + r2.unwrapOr(-1));
