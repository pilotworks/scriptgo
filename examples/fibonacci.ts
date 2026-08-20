// Fibonacci calculation demonstrating recursive and iterative performance

function fibIterative(n: number): number {
  if (n <= 1) return n;
  let prev = 0;
  let curr = 1;
  for (let i = 2; i <= n; i++) {
    const next = prev + curr;
    prev = curr;
    curr = next;
  }
  return curr;
}

function fibRecursive(n: number): number {
  if (n <= 1) return n;
  return fibRecursive(n - 1) + fibRecursive(n - 2);
}

console.log("=== Fibonacci Demo ===");
console.log("fibIterative(10): " + fibIterative(10));
console.log("fibIterative(30): " + fibIterative(30));
console.log("fibRecursive(15): " + fibRecursive(15));
