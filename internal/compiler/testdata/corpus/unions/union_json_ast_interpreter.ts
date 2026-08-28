// @expect: Evaluated AST result: 75
type LiteralExpr = { kind: "lit"; value: number };
type VarExpr = { kind: "var"; name: string };
type BinaryExpr = { kind: "bin"; op: "+" | "-" | "*" | "/"; left: Expr; right: Expr };
type LetExpr = { kind: "let"; name: string; value: Expr; body: Expr };

type Expr = LiteralExpr | VarExpr | BinaryExpr | LetExpr;

type Env = Map<string, number>;

function evalExpr(expr: Expr, env: Env): number {
    switch (expr.kind) {
        case "lit":
            return expr.value;
        case "var":
            if (!env.has(expr.name)) {
                throw new Error(`Unbound variable: ${expr.name}`);
            }
            return env.get(expr.name)!;
        case "bin": {
            const l = evalExpr(expr.left, env);
            const r = evalExpr(expr.right, env);
            switch (expr.op) {
                case "+": return l + r;
                case "-": return l - r;
                case "*": return l * r;
                case "/": return Math.floor(l / r);
            }
        }
        case "let": {
            const val = evalExpr(expr.value, env);
            const newEnv = new Map(env);
            newEnv.set(expr.name, val);
            return evalExpr(expr.body, newEnv);
        }
    }
}

// let x = 10 in let y = 5 in (x + y) * (x - y)
const program: Expr = {
    kind: "let",
    name: "x",
    value: { kind: "lit", value: 10 },
    body: {
        kind: "let",
        name: "y",
        value: { kind: "lit", value: 5 },
        body: {
            kind: "bin",
            op: "*",
            left: {
                kind: "bin",
                op: "+",
                left: { kind: "var", name: "x" },
                right: { kind: "var", name: "y" }
            },
            right: {
                kind: "bin",
                op: "-",
                left: { kind: "var", name: "x" },
                right: { kind: "var", name: "y" }
            }
        }
    }
};

const result = evalExpr(program, new Map());
console.log(`Evaluated AST result: ${result}`);
