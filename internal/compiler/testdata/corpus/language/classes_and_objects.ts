// ScriptGo Corpus: Language: Classes & Objects
// Consolidated test suite with inline assertions.

// --- Context Case: language_classes_abstract-interfaces ---
// @expect: Shape Box has area: 16
interface Shape_classes_abstract_interfaces_0 {
  area(): number;
}

abstract class BaseShape_classes_abstract_interfaces_0 implements Shape_classes_abstract_interfaces_0 {
  name: string = "";

  constructor(name: string) {
    this.name = name;
  }

  abstract area(): number;

  describe(): string {
    return "Shape " + this.name + " has area: " + this.area().toString();
  }
}

class Square_classes_abstract_interfaces_0 extends BaseShape_classes_abstract_interfaces_0 {
  side: number = 0;

  constructor(name: string, side: number) {
    super(name);
    this.side = side;
  }

  area(): number {
    return this.side * this.side;
  }
}

const sq_classes_abstract_interfaces_0 = new Square_classes_abstract_interfaces_0("Box", 4);
console.log(sq_classes_abstract_interfaces_0.describe());

// --- Context Case: language_classes_access-modifiers ---
// @expect: Alice
// @expect: 150
class Account_classes_access_modifiers_1 {
  public owner: string = "";
  private balance: number = 0;

  constructor(owner: string, initial: number) {
    this.owner = owner;
    this.balance = initial;
  }

  public deposit(amount: number): void {
    this.balance = this.balance + amount;
  }

  public getBalance(): number {
    return this.balance;
  }
}

const acc_classes_access_modifiers_1 = new Account_classes_access_modifiers_1("Alice", 100);
acc_classes_access_modifiers_1.deposit(50);
console.log(acc_classes_access_modifiers_1.owner);
console.log(acc_classes_access_modifiers_1.getBalance());

// --- Context Case: language_classes_constructor-params ---
// @expect: 42
class Point_classes_constructor_params_2 {
  x: number = 0;
  y: number = 0;

  constructor(x: number, y: number) {
    this.x = x;
    this.y = y;
  }

  sum(): number {
    return this.x + this.y;
  }
}

const p_classes_constructor_params_2: Point_classes_constructor_params_2 = new Point_classes_constructor_params_2(15, 27);
console.log(p_classes_constructor_params_2.sum());

// --- Context Case: language_classes_counter-state ---
// @expect: 10
// @expect: 11
// @expect: 12
// @expect: 11
// @expect: 11
class Counter_classes_counter_state_3 {
    private count: number;
    constructor(initial: number = 0) {
        this.count = initial;
    }
    increment(): number {
        this.count++;
        return this.count;
    }
    decrement(): number {
        this.count--;
        return this.count;
    }
    getValue(): number {
        return this.count;
    }
}

const c_classes_counter_state_3 = new Counter_classes_counter_state_3(10);
console.log(c_classes_counter_state_3.getValue());
console.log(c_classes_counter_state_3.increment());
console.log(c_classes_counter_state_3.increment());
console.log(c_classes_counter_state_3.decrement());
console.log(c_classes_counter_state_3.getValue());

// --- Context Case: language_classes_getter-setter-computed ---
// @expect: 10
// @expect: 20
// @expect: 200
// @expect: 60
// @expect: false
// @expect: 20
// @expect: 400
// @expect: true
// @expect: 20
// @expect: 400
class Rectangle_classes_getter_setter_computed_4 {
  private _width: number;
  private _height: number;

  constructor(w: number, h: number) {
    this._width = w;
    this._height = h;
  }

  get width(): number {
    return this._width;
  }

  set width(val: number) {
    if (val > 0) {
      this._width = val;
    }
  }

  get height(): number {
    return this._height;
  }

  set height(val: number) {
    if (val > 0) {
      this._height = val;
    }
  }

  get area(): number {
    return this._width * this._height;
  }

  get perimeter(): number {
    return 2 * (this._width + this._height);
  }

  get isSquare(): boolean {
    return this._width === this._height;
  }
}

