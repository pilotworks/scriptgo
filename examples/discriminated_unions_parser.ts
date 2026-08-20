type Action =
  | { type: "SET_COUNT"; payload: number }
  | { type: "INCREMENT"; step: number }
  | { type: "DECREMENT"; step: number }
  | { type: "RESET" };

interface AppState {
  count: number;
  history: string[];
}

function stateReducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case "SET_COUNT":
      return {
        count: action.payload,
        history: [...state.history, `Set to ${action.payload}`],
      };
    case "INCREMENT":
      return {
        count: state.count + action.step,
        history: [...state.history, `Incremented by ${action.step}`],
      };
    case "DECREMENT":
      return {
        count: state.count - action.step,
        history: [...state.history, `Decremented by ${action.step}`],
      };
    case "RESET":
      return {
        count: 0,
        history: [...state.history, "Reset to 0"],
      };
  }
}

console.log("=== Discriminated Unions Reducer ===");
let currentState: AppState = { count: 0, history: [] };

currentState = stateReducer(currentState, { type: "SET_COUNT", payload: 10 });
currentState = stateReducer(currentState, { type: "INCREMENT", step: 5 });
currentState = stateReducer(currentState, { type: "DECREMENT", step: 2 });
currentState = stateReducer(currentState, { type: "RESET" });

console.log(`Final Count: ${currentState.count}`);
console.log(`Event Log: ${currentState.history.join(" -> ")}`);
