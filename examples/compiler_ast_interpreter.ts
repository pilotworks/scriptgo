/**
 * Arithmetic & Logical Expression AST Interpreter
 * 
 * Demonstrates:
 * - Tokenizer / Lexer with string parsing & character classification
 * - Recursive Descent Parser constructing strongly-typed AST
 * - Discriminated Unions for AST Node shapes (`kind: "Binary" | "Unary" | "Literal" | "Variable" | "Call"`)
 * - Environment evaluation with lexical scope chaining
 * - Standard TypeScript error handling & diagnostics
 */

export type TokenType =
    | "NUMBER"
    | "IDENTIFIER"
    | "PLUS"
    | "MINUS"
    | "STAR"
    | "SLASH"
    | "PERCENT"
    | "CARET"
    | "LPAREN"
    | "RPAREN"
    | "COMMA"
    | "EQUAL"
    | "EQEQ"
    | "NOTEQ"
    | "LT"
    | "LTE"
    | "GT"
    | "GTE"
    | "AND"
    | "OR"
    | "NOT"
    | "EOF";

export interface Token {
    type: TokenType;
    value: string;
    pos: number;
}

export class Lexer {
    private pos: number = 0;
    private readonly source: string;

    constructor(source: string) {
        this.source = source;
    }

    private isDigit(ch: string): boolean {
        return ch >= "0" && ch <= "9";
    }

    private isAlpha(ch: string): boolean {
        return (ch >= "a" && ch <= "z") || (ch >= "A" && ch <= "Z") || ch === "_";
    }

    private peek(): string {
        if (this.pos >= this.source.length) return "\0";
        return this.source.charAt(this.pos);
    }

    private advance(): string {
        if (this.pos >= this.source.length) return "\0";
        const ch = this.source.charAt(this.pos);
        this.pos++;
        return ch;
    }

    nextToken(): Token {
        while (this.pos < this.source.length) {
            const ch = this.peek();

            if (ch === " " || ch === "\t" || ch === "\n" || ch === "\r") {
                this.advance();
                continue;
            }

            const startPos = this.pos;

            if (this.isDigit(ch)) {
                let numStr = "";
                let hasDot = false;
                while (this.isDigit(this.peek()) || (this.peek() === "." && !hasDot)) {
                    if (this.peek() === ".") hasDot = true;
                    numStr += this.advance();
                }
                return { type: "NUMBER", value: numStr, pos: startPos };
            }

            if (this.isAlpha(ch)) {
                let idStr = "";
                while (this.isAlpha(this.peek()) || this.isDigit(this.peek())) {
                    idStr += this.advance();
                }
                if (idStr === "and") return { type: "AND", value: idStr, pos: startPos };
                if (idStr === "or") return { type: "OR", value: idStr, pos: startPos };
                if (idStr === "not") return { type: "NOT", value: idStr, pos: startPos };
                return { type: "IDENTIFIER", value: idStr, pos: startPos };
            }

            this.advance();

            switch (ch) {
                case "+": return { type: "PLUS", value: "+", pos: startPos };
                case "-": return { type: "MINUS", value: "-", pos: startPos };
                case "*": return { type: "STAR", value: "*", pos: startPos };
                case "/": return { type: "SLASH", value: "/", pos: startPos };
                case "%": return { type: "PERCENT", value: "%", pos: startPos };
                case "^": return { type: "CARET", value: "^", pos: startPos };
                case "(": return { type: "LPAREN", value: "(", pos: startPos };
                case ")": return { type: "RPAREN", value: ")", pos: startPos };
                case ",": return { type: "COMMA", value: ",", pos: startPos };
                case "=":
                    if (this.peek() === "=") {
                        this.advance();
                        return { type: "EQEQ", value: "==", pos: startPos };
                    }
                    return { type: "EQUAL", value: "=", pos: startPos };
                case "!":
                    if (this.peek() === "=") {
                        this.advance();
                        return { type: "NOTEQ", value: "!=", pos: startPos };
                    }
                    return { type: "NOT", value: "!", pos: startPos };
                case "<":
                    if (this.peek() === "=") {
                        this.advance();
                        return { type: "LTE", value: "<=", pos: startPos };
                    }
                    return { type: "LT", value: "<", pos: startPos };
                case ">":
                    if (this.peek() === "=") {
                        this.advance();
                        return { type: "GTE", value: ">=", pos: startPos };
                    }
                    return { type: "GT", value: ">", pos: startPos };
                case "&":
                    if (this.peek() === "&") {
                        this.advance();
                        return { type: "AND", value: "&&", pos: startPos };
                    }
                    break;
                case "|":
                    if (this.peek() === "|") {
                        this.advance();
                        return { type: "OR", value: "||", pos: startPos };
                    }
                    break;
            }

            throw new Error(`Unexpected character '${ch}' at position ${startPos}`);
        }

        return { type: "EOF", value: "", pos: this.pos };
    }
}

