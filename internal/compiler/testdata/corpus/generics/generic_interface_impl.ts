// @expect: 2
// @expect: Alice
// @expect: Bob
// @expect: Not found
interface Repository<T> {
    add(item: T): void;
    getAll(): T[];
    findById(id: number): T | undefined;
}

interface Entity {
    id: number;
    name: string;
}

class InMemoryRepo<T extends Entity> implements Repository<T> {
    private list: T[] = [];

    add(item: T): void {
        this.list.push(item);
    }

    getAll(): T[] {
        return this.list;
    }

    findById(id: number): T | undefined {
        for (const item of this.list) {
            if (item.id === id) {
                return item;
            }
        }
        return undefined;
    }
}

const repo = new InMemoryRepo<Entity>();
repo.add({ id: 1, name: "Alice" });
repo.add({ id: 2, name: "Bob" });

console.log(repo.getAll().length);
console.log(repo.findById(1)?.name);
console.log(repo.findById(2)?.name);
console.log(repo.findById(3)?.name ?? "Not found");
