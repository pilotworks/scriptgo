const counter = {
  value: 2,
  get(): number {
    return this.value;
  },
};
console.log(counter.get());
