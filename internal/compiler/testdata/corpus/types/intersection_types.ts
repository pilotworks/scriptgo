// @expect: 101
// @expect: 1000
// @expect: 2000
// @expect: ScriptGo Architecture
type WithId = { id: number };
type WithTimestamps = { createdAt: number; updatedAt: number };
type Post = WithId & WithTimestamps & { title: string };

const p: Post = {
    id: 101,
    createdAt: 1000,
    updatedAt: 2000,
    title: "ScriptGo Architecture"
};

console.log(p.id);
console.log(p.createdAt);
console.log(p.updatedAt);
console.log(p.title);
