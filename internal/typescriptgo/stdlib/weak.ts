declare namespace __scriptgo {
    function weakrefNew(target: object): WeakRef<object>;
    function weakrefDeref(ref: WeakRef<object>): object | undefined;
    function finalizationRegistryNew(cleanupCallback: (heldValue: unknown) => void): FinalizationRegistry<unknown>;
    function finalizationRegistryRegister(registry: FinalizationRegistry<unknown>, target: object, heldValue: unknown, unregisterToken?: object): void;
    function finalizationRegistryUnregister(registry: FinalizationRegistry<unknown>, unregisterToken: object): boolean;
}

export class WeakRef<T extends object> {
    constructor(target: T) {
        return __scriptgo.weakrefNew(target) as unknown as WeakRef<T>;
    }

    deref(): T | undefined {
        return __scriptgo.weakrefDeref(this as unknown as WeakRef<object>) as T | undefined;
    }
}

export class FinalizationRegistry<T> {
    constructor(cleanupCallback: (heldValue: T) => void) {
        return __scriptgo.finalizationRegistryNew(cleanupCallback as (heldValue: unknown) => void) as unknown as FinalizationRegistry<T>;
    }

    register(target: object, heldValue: T, unregisterToken?: object): void {
        __scriptgo.finalizationRegistryRegister(this as unknown as FinalizationRegistry<unknown>, target, heldValue, unregisterToken);
    }

    unregister(unregisterToken: object): boolean {
        return __scriptgo.finalizationRegistryUnregister(this as unknown as FinalizationRegistry<unknown>, unregisterToken);
    }
}

export default {
    WeakRef,
    FinalizationRegistry,
};
