// @expect: 4
// @expect: [2,3,7,18]
function lengthOfLIS(nums: number[]): [number, number[]] {
    if (nums.length === 0) return [0, []];

    const tails: number[] = [];
    const parent: number[] = [];
    const tailIndices: number[] = [];

    for (let i = 0; i < nums.length; i++) {
        parent.push(-1);
    }

    for (let i = 0; i < nums.length; i++) {
        const x = nums[i];
        let l = 0;
        let r = tails.length;

        while (l < r) {
            const m = Math.floor((l + r) / 2);
            if (tails[m] < x) {
                l = m + 1;
            } else {
                r = m;
            }
        }

        if (l > 0) {
            parent[i] = tailIndices[l - 1];
        }

        if (l === tails.length) {
            tails.push(x);
            tailIndices.push(i);
        } else {
            tails[l] = x;
            tailIndices[l] = i;
        }
    }

    const lis: number[] = [];
    let curr = tailIndices[tails.length - 1];
    while (curr !== -1) {
        lis.push(nums[curr]);
        curr = parent[curr];
    }
    lis.reverse();

    return [tails.length, lis];
}

const sequence = [10, 9, 2, 5, 3, 7, 101, 18];
const [len, subsequence] = lengthOfLIS(sequence);
console.log(len);
console.log(JSON.stringify(subsequence));