const rect_classes_getter_setter_computed_4 = new Rectangle_classes_getter_setter_computed_4(10, 20);
console.log(rect_classes_getter_setter_computed_4.width);
console.log(rect_classes_getter_setter_computed_4.height);
console.log(rect_classes_getter_setter_computed_4.area);
console.log(rect_classes_getter_setter_computed_4.perimeter);
console.log(rect_classes_getter_setter_computed_4.isSquare);

rect_classes_getter_setter_computed_4.width = 20;
console.log(rect_classes_getter_setter_computed_4.width);
console.log(rect_classes_getter_setter_computed_4.area);
console.log(rect_classes_getter_setter_computed_4.isSquare);

// Invalid negative update ignored
rect_classes_getter_setter_computed_4.height = -5;
console.log(rect_classes_getter_setter_computed_4.height);
console.log(rect_classes_getter_setter_computed_4.area);

// --- Context Case: language_classes_getters-setters ---
// @expect: 50
// @expect: 80
class Rectangle_classes_getters_setters_5 {
  w: number = 0;
  h: number = 0;

  constructor(w: number, h: number) {
    this.w = w;
    this.h = h;
  }

  get area(): number {
    return this.w * this.h;
  }

  set width(val: number) {
    this.w = val;
  }
}

const rect_classes_getters_setters_5 = new Rectangle_classes_getters_setters_5(5, 10);
console.log(rect_classes_getters_setters_5.area);
rect_classes_getters_setters_5.width = 8;
console.log(rect_classes_getters_setters_5.area);

// --- Context Case: language_classes_inheritance ---
// @expect: Buddy makes a sound (bark from Golden)
class Animal_classes_inheritance_6 {
  name: string = "";

  constructor(name: string) {
    this.name = name;
  }

  speak(): string {
    return this.name + " makes a sound";
  }
}

class Dog_classes_inheritance_6 extends Animal_classes_inheritance_6 {
  breed: string = "";

  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }

  speak(): string {
    return super.speak() + " (bark from " + this.breed + ")";
  }
}

const dog_classes_inheritance_6 = new Dog_classes_inheritance_6("Buddy", "Golden");
console.log(dog_classes_inheritance_6.speak());

// --- Context Case: language_classes_instanceof ---
// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
class Animal_classes_instanceof_7 {
  name: string = "";
  constructor(name: string) {
    this.name = name;
  }
}

class Dog_classes_instanceof_7 extends Animal_classes_instanceof_7 {
  breed: string = "";
  constructor(name: string, breed: string) {
    super(name);
    this.breed = breed;
  }
}

class Cat_classes_instanceof_7 extends Animal_classes_instanceof_7 {
  isLazy: boolean = true;
}

class Vehicle_classes_instanceof_7 {
  wheels: number = 4;
}

const dog_classes_instanceof_7 = new Dog_classes_instanceof_7("Buddy", "Golden");
const cat_classes_instanceof_7 = new Cat_classes_instanceof_7("Whiskers");

console.log(dog_classes_instanceof_7 instanceof Dog_classes_instanceof_7);
console.log(dog_classes_instanceof_7 instanceof Animal_classes_instanceof_7);
console.log(dog_classes_instanceof_7 instanceof Vehicle_classes_instanceof_7);

console.log(cat_classes_instanceof_7 instanceof Cat_classes_instanceof_7);
console.log(cat_classes_instanceof_7 instanceof Animal_classes_instanceof_7);
console.log(cat_classes_instanceof_7 instanceof Dog_classes_instanceof_7);

// --- Context Case: language_classes_methods ---
// @expect: 42
class Calculator_classes_methods_8 {
  value: number = 10;

  add(n: number): number {
    return this.value + n;
  }
}

const c_classes_methods_8: Calculator_classes_methods_8 = new Calculator_classes_methods_8();
console.log(c_classes_methods_8.add(32));

