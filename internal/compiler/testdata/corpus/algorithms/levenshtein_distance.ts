// @expect: 3
// @expect: 0
// @expect: 4
function levenshteinDistance(s1: string, s2: string): number {
    const m = s1.length;
    const n = s2.length;
    const dp: number[][] = [];

    for (let i = 0; i <= m; i++) {
        const row: number[] = [];
        for (let j = 0; j <= n; j++) {
            row.push(0);
        }
        dp.push(row);
    }

    for (let i = 0; i <= m; i++) dp[i][0] = i;
    for (let j = 0; j <= n; j++) dp[0][j] = j;

    for (let i = 1; i <= m; i++) {
        for (let j = 1; j <= n; j++) {
            if (s1[i - 1] === s2[j - 1]) {
                dp[i][j] = dp[i - 1][j - 1];
            } else {
                const insertOp = dp[i][j - 1];
                const deleteOp = dp[i - 1][j];
                const replaceOp = dp[i - 1][j - 1];
                dp[i][j] = 1 + Math.min(insertOp, Math.min(deleteOp, replaceOp));
            }
        }
    }

    return dp[m][n];
}

console.log(levenshteinDistance("kitten", "sitting"));
console.log(levenshteinDistance("same", "same"));
console.log(levenshteinDistance("hello", "world"));
