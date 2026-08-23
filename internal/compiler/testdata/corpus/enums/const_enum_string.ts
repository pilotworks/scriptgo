// @expect: GET
// @expect: POST
// @expect: PUT
const enum HttpMethod {
    Get = "GET",
    Post = "POST",
    Put = "PUT",
}

console.log(HttpMethod.Get);
console.log(HttpMethod.Post);
console.log(HttpMethod.Put);
