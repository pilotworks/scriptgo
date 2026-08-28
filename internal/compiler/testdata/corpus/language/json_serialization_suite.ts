// @expect: null
// @expect: null
// @expect: null
// @expect: [1,null,3]
// @expect: [null,null,null]
// @expect: [true,false,null,"test"]
// @expect: {"id":101,"score":null,"rank":null}
// @expect: {"created":"2023-11-14T22:13:20.000Z","label":"release"}
// 1. Top-level primitive serialization
console.log(JSON.stringify(NaN));
console.log(JSON.stringify(Infinity));
console.log(JSON.stringify(-Infinity));

// 2. Array serialization containing nullish and non-finite numbers
console.log(JSON.stringify([1, undefined, 3]));
console.log(JSON.stringify([NaN, Infinity, -Infinity]));
console.log(JSON.stringify([true, false, null, "test"]));

// 3. Object serialization omitting undefined properties
interface UserRecord {
    id: number;
    nickname?: string;
    score: number;
    rank: number;
}

const user: UserRecord = {
    id: 101,
    nickname: undefined,
    score: NaN,
    rank: Infinity
};

console.log(JSON.stringify(user));

// 4. JSON.stringify on Date objects producing ISO 8601 timestamp
const timestampDate: Date = new Date(1700000000000);
const payload: { created: Date; label: string } = {
    created: timestampDate,
    label: "release"
};

console.log(JSON.stringify(payload));
