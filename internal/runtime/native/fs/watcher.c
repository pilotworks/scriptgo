#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdint.h>
#include <fcntl.h>
#include <errno.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__) || defined(__NetBSD__)
#define SCRIPTGO_USE_KQUEUE 1
#include <sys/event.h>
#include <sys/time.h>
#include <libgen.h>
#elif defined(__linux__)
#define SCRIPTGO_USE_INOTIFY 1
#include <sys/inotify.h>
#include <limits.h>
#endif

int scriptgo_runtime_set_error(const char *message);

typedef struct scriptgo_fs_watcher {
    int64_t id;
    int main_fd;    // kq fd on macOS, inotify fd on Linux
    int watch_fd;   // watched file fd on macOS, wd descriptor on Linux
    char *path;
    int closed;
    struct scriptgo_fs_watcher *next;
} scriptgo_fs_watcher;

static scriptgo_fs_watcher *watchers_head = NULL;
static int64_t next_watcher_id = 1;

static int watcher_fail(const char *msg) {
    return scriptgo_runtime_set_error(msg);
}

int scriptgo_fs_watch_create(const char *path, double *out_watch_id) {
    if (path == NULL || out_watch_id == NULL) {
        return watcher_fail("fs_watch invalid arguments");
    }

    scriptgo_fs_watcher *w = (scriptgo_fs_watcher *)malloc(sizeof(scriptgo_fs_watcher));
    if (w == NULL) {
        return watcher_fail("fs_watch allocation failed");
    }
    memset(w, 0, sizeof(scriptgo_fs_watcher));
    w->id = next_watcher_id++;
    w->path = strdup(path);
    w->closed = 0;

#if defined(SCRIPTGO_USE_KQUEUE)
    w->main_fd = kqueue();
    if (w->main_fd < 0) {
        free(w->path);
        free(w);
        return watcher_fail("fs_watch kqueue failed");
    }
    w->watch_fd = open(path, O_RDONLY);
    if (w->watch_fd < 0) {
        close(w->main_fd);
        free(w->path);
        free(w);
        return watcher_fail("fs_watch open file failed");
    }
    struct kevent change;
    EV_SET(&change, w->watch_fd, EVFILT_VNODE, EV_ADD | EV_ENABLE | EV_CLEAR,
           NOTE_WRITE | NOTE_DELETE | NOTE_RENAME | NOTE_ATTRIB | NOTE_EXTEND, 0, 0);
    if (kevent(w->main_fd, &change, 1, NULL, 0, NULL) < 0) {
        close(w->watch_fd);
        close(w->main_fd);
        free(w->path);
        free(w);
        return watcher_fail("fs_watch kevent registration failed");
    }
#elif defined(SCRIPTGO_USE_INOTIFY)
    w->main_fd = inotify_init1(IN_NONBLOCK);
    if (w->main_fd < 0) {
        free(w->path);
        free(w);
        return watcher_fail("fs_watch inotify_init failed");
    }
    w->watch_fd = inotify_add_watch(w->main_fd, path,
                                    IN_MODIFY | IN_ATTRIB | IN_CLOSE_WRITE |
                                    IN_MOVED_FROM | IN_MOVED_TO | IN_CREATE | IN_DELETE);
    if (w->watch_fd < 0) {
        close(w->main_fd);
        free(w->path);
        free(w);
        return watcher_fail("fs_watch inotify_add_watch failed");
    }
#else
    w->main_fd = -1;
    w->watch_fd = -1;
#endif

    w->next = watchers_head;
    watchers_head = w;
    *out_watch_id = (double)w->id;
    return 0;
}

int scriptgo_fs_watch_poll(double watch_id, char **out_event_type, char **out_filename, double *out_has_event) {
    if (out_event_type == NULL || out_filename == NULL || out_has_event == NULL) {
        return watcher_fail("fs_watch_poll invalid arguments");
    }
    *out_has_event = 0.0;
    *out_event_type = strdup("");
    *out_filename = strdup("");

    int64_t target_id = (int64_t)watch_id;
    scriptgo_fs_watcher *w = watchers_head;
    while (w != NULL) {
        if (w->id == target_id && !w->closed) {
            break;
        }
        w = w->next;
    }
    if (w == NULL) {
        return 0;
    }

#if defined(SCRIPTGO_USE_KQUEUE)
    struct timespec timeout = { 0, 0 };
    struct kevent event;
    int nev = kevent(w->main_fd, NULL, 0, &event, 1, &timeout);
    if (nev > 0) {
        free(*out_event_type);
        free(*out_filename);
        *out_has_event = 1.0;
        if (event.fflags & (NOTE_RENAME | NOTE_DELETE)) {
            *out_event_type = strdup("rename");
        } else {
            *out_event_type = strdup("change");
        }
        *out_filename = strdup(w->path);
        return 0;
    }
#elif defined(SCRIPTGO_USE_INOTIFY)
    char buf[sizeof(struct inotify_event) + NAME_MAX + 1];
    ssize_t len = read(w->main_fd, buf, sizeof(buf));
    if (len > 0) {
        struct inotify_event *event = (struct inotify_event *)buf;
        free(*out_event_type);
        free(*out_filename);
        *out_has_event = 1.0;
        if (event->mask & (IN_MOVE | IN_DELETE | IN_CREATE | IN_MOVED_FROM | IN_MOVED_TO)) {
            *out_event_type = strdup("rename");
        } else {
            *out_event_type = strdup("change");
        }
        if (event->len > 0) {
            *out_filename = strdup(event->name);
        } else {
            *out_filename = strdup(w->path);
        }
        return 0;
    }
#endif

    return 0;
}

int scriptgo_fs_watch_close(double watch_id) {
    int64_t target_id = (int64_t)watch_id;
    scriptgo_fs_watcher *w = watchers_head;
    while (w != NULL) {
        if (w->id == target_id && !w->closed) {
            w->closed = 1;
#if defined(SCRIPTGO_USE_KQUEUE)
            if (w->watch_fd >= 0) close(w->watch_fd);
            if (w->main_fd >= 0) close(w->main_fd);
#elif defined(SCRIPTGO_USE_INOTIFY)
            if (w->watch_fd >= 0 && w->main_fd >= 0) inotify_rm_watch(w->main_fd, w->watch_fd);
            if (w->main_fd >= 0) close(w->main_fd);
#endif
            w->watch_fd = -1;
            w->main_fd = -1;
            break;
        }
        w = w->next;
    }
    return 0;
}
