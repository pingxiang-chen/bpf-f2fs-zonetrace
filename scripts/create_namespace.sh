#!/bin/sh

ZNS_SSD=/dev/nvme3

SECTOR_SIZE=4096	# bytes
ZONE_SIZE=524288	# sectors (4KiB)
GB=1073741824 		# bytes
CONVENTION_SIZE=$((2*GB/SECTOR_SIZE))

NUM_OF_ZONE=0
NUM_OF_ZONE_EACH_NAMESPACE=0


if [ "$#" -ne 1 ]; then
	echo "Create namespace for ZNS SSD with given number of zones"
	echo "WARNING: This script will delete all namespaces of ZNS SSD at first and then create the new one"
	echo ""
	echo "Usage: ./create_namespace.sh [# of zones]"
	exit -1
fi

# 1. Delete all namespaces of ZNS SSD
nvme delete-ns $ZNS_SSD -n 0xffffffff

if [ $? -ne 0 ]
then
	echo "Delete namespace failed"
	exit 1
fi

# 2. Format ZNS SSD
nvme format $ZNS_SSD -n 0xffffffff -l 2 -f

if [ $? -ne 0 ]
then
	echo "Format namespace failed"
	exit 1
fi

# 3. Calculate number of zones of ZNS SSD
SIZE=$(nvme id-ctrl $ZNS_SSD | grep tnvmcap | sed 's/,//g'| sed 's/^tnvmcap   : //g' | awk '{print $1/4096}')

NUM_OF_ZONE=$((SIZE/ZONE_SIZE))
echo "Total zone in ZNS NVMe: $NUM_OF_ZONE"

NUM_OF_ZONE_OF_EACH_NAMESPACE=$1
echo "Number of zones of each namespace: $NUM_OF_ZONE_OF_EACH_NAMESPACE"

SIZE_OF_EACH_NAMESPACE=$((NUM_OF_ZONE_OF_EACH_NAMESPACE*ZONE_SIZE))

# 4. Create namespace
nvme create-ns -s $SIZE_OF_EACH_NAMESPACE -c $SIZE_OF_EACH_NAMESPACE -f 2 -d 0 --csi=2 $ZNS_SSD

if [ $? -ne 0 ]
then
	echo "Create namespace failed"
	exit 1
fi

# 5. Attach namespace
nvme attach-ns $ZNS_SSD -n 1 -c 0