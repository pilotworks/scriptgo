// @expect: true
// @expect: Alice
// @expect: false
// @expect: user:1,user:2
interface RecordEntry<K, V> {
    key: K;
    value: V;
}

class KeyValueStore<K extends string | number, V> {
    private entries: RecordEntry<K, V>[] = [];

    set(key: K, value: V): void {
        for (let i = 0; i < this.entries.length; i++) {
            if (this.entries[i].key === key) {
                this.entries[i].value = value;
                return;
            }
        }
        this.entries.push({ key, value });
    }

    get(key: K): V | undefined {
        for (let i = 0; i < this.entries.length; i++) {
            if (this.entries[i].key === key) {
                return this.entries[i].value;
            }
        }
        return undefined;
    }

    has(key: K): boolean {
        return this.get(key) !== undefined;
    }

    keys(): K[] {
        return this.entries.map(e => e.key);
    }

    values(): V[] {
        return this.entries.map(e => e.value);
    }
}

const store = new KeyValueStore<string, { id: number; name: string }>();
store.set("user:1", { id: 1, name: "Alice" });
store.set("user:2", { id: 2, name: "Bob" });

console.log(store.has("user:1"));
console.log(store.get("user:1")?.name);
console.log(store.has("user:3"));
console.log(store.keys().join(","));
