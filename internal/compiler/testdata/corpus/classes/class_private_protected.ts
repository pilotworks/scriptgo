// @expect: 1000
// @expect: 1050
class BaseAccount {
    protected balance: number;

    constructor(initial: number) {
        this.balance = initial;
    }

    getBalance(): number {
        return this.balance;
    }
}

class SavingsAccount extends BaseAccount {
    private interestRate: number;

    constructor(initial: number, rate: number) {
        super(initial);
        this.interestRate = rate;
    }

    addInterest(): void {
        this.balance += this.balance * this.interestRate;
    }
}

const acc = new SavingsAccount(1000, 0.05);
console.log(acc.getBalance());
acc.addInterest();
console.log(acc.getBalance());
