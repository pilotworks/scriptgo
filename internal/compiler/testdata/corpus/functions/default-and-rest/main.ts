function buildLog(level: string = "INFO", ...messages: string[]): string {
  let joined: string = "";
  for (let i = 0; i < messages.length; i = i + 1) {
    if (i > 0) {
      joined = joined + " ";
    }
    joined = joined + messages[i];
  }
  return "[" + level + "] " + joined;
}

function sumWithBase(base: number = 100, ...nums: number[]): number {
  let total: number = base;
  for (let i = 0; i < nums.length; i = i + 1) {
    total = total + nums[i];
  }
  return total;
}

console.log(buildLog("WARN", "Server", "starting", "on", "port", "8080"));
console.log(buildLog("ERROR", "Database", "connection", "lost"));
console.log(sumWithBase(0, 1, 2, 3, 4, 5));
console.log(sumWithBase(50, 10, 20, 30));