// ==========================================
// AST Node Hierarchy (Discriminated Unions)
// ==========================================

export interface LiteralNode {
    kind: "Literal";
    value: number;
}

export interface VariableNode {
    kind: "Variable";
    name: string;
}

export interface UnaryNode {
    kind: "Unary";
    op: string;
    right: ASTNode;
}

export interface BinaryNode {
    kind: "Binary";
    op: string;
    left: ASTNode;
    right: ASTNode;
}

export interface CallNode {
    kind: "Call";
    callee: string;
    args: ASTNode[];
}

export interface AssignNode {
    kind: "Assign";
    name: string;
    value: ASTNode;
}

export type ASTNode =
    | LiteralNode
    | VariableNode
    | UnaryNode
    | BinaryNode
    | CallNode
    | AssignNode;

// ==========================================
// Parser
// ==========================================

export class Parser {
    private currentToken: Token;
    private lexer: Lexer;

    constructor(source: string) {
        this.lexer = new Lexer(source);
        this.currentToken = this.lexer.nextToken();
    }

    private eat(tokenType: TokenType): Token {
        const token = this.currentToken;
        if (token.type !== tokenType) {
            throw new Error(`Expected token ${tokenType} but got ${token.type} ('${token.value}') at pos ${token.pos}`);
        }
        this.currentToken = this.lexer.nextToken();
        return token;
    }

    parse(): ASTNode {
        const node = this.parseAssignment();
        if (this.currentToken.type !== "EOF") {
            throw new Error(`Unexpected token after expression: ${this.currentToken.type} ('${this.currentToken.value}')`);
        }
        return node;
    }

    private parseAssignment(): ASTNode {
        const expr = this.parseLogicalOr();
        if (expr.kind === "Variable" && this.currentToken.type === "EQUAL") {
            this.eat("EQUAL");
            const val = this.parseAssignment();
            const assignNode: AssignNode = {
                kind: "Assign",
                name: expr.name,
                value: val
            };
            return assignNode;
        }
        return expr;
    }

