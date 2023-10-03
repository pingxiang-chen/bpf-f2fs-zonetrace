#!/bin/bash

MOUNT_POINT=/mnt/f2fs
FILE=test.0.0

[ -f $MOUNT_POINT/$FILE ] && watch -n 1 ../../fiemap/fiemap $MOUNT_POINT/$FILE