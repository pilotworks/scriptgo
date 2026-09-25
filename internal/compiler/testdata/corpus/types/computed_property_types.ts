// @expect: ScriptGo
const S: symbol = Symbol("tag");
type WithComputed = {
    [S]: string;
    name: string;
};
const item: WithComputed = {
    [S]: "tag_val",
    name: "ScriptGo"
};
console.log(item.name);
