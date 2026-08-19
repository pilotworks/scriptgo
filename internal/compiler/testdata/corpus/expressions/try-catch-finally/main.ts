function runTest(shouldThrow: number) {
  try {
    console.log("start try");
    if (shouldThrow === 1) {
      throw "boom";
    }
    console.log("end try");
  } catch (err) {
    console.log("in catch:");
    console.log(err);
  } finally {
    console.log("in finally");
  }
}

function main() {
  runTest(0);
  runTest(1);
}

main();
