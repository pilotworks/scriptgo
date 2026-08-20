import { createHash, randomUUID } from "node:crypto";

function sha256(content: string): string {
  const hash = createHash("sha256");
  return hash.update(content).digest("hex");
}

console.log("=== Node Crypto & Hash Utils ===");

const sessionToken: string = randomUUID();
console.log(`Generated UUID: ${sessionToken}`);

const rawPayload = `user:alice;ts:${Date.now()}`;
const hashHex = sha256(rawPayload);
console.log(`Payload: ${rawPayload}`);
console.log(`SHA256: ${hashHex}`);

const base64Encoded = btoa(rawPayload);
console.log(`Base64 Encoded: ${base64Encoded}`);
console.log(`Base64 Decoded: ${atob(base64Encoded)}`);
