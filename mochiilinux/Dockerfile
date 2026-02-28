FROM alpine AS stage0

# Stage0 - Building the cross-compiler

ARG VENDOR="mochii"
ARG ARCH="x86-64"
ARG BUILD_ARCH="x86_64"
ARG JOBS=16
ARG MUSSEL_VERSION="95dec40aee2077aa703b7abc7372ba4d34abb889"
ENV VENDOR=${VENDOR}
ENV ARCH=${ARCH}
ENV BUILD_ARCH=${BUILD_ARCH}
ENV JOBS=${JOBS}
ENV MUSSEL_VERSION=${MUSSEL_VERSION}

RUN apk update && \
  apk add git bash wget perl build-base make patch busybox-static curl && \
  git clone https://github.com/firasuke/mussel.git && \
  cd mussel && \
  git checkout ${MUSSEL_VERSION} -b build && \
  ./mussel ${ARCH} -k -l -o -p -s -T ${VENDOR}

ENV PATH=/mussel/toolchain/bin:$PATH
ENV LC_ALL=POSIX
ENV TARGET=${BUILD_ARCH}-${VENDOR}-linux-musl
ENV BUILD=${BUILD_ARCH}-pc-linux-musl

FROM stage0 AS sources-downloader

ENV CURL_VERSION=8.5.0
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://curl.se/download/curl-${CURL_VERSION}.tar.gz \
  -O curl-${CURL_VERSION}.tar.gz

ENV RSYNC_VERSION=3.4.1
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://download.samba.org/pub/rsync/rsync-${RSYNC_VERSION}.tar.gz \
  -O rsync-${RSYNC_VERSION}.tar.gz

ENV XXHASH_VERSION=0.8.3
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://github.com/Cyan4973/xxHash/archive/refs/tags/v${XXHASH_VERSION}.tar.gz \
  -O xxHash-${XXHASH_VERSION}.tar.gz

ENV ZSTD_VERSION=1.5.7
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://github.com/facebook/zstd/archive/v${ZSTD_VERSION}.tar.gz \
  -O zstd-${ZSTD_VERSION}.tar.gz

ENV LZ4_VERSION=1.10.0
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://github.com/lz4/lz4/archive/v${LZ4_VERSION}.tar.gz \
  -O lz4-${LZ4_VERSION}.tar.gz

ENV ZLIB_VERSION=1.3.1
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://zlib.net/fossils/zlib-${ZLIB_VERSION}.tar.gz \
  -O zlib-${ZLIB_VERSION}.tar.gz

ENV ACL_VERSION=2.3.2
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://download.savannah.gnu.org/releases/acl/acl-${ACL_VERSION}.tar.gz \
  -O acl-${ACL_VERSION}.tar.gz

ENV ATTR_VERSION=2.5.2
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://download.savannah.nongnu.org/releases/attr/attr-${ATTR_VERSION}.tar.gz \
  -O attr-${ATTR_VERSION}.tar.gz

ENV GAWK_VERSION=5.3.2
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://ftp.gnu.org/gnu/gawk/gawk-${GAWK_VERSION}.tar.xz \
  -O gawk-${GAWK_VERSION}.tar.xz

ENV CA_CERTIFICATES_VERSION=20250619
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://gitlab.alpinelinux.org/alpine/ca-certificates/-/archive/${CA_CERTIFICATES_VERSION}/ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 \
  -O ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2

ENV SYSTEMD_VERSION=257.6
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://github.com/systemd/systemd/archive/refs/tags/v${SYSTEMD_VERSION}.tar.gz \
  -O systemd-${SYSTEMD_VERSION}.tar.gz

ENV LIBCAP_VERSION=2.76
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-${LIBCAP_VERSION}.tar.xz \
  -O libcap-${LIBCAP_VERSION}.tar.xz

ENV UTIL_LINUX_VERSION=2.41.1
ENV UTIL_LINUX_VERSION_MAJOR=2.41
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://kernel.org/pub/linux/utils/util-linux/v${UTIL_LINUX_VERSION_MAJOR}/util-linux-${UTIL_LINUX_VERSION}.tar.xz \
  -O util-linux-${UTIL_LINUX_VERSION}.tar.xz

ENV PYTHON_VERSION=3.12.11
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tar.xz \
  -O python-${PYTHON_VERSION}.tar.xz

ENV SQLITE3_VERSION=3500400
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://www.sqlite.org/2025/sqlite-autoconf-${SQLITE3_VERSION}.tar.gz \
  -O sqlite-autoconf-${SQLITE3_VERSION}.tar.gz

ENV OPENSSL_VERSION=3.5.2
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz \
  -O openssl-${OPENSSL_VERSION}.tar.gz

ENV PKGCONFIG_VERSION=1.8.1
RUN mkdir -p /sources/downloads && cd /sources/downloads && \
  wget https://distfiles.dereferenced.org/pkgconf/pkgconf-${PKGCONFIG_VERSION}.tar.xz \
  -O pkgconf-${PKGCONFIG_VERSION}.tar.xz

RUN apk add git patch

ENV OE_CORE_VERSION=bed605e7cd095f210e24f1e07dab6f814d0bbb59

RUN cd /sources/downloads && tar -xvf systemd-${SYSTEMD_VERSION}.tar.gz && \
  mv systemd-${SYSTEMD_VERSION} systemd
RUN cd /sources/downloads && git clone https://github.com/openembedded/openembedded-core && \
  cd openembedded-core && \
  git checkout ${OE_CORE_VERSION}
COPY patches/apply_all.sh /apply_all.sh
RUN chmod +x /apply_all.sh
RUN /apply_all.sh /sources/downloads/openembedded-core/meta/recipes-core/systemd/systemd/ /sources/downloads/systemd

