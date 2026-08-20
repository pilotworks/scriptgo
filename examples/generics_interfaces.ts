// Generics, Interfaces, and Discriminated Unions Demo

interface Identifiable<T> {
  readonly id: T;
}

interface User extends Identifiable<number> {
  readonly name: string;
  readonly role: "admin" | "member";
}

type Shape =
  | { kind: "circle"; radius: number }
  | { kind: "rectangle"; width: number; height: number };

function calculateArea(shape: Shape): number {
  switch (shape.kind) {
    case "circle":
      return Math.PI * shape.radius * shape.radius;
    case "rectangle":
      return shape.width * shape.height;
  }
}

class Repository<T extends Identifiable<number>> {
  private items: T[] = [];

  public add(item: T): void {
    this.items.push(item);
  }

  public findById(id: number): T | undefined {
    return this.items.find((item: T): boolean => item.id === id);
  }

  public getAll(): readonly T[] {
    return this.items;
  }
}

console.log("=== Generics & Unions Demo ===");

const userRepo = new Repository<User>();
userRepo.add({ id: 1, name: "Alice", role: "admin" });
userRepo.add({ id: 2, name: "Bob", role: "member" });

const user = userRepo.findById(1);
console.log(`Found user: ${user ? user.name : "none"}`);

const circle: Shape = { kind: "circle", radius: 5 };
const rect: Shape = { kind: "rectangle", width: 4, height: 6 };
console.log(`Circle area: ${calculateArea(circle)}`);
console.log(`Rectangle area: ${calculateArea(rect)}`);
