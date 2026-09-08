export function summarize(input) {
  return { name: input.name, count: input.count + 1 };
}

export function bump(values) {
  return values.map((value) => value + 10);
}
