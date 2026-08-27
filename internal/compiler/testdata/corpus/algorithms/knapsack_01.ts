// @expect: 220
// @expect: [2,1]
function knapsack(weights: number[], values: number[], capacity: number): [number, number[]] {
    const n = weights.length;
    const dp: number[][] = [];

    for (let i = 0; i <= n; i++) {
        const row: number[] = [];
        for (let w = 0; w <= capacity; w++) {
            row.push(0);
        }
        dp.push(row);
    }

    for (let i = 1; i <= n; i++) {
        const wt = weights[i - 1];
        const val = values[i - 1];
        for (let w = 0; w <= capacity; w++) {
            if (wt <= w) {
                const pick = val + dp[i - 1][w - wt];
                const leave = dp[i - 1][w];
                dp[i][w] = Math.max(pick, leave);
            } else {
                dp[i][w] = dp[i - 1][w];
            }
        }
    }

    const maxVal = dp[n][capacity];
    const selected: number[] = [];
    let curW = capacity;
    for (let i = n; i > 0; i--) {
        if (dp[i][curW] !== dp[i - 1][curW]) {
            selected.push(i - 1);
            curW = curW - weights[i - 1];
        }
    }

    return [maxVal, selected];
}

const weights = [10, 20, 30];
const values = [60, 100, 120];
const capacity = 50;

const [maxProfit, items] = knapsack(weights, values, capacity);
console.log(maxProfit);
console.log(JSON.stringify(items));
