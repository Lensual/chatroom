#!/bin/bash

if [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
    echo "GNU/Linux"
    linux_params='--cross-prefix=x86_64-w64-mingw32- --enable-pthreads \'
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    #pkg-config问题可能需要pacman -S mingw-w64-x86_64-pkg-config
    echo "MINGW64_NT"
else
    echo "暂不支持从该环境编译 欢迎提交pr"
    exit
fi

build_path="$(pwd)/build"

cd FFmpeg

export PKG_CONFIG_PATH="$build_path/lib/pkgconfig"

make clean
./configure \
    --prefix="$build_path" \
    --arch=x86_64 \
    --target-os=mingw32 \
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
    --extra-libs="-lssp" \
    $linux_params

make -j$(nproc)
make install
