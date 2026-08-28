// @expect: i=0, j=0
// @expect: i=0, j=1
// @expect: i=1, j=0
// @expect: loop_broken_at_i_1_j_1
// @expect: 999

outerLoop: for (let i = 0; i < 3; i++) {
  for (let j = 0; j < 2; j++) {
    switch (i * 10 + j) {
      case 0:
        console.log("i=0, j=0");
        break;
      case 1:
        console.log("i=0, j=1");
        break;
      case 10:
        console.log("i=1, j=0");
        break;
      case 11:
        console.log("loop_broken_at_i_1_j_1");
        break outerLoop;
      default:
        console.log("unreachable");
        break;
    }
  }
}

console.log(999);