// --- Context Case: language_classes_multi-level-inheritance ---
// @expect: Vehicle(Generic)
// @expect: 100
// @expect: Car(Toyota, doors=4)
// @expect: 180
// @expect: ElectricCar(Tesla, battery=100kWh)
// @expect: 220
// @expect: Poly: Vehicle(Generic) speed: 100
// @expect: Poly: Vehicle(Toyota) speed: 100
// @expect: Poly: Vehicle(Tesla) speed: 100
// @expect: true
// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
// @expect: false
// @expect: true
class Vehicle_classes_multi_level_inheritance_9 {
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

class Car_classes_multi_level_inheritance_9 extends Vehicle_classes_multi_level_inheritance_9 {
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

class ElectricCar_classes_multi_level_inheritance_9 extends Car_classes_multi_level_inheritance_9 {
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

const v_classes_multi_level_inheritance_9 = new Vehicle_classes_multi_level_inheritance_9("Generic");
const c_classes_multi_level_inheritance_9 = new Car_classes_multi_level_inheritance_9("Toyota", 4);
const ec_classes_multi_level_inheritance_9 = new ElectricCar_classes_multi_level_inheritance_9("Tesla", 4, 100);

console.log(v_classes_multi_level_inheritance_9.describe());
console.log(v_classes_multi_level_inheritance_9.maxSpeed());
console.log(c_classes_multi_level_inheritance_9.describe());
console.log(c_classes_multi_level_inheritance_9.maxSpeed());
console.log(ec_classes_multi_level_inheritance_9.describe());
console.log(ec_classes_multi_level_inheritance_9.maxSpeed());

// Polymorphic calls through base references
function printVehicle_classes_multi_level_inheritance_9(veh: Vehicle_classes_multi_level_inheritance_9): void {
  console.log("Poly: " + veh.describe() + " speed: " + veh.maxSpeed());
}

printVehicle_classes_multi_level_inheritance_9(v_classes_multi_level_inheritance_9);
printVehicle_classes_multi_level_inheritance_9(c_classes_multi_level_inheritance_9);
printVehicle_classes_multi_level_inheritance_9(ec_classes_multi_level_inheritance_9);

// instanceof checks across 3 levels
console.log(ec_classes_multi_level_inheritance_9 instanceof ElectricCar_classes_multi_level_inheritance_9);
console.log(ec_classes_multi_level_inheritance_9 instanceof Car_classes_multi_level_inheritance_9);
console.log(ec_classes_multi_level_inheritance_9 instanceof Vehicle_classes_multi_level_inheritance_9);
console.log(c_classes_multi_level_inheritance_9 instanceof ElectricCar_classes_multi_level_inheritance_9);
console.log(c_classes_multi_level_inheritance_9 instanceof Car_classes_multi_level_inheritance_9);
console.log(c_classes_multi_level_inheritance_9 instanceof Vehicle_classes_multi_level_inheritance_9);
console.log(v_classes_multi_level_inheritance_9 instanceof ElectricCar_classes_multi_level_inheritance_9);
console.log(v_classes_multi_level_inheritance_9 instanceof Car_classes_multi_level_inheritance_9);
console.log(v_classes_multi_level_inheritance_9 instanceof Vehicle_classes_multi_level_inheritance_9);

// --- Context Case: language_classes_polymorphism ---
// @expect: 0
// @expect: Rectangle area = 0
// @expect: 0
// @expect: Circle area = 0
class Shape_classes_polymorphism_10 {
  name: string = "Generic Shape";

  area(): number {
    return 0;
  }

  describe(): string {
    return this.name + " area = " + this.area().toString();
  }
}

class Rectangle_classes_polymorphism_10 extends Shape_classes_polymorphism_10 {
  width: number = 0;
  height: number = 0;

  constructor(w: number, h: number) {
    super();
    this.name = "Rectangle";
    this.width = w;
    this.height = h;
  }

  area(): number {
    return this.width * this.height;
  }
}

class Circle_classes_polymorphism_10 extends Shape_classes_polymorphism_10 {
  radius: number = 0;

  constructor(r: number) {
    super();
    this.name = "Circle";
    this.radius = r;
  }

  area(): number {
    return 3.14 * this.radius * this.radius;
  }
}

const s1_classes_polymorphism_10: Shape_classes_polymorphism_10 = new Rectangle_classes_polymorphism_10(10, 5);
const s2_classes_polymorphism_10: Shape_classes_polymorphism_10 = new Circle_classes_polymorphism_10(2);

console.log(s1_classes_polymorphism_10.area());
console.log(s1_classes_polymorphism_10.describe());

console.log(s2_classes_polymorphism_10.area());
console.log(s2_classes_polymorphism_10.describe());

// --- Context Case: language_classes_stateful-encapsulation ---
// @expect: Account A-100: Balance=$500 (TxCount=0)
// @expect: Account B-200: Balance=$100 (TxCount=0)
// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: Account A-100: Balance=$350 (TxCount=3)
// @expect: Account B-200: Balance=$400 (TxCount=1)
class BankAccount_classes_stateful_encapsulation_11 {
    accountNumber: string;
    balance: number;
    transactionCount: number;

