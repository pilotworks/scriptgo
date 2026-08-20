interface Address {
    city: string;
    zip: number;
}

interface UserProfile {
    id: number;
    username: string;
    active: boolean;
    address: Address;
}

const user: UserProfile = {
    id: 101,
    username: "alice_dev",
    active: true,
    address: {
        city: "San Francisco",
        zip: 94105,
    },
};

console.log(user.id);
console.log(user.username);
console.log(user.active);
console.log(user.address.city);
console.log(user.address.zip);

function formatUserInfo(u: UserProfile): string {
    return `${u.username} (#${u.id}) lives in ${u.address.city} (${u.address.zip})`;
}

console.log(formatUserInfo(user));
