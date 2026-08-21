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

int scriptgo_date_to_date_string(double ms, char **out_str) {
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
    strftime(buf, 64, "%a %b %d %Y", tm_info);
    *out_str = buf;
    return 0;
}

int scriptgo_date_to_time_string(double ms, char **out_str) {
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
    strftime(buf, 64, "%H:%M:%S", tm_info);
    *out_str = buf;
    return 0;
}

int scriptgo_date_to_utc_string(double ms, char **out_str) {
    if (out_str == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (tm_info == NULL) {
        return date_fail("scriptgo date gmtime failed");
    }
    char *buf = malloc(64);
    if (buf == NULL) {
        return date_fail("scriptgo date allocation failed");
    }
    strftime(buf, 64, "%a, %d %b %Y %H:%M:%S GMT", tm_info);
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
    if (strptime(str, "%Y-%m-%d", &tm_info) != NULL) {
        time_t t = timegm(&tm_info);
        *out_ms = (double)t * 1000.0;
        return 0;
    }
    *out_ms = 0.0;
    return 0;
}

int scriptgo_date_utc(double year, double month, double date, double hours, double min, double sec, double ms, double *out_ms) {
    if (out_ms == NULL) {
        return date_fail("scriptgo date invalid arguments");
    }
    struct tm tm_info;
    memset(&tm_info, 0, sizeof(struct tm));
    int y = (int)year;
    if (y >= 0 && y <= 99) {
        y += 1900;
    }
    tm_info.tm_year = y - 1900;
    tm_info.tm_mon = (int)month;
    tm_info.tm_mday = (int)date;
    tm_info.tm_hour = (int)hours;
    tm_info.tm_min = (int)min;
    tm_info.tm_sec = (int)sec;
    time_t t = timegm(&tm_info);
    *out_ms = (double)t * 1000.0 + ms;
    return 0;
}

int scriptgo_date_get_date(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_mday;
    return 0;
}

int scriptgo_date_get_day(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_wday;
    return 0;
}

int scriptgo_date_get_full_year(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)(tm_info->tm_year + 1900);
    return 0;
}

int scriptgo_date_get_hours(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_hour;
    return 0;
}

int scriptgo_date_get_milliseconds(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    long long m = (long long)ms % 1000;
    if (m < 0) m += 1000;
    *out = (double)m;
    return 0;
}

int scriptgo_date_get_minutes(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_min;
    return 0;
}

int scriptgo_date_get_month(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_mon;
    return 0;
}

int scriptgo_date_get_seconds(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    *out = (double)tm_info->tm_sec;
    return 0;
}

int scriptgo_date_get_timezone_offset(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *loc = localtime(&sec);
    struct tm tm_loc = *loc;
    struct tm *utc = gmtime(&sec);
    struct tm tm_utc = *utc;
    time_t t_loc = timegm(&tm_loc);
    time_t t_utc = timegm(&tm_utc);
    *out = (double)(t_utc - t_loc) / 60.0;
    return 0;
}

int scriptgo_date_get_utc_date(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_mday;
    return 0;
}

int scriptgo_date_get_utc_day(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_wday;
    return 0;
}

int scriptgo_date_get_utc_full_year(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)(tm_info->tm_year + 1900);
    return 0;
}

int scriptgo_date_get_utc_hours(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_hour;
    return 0;
}

int scriptgo_date_get_utc_milliseconds(double ms, double *out) {
    return scriptgo_date_get_milliseconds(ms, out);
}

int scriptgo_date_get_utc_minutes(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_min;
    return 0;
}

int scriptgo_date_get_utc_month(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_mon;
    return 0;
}

int scriptgo_date_get_utc_seconds(double ms, double *out) {
    if (out == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    *out = (double)tm_info->tm_sec;
    return 0;
}

int scriptgo_date_set_date(double ms, double day, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_mday = (int)day;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_full_year(double ms, double year, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_year = (int)year - 1900;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_hours(double ms, double hours, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_hour = (int)hours;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_milliseconds(double ms, double millis, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    long long sec = (long long)(ms / 1000.0);
    *out_ms = (double)sec * 1000.0 + millis;
    return 0;
}

int scriptgo_date_set_minutes(double ms, double min, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_min = (int)min;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_month(double ms, double month, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_mon = (int)month;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_seconds(double ms, double sec_val, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = localtime(&sec);
    if (!tm_info) return date_fail("localtime failed");
    tm_info->tm_sec = (int)sec_val;
    time_t t = mktime(tm_info);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_date(double ms, double day, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_mday = (int)day;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_full_year(double ms, double year, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_year = (int)year - 1900;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_hours(double ms, double hours, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_hour = (int)hours;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_milliseconds(double ms, double millis, double *out_ms) {
    return scriptgo_date_set_milliseconds(ms, millis, out_ms);
}

int scriptgo_date_set_utc_minutes(double ms, double min, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_min = (int)min;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_month(double ms, double month, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_mon = (int)month;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}

int scriptgo_date_set_utc_seconds(double ms, double sec_val, double *out_ms) {
    if (out_ms == NULL) return date_fail("scriptgo date invalid arguments");
    time_t sec = (time_t)(ms / 1000.0);
    struct tm *tm_info = gmtime(&sec);
    if (!tm_info) return date_fail("gmtime failed");
    struct tm copy = *tm_info;
    copy.tm_sec = (int)sec_val;
    time_t t = timegm(&copy);
    long long rem = (long long)ms % 1000;
    if (rem < 0) rem += 1000;
    *out_ms = (double)t * 1000.0 + (double)rem;
    return 0;
}
