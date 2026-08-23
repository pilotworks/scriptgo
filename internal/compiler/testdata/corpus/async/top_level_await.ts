// @expect: step 1
// @expect: resolved: hello async
// @expect: step 2
export {};

function getGreeting(): Promise<string> {
    return Promise.resolve("hello async");
}

console.log("step 1");
const greeting = await getGreeting();
console.log("resolved: " + greeting);
console.log("step 2");
