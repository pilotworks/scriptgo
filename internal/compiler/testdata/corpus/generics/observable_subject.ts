// @expect: Obs 1 received: 10
// @expect: Obs 2 received: 10
// @expect: Obs 2 received: 20
// @expect: Final count: 1
type Observer<T> = (value: T) => void;
type Unsubscribe = () => void;

class Subject<T> {
    private observers: Observer<T>[] = [];

    subscribe(observer: Observer<T>): Unsubscribe {
        this.observers.push(observer);
        return () => {
            const index = this.observers.indexOf(observer);
            if (index !== -1) {
                this.observers.splice(index, 1);
            }
        };
    }

    next(value: T): void {
        for (let i = 0; i < this.observers.length; i++) {
            this.observers[i](value);
        }
    }

    observerCount(): number {
        return this.observers.length;
    }
}

const subject = new Subject<number>();

const obs1: Observer<number> = (val) => console.log("Obs 1 received: " + val);
const obs2: Observer<number> = (val) => console.log("Obs 2 received: " + val);

const unsub1 = subject.subscribe(obs1);
const unsub2 = subject.subscribe(obs2);

subject.next(10);
unsub1();
subject.next(20);

console.log("Final count: " + subject.observerCount());
