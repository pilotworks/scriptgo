// @api: requestinit.method
// @api: requestinit.headers
// @api: requestinit.body
// @expect: POST
// @expect: false
// @expect: test body
const init: RequestInit = {
    method: "POST",
    body: "test body"
};

console.log(init.method);
console.log(init.headers !== undefined);
console.log(init.body);

