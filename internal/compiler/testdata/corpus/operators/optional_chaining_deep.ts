// @expect: Hanoi
// @expect: 100000
// @expect: Xin chao
// @expect: No city
// @expect: No zip
// @expect: No greeting
type Address = {
    city?: {
        name: string;
        zip?: string;
    };
};

type User = {
    id: number;
    address?: Address;
    getGreeting?: () => string;
};

const user1: User = {
    id: 1,
    address: {
        city: {
            name: "Hanoi",
            zip: "100000"
        }
    },
    getGreeting: () => "Xin chao"
};

const user2: User = {
    id: 2
};

console.log(user1.address?.city?.name);
console.log(user1.address?.city?.zip);
console.log(user1.getGreeting?.());

console.log(user2.address?.city?.name ?? "No city");
console.log(user2.address?.city?.zip ?? "No zip");
console.log(user2.getGreeting?.() ?? "No greeting");
