class BankAccount {
    accountNumber: string;
    balance: number;
    transactionCount: number;

    constructor(accountNumber: string, initialDeposit: number) {
        this.accountNumber = accountNumber;
        this.balance = initialDeposit;
        this.transactionCount = 0;
    }

    deposit(amount: number): boolean {
        if (amount <= 0) {
            return false;
        }
        this.balance += amount;
        this.transactionCount++;
        return true;
    }

    withdraw(amount: number): boolean {
        if (amount <= 0 || amount > this.balance) {
            return false;
        }
        this.balance -= amount;
        this.transactionCount++;
        return true;
    }

    transferTo(target: BankAccount, amount: number): boolean {
        if (this.withdraw(amount)) {
            target.deposit(amount);
            return true;
        }
        return false;
    }

    getStatement(): string {
        return `Account ${this.accountNumber}: Balance=$${this.balance} (TxCount=${this.transactionCount})`;
    }
}

const acc1 = new BankAccount("A-100", 500);
const acc2 = new BankAccount("B-200", 100);

console.log(acc1.getStatement());
console.log(acc2.getStatement());

console.log(acc1.deposit(200));
console.log(acc1.withdraw(50));
console.log(acc1.withdraw(1000)); // Should fail

console.log(acc1.transferTo(acc2, 300));
console.log(acc1.transferTo(acc2, 9999)); // Should fail

console.log(acc1.getStatement());
console.log(acc2.getStatement());
