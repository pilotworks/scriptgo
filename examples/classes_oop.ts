// Object-Oriented Programming (OOP) demo: Classes, Static Blocks, Getters/Setters, and Inheritance

class Animal {
  static registryCount: number = 0;
  static defaultCategory: string = "Fauna";

  static {
    // Class Static Initialization Block
    Animal.registryCount = 100;
  }

  private _name: string;

  constructor(name: string) {
    this._name = name;
  }

  get name(): string {
    return this._name;
  }

  speak(): string {
    return this._name + " makes a generic sound.";
  }
}

class Dog extends Animal {
  private breed: string;

  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }

  override speak(): string {
    return this.name + " (" + this.breed + ") barks! Woof!";
  }
}

const dog = new Dog("Rex", "German Shepherd");
console.log("=== OOP & Classes Demo ===");
console.log("Dog Name: " + dog.name);
console.log("Dog Speak: " + dog.speak());
console.log("Is Animal? " + (dog instanceof Animal));
console.log("Default Category: " + Animal.defaultCategory);
console.log("Total Animal Registry Count: " + Animal.registryCount);
