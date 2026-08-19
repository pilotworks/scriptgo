class Account {
  public owner: string = "";
  private balance: number = 0;

  constructor(owner: string, initial: number) {
    this.owner = owner;
    this.balance = initial;
  }

  public deposit(amount: number): void {
    this.balance = this.balance + amount;
  }

  public getBalance(): number {
    return this.balance;
  }
}

const acc = new Account("Alice", 100);
acc.deposit(50);
console.log(acc.owner);
console.log(acc.getBalance());
