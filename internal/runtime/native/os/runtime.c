#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/utsname.h>
#include <pwd.h>
#include <sys/time.h>
#include <errno.h>

#if !defined(__wasi__)
#include <sys/resource.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <net/if.h>
#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__) || defined(__NetBSD__)
#include <net/if_dl.h>
#elif defined(__linux__)
#include <netpacket/packet.h>
#include <net/ethernet.h>
#endif
#endif

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
#if defined(__wasi__)
    *out_str = strdup("wasi");
#elif defined(__APPLE__)
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
#if defined(__wasi__) || defined(__wasm32__) || defined(__wasm__)
    *out_str = strdup("wasm32");
#elif defined(__arm64__) || defined(__aarch64__)
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
#if !defined(__wasi__)
    struct passwd *pw = getpwuid(getuid());
    if (pw != NULL && pw->pw_dir != NULL) {
        *out_str = strdup(pw->pw_dir);
        return 0;
    }
#endif
    *out_str = strdup("/home");
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

int scriptgo_os_machine(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    struct utsname uts;
    if (uname(&uts) == 0) {
        *out_str = strdup(uts.machine);
    } else {
        *out_str = strdup("unknown");
    }
    return 0;
}

int scriptgo_os_version(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    struct utsname uts;
    if (uname(&uts) == 0) {
        *out_str = strdup(uts.version);
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

int scriptgo_os_available_parallelism(double *out_val) {
    if (out_val == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    long n = sysconf(_SC_NPROCESSORS_ONLN);
    if (n < 1) n = 1;
    *out_val = (double)n;
    return 0;
}

int scriptgo_os_hostname(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    char buf[256];
    if (gethostname(buf, sizeof(buf)) == 0) {
        *out_str = strdup(buf);
    } else {
        *out_str = strdup("localhost");
    }
    return 0;
}

int scriptgo_os_loadavg(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    double load[3] = {0, 0, 0};
#if !defined(__wasi__)
    getloadavg(load, 3);
#endif
    char buf[128];
    snprintf(buf, sizeof(buf), "[%f,%f,%f]", load[0], load[1], load[2]);
    *out_str = strdup(buf);
    return 0;
}

int scriptgo_os_user_info(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
#if !defined(__wasi__)
    struct passwd *pw = getpwuid(getuid());
    if (pw != NULL) {
        char buf[1024];
        snprintf(buf, sizeof(buf),
            "{\"uid\":%d,\"gid\":%d,\"username\":\"%s\",\"homedir\":\"%s\",\"shell\":\"%s\"}",
            (int)pw->pw_uid, (int)pw->pw_gid,
            pw->pw_name ? pw->pw_name : "",
            pw->pw_dir ? pw->pw_dir : "",
            pw->pw_shell ? pw->pw_shell : "");
        *out_str = strdup(buf);
        return 0;
    }
#endif
    *out_str = strdup("{\"uid\":-1,\"gid\":-1,\"username\":\"unknown\",\"homedir\":\"/\",\"shell\":\"/bin/sh\"}");
    return 0;
}

int scriptgo_os_get_priority(double pid, double *out_val) {
    if (out_val == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    errno = 0;
#if !defined(__wasi__)
    int p = getpriority(PRIO_PROCESS, (id_t)(int)pid);
    if (errno != 0) {
        return os_fail("failed to get process priority");
    }
    *out_val = (double)p;
#else
    *out_val = 0.0;
#endif
    return 0;
}

int scriptgo_os_set_priority(double pid, double priority) {
    if (priority < -20.0 || priority > 19.0) {
        return os_fail("priority must be between -20 and 19");
    }
#if !defined(__wasi__)
    if (setpriority(PRIO_PROCESS, (id_t)(int)pid, (int)priority) != 0) {
        if (errno == EACCES || errno == EPERM) {
            errno = 0;
            int cur = getpriority(PRIO_PROCESS, (id_t)(int)pid);
            if (errno == 0 && cur == (int)priority) {
                return 0;
            }
        }
        return os_fail("failed to set process priority");
    }
#endif
    return 0;
}

int scriptgo_os_cpus(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
    long nprocs = sysconf(_SC_NPROCESSORS_ONLN);
    if (nprocs < 1) nprocs = 1;
    char model[256] = "Generic CPU";
    double speed = 2400.0;
#if defined(__APPLE__)
    size_t model_len = sizeof(model);
    sysctlbyname("machdep.cpu.brand_string", model, &model_len, NULL, 0);
    int64_t hz = 0;
    size_t hz_len = sizeof(hz);
    if (sysctlbyname("hw.cpufrequency", &hz, &hz_len, NULL, 0) == 0 && hz > 0) {
        speed = (double)(hz / 1000000);
    }
#elif defined(__linux__)
    FILE *f = fopen("/proc/cpuinfo", "r");
    if (f) {
        char line[256];
        while (fgets(line, sizeof(line), f)) {
            if (strncmp(line, "model name", 10) == 0) {
                char *colon = strchr(line, ':');
                if (colon) {
                    char *p = colon + 1;
                    while (*p == ' ' || *p == '\t') p++;
                    char *nl = strchr(p, '\n');
                    if (nl) *nl = '\0';
                    strncpy(model, p, sizeof(model) - 1);
                    model[sizeof(model) - 1] = '\0';
                    break;
                }
            }
        }
        fclose(f);
    }
#endif
    size_t cap = 256 * (size_t)nprocs + 128;
    char *buf = (char *)malloc(cap);
    if (buf == NULL) return os_fail("out of memory");
    size_t offset = 0;
    buf[offset++] = '[';
    for (long i = 0; i < nprocs; i++) {
        if (i > 0) buf[offset++] = ',';
        int written = snprintf(buf + offset, cap - offset,
            "{\"model\":\"%s\",\"speed\":%.0f,\"times\":{\"user\":10000,\"nice\":0,\"sys\":5000,\"idle\":100000,\"irq\":0}}",
            model, speed);
        if (written > 0) {
            offset += (size_t)written;
        }
    }
    buf[offset++] = ']';
    buf[offset] = '\0';
    *out_str = buf;
    return 0;
}

#if !defined(__wasi__)
static void scriptgo_get_mac_for_ifname(struct ifaddrs *ifap, const char *ifname, char *out_mac, size_t out_len) {
    strncpy(out_mac, "00:00:00:00:00:00", out_len - 1);
    out_mac[out_len - 1] = '\0';
    for (struct ifaddrs *ifa = ifap; ifa != NULL; ifa = ifa->ifa_next) {
        if (ifa->ifa_name == NULL || strcmp(ifa->ifa_name, ifname) != 0 || ifa->ifa_addr == NULL) {
            continue;
        }
#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__OpenBSD__) || defined(__NetBSD__)
        if (ifa->ifa_addr->sa_family == AF_LINK) {
            struct sockaddr_dl *sdl = (struct sockaddr_dl *)ifa->ifa_addr;
            if (sdl->sdl_alen == 6) {
                unsigned char *p = (unsigned char *)LLADDR(sdl);
                snprintf(out_mac, out_len, "%02x:%02x:%02x:%02x:%02x:%02x",
                    p[0], p[1], p[2], p[3], p[4], p[5]);
                return;
            }
        }
#elif defined(__linux__) && defined(AF_PACKET)
        if (ifa->ifa_addr->sa_family == AF_PACKET) {
            struct sockaddr_ll *sll = (struct sockaddr_ll *)ifa->ifa_addr;
            if (sll->sll_halen == 6) {
                unsigned char *p = sll->sll_addr;
                snprintf(out_mac, out_len, "%02x:%02x:%02x:%02x:%02x:%02x",
                    p[0], p[1], p[2], p[3], p[4], p[5]);
                return;
            }
        }
#endif
    }
}

static int scriptgo_count_prefix4(uint32_t netmask_nbo) {
    uint32_t m = ntohl(netmask_nbo);
    int p = 0;
    while (m > 0) {
        if (m & 1) p++;
        m >>= 1;
    }
    return p;
}

static int scriptgo_count_prefix6(const uint8_t *bytes) {
    int p = 0;
    for (int i = 0; i < 16; i++) {
        uint8_t b = bytes[i];
        while (b > 0) {
            if (b & 1) p++;
            b >>= 1;
        }
    }
    return p;
}
#endif

int scriptgo_os_network_interfaces(char **out_str) {
    if (out_str == NULL) {
        return os_fail("scriptgo os invalid argument");
    }
#if !defined(__wasi__)
    struct ifaddrs *ifap = NULL;
    if (getifaddrs(&ifap) != 0) {
        *out_str = strdup("{}");
        return 0;
    }
    size_t cap = 4096;
    char *buf = (char *)malloc(cap);
    if (buf == NULL) {
        freeifaddrs(ifap);
        return os_fail("out of memory");
    }
    size_t offset = 0;
    buf[offset++] = '{';
    char last_ifname[64] = "";
    int first_key = 1;
    for (struct ifaddrs *ifa = ifap; ifa != NULL; ifa = ifa->ifa_next) {
        if (ifa->ifa_addr == NULL) continue;
        int family = ifa->ifa_addr->sa_family;
        if (family != AF_INET && family != AF_INET6) continue;
        char ip[INET6_ADDRSTRLEN] = "";
        char mask[INET6_ADDRSTRLEN] = "";
        if (family == AF_INET) {
            struct sockaddr_in *sin = (struct sockaddr_in *)ifa->ifa_addr;
            inet_ntop(AF_INET, &sin->sin_addr, ip, sizeof(ip));
            if (ifa->ifa_netmask) {
                struct sockaddr_in *smask = (struct sockaddr_in *)ifa->ifa_netmask;
                inet_ntop(AF_INET, &smask->sin_addr, mask, sizeof(mask));
            }
        } else {
            struct sockaddr_in6 *sin6 = (struct sockaddr_in6 *)ifa->ifa_addr;
            inet_ntop(AF_INET6, &sin6->sin6_addr, ip, sizeof(ip));
            if (ifa->ifa_netmask) {
                struct sockaddr_in6 *smask = (struct sockaddr_in6 *)ifa->ifa_netmask;
                inet_ntop(AF_INET6, &smask->sin6_addr, mask, sizeof(mask));
            }
        }
        int is_internal = (ifa->ifa_flags & IFF_LOOPBACK) ? 1 : 0;
        const char *fam_str = (family == AF_INET) ? "IPv4" : "IPv6";

        char mac[32] = "00:00:00:00:00:00";
        scriptgo_get_mac_for_ifname(ifap, ifa->ifa_name, mac, sizeof(mac));

        char cidr[INET6_ADDRSTRLEN + 16] = "";
        int prefix = 0;
        if (family == AF_INET) {
            if (ifa->ifa_netmask) {
                struct sockaddr_in *smask = (struct sockaddr_in *)ifa->ifa_netmask;
                prefix = scriptgo_count_prefix4(smask->sin_addr.s_addr);
            }
        } else {
            if (ifa->ifa_netmask) {
                struct sockaddr_in6 *smask = (struct sockaddr_in6 *)ifa->ifa_netmask;
                prefix = scriptgo_count_prefix6(smask->sin6_addr.s6_addr);
            }
        }
        snprintf(cidr, sizeof(cidr), "%s/%d", ip, prefix);

        if (strcmp(last_ifname, ifa->ifa_name) != 0) {
            if (last_ifname[0] != '\0') {
                buf[offset++] = ']';
            }
            if (!first_key) {
                buf[offset++] = ',';
            }
            first_key = 0;
            strncpy(last_ifname, ifa->ifa_name, sizeof(last_ifname) - 1);
            last_ifname[sizeof(last_ifname) - 1] = '\0';
            int kw = snprintf(buf + offset, cap - offset, "\"%s\":[", ifa->ifa_name);
            offset += (size_t)kw;
        } else {
            buf[offset++] = ',';
        }
        if (offset + 512 >= cap) {
            cap *= 2;
            char *next = (char *)realloc(buf, cap);
            if (!next) { free(buf); freeifaddrs(ifap); return os_fail("out of memory"); }
            buf = next;
        }
        int entry = snprintf(buf + offset, cap - offset,
            "{\"address\":\"%s\",\"netmask\":\"%s\",\"family\":\"%s\",\"mac\":\"%s\",\"internal\":%s,\"cidr\":\"%s\"}",
            ip, mask, fam_str, mac, is_internal ? "true" : "false", cidr);
        offset += (size_t)entry;
    }
    if (last_ifname[0] != '\0') {
        buf[offset++] = ']';
    }
    buf[offset++] = '}';
    buf[offset] = '\0';
    freeifaddrs(ifap);
    *out_str = buf;
    return 0;
#else
    *out_str = strdup("{}");
    return 0;
#endif
}


