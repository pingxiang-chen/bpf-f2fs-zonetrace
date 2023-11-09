#!/bin/bash
MOUNT_POINT=/mnt/f2fs
REGULAR_DEVICE=nvme4n1p1
ZNS_DEVICE=nvme3n1
base_dir=fragdistance
size=$1

# 1. Check whether user run script under exploration folder
if [ $(basename $PWD) != "exploration" ]
then
    echo "Please go to exploration folder and execute the script"
    exit -1
fi

if [ "$#" -ne 1 ]; then
    echo "Usage: ./generate_distance [file_size]"
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
result_path=$base_dir
frag_distance=4

while [ $frag_distance -le 2048 ]
do
    ./generate_once.sh $size $frag_size $frag_distance $result_path
    let frag_distance=frag_distance*2
done