// @expect: Success: 50
// @expect: Failure: Division by zero
type Ok<T> = { ok: true; value: T };
type Err<E> = { ok: false; error: E };
type Result<T, E> = Ok<T> | Err<E>;

function divide(a: number, b: number): Result<number, string> {
    if (b === 0) {
        return { ok: false, error: "Division by zero" };
    }
    return { ok: true, value: a / b };
}

function handleResult(res: Result<number, string>): void {
    if (res.ok) {
        console.log("Success: " + res.value);
    } else {
        console.log("Failure: " + res.error);
    }
}

const r1 = divide(100, 2);
handleResult(r1);

const r2 = divide(100, 0);
handleResult(r2);
