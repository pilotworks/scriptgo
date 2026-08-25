// @expect: Alice
// @expect: 30
// @expect: 95000
interface Named {
    name: string;
}

interface Aged {
    age: number;
}

interface Employee extends Named, Aged {
    salary: number;
}

const emp: Employee = {
    name: "Alice",
    age: 30,
    salary: 95000
};

console.log(emp.name);
console.log(emp.age);
console.log(emp.salary);
