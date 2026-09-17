// ScriptGo Corpus: WebSocket RFC 6455 Client
// Verifies WebSocket construction, readyState, close, and constants.

import { WebSocket, CloseEvent, MessageEvent } from "ws";

// 1. Verify WebSocket constants
// @native.expected: ws_constants: 0 1 2 3
console.log("ws_constants: " + WebSocket.CONNECTING + " " + WebSocket.OPEN + " " + WebSocket.CLOSING + " " + WebSocket.CLOSED);

// 2. Instantiate client
// @native.expected: ws_initial_state: 0
const ws = new WebSocket("ws://127.0.0.1:9999/chat");
console.log("ws_initial_state: " + ws.readyState);

// 3. Verify CloseEvent instantiation
// @native.expected: close_event: close 1000 normal_closure true
const closeEv = new CloseEvent("close", { code: 1000, reason: "normal_closure", wasClean: true });
console.log("close_event: " + closeEv.type + " " + closeEv.code + " " + closeEv.reason + " " + closeEv.wasClean);

// 4. Verify MessageEvent instantiation
// @native.expected: message_event: message hello_ws ws://127.0.0.1:9999/chat
const msgEv = new MessageEvent<string>("message", { data: "hello_ws", origin: "ws://127.0.0.1:9999/chat" });
console.log("message_event: " + msgEv.type + " " + msgEv.data + " " + msgEv.origin);

// 5. Close client and check state
// @native.expected: ws_closed: true
ws.close(1000, "client_shutdown");
console.log("ws_closed: " + (ws.readyState === WebSocket.CLOSING || ws.readyState === WebSocket.CLOSED));
