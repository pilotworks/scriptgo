// Basic primitives, control flow, and template string interpolation

const appName: string = "ScriptGo Engine";
const version: number = 1.0;
const isProduction: boolean = true;
const maxConnections: bigint = 10000000000n;

function getStatusMessage(activeUsers: number): string {
  const loadStatus = activeUsers > 5000 ? "High" : "Normal";
  return `[${appName} v${version}] - Status: ${loadStatus} | Users: ${activeUsers} | Max: ${maxConnections} | Production: ${isProduction}`;
}

console.log(getStatusMessage(3500));
console.log(getStatusMessage(8200));
