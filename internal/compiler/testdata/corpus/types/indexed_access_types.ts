// @expect: Developer
// @expect: http://example.com/avatar.png
type User = {
    id: number;
    profile: {
        bio: string;
        avatarUrl: string;
    };
};

type Profile = User["profile"];

const prof: Profile = {
    bio: "Developer",
    avatarUrl: "http://example.com/avatar.png"
};

console.log(prof.bio);
console.log(prof.avatarUrl);
