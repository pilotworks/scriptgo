export function makeBox() {
  return {
    base: 4,
    add(value) {
      return this.base + value;
    },
  };
}
