#ifndef SCRIPTGO_RUNTIME_WEBSOCKET_H
#define SCRIPTGO_RUNTIME_WEBSOCKET_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// WebSocket readyState enum values
#define SCRIPTGO_WS_CONNECTING 0
#define SCRIPTGO_WS_OPEN       1
#define SCRIPTGO_WS_CLOSING    2
#define SCRIPTGO_WS_CLOSED     3

// Native C ABI for WebSocket client
int scriptgo_websocket_connect(const char *url, const char *protocol, double *out_handle);
int scriptgo_websocket_send_text(double handle, const char *data, double *out_sent);
int scriptgo_websocket_send_binary(double handle, const void *data, double length, double *out_sent);
int scriptgo_websocket_close(double handle, double code, const char *reason);
int scriptgo_websocket_poll(double handle, double *out_event_type, char **out_data, double *out_code, char **out_reason);
int scriptgo_websocket_ready_state(double handle, double *out_state);

#ifdef __cplusplus
}
#endif

#endif // SCRIPTGO_RUNTIME_WEBSOCKET_H
