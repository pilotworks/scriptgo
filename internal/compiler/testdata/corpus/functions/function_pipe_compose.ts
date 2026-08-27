// @expect: 23
// @expect: 18
// @expect: RESULT: 23
function pipe2<A, B, C>(f1: (a: A) => B, f2: (b: B) => C): (a: A) => C {
    return (a: A): C => f2(f1(a));
}

function compose2<A, B, C>(f2: (b: B) => C, f1: (a: A) => B): (a: A) => C {
    return (a: A): C => f2(f1(a));
}

const double = (x: number): number => x * 2;
const addThree = (x: number): number => x + 3;

const doubleThenAdd = pipe2(double, addThree);
const addThenDouble = compose2(double, addThree);

console.log(doubleThenAdd(10));
console.log(addThenDouble(6));

const formatResult = (n: number): string => "RESULT: " + n;
const fullPipeline = pipe2(doubleThenAdd, formatResult);
console.log(fullPipeline(10));
