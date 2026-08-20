type Direction = "asc" | "desc";
type HttpStatus = 200 | 404 | 500;

function sortOrder(dir: Direction): string {
    if (dir === "asc") {
        return "ascending";
    }
    return "descending";
}

function handleStatus(status: HttpStatus): string {
    switch (status) {
        case 200:
            return "OK";
        case 404:
            return "Not Found";
        case 500:
            return "Server Error";
        default:
            return "Unknown";
    }
}

console.log(sortOrder("asc"));
console.log(sortOrder("desc"));
console.log(handleStatus(200));
console.log(handleStatus(404));
console.log(handleStatus(500));
