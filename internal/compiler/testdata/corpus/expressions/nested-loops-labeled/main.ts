let hitCount = 0;
let earlyBreak = false;

outer: for (let i = 0; i < 4; i = i + 1) {
  middle: for (let j = 0; j < 4; j = j + 1) {
    if (i === 1 && j === 2) {
      console.log("skip middle at i=" + i + ", j=" + j);
      continue outer;
    }
    if (i === 2 && j === 2) {
      console.log("break outer at i=" + i + ", j=" + j);
      earlyBreak = true;
      break outer;
    }
    inner: for (let k = 0; k < 2; k = k + 1) {
      hitCount = hitCount + 1;
    }
  }
}

console.log("Total hits: " + hitCount);
console.log("Early break: " + earlyBreak);

// Labeled while loop test
let counter = 0;
loopA: while (counter < 10) {
  counter = counter + 1;
  if (counter % 2 === 0) {
    continue loopA;
  }
  if (counter > 6) {
    break loopA;
  }
  console.log("odd: " + counter);
}
