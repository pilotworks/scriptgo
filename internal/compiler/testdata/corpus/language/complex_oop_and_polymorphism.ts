// @expect: === Complex OOP & Polymorphism Test ===
// @expect: Initial Global Count: 1000
// @expect: 
// @expect: Entity List (Polymorphism):
// @expect: - User[user_1001] alice <alice@example.com> roles=[user]
// @expect: - User[user_1002] bob <bob@corp.internal> roles=[user,admin] permissions=[read,write]
// @expect: - [SUPERADMIN] User[user_1003] charlie <charlie@root.org> roles=[user,admin] permissions=[read,write,delete,sudo] token=root_token_charlie
// @expect: 
// @expect: Private field verification:
// @expect: Alice password check (correct): true
// @expect: Alice password check (wrong): false
// @expect: 
// @expect: Getter / Setter check:
// @expect: Alice initial email getter: alice@example.com
// @expect: Alice updated email getter: alice_new@example.org
// @expect: 
// @expect: Type hierarchy checks:
// @expect: Checking User u1:
// @expect: instanceof check: Entity=true, User=true, Admin=false, SuperAdmin=false
// @expect: Checking AdminUser admin:
// @expect: instanceof check: Entity=true, User=true, Admin=true, SuperAdmin=false
// @expect: Checking SuperAdminUser root:
// @expect: instanceof check: Entity=true, User=true, Admin=true, SuperAdmin=true
// @expect: 
// @expect: Final Global Count: 1003

// Complex OOP, Multi-Level Inheritance, Private Fields, Abstract Classes, Polymorphism & Static Blocks

abstract class Entity {
    public readonly id: string;
    public createdAt: number;
    private static entityCounter: number = 0;

    static {
        // Static initialization block
        Entity.entityCounter = 1000;
    }

    constructor(idPrefix: string) {
        Entity.entityCounter++;
        this.id = `${idPrefix}_${Entity.entityCounter}`;
        this.createdAt = 1700000000;
    }

    abstract describe(): string;

    getId(): string {
        return this.id;
    }

    static getGlobalCount(): number {
        return Entity.entityCounter;
    }
}

class User extends Entity {
    #passwordHash: string; // ECMAScript private field
    protected _email: string;
    public roles: string[];

    constructor(
        public username: string,
        email: string,
        passwordRaw: string,
        roles: string[] = ["user"]
    ) {
        super("user");
        this._email = email;
        this.#passwordHash = `hash_${passwordRaw}`;
        this.roles = roles;
    }

    get email(): string {
        return this._email.toLowerCase();
    }

    set email(val: string) {
        if (val.includes("@")) {
            this._email = val;
        }
    }

    verifyPassword(raw: string): boolean {
        return this.#passwordHash === `hash_${raw}`;
    }

    override describe(): string {
        return `User[${this.id}] ${this.username} <${this.email}> roles=[${this.roles.join(",")}]`;
    }
}

class AdminUser extends User {
    public permissions: string[];

    constructor(
        username: string,
        email: string,
        passwordRaw: string,
        permissions: string[]
    ) {
        super(username, email, passwordRaw, ["user", "admin"]);
        this.permissions = permissions;
    }

    override describe(): string {
        const baseDesc = super.describe();
        return `${baseDesc} permissions=[${this.permissions.join(",")}]`;
    }

    hasPermission(perm: string): boolean {
        return this.permissions.includes(perm);
    }
}

class SuperAdminUser extends AdminUser {
    public superToken: string;

    constructor(username: string, email: string, passwordRaw: string) {
        super(username, email, passwordRaw, ["read", "write", "delete", "sudo"]);
        this.superToken = `root_token_${username}`;
    }

    override describe(): string {
        return `[SUPERADMIN] ${super.describe()} token=${this.superToken}`;
    }
}

function printPolymorphicDescriptions(entities: Entity[]): void {
    for (let i = 0; i < entities.length; i++) {
        const ent = entities[i];
        console.log(`- ${ent.describe()}`);
    }
}

function testInstanceOfHierarchy(obj: Entity): void {
    const isEntity = obj instanceof Entity;
    const isUser = obj instanceof User;
    const isAdmin = obj instanceof AdminUser;
    const isSuperAdmin = obj instanceof SuperAdminUser;

    console.log(`instanceof check: Entity=${isEntity}, User=${isUser}, Admin=${isAdmin}, SuperAdmin=${isSuperAdmin}`);
}

function main(): void {
    console.log("=== Complex OOP & Polymorphism Test ===");
    console.log("Initial Global Count:", Entity.getGlobalCount());

    const u1 = new User("alice", "Alice@Example.com", "secret123");
    const admin = new AdminUser("bob", "bob@corp.internal", "admP@ss", ["read", "write"]);
    const root = new SuperAdminUser("charlie", "charlie@root.org", "toor123");

    console.log("\nEntity List (Polymorphism):");
    printPolymorphicDescriptions([u1, admin, root]);

    console.log("\nPrivate field verification:");
    console.log("Alice password check (correct):", u1.verifyPassword("secret123"));
    console.log("Alice password check (wrong):", u1.verifyPassword("wrongpass"));

    console.log("\nGetter / Setter check:");
    console.log("Alice initial email getter:", u1.email);
    u1.email = "alice_new@example.org";
    console.log("Alice updated email getter:", u1.email);

    console.log("\nType hierarchy checks:");
    console.log("Checking User u1:");
    testInstanceOfHierarchy(u1);
    console.log("Checking AdminUser admin:");
    testInstanceOfHierarchy(admin);
    console.log("Checking SuperAdminUser root:");
    testInstanceOfHierarchy(root);

    console.log("\nFinal Global Count:", Entity.getGlobalCount());
}

main();