    constructor(accountNumber: string, initialDeposit: number) {
        this.accountNumber = accountNumber;
        this.balance = initialDeposit;
        this.transactionCount = 0;
    }

    deposit(amount: number): boolean {
        if (amount <= 0) {
            return false;
        }
        this.balance += amount;
        this.transactionCount++;
        return true;
    }

    withdraw(amount: number): boolean {
        if (amount <= 0 || amount > this.balance) {
            return false;
        }
        this.balance -= amount;
        this.transactionCount++;
        return true;
    }

    transferTo(target: BankAccount_classes_stateful_encapsulation_11, amount: number): boolean {
        if (this.withdraw(amount)) {
            target.deposit(amount);
            return true;
        }
        return false;
    }

    getStatement(): string {
        return `Account ${this.accountNumber}: Balance=$${this.balance} (TxCount=${this.transactionCount})`;
    }
}

const acc1_classes_stateful_encapsulation_11 = new BankAccount_classes_stateful_encapsulation_11("A-100", 500);
const acc2_classes_stateful_encapsulation_11 = new BankAccount_classes_stateful_encapsulation_11("B-200", 100);

console.log(acc1_classes_stateful_encapsulation_11.getStatement());
console.log(acc2_classes_stateful_encapsulation_11.getStatement());

console.log(acc1_classes_stateful_encapsulation_11.deposit(200));
console.log(acc1_classes_stateful_encapsulation_11.withdraw(50));
console.log(acc1_classes_stateful_encapsulation_11.withdraw(1000)); // Should fail

console.log(acc1_classes_stateful_encapsulation_11.transferTo(acc2_classes_stateful_encapsulation_11, 300));
console.log(acc1_classes_stateful_encapsulation_11.transferTo(acc2_classes_stateful_encapsulation_11, 9999)); // Should fail

console.log(acc1_classes_stateful_encapsulation_11.getStatement());
console.log(acc2_classes_stateful_encapsulation_11.getStatement());

// --- Context Case: language_classes_static-blocks ---
// @expect: 8080
// @expect: localhost
// @expect: http://localhost:8080
// @expect: 20
// @expect: true
class ServerConfig_classes_static_blocks_12 {
  static port: number = 3000;
  static host: string = "localhost";
  static url: string = "";
  static isProduction: boolean = false;

  static {
    ServerConfig_classes_static_blocks_12.port = 8080;
    ServerConfig_classes_static_blocks_12.url = "http://" + ServerConfig_classes_static_blocks_12.host + ":" + ServerConfig_classes_static_blocks_12.port;
  }

  static counter: number = 10;
  static {
    this.counter = this.counter * 2;
    this.isProduction = true;
  }
}

console.log(ServerConfig_classes_static_blocks_12.port);
console.log(ServerConfig_classes_static_blocks_12.host);
console.log(ServerConfig_classes_static_blocks_12.url);
console.log(ServerConfig_classes_static_blocks_12.counter);
console.log(ServerConfig_classes_static_blocks_12.isProduction);

// --- Context Case: language_classes_static-fields ---
// @expect: 42
// @expect: point
class Point_classes_static_fields_13 {
  x: number = 42;
  label: string = "point";
}

const point_classes_static_fields_13 = new Point_classes_static_fields_13();
console.log(point_classes_static_fields_13.x);
console.log(point_classes_static_fields_13.label);

// --- Context Case: language_classes_static-inheritance ---
// @expect: 8080
// @expect: localhost
// @expect: http://localhost:8080
// @expect: v1
// @expect: http://localhost:8080/v1/users
// @expect: http://localhost:8080/v1/orders
class BaseConfig_classes_static_inheritance_14 {
  static defaultPort: number = 8080;
  static defaultHost: string = "localhost";

