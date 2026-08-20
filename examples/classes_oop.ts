// Object-Oriented Programming (OOP) demo: Interfaces, Abstract Classes, Static Blocks, Getters/Setters, and Inheritance

interface Describable {
  describe(): string;
}

abstract class Animal implements Describable {
  public static registryCount: number = 0;
  public static readonly defaultCategory: string = "Fauna";

  static {
    // Class Static Initialization Block (ES2022 / TS 4.4+)
    Animal.registryCount = 100;
  }

  protected _name: string;

  constructor(name: string) {
    this._name = name;
  }

  public get name(): string {
    return this._name;
  }

  public set name(newName: string) {
    if (newName.length > 0) {
      this._name = newName;
    }
  }

  public abstract speak(): string;

  public describe(): string {
    return `[${Animal.defaultCategory}] ${this._name}: ${this.speak()}`;
  }
}

class Dog extends Animal {
  private readonly breed: string;

  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }

  public override speak(): string {
    return `${this.name} (${this.breed}) barks! Woof!`;
  }
}

console.log("=== OOP & Classes Demo ===");

const dog = new Dog("Rex", "German Shepherd");
console.log(`Dog Name: ${dog.name}`);
console.log(`Dog Speak: ${dog.speak()}`);
console.log(`Dog Describe: ${dog.describe()}`);
console.log(`Is Animal? ${dog instanceof Animal}`);
console.log(`Default Category: ${Animal.defaultCategory}`);
console.log(`Total Animal Registry Count: ${Animal.registryCount}`);
