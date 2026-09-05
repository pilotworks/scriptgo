import { create, Domain, active } from "node:domain";

// @api: domain.create
// @expect: domain create passed
const d = create();
if (d instanceof Domain) {
    console.log("domain create passed");
}

// @api: domain.Domain
// @expect: Domain class passed
const d2 = new Domain();
if (d2 instanceof Domain) {
    console.log("Domain class passed");
}

// @api: Domain.enter
// @expect: enter passed
d.enter();
console.log("enter passed");

// @api: Domain.exit
// @expect: exit passed
d.exit();
console.log("exit passed");

// @api: Domain.run
// @expect: run passed
let runExecuted = false;
d.run(() => {
    runExecuted = true;
});
if (runExecuted) {
    console.log("run passed");
}

// @api: Domain.bind
// @expect: bind passed
let bindSuccess = false;
const boundFn = d.bind((): void => {
    bindSuccess = true;
});
(boundFn as () => void)();
if (bindSuccess) {
    console.log("bind passed");
}

// @api: Domain.intercept
// @expect: intercept passed
let interceptSuccess = false;
const interceptedFn = d.intercept((data: string): void => {
    if (data === "hello") {
        interceptSuccess = true;
    }
});
(interceptedFn as (err: unknown, data: string) => void)(null, "hello");
if (interceptSuccess) {
    console.log("intercept passed");
}
