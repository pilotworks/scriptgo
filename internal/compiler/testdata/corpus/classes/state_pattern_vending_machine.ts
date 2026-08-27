// @expect: Inserted 50
// @expect: Dispensed Soda
// @expect: Current state: NoCoinState
interface State {
    insertCoin(amount: number): void;
    selectItem(item: string): string | null;
    dispense(): string | null;
    getName(): string;
}

class VendingMachine {
    balance: number = 0;
    selectedItem: string | null = null;
    currentState: State;

    constructor() {
        this.currentState = new NoCoinState(this);
    }

    setState(state: State): void {
        this.currentState = state;
    }

    insertCoin(amount: number): void {
        this.currentState.insertCoin(amount);
    }

    selectItem(item: string): void {
        this.currentState.selectItem(item);
    }

    dispense(): void {
        const item = this.currentState.dispense();
        if (item) {
            console.log("Dispensed " + item);
        }
    }
}

class NoCoinState implements State {
    machine: VendingMachine;

    constructor(machine: VendingMachine) {
        this.machine = machine;
    }

    getName(): string {
        return "NoCoinState";
    }

    insertCoin(amount: number): void {
        this.machine.balance += amount;
        console.log("Inserted " + amount);
        this.machine.setState(new HasCoinState(this.machine));
    }

    selectItem(item: string): string | null {
        return null;
    }

    dispense(): string | null {
        return null;
    }
}

class HasCoinState implements State {
    machine: VendingMachine;

    constructor(machine: VendingMachine) {
        this.machine = machine;
    }

    getName(): string {
        return "HasCoinState";
    }

    insertCoin(amount: number): void {
        this.machine.balance += amount;
        console.log("Inserted " + amount);
    }

    selectItem(item: string): string | null {
        this.machine.selectedItem = item;
        this.machine.setState(new SoldState(this.machine));
        return item;
    }

    dispense(): string | null {
        return null;
    }
}

class SoldState implements State {
    machine: VendingMachine;

    constructor(machine: VendingMachine) {
        this.machine = machine;
    }

    getName(): string {
        return "SoldState";
    }

    insertCoin(amount: number): void {}

    selectItem(item: string): string | null {
        return null;
    }

    dispense(): string | null {
        const item = this.machine.selectedItem;
        this.machine.selectedItem = null;
        this.machine.balance = 0;
        this.machine.setState(new NoCoinState(this.machine));
        return item;
    }
}

const vm = new VendingMachine();
vm.insertCoin(50);
vm.selectItem("Soda");
vm.dispense();
console.log("Current state: " + vm.currentState.getName());
