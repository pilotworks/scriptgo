class Config {
    title: string;
    port: number;
    constructor(title: string, port: number) {
        this.title = title;
        this.port = port;
    }
}

function getPort(port: string): string {
    const fallback: string = "8080";
    return port ?? fallback;
}

console.log(getPort("3000"));
console.log(getPort("null"));
console.log(getPort("undefined"));

const cfg: Config = new Config("App", 5000);
console.log(cfg?.title);
console.log(cfg?.port);

const str: string = "hello world";
console.log(str?.length);
