// @expect: NONE
// @expect: READ|WRITE
// @expect: true
// @expect: false
// @expect: WRITE
// @expect: READ|WRITE|EXECUTE
// @expect: true
enum Permission {
    NONE = 0,
    READ = 1,
    WRITE = 2,
    EXECUTE = 4,
    ADMIN = 7
}

class UserAccess {
    private permissions: number;

    constructor(initial: number = Permission.NONE) {
        this.permissions = initial;
    }

    grant(p: Permission): void {
        this.permissions |= p;
    }

    revoke(p: Permission): void {
        this.permissions &= ~p;
    }

    has(p: Permission): boolean {
        return (this.permissions & p) === p;
    }

    list(): string {
        const perms: string[] = [];
        if (this.has(Permission.READ)) perms.push("READ");
        if (this.has(Permission.WRITE)) perms.push("WRITE");
        if (this.has(Permission.EXECUTE)) perms.push("EXECUTE");
        return perms.join("|") || "NONE";
    }
}

const user = new UserAccess();
console.log(user.list());

user.grant(Permission.READ);
user.grant(Permission.WRITE);
console.log(user.list());
console.log(user.has(Permission.WRITE));
console.log(user.has(Permission.EXECUTE));

user.revoke(Permission.READ);
console.log(user.list());

user.grant(Permission.ADMIN);
console.log(user.list());
console.log(user.has(Permission.ADMIN));
