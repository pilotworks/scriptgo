// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
interface UserProfile {
    name: string;
    email?: string;
    age: number;
}

const user: UserProfile = {
    name: "Alice",
    age: 28
};

console.log("name" in user);
console.log("email" in user);
console.log("age" in user);

const arr = [10, 20, 30];
console.log(1 in arr);
console.log(5 in arr);
