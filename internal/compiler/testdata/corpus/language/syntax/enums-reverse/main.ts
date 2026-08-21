enum Color {
  Red,
  Green,
  Blue,
}

enum HttpStatus {
  OK = 200,
  NotFound = 404,
}

enum LogLevel {
  Info = "INFO",
  Error = "ERROR",
}

console.log(Color.Red);
console.log(Color.Green);
console.log(Color.Blue);

console.log(Color[0]);
console.log(Color[1]);
console.log(Color[2]);

console.log(HttpStatus.OK);
console.log(HttpStatus[200]);
console.log(HttpStatus[404]);

console.log(LogLevel.Info);
console.log(LogLevel.Error);
