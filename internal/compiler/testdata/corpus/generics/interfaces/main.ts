interface Container<T> {
  get(): T;
}

class ItemHolder<T> implements Container<T> {
  item: T;

  constructor(item: T) {
    this.item = item;
  }

  get(): T {
    return this.item;
  }
}

const holder1 = new ItemHolder<number>(777);
const holder2 = new ItemHolder<string>("interface test");

console.log(holder1.get());
console.log(holder2.get());
