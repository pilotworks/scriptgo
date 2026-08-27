// @expect: Tokens: 5
// @expect: Result: 23
// @expect: Result: 42
enum TokenType {
    NUMBER,
    PLUS,
    MINUS,
    STAR,
    SLASH,
    LPAREN,
    RPAREN,
    EOF
}

interface Token {
    type: TokenType;
    value: number;
}

class Lexer {
    private src: string;
    private pos: number = 0;

    constructor(src: string) {
        this.src = src;
    }

    nextToken(): Token {
        while (this.pos < this.src.length && this.src[this.pos] === " ") {
            this.pos++;
        }

        if (this.pos >= this.src.length) {
            return { type: TokenType.EOF, value: 0 };
        }

        const ch = this.src[this.pos];
        if (ch >= "0" && ch <= "9") {
            let num = 0;
            while (this.pos < this.src.length && this.src[this.pos] >= "0" && this.src[this.pos] <= "9") {
                num = num * 10 + (this.src.charCodeAt(this.pos) - 48);
                this.pos++;
            }
            return { type: TokenType.NUMBER, value: num };
        }

        this.pos++;
        switch (ch) {
            case "+":
                return { type: TokenType.PLUS, value: 0 };
            case "-":
                return { type: TokenType.MINUS, value: 0 };
            case "*":
                return { type: TokenType.STAR, value: 0 };
            case "/":
                return { type: TokenType.SLASH, value: 0 };
            case "(":
                return { type: TokenType.LPAREN, value: 0 };
            case ")":
                return { type: TokenType.RPAREN, value: 0 };
            default:
                return { type: TokenType.EOF, value: 0 };
        }
    }
}

class Parser {
    private lexer: Lexer;
    private currentToken: Token;

    constructor(lexer: Lexer) {
        this.lexer = lexer;
        this.currentToken = this.lexer.nextToken();
    }

    private eat(type: TokenType): void {
        if (this.currentToken.type === type) {
            this.currentToken = this.lexer.nextToken();
        }
    }

    private factor(): number {
        const token = this.currentToken;
        if (token.type === TokenType.NUMBER) {
            this.eat(TokenType.NUMBER);
            return token.value;
        } else if (token.type === TokenType.LPAREN) {
            this.eat(TokenType.LPAREN);
            const result = this.expr();
            this.eat(TokenType.RPAREN);
            return result;
        }
        return 0;
    }

    private term(): number {
        let result = this.factor();
        while (this.currentToken.type === TokenType.STAR || this.currentToken.type === TokenType.SLASH) {
            const token = this.currentToken;
            if (token.type === TokenType.STAR) {
                this.eat(TokenType.STAR);
                result = result * this.factor();
            } else if (token.type === TokenType.SLASH) {
                this.eat(TokenType.SLASH);
                result = Math.floor(result / this.factor());
            }
        }
        return result;
    }

    expr(): number {
        let result = this.term();
        while (this.currentToken.type === TokenType.PLUS || this.currentToken.type === TokenType.MINUS) {
            const token = this.currentToken;
            if (token.type === TokenType.PLUS) {
                this.eat(TokenType.PLUS);
                result = result + this.term();
            } else if (token.type === TokenType.MINUS) {
                this.eat(TokenType.MINUS);
                result = result - this.term();
            }
        }
        return result;
    }
}

const input1 = "3 + 4 * 5";
const lex1 = new Lexer(input1);
const tokensCount = (() => {
    const lex = new Lexer(input1);
    let count = 0;
    let t = lex.nextToken();
    while (t.type !== TokenType.EOF) {
        count++;
        t = lex.nextToken();
    }
    return count;
})();

console.log("Tokens: " + tokensCount);

const parser1 = new Parser(new Lexer(input1));
console.log("Result: " + parser1.expr());

const input2 = "(2 + 4) * (10 - 3)";
const parser2 = new Parser(new Lexer(input2));
console.log("Result: " + parser2.expr());
