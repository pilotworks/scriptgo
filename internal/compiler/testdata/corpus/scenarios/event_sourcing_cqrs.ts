// @expect: Balance: 350
// @expect: Events count: 3
// @expect: Active accounts: 1
interface BankAccountCreatedEvent {
    type: "ACCOUNT_CREATED";
    accountId: string;
    owner: string;
    initialBalance: number;
}

interface MoneyDepositedEvent {
    type: "MONEY_DEPOSITED";
    accountId: string;
    amount: number;
}

interface MoneyWithdrawnEvent {
    type: "MONEY_WITHDRAWN";
    accountId: string;
    amount: number;
}

interface AccountClosedEvent {
    type: "ACCOUNT_CLOSED";
    accountId: string;
}

type BankEvent =
    | BankAccountCreatedEvent
    | MoneyDepositedEvent
    | MoneyWithdrawnEvent
    | AccountClosedEvent;

class BankAccountAggregate {
    id: string = "";
    owner: string = "";
    balance: number = 0;
    isOpen: boolean = false;

    apply(event: BankEvent): void {
        switch (event.type) {
            case "ACCOUNT_CREATED":
                this.id = event.accountId;
                this.owner = event.owner;
                this.balance = event.initialBalance;
                this.isOpen = true;
                break;
            case "MONEY_DEPOSITED":
                this.balance += event.amount;
                break;
            case "MONEY_WITHDRAWN":
                this.balance -= event.amount;
                break;
            case "ACCOUNT_CLOSED":
                this.isOpen = false;
                break;
        }
    }
}

class EventStore {
    private events: BankEvent[] = [];

    append(event: BankEvent): void {
        this.events.push(event);
    }

    getEvents(accountId: string): BankEvent[] {
        return this.events.filter((e) => e.accountId === accountId);
    }

    getAllEvents(): BankEvent[] {
        return this.events;
    }
}

// Projections / Query side
class AccountSummaryView {
    activeAccountsCount: number = 0;

    project(event: BankEvent): void {
        if (event.type === "ACCOUNT_CREATED") {
            this.activeAccountsCount++;
        } else if (event.type === "ACCOUNT_CLOSED") {
            this.activeAccountsCount--;
        }
    }
}

const store = new EventStore();
const view = new AccountSummaryView();

function dispatch(event: BankEvent): void {
    store.append(event);
    view.project(event);
}

dispatch({
    type: "ACCOUNT_CREATED",
    accountId: "ACC-101",
    owner: "Alice",
    initialBalance: 100
});

dispatch({
    type: "MONEY_DEPOSITED",
    accountId: "ACC-101",
    amount: 300
});

dispatch({
    type: "MONEY_WITHDRAWN",
    accountId: "ACC-101",
    amount: 50
});

dispatch({
    type: "ACCOUNT_CREATED",
    accountId: "ACC-102",
    owner: "Bob",
    initialBalance: 50
});

// Replay aggregate state for ACC-101
const acc101 = new BankAccountAggregate();
const history101 = store.getEvents("ACC-101");
for (let i = 0; i < history101.length; i++) {
    acc101.apply(history101[i]);
}

dispatch({
    type: "ACCOUNT_CLOSED",
    accountId: "ACC-102"
});

console.log("Balance: " + acc101.balance);
console.log("Events count: " + history101.length);
console.log("Active accounts: " + view.activeAccountsCount);
