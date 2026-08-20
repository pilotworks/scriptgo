interface HasId {
  id: number;
  name: string;
}

interface HasExtra extends HasId {
  role: string;
}

function getEntityName<T extends HasId>(entity: T): string {
  return entity.name;
}

function compareEntityIds<T extends HasId, U extends HasId>(a: T, b: U): boolean {
  return a.id === b.id;
}

class UserAccount implements HasExtra {
  constructor(public id: number, public name: string, public role: string) {}
}

class ProductItem implements HasId {
  constructor(public id: number, public name: string) {}
}

const user = new UserAccount(101, "Alice", "Admin");
const product = new ProductItem(101, "Widget");
const product2 = new ProductItem(202, "Gadget");

console.log(getEntityName<UserAccount>(user));
console.log(getEntityName<ProductItem>(product));
console.log(compareEntityIds<UserAccount, ProductItem>(user, product));
console.log(compareEntityIds<ProductItem, ProductItem>(product, product2));
