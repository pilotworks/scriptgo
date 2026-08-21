function identity<T>(x: T): T {
    return x;
}

function pair<A, B>(first: A, second: B): string {
    return `${first} - ${second}`;
}

const num = identity(42);
console.log(num);

const str = identity("scriptgo");
console.log(str);

const p = pair(123, "apple");
console.log(p);
