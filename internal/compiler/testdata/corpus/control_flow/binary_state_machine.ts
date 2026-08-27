// @expect: Valid: true
// @expect: Valid: false
// @expect: Valid: true
// Accepts strings with an even number of '0's and even number of '1's
// States: 0 (even 0, even 1), 1 (odd 0, even 1), 2 (even 0, odd 1), 3 (odd 0, odd 1)
function validateEvenZerosAndOnes(input: string): boolean {
    let state = 0;
    for (let i = 0; i < input.length; i++) {
        const ch = input[i];
        switch (state) {
            case 0:
                if (ch === "0") state = 1;
                else if (ch === "1") state = 2;
                else return false;
                break;
            case 1:
                if (ch === "0") state = 0;
                else if (ch === "1") state = 3;
                else return false;
                break;
            case 2:
                if (ch === "0") state = 3;
                else if (ch === "1") state = 0;
                else return false;
                break;
            case 3:
                if (ch === "0") state = 2;
                else if (ch === "1") state = 1;
                else return false;
                break;
        }
    }
    return state === 0;
}

console.log("Valid: " + validateEvenZerosAndOnes("0011"));
console.log("Valid: " + validateEvenZerosAndOnes("01011"));
console.log("Valid: " + validateEvenZerosAndOnes("01100011"));
