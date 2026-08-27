// @expect: 5000050000
// @expect: 120
type StepFn<T> = () => Step<T>;
type Step<T> = { done: true; value: T } | { done: false; next: StepFn<T> };

function runTrampoline<T>(initial: Step<T>): T {
    let current = initial;
    while (!current.done) {
        current = current.next();
    }
    return current.value;
}

function sumTail(n: number, acc: number = 0): Step<number> {
    if (n === 0) {
        return { done: true, value: acc };
    }
    return {
        done: false,
        next: () => sumTail(n - 1, acc + n)
    };
}

console.log(runTrampoline(sumTail(100000)));

function factTail(n: number, acc: number = 1): Step<number> {
    if (n <= 1) {
        return { done: true, value: acc };
    }
    return {
        done: false,
        next: () => factTail(n - 1, acc * n)
    };
}

console.log(runTrampoline(factTail(5)));
