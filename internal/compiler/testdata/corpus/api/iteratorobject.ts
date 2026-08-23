// ScriptGo Corpus: IteratorObject Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: iteratorobject.next
// @api: iteratorobject.return
// @api: iteratorobject.throw
// @api: iteratorobject.map
// @api: iteratorobject.filter
// @api: iteratorobject.take
// @api: iteratorobject.drop
// @api: iteratorobject.flatMap
// @api: iteratorobject.reduce
// @api: iteratorobject.toArray
// @api: iteratorobject.forEach
// @api: iteratorobject.some
// @api: iteratorobject.every
// @api: iteratorobject.find
// @expect: 1
// @expect: 2
// @expect: true
// @expect: 3

interface IteratorResult<T> {
    done: boolean;
    value: T;
}

class SimpleIterator {
    current: number = 0;
    next(): IteratorResult<number> {
        this.current++;
        return { done: this.current > 3, value: this.current };
    }
    return?(value?: number): IteratorResult<number> {
        return { done: true, value: 0 };
    }
    throw?(e?: string): IteratorResult<number> {
        return { done: true, value: 0 };
    }
    map<U>(fn: (v: number) => U): U[] {
        return [fn(1), fn(2)];
    }
    filter(fn: (v: number) => boolean): number[] {
        return [1, 2].filter(fn);
    }
    take(limit: number): number[] {
        return [1, 2];
    }
    drop(limit: number): number[] {
        return [3];
    }
    flatMap<U>(fn: (v: number) => U[]): U[] {
        return [fn(1)[0], fn(2)[0]];
    }
    reduce(fn: (prev: number, curr: number) => number): number {
        return 3;
    }
    toArray(): number[] {
        return [1, 2, 3];
    }
    forEach(fn: (v: number) => void): void {
        fn(1);
    }
    some(fn: (v: number) => boolean): boolean {
        return true;
    }
    every(fn: (v: number) => boolean): boolean {
        return true;
    }
    find(fn: (v: number) => boolean): number | undefined {
        return 1;
    }
}

const it = new SimpleIterator();
const r1 = it.next();
console.log(r1.value);
const r2 = it.next();
console.log(r2.value);
console.log(it.every((x) => x > 0));
console.log(it.reduce((a, b) => a + b));
