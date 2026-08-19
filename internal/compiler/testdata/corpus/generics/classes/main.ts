class Box<T> {
  value: T;

  constructor(value: T) {
    this.value = value;
  }

  getValue(): T {
    return this.value;
  }

  setValue(newValue: T): void {
    this.value = newValue;
  }
}

const numBox = new Box<number>(123);
const strBox = new Box<string>("scriptgo box");

console.log(numBox.getValue());
console.log(strBox.getValue());

numBox.setValue(456);
strBox.setValue("updated box");

console.log(numBox.getValue());
console.log(strBox.getValue());