  static getUrl(): string {
    return "http://" + BaseConfig_classes_static_inheritance_14.defaultHost + ":" + BaseConfig_classes_static_inheritance_14.defaultPort;
  }
}

class AppConfig_classes_static_inheritance_14 extends BaseConfig_classes_static_inheritance_14 {
  static apiVersion: string = "v1";

  static getApiEndpoint(path: string): string {
    return BaseConfig_classes_static_inheritance_14.getUrl() + "/" + AppConfig_classes_static_inheritance_14.apiVersion + "/" + path;
  }
}

console.log(BaseConfig_classes_static_inheritance_14.defaultPort);
console.log(BaseConfig_classes_static_inheritance_14.defaultHost);
console.log(BaseConfig_classes_static_inheritance_14.getUrl());

console.log(AppConfig_classes_static_inheritance_14.apiVersion);
console.log(AppConfig_classes_static_inheritance_14.getApiEndpoint("users"));
console.log(AppConfig_classes_static_inheritance_14.getApiEndpoint("orders"));

// --- Context Case: language_classes_static-members ---
// @expect: 3.14
// @expect: 42
class MathUtils_classes_static_members_15 {
  static pi: number = 3.14;

  static multiply(a: number, b: number): number {
    return a * b;
  }
}

console.log(MathUtils_classes_static_members_15.pi);
console.log(MathUtils_classes_static_members_15.multiply(6, 7));

// --- Context Case: language_classes_three-tier-inheritance ---
// @expect: Entity[1]: NPC
// @expect: 10
// @expect: Monster[2]: Goblin (Lvl 5)
// @expect: 50
// @expect: BOSS[Rank 3] -> Monster[3]: Dragon (Lvl 50)
// @expect: 3000
// @expect: Summary: Entity[1]: NPC with power 10
// @expect: Summary: Entity[2]: Goblin with power 10
// @expect: Summary: Entity[3]: Dragon with power 10
class Entity_classes_three_tier_inheritance_16 {
    id: number;
    name: string;

    constructor(id: number, name: string) {
        this.id = id;
        this.name = name;
    }

    describe(): string {
        return `Entity[${this.id}]: ${this.name}`;
    }

    getPower(): number {
        return 10;
    }
}

class Monster_classes_three_tier_inheritance_16 extends Entity_classes_three_tier_inheritance_16 {
    level: number;

    constructor(id: number, name: string, level: number) {
        super(id, name);
        this.level = level;
    }

    describe(): string {
        return `Monster[${this.id}]: ${this.name} (Lvl ${this.level})`;
    }

    getPower(): number {
        return super.getPower() * this.level;
    }
}

class BossMonster_classes_three_tier_inheritance_16 extends Monster_classes_three_tier_inheritance_16 {
    bossRank: number;

    constructor(id: number, name: string, level: number, bossRank: number) {
        super(id, name, level);
        this.bossRank = bossRank;
    }

    describe(): string {
        return `BOSS[Rank ${this.bossRank}] -> ${super.describe()}`;
    }

