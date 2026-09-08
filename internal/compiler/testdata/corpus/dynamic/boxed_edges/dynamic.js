export function inspect(input) {
  return {
    nested: { active: input.active },
    values: [input.count, null, undefined],
  };
}

export function empty() {
  return { objectValue: {}, arrayValue: [] };
}
