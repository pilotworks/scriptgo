// @expect: 3 + 5 * 2 = 13
// @expect: (3 + 5) * 2 = 16
// @expect: 100 * ( 2 + 12 ) / 14 = 100
// @expect: 50 + 20 * 3 - 40 / 2 = 90
function getPrecedence(op: string): number {
    if (op === "+" || op === "-") return 1;
    if (op === "*" || op === "/") return 2;
    return 0;
}

function applyOp(a: number, b: number, op: string): number {
    if (op === "+") return a + b;
    if (op === "-") return a - b;
    if (op === "*") return a * b;
    if (op === "/") return Math.floor(a / b);
    return 0;
}

function isDigit(ch: string): boolean {
    return ch >= "0" && ch <= "9";
}

function evaluateInfix(expr: string): number {
    const values: number[] = [];
    const ops: string[] = [];

    let i = 0;
    while (i < expr.length) {
        const ch = expr[i];

        if (ch === " ") {
            i++;
            continue;
        }

        if (ch === "(") {
            ops.push(ch);
            i++;
        } else if (isDigit(ch)) {
            let val = 0;
            while (i < expr.length && isDigit(expr[i])) {
                val = val * 10 + (expr.charCodeAt(i) - 48);
                i++;
            }
            values.push(val);
        } else if (ch === ")") {
            while (ops.length > 0 && ops[ops.length - 1] !== "(") {
                const op = ops.pop()!;
                const val2 = values.pop()!;
                const val1 = values.pop()!;
                values.push(applyOp(val1, val2, op));
            }
            ops.pop(); // remove '('
            i++;
        } else {
            // Operator
            while (ops.length > 0 && getPrecedence(ops[ops.length - 1]) >= getPrecedence(ch)) {
                const op = ops.pop()!;
                const val2 = values.pop()!;
                const val1 = values.pop()!;
                values.push(applyOp(val1, val2, op));
            }
            ops.push(ch);
            i++;
        }
    }

    while (ops.length > 0) {
        const op = ops.pop()!;
        const val2 = values.pop()!;
        const val1 = values.pop()!;
        values.push(applyOp(val1, val2, op));
    }

    return values.length > 0 ? values[0] : 0;
}

console.log("3 + 5 * 2 = " + evaluateInfix("3 + 5 * 2"));
console.log("(3 + 5) * 2 = " + evaluateInfix("(3 + 5) * 2"));
console.log("100 * ( 2 + 12 ) / 14 = " + evaluateInfix("100 * ( 2 + 12 ) / 14"));
console.log("50 + 20 * 3 - 40 / 2 = " + evaluateInfix("50 + 20 * 3 - 40 / 2"));
