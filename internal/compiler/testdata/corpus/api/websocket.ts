// ScriptGo Corpus: Web Standards WebSocket API
// Consolidated test suite verifying WebSocket properties, state constants, event dispatch, and lifecycle.

import { WebSocket } from "node:http";

// @api: WebSocket.CONNECTING
// @api: WebSocket.OPEN
// @api: WebSocket.CLOSING
// @api: WebSocket.CLOSED
// @expect: 0
// @expect: 1
// @expect: 2
// @expect: 3
console.log(WebSocket.CONNECTING);
console.log(WebSocket.OPEN);
console.log(WebSocket.CLOSING);
console.log(WebSocket.CLOSED);

// @api: new WebSocket
// @expect: ws://localhost:8080
// @expect: 1
const ws = new WebSocket("ws://localhost:8080", "chat");
console.log(ws.url);
console.log(ws.readyState);

// @api: WebSocket.send
ws.send("hello server");

// @api: WebSocket.addEventListener
// @api: WebSocket.dispatchEvent
// @expect: received_message: message
ws.addEventListener("message", (e: { type: string }) => {
    console.log("received_message: " + e.type);
});
ws.dispatchEvent({ type: "message" });

// @api: WebSocket.close
// @expect: close_event: 1000
// @expect: ws_closed: 3
ws.onclose = (e: { code?: number }) => {
    console.log("close_event: " + e.code);
};
ws.close(1000, "normal");
console.log("ws_closed: " + ws.readyState);
