// @expect: true
// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
class Animal {
    name: string;
    constructor(name: string) {
        this.name = name;
    }
}

class Mammal extends Animal {
    warmBlooded: boolean = true;
}

class Dog extends Mammal {
    bark(): string {
        return "Woof";
    }
}

class Bird extends Animal {
    canFly: boolean = true;
}

const dog = new Dog("Buddy");
const bird = new Bird("Tweety");

console.log(dog instanceof Dog);
console.log(dog instanceof Mammal);
console.log(dog instanceof Animal);
console.log((dog as unknown) instanceof Bird);

console.log(bird instanceof Bird);
console.log(bird instanceof Animal);
console.log((bird as unknown) instanceof Mammal);