    getPower(): number {
        return super.getPower() * this.bossRank * 2;
    }
}

const e_classes_three_tier_inheritance_16: Entity_classes_three_tier_inheritance_16 = new Entity_classes_three_tier_inheritance_16(1, "NPC");
console.log(e_classes_three_tier_inheritance_16.describe());
console.log(e_classes_three_tier_inheritance_16.getPower());

const m_classes_three_tier_inheritance_16: Monster_classes_three_tier_inheritance_16 = new Monster_classes_three_tier_inheritance_16(2, "Goblin", 5);
console.log(m_classes_three_tier_inheritance_16.describe());
console.log(m_classes_three_tier_inheritance_16.getPower());

const b_classes_three_tier_inheritance_16: BossMonster_classes_three_tier_inheritance_16 = new BossMonster_classes_three_tier_inheritance_16(3, "Dragon", 50, 3);
console.log(b_classes_three_tier_inheritance_16.describe());
console.log(b_classes_three_tier_inheritance_16.getPower());

function printEntitySummary_classes_three_tier_inheritance_16(entity: Entity_classes_three_tier_inheritance_16): void {
    console.log(`Summary: ${entity.describe()} with power ${entity.getPower()}`);
}

printEntitySummary_classes_three_tier_inheritance_16(e_classes_three_tier_inheritance_16);
printEntitySummary_classes_three_tier_inheritance_16(m_classes_three_tier_inheritance_16);
printEntitySummary_classes_three_tier_inheritance_16(b_classes_three_tier_inheritance_16);

// --- Context Case: language_objects_bracket-access ---
// @expect: Alice
// @expect: 30
// @expect: Hanoi
// @expect: 31
// @expect: Saigon
const person_objects_bracket_access_17 = { name: "Alice", age: 30, city: "Hanoi" };
console.log(person_objects_bracket_access_17["name"]);
console.log(person_objects_bracket_access_17["age"]);
console.log(person_objects_bracket_access_17["city"]);

person_objects_bracket_access_17["age"] = 31;
person_objects_bracket_access_17["city"] = "Saigon";
console.log(person_objects_bracket_access_17["age"]);
console.log(person_objects_bracket_access_17["city"]);

// --- Context Case: language_objects_discriminated-union ---
// @expect: 10
// @expect: 5
class Circle_objects_discriminated_union_18 {
    kind: string = "circle";
    radius: number;
    constructor(r: number) {
        this.radius = r;
    }
}

class Square_objects_discriminated_union_18 {
    kind: string = "square";
    size: number;
    constructor(s: number) {
        this.size = s;
    }
}

type Shape_objects_discriminated_union_18 = Circle_objects_discriminated_union_18 | Square_objects_discriminated_union_18;

const c_objects_discriminated_union_18: Shape_objects_discriminated_union_18 = new Circle_objects_discriminated_union_18(10);
if (c_objects_discriminated_union_18.kind === "circle") {
    console.log(c_objects_discriminated_union_18.radius);
}

const sq_objects_discriminated_union_18: Shape_objects_discriminated_union_18 = new Square_objects_discriminated_union_18(5);
if (sq_objects_discriminated_union_18.kind === "square") {
    console.log(sq_objects_discriminated_union_18.size);
}

// --- Context Case: language_objects_literals ---
// @expect: 10
// @expect: 20
// @expect: 30
// @expect: Alice
// @expect: 30
// @expect: 100
// @expect: cool
const point_objects_literals_19 = { x: 10, y: 20 };
console.log(point_objects_literals_19.x);
console.log(point_objects_literals_19.y);
console.log(point_objects_literals_19.x + point_objects_literals_19.y);

const person_objects_literals_19 = { name: "Alice", age: 30 };
console.log(person_objects_literals_19.name);
console.log(person_objects_literals_19.age);

const a_objects_literals_19: number = 100;
const b_objects_literals_19: string = "cool";
const shorthand_objects_literals_19 = { a: a_objects_literals_19, b: b_objects_literals_19 };
console.log(shorthand_objects_literals_19.a);
console.log(shorthand_objects_literals_19.b);

// --- Context Case: language_objects_nested-object-records ---
// @expect: 101
// @expect: alice_dev
// @expect: true
// @expect: San Francisco
// @expect: 94105
// @expect: alice_dev (#101) lives in San Francisco (94105)
interface Address_objects_nested_object_records_20 {
    city: string;
    zip: number;
}

interface UserProfile_objects_nested_object_records_20 {
    id: number;
    username: string;
    active: boolean;
    address: Address_objects_nested_object_records_20;
}

const user_objects_nested_object_records_20: UserProfile_objects_nested_object_records_20 = {
    id: 101,
    username: "alice_dev",
    active: true,
    address: {
        city: "San Francisco",
        zip: 94105,
    },
};

console.log(user_objects_nested_object_records_20.id);
console.log(user_objects_nested_object_records_20.username);
console.log(user_objects_nested_object_records_20.active);
console.log(user_objects_nested_object_records_20.address.city);
console.log(user_objects_nested_object_records_20.address.zip);

function formatUserInfo_objects_nested_object_records_20(u: UserProfile_objects_nested_object_records_20): string {
    return `${u.username} (#${u.id}) lives in ${u.address.city} (${u.address.zip})`;
}

console.log(formatUserInfo_objects_nested_object_records_20(user_objects_nested_object_records_20));
