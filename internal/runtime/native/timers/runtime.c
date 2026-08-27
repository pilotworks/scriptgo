#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

typedef struct {
    void *fn_ptr;
    void *env;
} scriptgo_timer_closure;

typedef struct scriptgo_timer {
    int64_t id;
    scriptgo_timer_closure *closure;
    double delay_ms;
    int is_interval;
    int canceled;
    struct timespec fire_time;
    struct scriptgo_timer *next;
} scriptgo_timer;

static scriptgo_timer *timers_head = NULL;
static int64_t next_timer_id = 1;

static int timer_fail(const char *msg) {
    fprintf(stderr, "%s\n", msg);
    return -1;
}

static void get_current_time(struct timespec *ts) {
    clock_gettime(CLOCK_MONOTONIC, ts);
}

static void add_ms_to_timespec(struct timespec *ts, double ms) {
    long sec = (long)(ms / 1000.0);
    long nsec = (long)((ms - (double)(sec * 1000)) * 1000000.0);
    ts->tv_sec += sec;
    ts->tv_nsec += nsec;
    if (ts->tv_nsec >= 1000000000L) {
        ts->tv_sec += 1;
        ts->tv_nsec -= 1000000000L;
    }
}

static int compare_timespec(const struct timespec *a, const struct timespec *b) {
    if (a->tv_sec < b->tv_sec) return -1;
    if (a->tv_sec > b->tv_sec) return 1;
    if (a->tv_nsec < b->tv_nsec) return -1;
    if (a->tv_nsec > b->tv_nsec) return 1;
    return 0;
}

int scriptgo_timer_set_timeout(void *closure_handle, double delay_ms, double *out_id) {
    if (closure_handle == NULL) return timer_fail("scriptgo setTimeout missing closure");
    scriptgo_timer *t = malloc(sizeof(scriptgo_timer));
    if (t == NULL) return timer_fail("scriptgo setTimeout allocation failed");
    t->id = next_timer_id++;
    t->closure = (scriptgo_timer_closure *)closure_handle;
    t->delay_ms = delay_ms;
    t->is_interval = 0;
    t->canceled = 0;
    get_current_time(&t->fire_time);
    add_ms_to_timespec(&t->fire_time, delay_ms);
    t->next = timers_head;
    timers_head = t;
    if (out_id != NULL) {
        *out_id = (double)t->id;
    }
    return 0;
}

static scriptgo_timer *currently_running_timer = NULL;

int scriptgo_timer_clear_timeout(double id) {
    int64_t target_id = (int64_t)id;
    if (target_id <= 0 && currently_running_timer != NULL) {
        currently_running_timer->canceled = 1;
        return 0;
    }
    scriptgo_timer *curr = timers_head;
    while (curr != NULL) {
        if (curr->id == target_id) {
            curr->canceled = 1;
            break;
        }
        curr = curr->next;
    }
    return 0;
}

int scriptgo_timer_set_interval(void *closure_handle, double delay_ms, double *out_id) {
    if (closure_handle == NULL) return timer_fail("scriptgo setInterval missing closure");
    scriptgo_timer *t = malloc(sizeof(scriptgo_timer));
    if (t == NULL) return timer_fail("scriptgo setInterval allocation failed");
    t->id = next_timer_id++;
    t->closure = (scriptgo_timer_closure *)closure_handle;
    t->delay_ms = delay_ms;
    t->is_interval = 1;
    t->canceled = 0;
    get_current_time(&t->fire_time);
    add_ms_to_timespec(&t->fire_time, delay_ms);
    t->next = timers_head;
    timers_head = t;
    if (out_id != NULL) {
        *out_id = (double)t->id;
    }
    return 0;
}

int scriptgo_timer_clear_interval(double id) {
    return scriptgo_timer_clear_timeout(id);
}

int scriptgo_timer_set_immediate(void *closure_handle, double *out_id) {
    return scriptgo_timer_set_timeout(closure_handle, 0.0, out_id);
}

int scriptgo_timer_clear_immediate(double id) {
    return scriptgo_timer_clear_timeout(id);
}

int scriptgo_timers_drain(void) {
    extern int scriptgo_event_loop_run(void);
    while (1) {
        scriptgo_event_loop_run();
        
        scriptgo_timer *earliest = NULL;
        scriptgo_timer *curr = timers_head;
        while (curr != NULL) {
            if (!curr->canceled) {
                if (earliest == NULL || compare_timespec(&curr->fire_time, &earliest->fire_time) < 0) {
                    earliest = curr;
                }
            }
            curr = curr->next;
        }

        if (earliest == NULL) {
            break;
        }

        struct timespec now;
        get_current_time(&now);
        if (compare_timespec(&earliest->fire_time, &now) > 0) {
            long sleep_sec = earliest->fire_time.tv_sec - now.tv_sec;
            long sleep_nsec = earliest->fire_time.tv_nsec - now.tv_nsec;
            if (sleep_nsec < 0) {
                sleep_sec -= 1;
                sleep_nsec += 1000000000L;
            }
            struct timespec req = { sleep_sec, sleep_nsec };
            nanosleep(&req, NULL);
        }

        if (earliest->canceled) {
            continue;
        }

        if (earliest->closure != NULL && earliest->closure->fn_ptr != NULL) {
            currently_running_timer = earliest;
            void (*fn)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t) =
                (void (*)(void *, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t, uint32_t, uint32_t, uint64_t))earliest->closure->fn_ptr;
            fn(earliest->closure->env, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0);
            currently_running_timer = NULL;
        }

        if (earliest->is_interval && !earliest->canceled) {
            get_current_time(&earliest->fire_time);
            add_ms_to_timespec(&earliest->fire_time, earliest->delay_ms);
        } else {
            earliest->canceled = 1;
        }
    }
    return 0;
}
