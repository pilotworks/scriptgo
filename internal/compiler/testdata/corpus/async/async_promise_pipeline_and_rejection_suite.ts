// @expect: caught:rejected_error
// @expect: status:recovered
// @expect: 52
// 1. Promise.then value transformation chaining
Promise.resolve(21)
    .then((val: number) => val * 2)
    .then((val: number) => val + 10)
    .then((val: number) => console.log(val));

// 2. Promise.reject error propagation into .catch handler
Promise.reject(new Error("rejected_error"))
    .catch((err: Error) => {
        console.log("caught:" + err.message);
        return 0;
    });

// 3. Chained then after catch recovery
Promise.reject(new Error("initial_fail"))
    .catch(() => "recovered")
    .then((val: string) => console.log("status:" + val));
