// @expect: Toyota car: Vehicle moving at 120 km/h
class Vehicle {
    speed: number;

    constructor(speed: number) {
        this.speed = speed;
    }

    describe(): string {
        return "Vehicle moving at " + this.speed + " km/h";
    }
}

class Car extends Vehicle {
    brand: string;

    constructor(speed: number, brand: string) {
        super(speed);
        this.brand = brand;
    }

    describe(): string {
        return this.brand + " car: " + super.describe();
    }
}

const c = new Car(120, "Toyota");
console.log(c.describe());
