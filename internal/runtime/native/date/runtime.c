#ifndef _GNU_SOURCE
#define _GNU_SOURCE 1
#endif
#ifndef _DEFAULT_SOURCE
#define _DEFAULT_SOURCE 1
#endif
#ifdef __APPLE__
#ifndef _DARWIN_C_SOURCE
#define _DARWIN_C_SOURCE 1
#endif
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <time.h>

int scriptgo_runtime_set_error(const char *message);

static int date_fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_date_now(double *out_ms) {
    if (out_ms == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    struct timeval tv;
    gettimeofday(&tv, NULL);
    *out_ms = (double)tv.tv_sec * 1000.0 + (double)tv.tv_usec / 1000.0;
    return 0;
}

int scriptgo_date_to_iso_string(double ms, char **out_str) {
    if (out_str == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    time_t sec = (time_t)(ms / 1000.0);
    int millis = (int)((long long)ms % 1000);
    if (millis < 0) millis += 1000;
    struct tm *tm_info = gmtime(&sec);
    if (tm_info == NULL) {
        return date_fail("scriptgo date gmtime failed");
    }
    char *buf = malloc(32);
    if (buf == NULL) {
        return date_fail("scriptgo date allocation failed");
    }
    snprintf(buf, 32, "%04d-%02d-%02dT%02d:%02d:%02d.%03dZ",
             tm_info->tm_year + 1900,
             tm_info->tm_mon + 1,
             tm_info->tm_mday,
             tm_info->tm_hour,
             tm_info->tm_min,
             tm_info->tm_sec,
             millis);
    *out_str = buf;
    return 0;
}

int scriptgo_date_to_string(double ms, char **out_str) {
    if (out_str == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (tm_info == NULL) {
        return date_fail("scriptgo date localtime failed");
    }
    char *buf = malloc(64);
    if (buf == NULL) {
        return date_fail("scriptgo date allocation failed");
    }
    strftime(buf, 64, "%a %b %d %Y %H:%M:%S", tm_info);
    *out_str = buf;
    return 0;
}

int scriptgo_date_parse(const char *str, double *out_ms) {
    if (str == NULL || out_ms == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    struct tm tm_info;
    memset(&tm_info, 0, sizeof(struct tm));
    if (strptime(str, "%Y-%m-%dT%H:%M:%S", &tm_info) != NULL) {
        time_t t = timegm(&tm_info);
        *out_ms = (double)t * 1000.0;
        return 0;
    }
    *out_ms = 0.0;
    return 0;
}
