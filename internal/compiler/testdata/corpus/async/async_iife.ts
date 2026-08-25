// @expect: Hello from async IIFE
(async () => {
    const greeting = await Promise.resolve("Hello from async IIFE");
    console.log(greeting);
})();
