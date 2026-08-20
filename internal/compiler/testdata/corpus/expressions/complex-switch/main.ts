function categorizeHttpStatus(code: number): string {
  let category = "";
  switch (code) {
    case 200:
    case 201:
    case 204:
      category = "success";
      break;
    case 301:
    case 302:
    case 304:
      category = "redirection";
      break;
    case 400:
    case 401:
    case 403:
    case 404:
      category = "client_error";
      break;
    case 500:
    case 502:
    case 503:
      category = "server_error";
      break;
    default:
      category = "unknown";
      break;
  }
  return category;
}

function processCommand(cmd: string): number {
  let priority = 0;
  switch (cmd) {
    case "start":
      priority = 10;
      break;
    case "stop":
      priority = 20;
      break;
    case "restart":
      priority = 15;
      break;
    default:
      priority = 0;
      break;
  }
  return priority;
}

console.log(categorizeHttpStatus(200));
console.log(categorizeHttpStatus(204));
console.log(categorizeHttpStatus(301));
console.log(categorizeHttpStatus(404));
console.log(categorizeHttpStatus(500));
console.log(categorizeHttpStatus(418));

console.log(processCommand("start"));
console.log(processCommand("restart"));
console.log(processCommand("stop"));
console.log(processCommand("status"));
