// SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause)
/* Copyright (c) 2021 Sartura
 * Based on minimal.c by Facebook */

#include <bpf/libbpf.h>
#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>
#include <sys/resource.h>
#include <unistd.h>

#include "f2fs.skel.h"

int nr_zone;
int nr_op_zone;
int zone_size;
int zone_blocks;
int zone_cap_blocks;

struct event {
    int segno;
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

    const unsigned int segment_size = 2;  // 2 MiB
    unsigned int seg_per_zone = zone_size / segment_size;
    unsigned int cur_zone = e->segno / seg_per_zone;

    printf("update_sit_entry segno: %d cur_zone:%d\n", e->segno % 1024, cur_zone);
    fflush(stdout);
    write(1, e->cur_valid_map, 64);
    printf("\n");
    return 0;
}

int main(int argc, char **argv) {
    if (argc < 4) {
        printf("Usage: sudo %s <nr_zone> <nr_op_zone> <zone_size(MiB)>\n", argv[0]);
        return 1;
    }
    struct ring_buffer *rb = NULL;
    struct f2fs_bpf *skel;
    int err;

    nr_zone = atoi(argv[1]);
    nr_op_zone = atoi(argv[2]);
    zone_size = atoi(argv[3]);

    zone_blocks = 524288;
    zone_cap_blocks = 275712;

    printf("info: total_zone=%d zone_blocks=%d zone_cap_blocks=%d\n", nr_zone, zone_blocks, zone_cap_blocks);
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
    skel = f2fs_bpf__open_and_load();
    if (!skel) {
        fprintf(stderr, "Failed to open BPF skeleton\n");
        return 1;
    }

    /* Attach tracepoint handler */
    err = f2fs_bpf__attach(skel);
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
    f2fs_bpf__destroy(skel);
    return -err;
}
