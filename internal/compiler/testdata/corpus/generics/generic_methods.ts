// @expect: LOG:123
// @expect: LOG:message
// @expect: LOG:true
// @expect: 4
// @expect: 1
class Serializer {
    prefix: string;

    constructor(prefix: string) {
        this.prefix = prefix;
    }

    wrap<T>(item: T): string {
        return this.prefix + ":" + String(item);
    }

    toArray<T>(...items: T[]): T[] {
        return items;
    }
}

const s = new Serializer("LOG");
console.log(s.wrap(123));
console.log(s.wrap("message"));
console.log(s.wrap(true));

const arr = s.toArray<number>(1, 2, 3, 4);
console.log(arr.length);
console.log(arr[0]);
