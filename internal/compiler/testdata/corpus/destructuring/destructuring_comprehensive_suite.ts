// @expect: 127.0.0.1 8080 5 3000
// @expect: 100 2 3
// @expect: item1 2 item2,item3
// @expect: auth_token_xyz
// @expect: 101 9500 42
// @expect: 10 30 50
// 1. Nested object destructuring with default values
interface ServerOptions {
    host: string;
    port?: number;
    settings?: {
        retries?: number;
        timeout?: number;
    };
}

const serverConfig: ServerOptions = {
    host: "127.0.0.1",
    settings: {
        timeout: 3000
    }
};

const {
    host,
    port = 8080,
    settings: { retries = 5, timeout = 1000 } = {}
} = serverConfig;

console.log(host, port, retries, timeout);

// 2. Nested array destructuring with rest and defaults
const [firstVal = 1, [nestedVal = 2, deepVal = 3] = []] = [100];
console.log(firstVal, nestedVal, deepVal);

const [head = "default", ...tail] = ["item1", "item2", "item3"];
console.log(head, tail.length, tail.join(","));

// 3. Object rest destructuring excluding selected fields
interface UserRecord {
    id: number;
    token: string;
    score: number;
    level: number;
}

const user: UserRecord = {
    id: 101,
    token: "auth_token_xyz",
    score: 9500,
    level: 42
};

const { token, ...publicProfile } = user;
console.log(token);
console.log(publicProfile.id, publicProfile.score, publicProfile.level);

// 4. Array destructuring skipping elements with holes
const [firstElem, , thirdElem, , fifthElem] = [10, 20, 30, 40, 50];
console.log(firstElem, thirdElem, fifthElem);
