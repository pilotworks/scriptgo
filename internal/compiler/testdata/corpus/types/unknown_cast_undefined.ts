// @expect: true
function castUndefined(value: unknown): string {
    return value as string;
}

const value = castUndefined(undefined);
console.log(value === undefined);
