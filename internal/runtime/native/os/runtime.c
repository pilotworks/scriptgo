#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/utsname.h>
#include <pwd.h>
#include <sys/time.h>

#if defined(__APPLE__)
#include <sys/sysctl.h>
#include <mach/mach_host.h>
#include <mach/mach_init.h>
#include <mach/host_info.h>
#elif defined(__linux__)
#include <sys/sysinfo.h>
#endif

int scriptgo_runtime_set_error(const char *message);

static int os_fail(const char *message) { return scriptgo_runtime_set_error(message); }

int scriptgo_os_platform(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
#if defined(__APPLE__)
    *out_str = strdup("darwin");
#elif defined(__linux__)
    *out_str = strdup("linux");
#elif defined(_WIN32)
    *out_str = strdup("win32");
#else
    *out_str = strdup("unknown");
#endif
    return 0;
}

int scriptgo_os_arch(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
#if defined(__arm64__) || defined(__aarch64__)
    *out_str = strdup("arm64");
#elif defined(__x86_64__) || defined(__amd64__)
    *out_str = strdup("x64");
#elif defined(__i386__)
    *out_str = strdup("ia32");
#elif defined(__arm__)
    *out_str = strdup("arm");
#else
    *out_str = strdup("unknown");
#endif
    return 0;
}

int scriptgo_os_homedir(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    const char *home = getenv("HOME");
    if (home != NULL && strlen(home) > 0) {
        *out_str = strdup(home);
        return 0;
    }
    struct passwd *pw = getpwuid(getuid());
    if (pw != NULL && pw->pw_dir != NULL) {
        *out_str = strdup(pw->pw_dir);
        return 0;
    }
    *out_str = strdup("");
    return 0;
}

int scriptgo_os_type(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    struct utsname uts;
    if (uname(&uts) == 0) {
        *out_str = strdup(uts.sysname);
    } else {
        *out_str = strdup("unknown");
    }
    return 0;
}

int scriptgo_os_release(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    struct utsname uts;
    if (uname(&uts) == 0) {
        *out_str = strdup(uts.release);
    } else {
        *out_str = strdup("unknown");
    }
    return 0;
}

int scriptgo_os_uptime(double *out_val) {
    if (out_val == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    *out_val = 0.0;
#if defined(__APPLE__)
    struct timeval boottime;
    size_t len = sizeof(boottime);
    int mib[2] = {CTL_KERN, KERN_BOOTTIME};
    if (sysctl(mib, 2, &boottime, &len, NULL, 0) == 0) {
        struct timeval now;
        gettimeofday(&now, NULL);
        *out_val = (double)(now.tv_sec - boottime.tv_sec);
    }
#elif defined(__linux__)
    struct sysinfo info;
    if (sysinfo(&info) == 0) {
        *out_val = (double)info.uptime;
    }
#endif
    return 0;
}

int scriptgo_os_totalmem(double *out_val) {
    if (out_val == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    *out_val = 0.0;
#if defined(__APPLE__)
    int64_t mem = 0;
    size_t len = sizeof(mem);
    int mib[2] = {CTL_HW, HW_MEMSIZE};
    if (sysctl(mib, 2, &mem, &len, NULL, 0) == 0) {
        *out_val = (double)mem;
    }
#elif defined(__linux__)
    struct sysinfo info;
    if (sysinfo(&info) == 0) {
        *out_val = (double)info.totalram * (double)info.mem_unit;
    }
#endif
    return 0;
}

int scriptgo_os_freemem(double *out_val) {
    if (out_val == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    *out_val = 0.0;
#if defined(__APPLE__)
    mach_port_t host = mach_host_self();
    vm_size_t page_size = 4096;
    host_page_size(host, &page_size);
    vm_statistics64_data_t vm_stat;
    mach_msg_type_number_t count = HOST_VM_INFO64_COUNT;
    if (host_statistics64(host, HOST_VM_INFO64, (host_info64_t)&vm_stat, &count) == KERN_SUCCESS) {
        *out_val = (double)((int64_t)vm_stat.free_count * (int64_t)page_size);
    }
#elif defined(__linux__)
    struct sysinfo info;
    if (sysinfo(&info) == 0) {
        *out_val = (double)info.freeram * (double)info.mem_unit;
    }
#endif
    return 0;
}

int scriptgo_os_tmpdir(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    const char *tmp = getenv("TMPDIR");
    if (tmp == NULL || strlen(tmp) == 0) {
        tmp = getenv("TMP");
    }
    if (tmp == NULL || strlen(tmp) == 0) {
        tmp = getenv("TEMP");
    }
    if (tmp == NULL || strlen(tmp) == 0) {
        tmp = "/tmp";
    }
    *out_str = strdup(tmp);
    return 0;
}

