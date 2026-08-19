class Animal {
  name: string = "";

  constructor(name: string) {
    this.name = name;
  }

  speak(): string {
    return this.name + " makes a sound";
  }
}

class Dog extends Animal {
  breed: string = "";

  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }

  speak(): string {
    return super.speak() + " (bark from " + this.breed + ")";
  }
}

const dog = new Dog("Buddy", "Golden");
console.log(dog.speak());
