// @expect: 1,2
// @expect: 1-2

async function makeValues(): Promise<number[]> {
    return [1, 2];
}

makeValues()
    .then((values: number[]) => {
        console.log(values.join(","));
        return values;
    })
    .then((values: number[]) => console.log(values.join("-")));
