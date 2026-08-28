// @expect: 2
// @expect: 0
// @expect: 9
function rabinKarp(text: string, pattern: string): number[] {
    const matches: number[] = [];
    const n = text.length;
    const m = pattern.length;
    if (m === 0 || n < m) return matches;

    const base = 256;
    const prime = 101;
    let h = 1;

    for (let i = 0; i < m - 1; i++) {
        h = (h * base) % prime;
    }

    let pHash = 0;
    let tHash = 0;

    for (let i = 0; i < m; i++) {
        pHash = (base * pHash + pattern.charCodeAt(i)) % prime;
        tHash = (base * tHash + text.charCodeAt(i)) % prime;
    }

    for (let i = 0; i <= n - m; i++) {
        if (pHash === tHash) {
            let match = true;
            for (let j = 0; j < m; j++) {
                if (text.charCodeAt(i + j) !== pattern.charCodeAt(j)) {
                    match = false;
                    break;
                }
            }
            if (match) {
                matches.push(i);
            }
        }

        if (i < n - m) {
            tHash = (base * (tHash - text.charCodeAt(i) * h) + text.charCodeAt(i + m)) % prime;
            if (tHash < 0) {
                tHash = tHash + prime;
            }
        }
    }

    return matches;
}

const txt = "ABCCDABCEABCCDABCF";
const pat = "ABCC";
const results = rabinKarp(txt, pat);
console.log(results.length);
for (const idx of results) {
    console.log(idx);
}
