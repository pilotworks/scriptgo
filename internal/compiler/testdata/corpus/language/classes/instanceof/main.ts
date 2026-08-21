class Animal {
  name: string = "";
  constructor(name: string) {
    this.name = name;
  }
}

class Dog extends Animal {
  breed: string = "";
  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }
}

class Cat extends Animal {
  isLazy: boolean = true;
}

class Vehicle {
  wheels: number = 4;
}

const dog = new Dog("Buddy", "Golden");
const cat = new Cat("Whiskers");

console.log(dog instanceof Dog);
console.log(dog instanceof Animal);
console.log(dog instanceof Vehicle);

console.log(cat instanceof Cat);
console.log(cat instanceof Animal);
console.log(cat instanceof Dog);
