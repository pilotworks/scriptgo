// @expect: prefix
// @expect: 7
async function timerValue(): Promise<number> {
    const value = await new Promise<number>((resolve) => {
        setTimeout(() => resolve(7), 0);
    });
    return value;
}

console.log("prefix");
timerValue().then((value: number) => console.log(value));
