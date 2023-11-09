#!/bin/bash
MOUNT_POINT=/mnt/f2fs
REGULAR_DEVICE=nvme4n1p1

if [ $(basename $PWD) != "exploration" ]
then
    echo "Please go to exploration folder and execute the script"
    exit -1
fi

if [ "$#" -ne 4 ]; then
    echo "Usage: ./generate_once [file_size] [fragsize] [fragdistance] [result_path]"
	exit
fi

size=$1
result_path="results/$4"
[ ! -d results ] && mkdir results
[ ! -d $result_path ] && mkdir $result_path

let req_size=$size*1024

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

for frag_size in $2
do
    let counts=$size*1024/$frag_size

    if [ "$counts" -eq 0 ]
    then
        continue
    fi

    # Unmount the tested device and remove block folder
    if $(findmnt --source /dev/$REGULAR_DEVICE >/dev/null);
    then echo "unmount /dev/$REGULAR_DEVICE" && sudo umount /dev/$REGULAR_DEVICE
    fi 
    sleep 1
    ../scripts/mount_real.sh 0 &> /dev/null
    sleep 1
    sync
    sleep 1

    for frag_distance in $3
    do
        while (( --counts >= 0 )); do
            dd if=/dev/zero of=$MOUNT_POINT/target_file.png count=1 bs=${frag_size}K oflag=sync,append,nocache conv=notrunc &> /dev/null
            dd if=/dev/zero of=$MOUNT_POINT/dummy.png count=1 bs=${frag_distance}K oflag=sync,append,nocache conv=notrunc &> /dev/null
        done
        rm -f $MOUNT_POINT/dummy.png
        sync
        sync; sudo sh -c "/usr/bin/echo 3 > /proc/sys/vm/drop_caches" && sleep 1
        val=$(./read_seq $MOUNT_POINT/target_file.png $req_size | awk '{print $3}')
        printf "Throughput = %f\n" $val > $result_path/${frag_size}_${frag_distance}.result > $result_path/${frag_size}_${frag_distance}.result
        echo "frag_size = $frag_size" >> $result_path/${frag_size}_${frag_distance}.result
        ../f2fs-tools/tools/fibmap.f2fs $MOUNT_POINT/target_file.png >> $result_path/${frag_size}_${frag_distance}.result
        sleep 1
    done
done