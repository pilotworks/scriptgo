// @api: responseinit.status
// @api: responseinit.statusText
// @api: responseinit.headers
// @expect: 200
// @expect: OK
// @expect: false
const init: ResponseInit = {
    status: 200,
    statusText: "OK"
};

console.log(init.status);
console.log(init.statusText);
console.log(init.headers !== undefined);

