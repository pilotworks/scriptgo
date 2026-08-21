function testFindIndex(): void {
  console.log("=== Array.findIndex ===");
  const nums = [10, 20, 30];
  console.log(nums.findIndex((x) => x > 15));
  console.log(nums.findIndex((x) => x > 50));
  const words = ["apple", "banana", "cherry"];
  console.log(words.findIndex((s) => s.startsWith("b")));
  console.log(words.findIndex((s) => s === "date"));
}

function testFill(): void {
  console.log("=== Array.fill ===");
  const a = [1, 2, 3, 4];
  a.fill(0);
  for (const x of a) {
    console.log(x);
  }
  const b = [1, 2, 3, 4];
  b.fill(9, 1, 3);
  for (const x of b) {
    console.log(x);
  }
}

function testToReversed(): void {
  console.log("=== Array.toReversed ===");
  const original = [1, 2, 3];
  const reversed = original.toReversed();
  for (const x of reversed) {
    console.log(x);
  }
  for (const x of original) {
    console.log(x);
  }
}

function testToSorted(): void {
  console.log("=== Array.toSorted ===");
  const nums = [30, 10, 50, 20, 40];
  const sortedNums = nums.toSorted();
  for (const x of sortedNums) {
    console.log(x);
  }
  for (const x of nums) {
    console.log(x);
  }
  const words = ["cherry", "apple", "banana"];
  const sortedWords = words.toSorted();
  for (const s of sortedWords) {
    console.log(s);
  }
}

function main(): void {
  testFindIndex();
  testFill();
  testToReversed();
  testToSorted();
}

main();
