#!/bin/bash

if [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    echo "GNU/Linux"
    linux_params='--enable-pthreads'
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    #pkg-config问题可能需要pacman -S mingw-w64-x86_64-pkg-config
    echo "MINGW64_NT"
    mingw64_params='--arch=x86_64 --target-os=mingw32 --extra-libs="-lssp"'
else
    echo "暂不支持从该环境编译 欢迎提交pr"
    exit
fi

build_path="$(pwd)/build"

cd FFmpeg

export PKG_CONFIG_PATH="$build_path/lib/pkgconfig"

git clean -xfd

#make clean
./configure \
    --prefix="$build_path" \
    --disable-doc \
    --disable-programs \
    --disable-network \
    --disable-everything \
    --disable-autodetect \
    --enable-libopus \
    --enable-decoder=libopus \
    --enable-encoder=libopus \
    --enable-static \
    --disable-shared \
    --pkg-config-flags="--static" \
    --extra-ldflags="-static" \
    --extra-cflags="-static -pipe -Ofast -funroll-loops -ffast-math" \
    $linux_params \
    $mingw64_params

make -j$(nproc)
make install
