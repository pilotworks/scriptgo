// @expect: admin
// @expect: user
function firstRoute<const T extends string[]>(routes: T): string {
    return routes[0];
}

const routes = ["admin", "user", "guest"];
console.log(firstRoute(routes));
console.log(routes[1]);
