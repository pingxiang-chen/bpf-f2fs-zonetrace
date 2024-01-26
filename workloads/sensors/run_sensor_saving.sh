#!/bin/bash
MOUNT_POINT=/mnt/f2fs
REGULAR_DEVICE=nvme4n1p1
NUM_OF_SENSOR=$1

let GB=1024*1024*1024
let MB=1024*1024
let KB=1024

let num_of_blks=50*$MB/128/$KB
let counts=$NUM_OF_SENSOR

if [ "$#" -ne 1 ]; then
    echo "Usage: ./run_sensors_saving.sh [num_of_streams]"
	exit
fi

while (( --counts >= 0 )); do
	dd if=/dev/random of=$MOUNT_POINT/target_file_$counts.mov count=$num_of_blks bs=128K oflag=sync,append,nocache conv=notrunc &
done

let counts=$NUM_OF_SENSOR
let num_of_kb_read=50*$MB/$KB
avg_throughput=0.0

while (( --counts >= 0 )); do
	sync; sudo sh -c "/usr/bin/echo 3 > /proc/sys/vm/drop_caches" && sleep 2
	echo "Reading target_file_$counts.mov..." 
	throughput=$(./read_seq /mnt/f2fs/target_file_$counts.mov $num_of_kb_read | grep Throughput | tr -cd '[:digit:].')
	avg_throughput=$(echo "$avg_throughput + $throughput" | bc)
done

let counts=$NUM_OF_SENSOR
counts_f="$counts.0"

avg_throughput=$(echo "$avg_throughput / $counts_f" | bc)

echo "avg throughput = $avg_throughput"
