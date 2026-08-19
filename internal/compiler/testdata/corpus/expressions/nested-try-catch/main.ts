function compute(x: number) {
  try {
    console.log("outer try");
    try {
      console.log("inner try");
      if (x === 1) {
        throw "inner error";
      }
      console.log("inner success");
    } catch (e1) {
      console.log("caught inner:");
      console.log(e1);
      throw "re-thrown from inner";
    } finally {
      console.log("inner finally");
    }
  } catch (e2) {
    console.log("caught outer:");
    console.log(e2);
  } finally {
    console.log("outer finally");
  }
}

function main() {
  compute(0);
  compute(1);
}

main();
