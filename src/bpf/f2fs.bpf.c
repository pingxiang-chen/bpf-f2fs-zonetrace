// SPDX-License-Identifier: GPL-2.0 OR BSD-3-Clause
/* Copyright (c) 2021 Sartura */
#include "f2fs.h"

#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

/* BPF ringbuf map */
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024 /* 256KB */);
} rb SEC(".maps");

struct event {
    int segno;
    unsigned char cur_valid_map[65];
};

SEC("fexit/update_sit_entry")
int BPF_PROG(update_sit_entry, struct f2fs_sb_info *sbi, block_t blkaddr) {
    struct event *e;
    e = bpf_ringbuf_reserve(&rb, sizeof(*e), 0);
    if (!e) {
        bpf_printk("kprobe:update_sit_entry: ringbuf_reserve failed\n");
        return 0;
    }

    struct seg_entry *ses = BPF_CORE_READ(sbi, sm_info, sit_info, sentries);

    struct f2fs_sm_info *SM_I = BPF_CORE_READ(sbi, sm_info);

    // The BTF doesn't record #define macros, so we unfold the fs/f2fs/segment.h GET_SEGNO macro here
    int segno = ((blkaddr <= 0) ? (4294967295U) : (((((blkaddr) - (SM_I != 0 ? BPF_CORE_READ(SM_I, seg0_blkaddr) : BPF_CORE_READ(sbi, raw_super, segment0_blkaddr))) >> BPF_CORE_READ(sbi, log_blocks_per_seg))) - BPF_CORE_READ(SM_I, free_info, start_segno)));

    struct seg_entry *se = ses + segno;
    unsigned char *bitmap = BPF_CORE_READ(se, cur_valid_map);

    e->segno = segno;
    bpf_core_read(&e->cur_valid_map, sizeof(e->cur_valid_map), (void *)bitmap);

    bpf_ringbuf_submit(e, 0);
    return 0;
}
