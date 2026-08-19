function mayThrow(flag: number) {
  if (flag === 1) {
    throw "custom error message";
  }
  console.log("no throw");
}

function main() {
  try {
    mayThrow(0);
    mayThrow(1);
    console.log("unreachable");
  } catch (e) {
    console.log("caught:");
    console.log(e);
  }
  console.log("recovered");
}

main();
