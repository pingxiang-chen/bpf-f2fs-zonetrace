// SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)
/* Copyright (c) 2021 Sartura
 * Based on minimal.c by Facebook */

#include <bpf/libbpf.h>
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>

#include "f2fszonetracer.skel.h"

#define DEBUG 0

int nr_zones;
int segment_size = 2; // 2 MiB
int zone_size;
int zone_blocks;
int f2fs_main_blkaddr;
int zone_start_blkaddr;
int zone_segno_offset;

unsigned int buf[3];

struct event {
    int segno;
    unsigned int seg_type;
    unsigned char cur_valid_map[65];
};

static int libbpf_print_fn(enum libbpf_print_level level, const char *format, va_list args) {
    return vfprintf(stderr, format, args);
}

static volatile sig_atomic_t stop;

static void sig_int(int signo) {
    fprintf(stderr, "signal received: %d\n", signo);
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

/**
 * handle BPF event
*/
int handle_event(void *ctx, void *data, size_t data_sz) {
    const struct event *e = data;

    int calculated_segno = e->segno - zone_segno_offset;
    if (calculated_segno < 0) {
        return 0;
    }

    unsigned int seg_per_zone = zone_size / segment_size;
    unsigned int cur_zone = calculated_segno / seg_per_zone;

    buf[0] = calculated_segno;
    buf[1] = cur_zone;
    buf[2] = e->seg_type; // __builtin_ctz(e->seg_type);

    write(1, buf, 12);
    write(1, e->cur_valid_map, 64);
    return 0;
}

/**
 * Read values from sysfs path
*/
int read_sysfs_device_queue(const char *device_path, const char *filename) {
    FILE *proc;
    char cmd[1024];
    int value;

    /* since sysfs is not a regular file, use popen instead. */
    snprintf(cmd, sizeof(cmd), "cat /sys/block/%s/queue/%s", device_path, filename);
    if (DEBUG)
        fprintf(stderr, "debug: cmd=%s\n", cmd);

    proc = popen(cmd, "r");

    if (!proc) { /* validate pipe open for reading */
        fprintf(stderr, "error: process open failed.\n");
        return 1;
    }

    if (fscanf(proc, "%d", &value) == 1) { /* read/validate value */
        if (DEBUG)
            fprintf(stderr, "debug: value: %d\n", value);
        pclose(proc);
        return value;
    }

    fprintf(stderr, "error: invalid response.\n");
    pclose(proc);
    return -1;
}

int main(int argc, char **argv) {
    if (argc != 5) {
        printf("Usage: sudo %s <device_name> <mount> <f2fs_main_blkaddr> <zoned_start_blkaddr>\nex) sudo %s nvme0n1 32768 65536\n", argv[0], argv[0]);
        return 1;
    }
    struct ring_buffer *rb = NULL;
    struct f2fszonetracer_bpf *skel;
    int err;

    struct sigaction sa;
    sa.sa_handler = sig_int;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;

    if (sigaction(SIGINT, &sa, NULL) == -1) {
        printf("signo %d ", SIGINT);
        perror("sigaction");
        return 1;
    }

    if (sigaction(SIGTERM, &sa, NULL) == -1) {
        printf("signo %d ", SIGTERM);
        perror("sigaction");
        return 1;
    }

    nr_zones = read_sysfs_device_queue(argv[1], "nr_zones");
    zone_blocks = read_sysfs_device_queue(argv[1], "chunk_sectors") / 8;
    zone_size = zone_blocks * 4 / 1024;                                 // MiB
    f2fs_main_blkaddr = atoi(argv[3]);                                  // f2fs main block address
    zone_start_blkaddr = atoi(argv[4]);                                 // zoned start block address
    zone_segno_offset = (zone_start_blkaddr - f2fs_main_blkaddr) / 512; // total amount of regular block device segments

    if (DEBUG) {
        fprintf(stderr, "zone_segno_offset %d\n", zone_segno_offset);
    }

    if (DEBUG)
        fprintf(stderr, "debug: nr_zones=%d zone_blocks=%d\n", nr_zones, zone_blocks);

    if (nr_zones < 0 || zone_blocks < 0) {
        printf("error: failed to read sysfs\n");
        return 1;
    }

    printf("info: mount=%s total_zone=%d zone_blocks=%d\n", argv[2], nr_zones, zone_blocks);
    fflush(stdout);

    /* Set up libbpf errors and debug info callback */
    if (DEBUG)
        libbpf_set_print(libbpf_print_fn);

    /* Bump RLIMIT_MEMLOCK to create BPF maps */
    bump_memlock_rlimit();

    if (signal(SIGPIPE, sig_int) == SIG_ERR) {
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
            fprintf(stderr, "Error consuming ring buffer: %d\n", err);
            break;
        }
    }

cleanup:
    ring_buffer__free(rb);
    f2fszonetracer_bpf__destroy(skel);
    return -err;
}
