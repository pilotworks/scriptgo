// @expect: 127.0.0.1
// @expect: 5432
interface DatabaseConfig {
    readonly host: string;
    readonly port: number;
}

const db: DatabaseConfig = {
    host: "127.0.0.1",
    port: 5432
};

console.log(db.host);
console.log(db.port);
