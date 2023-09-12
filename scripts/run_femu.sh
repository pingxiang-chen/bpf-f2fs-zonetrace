#!/bin/sh
set -e

IMAGE_DIR=$HOME/images      # FEMU Image directory
IMAGE_FILE=debian.qcow2     # Virtual machine disk image

# 1. Set directory to the root of the repository
ROOT=$(pwd)
if [ $(basename $PWD) = "scripts" ]
then
    ROOT=$(dirname $PWD)
fi
cd $ROOT

# 2. Check if FEMU is installed
if [ ! -d ./femu/build-femu ]
then
    # Install FEMU
    echo "Installing FEMU..."
    wget -q --show-progress "https://github.com/vtess/FEMU/archive/refs/tags/femu-v8.0.0.tar.gz" -O "femu.tar.gz"
    mkdir -p femu/build-femu
    tar zxf femu.tar.gz -C femu --strip-components=1
    rm femu.tar.gz

    # Build FEMU
    cd femu/build-femu
    cp ../femu-scripts/femu-copy-scripts.sh .
    ./femu-copy-scripts.sh .
    sudo ./pkgdep.sh
    ./femu-compile.sh
else
    echo "FEMU is already installed."
fi

cd $ROOT/femu/build-femu

# 3. Fix image directory of FEMU scripts
sed -i "s|IMGDIR=.*|IMGDIR=$IMAGE_DIR|g" run-zns.sh
sed -i "s|OSIMGF=.*|OSIMGF=\$IMGDIR/$IMAGE_FILE|g" run-zns.sh

# 4. Run FEMU
echo "Starting FEMU..."
./run-zns.sh
