interface UserProfile {
    id: number;
    name: string;
    active: boolean;
    scores: number[];
    tags: string[];
}

const profile: UserProfile = {
    id: 101,
    name: "Alice",
    active: true,
    scores: [95, 100],
    tags: ["admin", "dev"],
};

console.log(JSON.stringify(profile));
