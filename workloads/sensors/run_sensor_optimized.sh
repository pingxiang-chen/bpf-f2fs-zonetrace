#!/bin/bash
MOUNT_POINT=/mnt/f2fs
REGULAR_DEVICE=nvme4n1p1

let GB=1024*1024*1024
let MB=1024*1024
let KB=1024
let num_of_blks=50*$MB/128/$KB

dd if=/dev/random of=$MOUNT_POINT/target_file_1.mov count=$num_of_blks bs=128K oflag=sync,append,nocache conv=notrunc &
dd if=/dev/random of=$MOUNT_POINT/target_file_1.db count=$num_of_blks bs=128K oflag=sync,append,nocache conv=notrunc &

let num_of_kb_read=50*$MB/$KB
avg_throughput=0.0

# read cold *.mov
sync; sudo sh -c "/usr/bin/echo 3 > /proc/sys/vm/drop_caches" && sleep 2
echo "Reading target_file_1.mov..." 
throughput=$(./read_seq /mnt/f2fs/target_file_1.mov $num_of_kb_read | grep Throughput | tr -cd '[:digit:].')
avg_throughput=$(echo "$avg_throughput + $throughput" | bc)

# read hot *.db
sync; sudo sh -c "/usr/bin/echo 3 > /proc/sys/vm/drop_caches" && sleep 2
echo "Reading target_file_1.db..." 
throughput=$(./read_seq /mnt/f2fs/target_file_1.db $num_of_kb_read | grep Throughput | tr -cd '[:digit:].')
avg_throughput=$(echo "$avg_throughput + $throughput" | bc)


counts_f="2.0"
avg_throughput=$(echo "$avg_throughput / $counts_f" | bc)
echo "avg throughput = $avg_throughput"
