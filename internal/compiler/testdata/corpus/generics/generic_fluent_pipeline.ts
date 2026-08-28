// @expect: 31
// @expect: ALPHA,BETA,GAMMA
class Pipeline<T> {
    private items: T[];

    constructor(items: T[]) {
        this.items = items;
    }

    map<U>(fn: (item: T) => U): Pipeline<U> {
        const result: U[] = [];
        for (const item of this.items) {
            result.push(fn(item));
        }
        return new Pipeline<U>(result);
    }

    filter(predicate: (item: T) => boolean): Pipeline<T> {
        const result: T[] = [];
        for (const item of this.items) {
            if (predicate(item)) {
                result.push(item);
            }
        }
        return new Pipeline<T>(result);
    }

    reduce<R>(reducer: (acc: R, item: T) => R, initial: R): R {
        let acc = initial;
        for (const item of this.items) {
            acc = reducer(acc, item);
        }
        return acc;
    }

    toArray(): T[] {
        return this.items;
    }
}

const rawNumbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

const finalResult = new Pipeline<number>(rawNumbers)
    .filter((n) => n % 2 === 0)
    .map((n) => `num_${n * 10}`)
    .map((str) => str.length)
    .reduce((sum, len) => sum + len, 0);

console.log(finalResult);

const stringPipeline = new Pipeline<string>(["alpha", "beta", "gamma"])
    .map((s) => s.toUpperCase())
    .toArray();

console.log(stringPipeline.join(","));
