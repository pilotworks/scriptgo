// Object-Oriented Programming: Abstract classes, inheritance, access modifiers

abstract class Account {
  protected balance: number;

  constructor(
    public readonly accountNumber: string,
    initialBalance: number
  ) {
    this.balance = initialBalance;
  }

  public getBalance(): number {
    return this.balance;
  }

  public abstract withdraw(amount: number): boolean;

  public deposit(amount: number): void {
    if (amount > 0) {
      this.balance += amount;
      console.log(`Deposited $${amount} to ${this.accountNumber}. New balance: $${this.balance}`);
    }
  }
}

class SavingsAccount extends Account {
  constructor(
    accountNumber: string,
    initialBalance: number,
    private interestRate: number
  ) {
    super(accountNumber, initialBalance);
  }

  public withdraw(amount: number): boolean {
    if (amount <= this.balance) {
      this.balance -= amount;
      console.log(`Withdrew $${amount} successfully. Remaining balance: $${this.balance}`);
      return true;
    }
    console.log("Withdrawal failed: Insufficient funds.");
    return false;
  }

  public applyInterest(): void {
    const interest = this.balance * this.interestRate;
    this.balance += interest;
    console.log(`Applied interest +$${interest}. New balance: $${this.balance}`);
  }
}

const myAccount = new SavingsAccount("ACC-9921", 1000, 0.05);
myAccount.deposit(500);
myAccount.applyInterest();
myAccount.withdraw(200);
