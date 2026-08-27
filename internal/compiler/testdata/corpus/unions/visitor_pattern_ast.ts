// @expect: 14
// @expect: ((1 + (2 * 3)) + (14 / 2))
// @expect: true
type Expr =
    | { kind: "num"; value: number }
    | { kind: "add"; left: Expr; right: Expr }
    | { kind: "mul"; left: Expr; right: Expr }
    | { kind: "div"; left: Expr; right: Expr };

interface ExprVisitor<R> {
    visitNum(value: number): R;
    visitAdd(left: Expr, right: Expr): R;
    visitMul(left: Expr, right: Expr): R;
    visitDiv(left: Expr, right: Expr): R;
}

function accept<R>(expr: Expr, visitor: ExprVisitor<R>): R {
    switch (expr.kind) {
        case "num":
            return visitor.visitNum(expr.value);
        case "add":
            return visitor.visitAdd(expr.left, expr.right);
        case "mul":
            return visitor.visitMul(expr.left, expr.right);
        case "div":
            return visitor.visitDiv(expr.left, expr.right);
    }
}

class Evaluator implements ExprVisitor<number> {
    visitNum(value: number): number {
        return value;
    }
    visitAdd(left: Expr, right: Expr): number {
        return accept(left, this) + accept(right, this);
    }
    visitMul(left: Expr, right: Expr): number {
        return accept(left, this) * accept(right, this);
    }
    visitDiv(left: Expr, right: Expr): number {
        return accept(left, this) / accept(right, this);
    }
}

class Printer implements ExprVisitor<string> {
    visitNum(value: number): string {
        return String(value);
    }
    visitAdd(left: Expr, right: Expr): string {
        return "(" + accept(left, this) + " + " + accept(right, this) + ")";
    }
    visitMul(left: Expr, right: Expr): string {
        return "(" + accept(left, this) + " * " + accept(right, this) + ")";
    }
    visitDiv(left: Expr, right: Expr): string {
        return "(" + accept(left, this) + " / " + accept(right, this) + ")";
    }
}

// Expr: (1 + 2 * 3) + (14 / 2) = (1 + 6) + 7 = 14
const ast: Expr = {
    kind: "add",
    left: {
        kind: "add",
        left: { kind: "num", value: 1 },
        right: {
            kind: "mul",
            left: { kind: "num", value: 2 },
            right: { kind: "num", value: 3 }
        }
    },
    right: {
        kind: "div",
        left: { kind: "num", value: 14 },
        right: { kind: "num", value: 2 }
    }
};

const evalVisitor = new Evaluator();
const printVisitor = new Printer();

console.log(accept(ast, evalVisitor));
console.log(accept(ast, printVisitor));
console.log(accept(ast, evalVisitor) === 14);
