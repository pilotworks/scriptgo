function concatAll(prefix: string, ...items: string[]): string {
    let result: string = prefix;
    for (let i: number = 0; i < items.length; i += 1) {
        result += ":" + items[i];
    }
    return result;
}

console.log(concatAll("start", "a", "b", "c"));
