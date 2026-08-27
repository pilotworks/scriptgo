// @expect: state: idle
// @expect: state: loading (progress: 0)
// @expect: state: loading (progress: 50)
// @expect: state: success (data: done)
type State =
    | { status: "idle" }
    | { status: "loading"; progress: number }
    | { status: "success"; data: string }
    | { status: "error"; error: string };

type MachineEvent =
    | { type: "FETCH" }
    | { type: "PROGRESS"; value: number }
    | { type: "RESOLVE"; data: string }
    | { type: "REJECT"; error: string };

function transition(state: State, event: MachineEvent): State {
    switch (state.status) {
        case "idle":
            if (event.type === "FETCH") {
                return { status: "loading", progress: 0 };
            }
            return state;
        case "loading":
            if (event.type === "PROGRESS") {
                return { status: "loading", progress: event.value };
            }
            if (event.type === "RESOLVE") {
                return { status: "success", data: event.data };
            }
            if (event.type === "REJECT") {
                return { status: "error", error: event.error };
            }
            return state;
        default:
            return state;
    }
}

function printState(state: State): void {
    if (state.status === "idle") {
        console.log("state: idle");
    } else if (state.status === "loading") {
        console.log("state: loading (progress: " + state.progress + ")");
    } else if (state.status === "success") {
        console.log("state: success (data: " + state.data + ")");
    } else {
        console.log("state: error (" + state.error + ")");
    }
}

let current: State = { status: "idle" };
printState(current);

current = transition(current, { type: "FETCH" });
printState(current);

current = transition(current, { type: "PROGRESS", value: 50 });
printState(current);

current = transition(current, { type: "RESOLVE", data: "done" });
printState(current);
