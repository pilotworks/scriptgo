// 1. for...in on array
const arr: string[] = ["x", "y", "z"];
for (const key in arr) {
    console.log(key);
}

// 2. for...in on static object shape
class Config {
    host: string;
    port: number;
    constructor(host: string, port: number) {
        this.host = host;
        this.port = port;
    }
}

const cfg: Config = new Config("localhost", 8080);
for (const prop in cfg) {
    console.log(prop);
}
