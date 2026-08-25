// @expect: true
// @expect: true
// @expect: false
enum Permissions {
    None = 0,
    Execute = 1,
    Write = 2,
    Read = 4
}

function hasPermission(userPerms: number, perm: Permissions): boolean {
    return (userPerms & perm) === perm;
}

const userRole = Permissions.Read | Permissions.Write;

console.log(hasPermission(userRole, Permissions.Read));
console.log(hasPermission(userRole, Permissions.Write));
console.log(hasPermission(userRole, Permissions.Execute));
