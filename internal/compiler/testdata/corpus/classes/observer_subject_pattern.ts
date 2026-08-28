// @expect: Observer Alpha received event 'price_change' with data: 100
// @expect: ObserverB computed: 500
// @expect: Observer Alpha received event 'volume_spike' with data: 250
// @expect: ObserverB computed: 1250
interface Observer {
    update(event: string, data: number): void;
}

class ConcreteObserverA implements Observer {
    name: string;

    constructor(name: string) {
        this.name = name;
    }

    update(event: string, data: number): void {
        console.log("Observer " + this.name + " received event '" + event + "' with data: " + data);
    }
}

class ConcreteObserverB implements Observer {
    factor: number;

    constructor(factor: number) {
        this.factor = factor;
    }

    update(event: string, data: number): void {
        console.log("ObserverB computed: " + (data * this.factor));
    }
}

class Subject {
    private observers: Observer[] = [];

    attach(observer: Observer): void {
        this.observers.push(observer);
    }

    notify(event: string, data: number): void {
        for (let i = 0; i < this.observers.length; i++) {
            this.observers[i].update(event, data);
        }
    }
}

const subject = new Subject();
const obsA = new ConcreteObserverA("Alpha");
const obsB = new ConcreteObserverB(5);

subject.attach(obsA);
subject.attach(obsB);

subject.notify("price_change", 100);
subject.notify("volume_spike", 250);
