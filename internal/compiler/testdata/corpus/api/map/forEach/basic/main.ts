const m = new Map<string, number>();
m.set("x", 5);
m.forEach((val: number, key: string) => {
    console.log(key + "=" + val);
});
