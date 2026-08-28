// @expect: true
// @expect: true
// @expect: false
// @expect: true
// @expect: 1
class TokenBucket {
    private capacity: number;
    private refillRatePerSec: number;
    private tokens: number;
    private lastRefillTimestamp: number;

    constructor(capacity: number, refillRatePerSec: number) {
        this.capacity = capacity;
        this.refillRatePerSec = refillRatePerSec;
        this.tokens = capacity;
        this.lastRefillTimestamp = 0;
    }

    refill(currentTimestamp: number): void {
        if (this.lastRefillTimestamp === 0) {
            this.lastRefillTimestamp = currentTimestamp;
            return;
        }
        const deltaSec = (currentTimestamp - this.lastRefillTimestamp) / 1000;
        const added = deltaSec * this.refillRatePerSec;
        this.tokens = Math.min(this.capacity, this.tokens + added);
        this.lastRefillTimestamp = currentTimestamp;
    }

    tryConsume(tokens: number, currentTimestamp: number): boolean {
        this.refill(currentTimestamp);
        if (this.tokens >= tokens) {
            this.tokens -= tokens;
            return true;
        }
        return false;
    }

    getTokens(): number {
        return Math.floor(this.tokens);
    }
}

const limiter = new TokenBucket(10, 2); // 10 tokens max, 2 tokens/sec refill

console.log(limiter.tryConsume(5, 1000));  // true, left 5
console.log(limiter.tryConsume(4, 1000));  // true, left 1
console.log(limiter.tryConsume(3, 1000));  // false, not enough tokens

console.log(limiter.tryConsume(3, 2500));  // 1.5s elapsed -> 3 tokens added -> 4 total -> true
console.log(limiter.getTokens());