    private parseLogicalOr(): ASTNode {
        let node = this.parseLogicalAnd();
        while (this.currentToken.type === "OR") {
            const op = this.currentToken.value;
            this.eat("OR");
            const right = this.parseLogicalAnd();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseLogicalAnd(): ASTNode {
        let node = this.parseEquality();
        while (this.currentToken.type === "AND") {
            const op = this.currentToken.value;
            this.eat("AND");
            const right = this.parseEquality();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseEquality(): ASTNode {
        let node = this.parseRelational();
        while (this.currentToken.type === "EQEQ" || this.currentToken.type === "NOTEQ") {
            const op = this.currentToken.value;
            this.eat(this.currentToken.type);
            const right = this.parseRelational();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseRelational(): ASTNode {
        let node = this.parseAdditive();
        while (
            this.currentToken.type === "LT" ||
            this.currentToken.type === "LTE" ||
            this.currentToken.type === "GT" ||
            this.currentToken.type === "GTE"
        ) {
            const op = this.currentToken.value;
            this.eat(this.currentToken.type);
            const right = this.parseAdditive();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseAdditive(): ASTNode {
        let node = this.parseMultiplicative();
        while (this.currentToken.type === "PLUS" || this.currentToken.type === "MINUS") {
            const op = this.currentToken.value;
            this.eat(this.currentToken.type);
            const right = this.parseMultiplicative();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseMultiplicative(): ASTNode {
        let node = this.parseExponential();
        while (
            this.currentToken.type === "STAR" ||
            this.currentToken.type === "SLASH" ||
            this.currentToken.type === "PERCENT"
        ) {
            const op = this.currentToken.value;
            this.eat(this.currentToken.type);
            const right = this.parseExponential();
            node = { kind: "Binary", op, left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseExponential(): ASTNode {
        let node = this.parseUnary();
        if (this.currentToken.type === "CARET") {
            this.eat("CARET");
            const right = this.parseExponential(); // Right-associative
            node = { kind: "Binary", op: "^", left: node, right } as BinaryNode;
        }
        return node;
    }

    private parseUnary(): ASTNode {
        if (this.currentToken.type === "MINUS" || this.currentToken.type === "PLUS" || this.currentToken.type === "NOT") {
            const op = this.currentToken.value;
            this.eat(this.currentToken.type);
            const right = this.parseUnary();
            const unaryNode: UnaryNode = { kind: "Unary", op, right };
            return unaryNode;
        }
        return this.parsePrimary();
    }

    private parsePrimary(): ASTNode {
        const token = this.currentToken;

        if (token.type === "NUMBER") {
            this.eat("NUMBER");
            const literal: LiteralNode = { kind: "Literal", value: parseFloat(token.value) };
            return literal;
        }

        if (token.type === "IDENTIFIER") {
            const name = token.value;
            this.eat("IDENTIFIER");

            // Function Call: name(arg1, arg2, ...)
            if (this.currentToken.type === "LPAREN") {
                this.eat("LPAREN");
                const args: ASTNode[] = [];
                if ((this.currentToken.type as TokenType) !== "RPAREN") {
                    args.push(this.parseAssignment());
                    while ((this.currentToken.type as TokenType) === "COMMA") {
                        this.eat("COMMA");
                        args.push(this.parseAssignment());
                    }
                }
                this.eat("RPAREN");
                const callNode: CallNode = { kind: "Call", callee: name, args };
                return callNode;
            }

            const varNode: VariableNode = { kind: "Variable", name };
            return varNode;
        }

        if (token.type === "LPAREN") {
            this.eat("LPAREN");
            const node = this.parseAssignment();
            this.eat("RPAREN");
            return node;
        }

        throw new Error(`Unexpected token in expression: ${token.type} ('${token.value}') at pos ${token.pos}`);
    }
}

// ==========================================
// Environment & Evaluator
// ==========================================

export class Environment {
    private vars: Map<string, number> = new Map();
    private parent?: Environment;

    constructor(parent?: Environment) {
        this.parent = parent;
    }

    set(name: string, value: number): void {
        this.vars.set(name, value);
    }

    get(name: string): number {
        if (this.vars.has(name)) {
            return this.vars.get(name)!;
        }
        if (this.parent) {
            return this.parent.get(name);
        }
        throw new Error(`Undefined variable '${name}'`);
    }
}

export class Evaluator {
    private env: Environment;

    constructor(env?: Environment) {
        this.env = env || new Environment();
        // Setup standard constants
        this.env.set("PI", Math.PI);
        this.env.set("E", Math.E);
    }

    getEnvironment(): Environment {
        return this.env;
    }

    evaluate(node: ASTNode): number {
        switch (node.kind) {
            case "Literal":
                return node.value;

            case "Variable":
                return this.env.get(node.name);

            case "Assign": {
                const val = this.evaluate(node.value);
                this.env.set(node.name, val);
                return val;
            }

            case "Unary": {
                const r = this.evaluate(node.right);
                if (node.op === "-") return -r;
                if (node.op === "+") return +r;
                if (node.op === "!" || node.op === "not") return r === 0 ? 1 : 0;
                throw new Error(`Unknown unary operator '${node.op}'`);
            }

            case "Binary": {
                const left = this.evaluate(node.left);
                // Short-circuit logical ops
                if (node.op === "&&" || node.op === "and") {
                    return (left !== 0 && this.evaluate(node.right) !== 0) ? 1 : 0;
                }
                if (node.op === "||" || node.op === "or") {
                    return (left !== 0 || this.evaluate(node.right) !== 0) ? 1 : 0;
                }

                const right = this.evaluate(node.right);
                switch (node.op) {
                    case "+": return left + right;
                    case "-": return left - right;
                    case "*": return left * right;
                    case "/":
                        if (right === 0) throw new Error("Division by zero");
                        return left / right;
                    case "%": return left % right;
                    case "^": return Math.pow(left, right);
                    case "==": return left === right ? 1 : 0;
                    case "!=": return left !== right ? 1 : 0;
                    case "<": return left < right ? 1 : 0;
                    case "<=": return left <= right ? 1 : 0;
                    case ">": return left > right ? 1 : 0;
                    case ">=": return left >= right ? 1 : 0;
                    default:
                        throw new Error(`Unknown binary operator '${node.op}'`);
                }
            }

            case "Call": {
                const evaluatedArgs = node.args.map(arg => this.evaluate(arg));
                switch (node.callee) {
                    case "sin": return Math.sin(evaluatedArgs[0]);
                    case "cos": return Math.cos(evaluatedArgs[0]);
                    case "tan": return Math.tan(evaluatedArgs[0]);
                    case "sqrt": return Math.sqrt(evaluatedArgs[0]);
                    case "abs": return Math.abs(evaluatedArgs[0]);
                    case "floor": return Math.floor(evaluatedArgs[0]);
                    case "ceil": return Math.ceil(evaluatedArgs[0]);
                    case "round": return Math.round(evaluatedArgs[0]);
                    case "min": return Math.min(...evaluatedArgs);
                    case "max": return Math.max(...evaluatedArgs);
                    case "pow": return Math.pow(evaluatedArgs[0], evaluatedArgs[1]);
                    default:
                        throw new Error(`Unknown function '${node.callee}'`);
                }
            }
        }
    }

    eval(expressionText: string): number {
        const parser = new Parser(expressionText);
        const ast = parser.parse();
        return this.evaluate(ast);
    }
}

// ==========================================
// Demonstration
// ==========================================

function main(): void {
    console.log("=== Arithmetic & Logical Expression Interpreter ===");

    const evaluator = new Evaluator();

    const testExpressions = [
        "10 + 20 * 3",
        "(10 + 20) * 3",
        "2 ^ 3 ^ 2", // right associative: 2^(3^2) = 2^9 = 512
        "100 / 5 / 2", // left associative: (100/5)/2 = 10
        "sqrt(16) + abs(-25)",
        "min(10, 5, 20, 3) * max(2, 8, 4)",
        "sin(PI / 2) + cos(0)",
        "10 > 5 && 20 <= 30",
        "!(5 == 5) || (10 != 2)",
        "x = 42",
        "y = x * 2 + 6",
        "sqrt(y + x + 10)"
    ];

    for (let i = 0; i < testExpressions.length; i++) {
        const expr = testExpressions[i];
        try {
            const result = evaluator.eval(expr);
            console.log(`[EVAL] "${expr}" => ${result}`);
        } catch (err: any) {
            console.error(`[ERROR] "${expr}" => ${err.message}`);
        }
    }
}

main();
