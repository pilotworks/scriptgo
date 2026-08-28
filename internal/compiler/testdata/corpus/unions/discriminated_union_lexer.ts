// @expect: IDENT(let)@0
// @expect: IDENT(result)@4
// @expect: OP(=)@11
// @expect: OP(()@13
// @expect: NUMBER(42)@14
// @expect: OP(+)@17
// @expect: IDENT(count)@19
// @expect: OP())@24
// @expect: OP(*)@26
// @expect: NUMBER(10)@28
// @expect: EOF@30
type TokenNumber = { type: "NUMBER"; value: number; pos: number };
type TokenIdent = { type: "IDENT"; name: string; pos: number };
type TokenOperator = { type: "OPERATOR"; symbol: string; pos: number };
type TokenEOF = { type: "EOF"; pos: number };

type Token = TokenNumber | TokenIdent | TokenOperator | TokenEOF;

function tokenize(input: string): Token[] {
    const tokens: Token[] = [];
    let i = 0;

    while (i < input.length) {
        const ch = input.charAt(i);

        if (ch === " " || ch === "\t" || ch === "\n" || ch === "\r") {
            i++;
            continue;
        }

        if (ch >= "0" && ch <= "9") {
            const start = i;
            let numStr = "";
            while (i < input.length && input.charAt(i) >= "0" && input.charAt(i) <= "9") {
                numStr += input.charAt(i);
                i++;
            }
            tokens.push({ type: "NUMBER", value: Number(numStr), pos: start });
            continue;
        }

        if ((ch >= "a" && ch <= "z") || (ch >= "A" && ch <= "Z") || ch === "_") {
            const start = i;
            let identStr = "";
            while (i < input.length && (
                (input.charAt(i) >= "a" && input.charAt(i) <= "z") ||
                (input.charAt(i) >= "A" && input.charAt(i) <= "Z") ||
                (input.charAt(i) >= "0" && input.charAt(i) <= "9") ||
                input.charAt(i) === "_"
            )) {
                identStr += input.charAt(i);
                i++;
            }
            tokens.push({ type: "IDENT", name: identStr, pos: start });
            continue;
        }

        if (ch === "+" || ch === "-" || ch === "*" || ch === "/" || ch === "=" || ch === "(" || ch === ")") {
            tokens.push({ type: "OPERATOR", symbol: ch, pos: i });
            i++;
            continue;
        }

        throw new Error(`Unexpected character at ${i}: ${ch}`);
    }

    tokens.push({ type: "EOF", pos: i });
    return tokens;
}

const tokens = tokenize("let result = (42 + count) * 10");
for (const tok of tokens) {
    switch (tok.type) {
        case "NUMBER":
            console.log(`NUMBER(${tok.value})@${tok.pos}`);
            break;
        case "IDENT":
            console.log(`IDENT(${tok.name})@${tok.pos}`);
            break;
        case "OPERATOR":
            console.log(`OP(${tok.symbol})@${tok.pos}`);
            break;
        case "EOF":
            console.log(`EOF@${tok.pos}`);
            break;
    }
}
