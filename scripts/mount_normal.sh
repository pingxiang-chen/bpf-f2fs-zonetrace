#!/bin/bash

DISABLE_CHECK_POINT=$1
REGULAR_DEVICE=nvme1n1p1
MOUNT_POINT=/mnt/normal

if [ "$EUID" -ne 0 ]
then
    echo "Please run as root"
    exit
fi

if [ "$#" -ne 1 ]; then
	echo "Usage: ./mount_normal.sh [diable_checkpint]"
	exit -1
fi

mkdir -p $MOUNT_POINT
umount -q $MOUNT_POINT

if [ $DISABLE_CHECK_POINT -eq 1 ]; then
	echo "Disable checkpointing in f2fs"
	mkfs.f2fs -o checkpoint=disable:100% /dev/$REGULAR_DEVICE 
else
	echo "Enable checkpointing in f2fs"
	mkfs.f2fs /dev/$REGULAR_DEVICE 
fi

mount -t f2fs /dev/$REGULAR_DEVICE $MOUNT_POINT
