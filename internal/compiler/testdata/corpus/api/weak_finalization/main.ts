class User {
    name: string;
    constructor(name: string) {
        this.name = name;
    }
}

const user = new User("Alice");
const ref = new WeakRef(user);
const derefUser = ref.deref();
if (derefUser) {
    console.log("WeakRef deref user name:", derefUser.name);
}

const registry = new FinalizationRegistry((held: string) => {
    console.log("Finalizer called with heldValue:", held);
});

const token = { id: 1 };
registry.register(user, "user-alice-token", token);
const unregistered = registry.unregister(token);
console.log("FinalizationRegistry unregister success:", unregistered);

const unregisteredSecond = registry.unregister(token);
console.log("FinalizationRegistry unregister second time:", unregisteredSecond);
