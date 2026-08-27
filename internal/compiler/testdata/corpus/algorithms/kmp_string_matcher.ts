// @expect: [0,9,12]
// @expect: -1
function buildKMPTable(pattern: string): number[] {
    const lps: number[] = [];
    for (let i = 0; i < pattern.length; i++) {
        lps.push(0);
    }

    let len = 0;
    let i = 1;

    while (i < pattern.length) {
        if (pattern[i] === pattern[len]) {
            len++;
            lps[i] = len;
            i++;
        } else {
            if (len !== 0) {
                len = lps[len - 1];
            } else {
                lps[i] = 0;
                i++;
            }
        }
    }
    return lps;
}

function kmpSearch(text: string, pattern: string): number[] {
    const matches: number[] = [];
    if (pattern.length === 0) return matches;

    const lps = buildKMPTable(pattern);
    let i = 0; // index for text
    let j = 0; // index for pattern

    while (i < text.length) {
        if (pattern[j] === text[i]) {
            i++;
            j++;
        }

        if (j === pattern.length) {
            matches.push(i - j);
            j = lps[j - 1];
        } else if (i < text.length && pattern[j] !== text[i]) {
            if (j !== 0) {
                j = lps[j - 1];
            } else {
                i++;
            }
        }
    }
    return matches;
}

const text = "AABAACAADAABAABA";
const pattern = "AABA";
const occurrences = kmpSearch(text, pattern);
console.log(JSON.stringify(occurrences));

const notFound = kmpSearch("ABCDEF", "XYZ");
console.log(notFound.length > 0 ? notFound[0] : -1);
