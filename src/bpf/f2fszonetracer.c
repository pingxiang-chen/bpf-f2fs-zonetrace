// SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)
/* Copyright (c) 2021 Sartura
 * Based on minimal.c by Facebook */

#include <bpf/libbpf.h>
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>

#include "f2fszonetracer.skel.h"

#define DEBUG 0

int nr_zones;
int segment_size = 2;  // 2 MiB
int zone_size;
int zone_blocks;

struct event {
    int segno;
    unsigned int seg_type:6;
    unsigned char cur_valid_map[65];
};

static int libbpf_print_fn(enum libbpf_print_level level, const char *format, va_list args) {
    return vfprintf(stderr, format, args);
}

static volatile sig_atomic_t stop;

static void sig_int(int signo) {
    stop = 1;
}

void bump_memlock_rlimit(void) {
    struct rlimit rlim_new = {
        .rlim_cur = RLIM_INFINITY,
        .rlim_max = RLIM_INFINITY,
    };

    if (setrlimit(RLIMIT_MEMLOCK, &rlim_new)) {
        fprintf(stderr, "Failed to increase RLIMIT_MEMLOCK limit!\n");
        exit(1);
    }
}

int handle_event(void *ctx, void *data, size_t data_sz) {
    const struct event *e = data;

    unsigned int seg_per_zone = zone_size / segment_size;
    unsigned int cur_zone = e->segno / seg_per_zone;

    printf("update_sit_entry segno: %d cur_zone:%d seg_type:%u\n", e->segno % 1024, cur_zone, e->seg_type);
    fflush(stdout);
    write(1, e->cur_valid_map, 64);
    printf("\n");
    return 0;
}

int read_sysfs_device_queue(const char *device_path, const char *filename) {
    FILE *proc;
    char cmd[1024];
    int value;

    snprintf(cmd, sizeof(cmd), "cat /sys/block/%s/queue/%s", device_path, filename);
    if (DEBUG)
        printf("debug: cmd=%s\n", cmd);

    proc = popen(cmd, "r");

    if (!proc) { /* validate pipe open for reading */
        fprintf(stderr, "error: process open failed.\n");
        return 1;
    }

    if (fscanf(proc, "%d", &value) == 1) { /* read/validate value */
        if (DEBUG)
            printf("debug: value: %d\n", value);
        pclose(proc);
        return value;
    }

    fprintf(stderr, "error: invalid response.\n");
    pclose(proc);
    return -1;
}

int main(int argc, char **argv) {
    if (argc != 2) {
        printf("Usage: sudo %s <device_name>\nex) sudo %s nvme0n1\n", argv[0], argv[0]);
        return 1;
    }
    struct ring_buffer *rb = NULL;
    struct f2fszonetracer_bpf *skel;
    int err;

    nr_zones = read_sysfs_device_queue(argv[1], "nr_zones");
    zone_blocks = read_sysfs_device_queue(argv[1], "chunk_sectors") / 8;
    zone_size = zone_blocks * 4 / 1024; // MiB

    if (DEBUG)
        printf("debug: nr_zones=%d zone_blocks=%d\n", nr_zones, zone_blocks);

    if (nr_zones < 0 || zone_blocks < 0) {
        printf("error: failed to read sysfs\n");
        return 1;
    }

    printf("info: total_zone=%d zone_blocks=%d\n", nr_zones, zone_blocks);
    fflush(stdout);

    /* Set up libbpf errors and debug info callback */
    libbpf_set_print(libbpf_print_fn);

    /* Bump RLIMIT_MEMLOCK to create BPF maps */
    bump_memlock_rlimit();

    if (signal(SIGINT, sig_int) == SIG_ERR) {
        fprintf(stderr, "can't set signal handler: %s\n", strerror(errno));
        goto cleanup;
    }

    /* Open load and verify BPF application */
    skel = f2fszonetracer_bpf__open_and_load();
    if (!skel) {
        fprintf(stderr, "Failed to open BPF skeleton\n");
        return 1;
    }

    /* Attach tracepoint handler */
    err = f2fszonetracer_bpf__attach(skel);
    if (err) {
        fprintf(stderr, "Failed to attach BPF skeleton\n");
        goto cleanup;
    }

    /* Set up ring buffer */
    rb = ring_buffer__new(bpf_map__fd(skel->maps.rb), handle_event, NULL, NULL);
    if (!rb) {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }

    while (!stop) {
        err = ring_buffer__consume(rb);
        /* Ctrl-C will cause -EINTR */
        if (err == -EINTR) {
            err = 0;
            break;
        }
        if (err < 0) {
            printf("Error consuming ring buffer: %d\n", err);
            break;
        }
    }

cleanup:
    ring_buffer__free(rb);
    f2fszonetracer_bpf__destroy(skel);
    return -err;
}
