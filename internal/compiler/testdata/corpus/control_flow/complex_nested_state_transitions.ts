// @expect: 31
// @expect: 40
// @expect: 13
function evaluateStateMachine(inputs: string[]): number {
    let state = 0;
    let counter = 0;

    outer: for (let i = 0; i < inputs.length; i++) {
        const token = inputs[i];

        switch (state) {
            case 0:
                if (token === "START") {
                    state = 1;
                    counter += 10;
                } else if (token === "SKIP") {
                    continue outer;
                } else {
                    counter -= 1;
                }
                break;
            case 1:
                switch (token) {
                    case "INC":
                        counter += 5;
                        break;
                    case "DOUBLE":
                        counter *= 2;
                        break;
                    case "RESET":
                        state = 0;
                        break;
                    case "STOP":
                        break outer;
                    default:
                        counter += 1;
                        break;
                }
                break;
            case 2:
                counter += 100;
                break;
            default:
                break;
        }
    }

    return counter;
}

console.log(evaluateStateMachine(["SKIP", "START", "INC", "DOUBLE", "EXTRA", "STOP", "INC"]));
console.log(evaluateStateMachine(["START", "RESET", "START", "DOUBLE"]));
console.log(evaluateStateMachine(["NOOP", "NOOP", "START", "INC"]));
