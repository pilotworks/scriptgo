// @expect: HTTP 200 OK
// @expect: Header: Content-Type: application/json
// @expect: Header: Content-Length: 42
// @expect: HTTP 404 Not Found
// @expect: Header: Content-Type: text/plain
// @expect: HTTP 500 Internal Server Error
type HttpResponse = [number, string, string[]];

function formatResponse(resp: HttpResponse): string {
    const [statusCode, statusText, headers] = resp;
    let out = `HTTP ${statusCode} ${statusText}\n`;
    for (const h of headers) {
        out += `Header: ${h}\n`;
    }
    return out.trim();
}

const resps: HttpResponse[] = [
    [200, "OK", ["Content-Type: application/json", "Content-Length: 42"]],
    [404, "Not Found", ["Content-Type: text/plain"]],
    [500, "Internal Server Error", []]
];

for (const r of resps) {
    console.log(formatResponse(r));
}
