// @expect: 10
// @expect: 6
// @expect: 0
type IncrementAction = { type: "INCREMENT"; amount: number };
type DecrementAction = { type: "DECREMENT"; amount: number };
type ResetAction = { type: "RESET" };

type Action = IncrementAction | DecrementAction | ResetAction;

function reducer(state: number, action: Action): number {
    switch (action.type) {
        case "INCREMENT":
            return state + action.amount;
        case "DECREMENT":
            return state - action.amount;
        case "RESET":
            return 0;
    }
}

let state = 0;
state = reducer(state, { type: "INCREMENT", amount: 10 });
console.log(state);
state = reducer(state, { type: "DECREMENT", amount: 4 });
console.log(state);
state = reducer(state, { type: "RESET" });
console.log(state);
