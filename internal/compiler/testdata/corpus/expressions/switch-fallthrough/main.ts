function classify(digit: number): string {
  switch (digit) {
    case 2:
    case 3:
    case 5:
    case 7:
      return "prime";
    case 4:
    case 6:
    case 8:
    case 9:
      return "composite";
    default:
      return "other";
  }
}

console.log(classify(2));
console.log(classify(3));
console.log(classify(4));
console.log(classify(7));
console.log(classify(9));
console.log(classify(1));
