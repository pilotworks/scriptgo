class Vehicle {
  public brand: string;

  constructor(brand: string) {
    this.brand = brand;
  }

  describe(): string {
    return "Vehicle(" + this.brand + ")";
  }

  maxSpeed(): number {
    return 100;
  }
}

class Car extends Vehicle {
  public doors: number;

  constructor(brand: string, doors: number) {
    super(brand);
    this.doors = doors;
  }

  describe(): string {
    return "Car(" + this.brand + ", doors=" + this.doors + ")";
  }

  maxSpeed(): number {
    return 180;
  }
}

class ElectricCar extends Car {
  public batteryCapacity: number;

  constructor(brand: string, doors: number, batteryCapacity: number) {
    super(brand, doors);
    this.batteryCapacity = batteryCapacity;
  }

  describe(): string {
    return "ElectricCar(" + this.brand + ", battery=" + this.batteryCapacity + "kWh)";
  }

  maxSpeed(): number {
    return 220;
  }
}

const v = new Vehicle("Generic");
const c = new Car("Toyota", 4);
const ec = new ElectricCar("Tesla", 4, 100);

console.log(v.describe());
console.log(v.maxSpeed());
console.log(c.describe());
console.log(c.maxSpeed());
console.log(ec.describe());
console.log(ec.maxSpeed());

// Polymorphic calls through base references
function printVehicle(veh: Vehicle): void {
  console.log("Poly: " + veh.describe() + " speed: " + veh.maxSpeed());
}

printVehicle(v);
printVehicle(c);
printVehicle(ec);

// instanceof checks across 3 levels
console.log(ec instanceof ElectricCar);
console.log(ec instanceof Car);
console.log(ec instanceof Vehicle);
console.log(c instanceof ElectricCar);
console.log(c instanceof Car);
console.log(c instanceof Vehicle);
console.log(v instanceof ElectricCar);
console.log(v instanceof Car);
console.log(v instanceof Vehicle);
