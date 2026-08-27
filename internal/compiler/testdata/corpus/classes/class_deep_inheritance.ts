// @expect: Living: Alice
// @expect: Human: Alice (age: 30)
// @expect: Employee: Alice (role: Engineer)
// @expect: Base salary: 100000
// @expect: Bonus: 15000
class LivingEntity {
    name: string;
    constructor(name: string) {
        this.name = name;
    }

    describe(): string {
        return "Living: " + this.name;
    }
}

class Human extends LivingEntity {
    age: number;
    constructor(name: string, age: number) {
        super(name);
        this.age = age;
    }

    describe(): string {
        return "Human: " + this.name + " (age: " + this.age + ")";
    }
}

class Employee extends Human {
    role: string;
    salary: number;

    constructor(name: string, age: number, role: string, salary: number) {
        super(name, age);
        this.role = role;
        this.salary = salary;
    }

    describe(): string {
        return "Employee: " + this.name + " (role: " + this.role + ")";
    }

    calculateBonus(percentage: number): number {
        return (this.salary * percentage) / 100;
    }
}

const e = new Employee("Alice", 30, "Engineer", 100000);
const living: LivingEntity = new LivingEntity("Alice");
const human: LivingEntity = new Human("Alice", 30);

console.log(living.describe());
console.log(human.describe());
console.log(e.describe());
console.log("Base salary: " + e.salary);
console.log("Bonus: " + e.calculateBonus(15));
