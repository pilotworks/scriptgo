// @expect: Alice 100 default-tag USD
// @expect: [10,20,30,[40,50]]
interface ComplexUser {
    info: {
        name: string;
        scores: number[];
        tag?: string;
    };
    wallet: {
        currency: string;
        amount: number;
    };
}

function processUser({
    info: { name, scores: [firstScore], tag = "default-tag" },
    wallet: { currency }
}: ComplexUser): string {
    return `${name} ${firstScore} ${tag} ${currency}`;
}

const user: ComplexUser = {
    info: {
        name: "Alice",
        scores: [100, 95, 88]
    },
    wallet: {
        currency: "USD",
        amount: 500
    }
};

console.log(processUser(user));

const matrixNested: number[][] = [[10, 20], [30, 40, 50]];
const [[a, b], [c, ...restD]] = matrixNested;
console.log(JSON.stringify([a, b, c, restD]));
