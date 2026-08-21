function greet(name: string | null): void {
    if (name != null) {
        console.log("hello " + name);
    } else {
        console.log("hello stranger");
    }
}
greet("alice");
greet(null);
