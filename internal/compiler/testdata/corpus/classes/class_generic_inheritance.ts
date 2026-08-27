// @expect: base item: 20
// @expect: base item: 10
// @expect: sorted item: 5
// @expect: sorted item: 15
class Container<T> {
    protected items: T[] = [];

    add(item: T): void {
        this.items.push(item);
    }

    dump(prefix: string): void {
        for (let i = 0; i < this.items.length; i++) {
            console.log(prefix + ": " + this.items[i]);
        }
    }
}

class SortedContainer extends Container<number> {
    override add(item: number): void {
        super.add(item);
        this.items.sort((a, b) => a - b);
    }
}

const c1 = new Container<number>();
c1.add(20);
c1.add(10);
c1.dump("base item");

const c2 = new SortedContainer();
c2.add(15);
c2.add(5);
c2.dump("sorted item");
