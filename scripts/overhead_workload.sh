#!/bin/bash

DIR=../.record/overhead
mkdir -p ${DIR}

for i in 1 2 4 8 16 32
do
    echo "########## numjobs=$i Start ##########"

    for j in 1 2 3 4 5
    do
        echo -n "Mount the ZNS SSD with f2fs..."
        sudo ./mount_real.sh 2>&1 > /dev/null
        echo "Done"

        echo -n "run fio..."
        sudo NUMJOBS=$i fio --output="${DIR}/j${i}_${j}_fio_with_trace.log" ../overhead.fio
        echo "Done"
    done

    echo "########## numjobs=$i End ##########"

done