FROM stage0 AS skeleton

COPY ./setup_rootfs.sh ./setup_rootfs.sh
RUN chmod +x ./setup_rootfs.sh && SYSROOT=/sysroot ./setup_rootfs.sh

# Stage0 - building the packages using the cross-compiler

# Busybox
FROM stage0 AS busybox-stage0

ENV BUSYBOX_VERSION=1.37.0

RUN mkdir /sources && cd /sources && \
  wget https://busybox.net/downloads/busybox-${BUSYBOX_VERSION}.tar.bz2

RUN cd /sources && tar -xvf busybox-${BUSYBOX_VERSION}.tar.bz2 && \
  cd busybox-${BUSYBOX_VERSION} && \
  make distclean && \
  make ARCH="${ARCH}" defconfig && \
  sed -i 's/\(CONFIG_\)\(.*\)\(INETD\)\(.*\)=y/# \1\2\3\4 is not set/g' .config && \
  sed -i 's/\(CONFIG_IFPLUGD\)=y/# \1 is not set/g' .config && \
  sed -i 's/\(CONFIG_FEATURE_WTMP\)=y/# \1 is not set/g' .config && \
  sed -i 's/\(CONFIG_FEATURE_UTMP\)=y/# \1 is not set/g' .config && \
  sed -i 's/\(CONFIG_UDPSVD\)=y/# \1 is not set/g' .config && \
  sed -i 's/\(CONFIG_TCPSVD\)=y/# \1 is not set/g' .config && \
  sed -i 's/\(CONFIG_TC\)=y/# \1 is not set/g' .config && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" CONFIG_PREFIX="/sysroot" install

# Musl
FROM stage0 AS musl-stage0

ENV MUSL_VERSION=1.2.5

RUN mkdir /sources && cd /sources && \
  wget https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz

RUN cd /sources && tar -xvf musl-${MUSL_VERSION}.tar.gz && \
  cd musl-${MUSL_VERSION} && \
  ./configure \
    CROSS_COMPILE=${TARGET}- \
    --prefix=/usr \
    --disable-static \
    --target=${TARGET} && \
  make -j${JOBS} &&\
  DESTDIR=/sysroot make -j${JOBS} install

# GCC
FROM stage0 AS gcc-stage0

ENV GCC_VERSION=14.3.0
ENV GMP_VERSION=6.3.0
ENV MPC_VERSION=1.3.1
ENV MPFR_VERSION=4.2.2

RUN mkdir /sources && cd /sources && \
  wget \
    https://ftp.gnu.org/gnu/gcc/gcc-${GCC_VERSION}/gcc-${GCC_VERSION}.tar.xz \
    https://ftp.gnu.org/gnu/gmp/gmp-${GMP_VERSION}.tar.bz2 \
    https://ftp.gnu.org/gnu/mpc/mpc-${MPC_VERSION}.tar.gz \
    https://ftp.gnu.org/gnu/mpfr/mpfr-${MPFR_VERSION}.tar.bz2 && \
  tar -xvf gcc-${GCC_VERSION}.tar.xz && \
  tar -xvf gmp-${GMP_VERSION}.tar.bz2 && \
  tar -xvf mpc-${MPC_VERSION}.tar.gz && \
  tar -xvf mpfr-${MPFR_VERSION}.tar.bz2 && \
  mv -v gmp-${GMP_VERSION} gcc-${GCC_VERSION}/gmp && \
  mv -v mpc-${MPC_VERSION} gcc-${GCC_VERSION}/mpc && \
  mv -v mpfr-${MPFR_VERSION} gcc-${GCC_VERSION}/mpfr

RUN cd /sources && \
  mkdir -p /sysroot/usr/include && \
  cd gcc-${GCC_VERSION} && mkdir -v build && cd build && ls -liah /gcc-${GCC_VERSION}/mpfr/src/ && \
  ../configure \
    --prefix=/usr \
    --build=${BUILD_ARCH} \
    --host=${TARGET} \
    --target=${TARGET} \
    --with-sysroot=/ \
    --disable-nls \
    --enable-languages=c,c++ \
    --enable-c99 \
    --enable-long-long \
    --disable-libmudflap \
    --disable-multilib \
    --disable-libsanitizer && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" -j${JOBS} && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" DESTDIR=/sysroot install ;

# Make
FROM stage0 AS make-stage0

ENV MAKE_VERSION=4.4.1
RUN mkdir /sources && cd /sources && \
  wget https://ftp.gnu.org/gnu/make/make-${MAKE_VERSION}.tar.gz

RUN cd /sources && tar -xvf make-${MAKE_VERSION}.tar.gz && \
  cd make-${MAKE_VERSION} && \
  ./configure --prefix=/usr \
  --build=${BUILD_ARCH} --host=${TARGET} && \
  make -j${JOBS} && \
  make -j${JOBS} DESTDIR=/sysroot install

# Binutils
FROM stage0 AS binutils-stage0

ENV BINUTILS_VERSION=2.44
RUN mkdir /sources && cd /sources && \
  wget https://ftp.gnu.org/gnu/binutils/binutils-${BINUTILS_VERSION}.tar.xz

RUN cd /sources && tar -xvf binutils-${BINUTILS_VERSION}.tar.xz && \
  cd binutils-${BINUTILS_VERSION} && \
  ./configure \
    --prefix=/usr \
    --build=${BUILD_ARCH} \
    --host=${TARGET} \
    --target=${TARGET} \
    --with-sysroot=/ \
    --disable-nls \
    --disable-multilib \
    --enable-shared && \
  make -j${JOBS} && \
  make DESTDIR=/sysroot install ;
