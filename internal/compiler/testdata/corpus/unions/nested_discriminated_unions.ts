// @expect: -20
// @expect: 200
type LiteralExpr = {
    kind: "literal";
    value: number;
};

type BinaryExpr = {
    kind: "binary";
    op: "+" | "-" | "*";
    left: Expr;
    right: Expr;
};

type UnaryExpr = {
    kind: "unary";
    op: "-";
    expr: Expr;
};

type ConditionalExpr = {
    kind: "cond";
    cond: Expr;
    thenBranch: Expr;
    elseBranch: Expr;
};

type Expr = LiteralExpr | BinaryExpr | UnaryExpr | ConditionalExpr;

function evalExpr(expr: Expr): number {
    switch (expr.kind) {
        case "literal":
            return expr.value;
        case "unary":
            if (expr.op === "-") {
                return -evalExpr(expr.expr);
            }
            return 0;
        case "binary": {
            const l = evalExpr(expr.left);
            const r = evalExpr(expr.right);
            if (expr.op === "+") return l + r;
            if (expr.op === "-") return l - r;
            if (expr.op === "*") return l * r;
            return 0;
        }
        case "cond": {
            const c = evalExpr(expr.cond);
            return c !== 0 ? evalExpr(expr.thenBranch) : evalExpr(expr.elseBranch);
        }
    }
}

// (2 + 3) * -4 = -20
const ast1: Expr = {
    kind: "binary",
    op: "*",
    left: {
        kind: "binary",
        op: "+",
        left: { kind: "literal", value: 2 },
        right: { kind: "literal", value: 3 }
    },
    right: {
        kind: "unary",
        op: "-",
        expr: { kind: "literal", value: 4 }
    }
};

console.log(evalExpr(ast1));

// cond: (5 - 5) ? 100 : 200 = 200
const ast2: Expr = {
    kind: "cond",
    cond: {
        kind: "binary",
        op: "-",
        left: { kind: "literal", value: 5 },
        right: { kind: "literal", value: 5 }
    },
    thenBranch: { kind: "literal", value: 100 },
    elseBranch: { kind: "literal", value: 200 }
};

console.log(evalExpr(ast2));
