#!/bin/bash
MOUNT_POINT=/mnt/f2fs
REGULAR_DEVICE=nvme4n1p1
ZNS_DEVICE=nvme3n1

base_dir=fragsize
size=$1

if [ "$#" -ne 1 ]; then
    echo "Usage: ./generate_frag [file_size]"
	exit -1
fi

let req_size=$size*1024

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

max_segments="$(cat /sys/block/$ZNS_DEVICE/queue/max_segments)"
hw_sector_size="$(cat /sys/block/$ZNS_DEVICE/queue/hw_sector_size)"
let max_explore_size=$max_segments*$hw_sector_size/1024

frag_size=4
frag_distance=1024

mkdir $base_dir -p
result_path=$base_dir/$size"_"$frag_distance


while [ $frag_size -le $max_explore_size ]
do
    ./generate_once.sh $size $frag_size $frag_distance $result_path
    let frag_size=frag_size*2
done

./generate_once.sh $size $max_explore_size $frag_distance $result_path