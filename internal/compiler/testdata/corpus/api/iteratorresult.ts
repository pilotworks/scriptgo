// ScriptGo Corpus: IteratorResult Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: iteratorresult.done
// @api: iteratorresult.value
// @expect: false
// @expect: 42
interface IteratorResult<T> {
    done: boolean;
    value: T;
}

const res: IteratorResult<number> = { done: false, value: 42 };
console.log(res.done);
console.log(res.value);

