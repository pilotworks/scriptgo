interface ServerConfig {
    host: string;
    port: number;
    enabled: boolean;
    tags: string[];
}

const config: ServerConfig = {
    host: "localhost",
    port: 8080,
    enabled: true,
    tags: ["api", "v1", "production"],
};

const jsonStr: string = JSON.stringify(config);
console.log(jsonStr);

console.log(JSON.stringify([10, 20, 30]));
console.log(JSON.stringify(["alpha", "beta", "gamma"]));
console.log(JSON.stringify(true));
console.log(JSON.stringify(false));
console.log(JSON.stringify(12345));
console.log(JSON.stringify("hello world"));
