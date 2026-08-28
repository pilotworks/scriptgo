// @expect: case1
// @expect: case2_after_default
// @expect: default_fallthrough
// @expect: case2_after_default
// @expect: default_break

function testMiddleDefault(val: number): void {
  switch (val) {
    case 1:
      console.log("case1");
      break;
    default:
      console.log("default_fallthrough");
    case 2:
      console.log("case2_after_default");
      break;
    case 3:
      console.log("case3");
      break;
  }
}

function testMiddleDefaultWithBreak(val: number): void {
  switch (val) {
    case 1:
      console.log("case1");
      break;
    default:
      console.log("default_break");
      break;
    case 2:
      console.log("case2");
      break;
  }
}

testMiddleDefault(1);
testMiddleDefault(2);
testMiddleDefault(99);
testMiddleDefaultWithBreak(99);
