// ScriptGo Corpus: ECMAScript Reflect API Complete Test Suite
// Verifying all 19 standard Reflect operations & reflection metadata in Static Tier.

// @api: Reflect.defineMetadata
// @api: Reflect.getMetadata
// @api: Reflect.getOwnMetadata
// @api: Reflect.hasMetadata
// @api: Reflect.hasOwnMetadata
// @api: Reflect.metadata
// @api: Reflect.get
// @api: Reflect.set
// @api: Reflect.has
// @api: Reflect.deleteProperty
// @api: Reflect.ownKeys
// @api: Reflect.defineProperty
// @api: Reflect.getOwnPropertyDescriptor
// @api: Reflect.getPrototypeOf
// @api: Reflect.setPrototypeOf
// @api: Reflect.isExtensible
// @api: Reflect.preventExtensions
// @api: Reflect.apply
// @api: Reflect.construct

// --- 1. Class with Metadata and Properties ---
class Account {
    @Reflect.metadata("security", "high")
    balance: number = 1000;
    owner: string = "Alice";

    @Reflect.metadata("audit", "true")
    deposit(amount: number): number {
        this.balance = this.balance + amount;
        return this.balance;
    }
}

// --- 2. Test Metadata APIs ---
// @expect: true
console.log(Reflect.hasMetadata("security", Account, "balance"));
// @expect: high
console.log(Reflect.getMetadata("security", Account, "balance"));
// @expect: true
console.log(Reflect.hasOwnMetadata("security", Account, "balance"));
// @expect: high
console.log(Reflect.getOwnMetadata("security", Account, "balance"));

Reflect.defineMetadata("version", "2.0", Account, "deposit");
// @expect: true
console.log(Reflect.hasMetadata("version", Account, "deposit"));
// @expect: 2.0
console.log(Reflect.getMetadata("version", Account, "deposit"));

// --- 3. Test Object Properties (get, set, has, ownKeys) ---
let acc = new Account();

// @expect: Alice
console.log(Reflect.get(acc, "owner"));
// @expect: 1000
console.log(Reflect.get(acc, "balance"));

// @expect: true
console.log(Reflect.has(acc, "owner"));
// @expect: true
console.log(Reflect.has(acc, "balance"));
// @expect: false
console.log(Reflect.has(acc, "nonExistent"));

// @expect: true
console.log(Reflect.set(acc, "owner", "Bob"));
// @expect: Bob
console.log(Reflect.get(acc, "owner"));

let keys = Reflect.ownKeys(acc);
// @expect: balance
console.log(keys[0]);
// @expect: owner
console.log(keys[1]);

// --- 4. Test deleteProperty, defineProperty, getOwnPropertyDescriptor ---
// @expect: true
console.log(Reflect.deleteProperty(acc, "owner"));

// @expect: true
console.log(Reflect.defineProperty(acc, "owner", { value: "Charlie" }));
// @expect: Charlie
console.log(Reflect.get(acc, "owner"));

let desc = Reflect.getOwnPropertyDescriptor(acc, "owner");
// @expect: true
console.log(desc!.writable);
// @expect: true
console.log(desc!.enumerable);
// @expect: true
console.log(desc!.configurable);

// --- 5. Test prototype & extension checks ---
// @expect: Object.prototype
console.log(Reflect.getPrototypeOf(acc));
// @expect: true
console.log(Reflect.setPrototypeOf(acc, null));
// @expect: true
console.log(Reflect.isExtensible(acc));
// @expect: true
console.log(Reflect.preventExtensions(acc));

// --- 6. Test apply and construct ---
function add(a: number, b: number): number {
    return a + b;
}

// @expect: 30
console.log(Reflect.apply(add, undefined, [10, 20]));

let newAcc = Reflect.construct(Account, []);
// @expect: Alice
console.log(Reflect.get(newAcc, "owner"));
