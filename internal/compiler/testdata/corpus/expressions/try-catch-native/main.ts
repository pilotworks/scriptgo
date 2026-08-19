try {
  throw "boom";
} catch (error) {
  console.log(error);
} finally {
  console.log("finally");
}
