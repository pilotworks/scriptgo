// @expect: John
// @expect: 35
// @expect: ID-1234
// @expect: John (35) [ID-1234]
class Person {
    constructor(
        public name: string,
        public age: number,
        readonly id: string
    ) {}

    getInfo(): string {
        return this.name + " (" + this.age + ") [" + this.id + "]";
    }
}

const p = new Person("John", 35, "ID-1234");
console.log(p.name);
console.log(p.age);
console.log(p.id);
console.log(p.getInfo());
