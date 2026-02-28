# =============================================================================
# Mochiilinux - A minimal Linux distribution built from scratch
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 0: Base image with cross-compiler toolchain
# -----------------------------------------------------------------------------
# This stage uses mussel to build a cross-compiler toolchain for the target
# architecture. The toolchain includes gcc, binutils, musl libc, and kernel
# headers for the specified architecture.
# -----------------------------------------------------------------------------
FROM alpine AS stage0

# Build arguments for customizing the target architecture and vendor
ARG VENDOR="mochii"
ARG ARCH="x86-64"
ARG BUILD_ARCH="x86_64"
ARG JOBS=16
ARG MUSSEL_VERSION="95dec40aee2077aa703b7abc7372ba4d34abb889"

# Environment variables passed to subsequent stages
ENV VENDOR=${VENDOR}
ENV ARCH=${ARCH}
ENV BUILD_ARCH=${BUILD_ARCH}
ENV JOBS=${JOBS}
ENV MUSSEL_VERSION=${MUSSEL_VERSION}

# Install required build dependencies and clone mussel toolchain builder
RUN apk update && \
  apk add git bash wget perl build-base make patch busybox-static curl && \
  git clone https://github.com/firasuke/mussel.git && \
  cd mussel && \
  git checkout ${MUSSEL_VERSION} -b build && \
  ./mussel ${ARCH} -k -l -o -p -s -T ${VENDOR}

# Configure environment for cross-compilation
ENV PATH=/mussel/toolchain/bin:$PATH
ENV LC_ALL=POSIX
ENV TARGET=${BUILD_ARCH}-${VENDOR}-linux-musl
ENV BUILD=${BUILD_ARCH}-pc-linux-musl


# -----------------------------------------------------------------------------
# Stage 0: Source code downloads
# -----------------------------------------------------------------------------
# Downloads all source tarballs needed for building the distribution.
# These sources are cached in this stage for use by later build stages.
# -----------------------------------------------------------------------------
FROM stage0 AS sources-downloader

# Download source tarballs - each package has its own versioned tarball
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

# Install additional tools needed for patching
RUN apk add git patch

# Apply patches to systemd from OpenEmbedded Core
ENV OE_CORE_VERSION=bed605e7cd095f210e24f1e07dab6f814d0bbb59

RUN cd /sources/downloads && tar -xvf systemd-${SYSTEMD_VERSION}.tar.gz && \
  mv systemd-${SYSTEMD_VERSION} systemd
RUN cd /sources/downloads && git clone https://github.com/openembedded/openembedded-core && \
  cd openembedded-core && \
  git checkout ${OE_CORE_VERSION}
COPY patches/apply_all.sh /apply_all.sh
RUN chmod +x /apply_all.sh
RUN /apply_all.sh /sources/downloads/openembedded-core/meta/recipes-core/systemd/systemd/ /sources/downloads/systemd


# -----------------------------------------------------------------------------
# Skeleton Stage: Base root filesystem
# -----------------------------------------------------------------------------
# Creates the initial root filesystem structure using the setup script.
# This provides the basic directory layout and essential files.
# -----------------------------------------------------------------------------
FROM stage0 AS skeleton

COPY ./setup_rootfs.sh ./setup_rootfs.sh
RUN chmod +x ./setup_rootfs.sh && SYSROOT=/sysroot ./setup_rootfs.sh


# =============================================================================
# Stage 0: Package Build Stages
# =============================================================================
# Each package is built in its own stage using the cross-compiler from stage0.
# The resulting binaries are installed to /sysroot for later merging.
# -----------------------------------------------------------------------------

# Busybox: Tiny Swiss Army knife of embedded Linux
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


# Musl: Lightweight, fast, simple libc implementation
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
  make -j${JOBS} && \
  DESTDIR=/sysroot make -j${JOBS} install


# GCC: GNU Compiler Collection (C/C++ compiler)
FROM stage0 AS gcc-stage0

ENV GCC_VERSION=14.3.0
ENV GMP_VERSION=6.3.0
ENV MPC_VERSION=1.3.1
ENV MPFR_VERSION=4.2.2

# Download GCC and its dependencies (GMP, MPC, MPFR)
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

# Configure and build GCC as a cross-compiler
RUN cd /sources && \
  mkdir -p /sysroot/usr/include && \
  cd gcc-${GCC_VERSION} && mkdir -v build && cd build && \
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
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" DESTDIR=/sysroot install


# GNU Make: Build automation tool
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


# Binutils: Collection of binary tools (ld, as, ar, etc.)
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
  make DESTDIR=/sysroot install


# -----------------------------------------------------------------------------
# Kernel Headers Provider
# -----------------------------------------------------------------------------
# Hack: Use Alpine's kernel headers to avoid building them from scratch.
# These headers are needed to compile userspace programs like busybox.
# -----------------------------------------------------------------------------
FROM alpine AS alpine-hack
RUN apk add linux-headers


# -----------------------------------------------------------------------------
# Stage 1 Merge: Combine all built packages into single rootfs
# -----------------------------------------------------------------------------
# This stage merges all the individually built packages into a single
# root filesystem that forms the basis of the final image.
# -----------------------------------------------------------------------------
FROM stage0 AS stage1-merge

RUN apk add rsync

# Copy skeleton rootfs as the base
COPY --from=skeleton /sysroot /skeleton

# Merge each package into the skeleton rootfs using rsync
# This preserves permissions, attributes, and handles hard links correctly

# GCC (C/C++ compiler)
COPY --from=gcc-stage0 /sysroot /gcc
RUN rsync -aHAX --keep-dirlinks /gcc/. /skeleton

# Musl (C standard library)
COPY --from=musl-stage0 /sysroot /musl
RUN rsync -aHAX --keep-dirlinks /musl/. /skeleton

# Busybox (coreutils replacement)
COPY --from=busybox-stage0 /sysroot /busybox
RUN rsync -aHAX --keep-dirlinks /busybox/. /skeleton

# GNU Make (build tool)
COPY --from=make-stage0 /sysroot /make
RUN rsync -aHAX --keep-dirlinks /make/. /skeleton

# Binutils (binary tools)
COPY --from=binutils-stage0 /sysroot /binutils
RUN rsync -aHAX --keep-dirlinks /binutils/. /skeleton

# Kernel headers (required for compiling userspace programs)
COPY --from=alpine-hack /usr/include/linux /linux
RUN mkdir -p /skeleton/usr/include/linux && rsync -aHAX --keep-dirlinks /linux/. /skeleton/usr/include/linux

COPY --from=alpine-hack /usr/include/asm /asm
RUN mkdir -p /skeleton/usr/include/asm && rsync -aHAX --keep-dirlinks /asm/. /skeleton/usr/include/asm

COPY --from=alpine-hack /usr/include/asm-generic /asm-generic
RUN mkdir -p /skeleton/usr/include/asm-generic && rsync -aHAX --keep-dirlinks /asm-generic/. /skeleton/usr/include/asm-generic

COPY --from=alpine-hack /usr/include/mtd /mtd
RUN mkdir -p /skeleton/usr/include/mtd && rsync -aHAX --keep-dirlinks /mtd/. /skeleton/usr/include/mtd


# -----------------------------------------------------------------------------
# Final Image Stage: Clean minimal image
# -----------------------------------------------------------------------------
# The final minimal image containing only the built root filesystem.
# This serves as the base for further customization or testing.
# -----------------------------------------------------------------------------
FROM scratch AS stage1

ARG VENDOR="mochii"
ARG ARCH="x86-64"
ARG BUILD_ARCH="x86_64"

ENV VENDOR=${VENDOR}
ENV ARCH=${ARCH}
ENV BUILD_ARCH=${BUILD_ARCH}
ENV TARGET=${BUILD_ARCH}-${VENDOR}-linux-musl
ENV BUILD=${BUILD_ARCH}-pc-linux-musl
ENV COMMON_ARGS="--prefix=/usr --host=${TARGET} --build=${BUILD}"

# Copy the merged root filesystem into the final image
COPY --from=stage1-merge /skeleton /


# -----------------------------------------------------------------------------
# Build Dependencies Stage: Install build tools needed for compiling packages
# -----------------------------------------------------------------------------
# This stage installs additional build dependencies needed by autoconf and
# other autotools. We need Alpine Linux for this, so we use a hybrid approach.
# -----------------------------------------------------------------------------
FROM alpine AS build-deps
RUN apk add m4


# -----------------------------------------------------------------------------
# Test Stage: Verify the build environment works correctly
# -----------------------------------------------------------------------------
# A test stage to verify that the toolchain and basic compilation work.
# This stage compiles a simple C program to validate the build environment.
# -----------------------------------------------------------------------------
FROM stage1 AS test1

# Verify basic tools are available
RUN ls -laih /
RUN gcc --version
RUN make --version

# Compile and run a simple test program
COPY ./tests/gcc/test.c test.c
RUN gcc -Wall test.c -o test
RUN ./test

# Stage 1.5 - Building the packages for the final image

# Musl
FROM stage1 AS musl

ENV MUSL_VERSION=1.2.5

RUN mkdir /sources && cd /sources && \
  wget https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz

RUN cd /sources && tar -xvf musl-${MUSL_VERSION}.tar.gz && \
  cd musl-${MUSL_VERSION} && \
  ./configure \
    --prefix=/usr \
    --disable-static && \
  make -j${JOBS} && \
  DESTDIR=/sysroot make -j${JOBS} install

# Pkgconfig
FROM stage1 AS pkgconfig

ENV PKGCONFIG_VERSION=1.8.1

COPY --from=sources-downloader /sources/downloads/pkgconf-${PKGCONFIG_VERSION}.tar.xz /sources/

RUN mkdir -p /sources && cd /sources && tar -xvf pkgconf-${PKGCONFIG_VERSION}.tar.xz && \
  cd pkgconf-${PKGCONFIG_VERSION} && \
  mkdir -p /pkgconfig && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --localstatedir=/var \
    --with-pkg-config-dir=/usr/local/lib/pkgconfig:/usr/local/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig && \
  make && \
  make DESTDIR=/pkgconfig install && \
  make install && \
  ln -s pkgconf /pkgconfig/usr/bin/pkg-config

# Autoconf
FROM build-deps AS autoconf

COPY --from=stage1-merge /skeleton /

ENV AUTOCONF_VERSION=2.71

RUN mkdir -p /sources && cd /sources && wget http://ftp.gnu.org/gnu/autoconf/autoconf-${AUTOCONF_VERSION}.tar.xz && \
  tar -xvf autoconf-${AUTOCONF_VERSION}.tar.xz && \
  cd autoconf-${AUTOCONF_VERSION} && \
  mkdir -p /autoconf && \
  ./configure ${COMMON_ARGS} && \
  make DESTDIR=/autoconf && \
  make DESTDIR=/autoconf install && \
  make install

# Automake
FROM build-deps AS automake

COPY --from=stage1-merge /skeleton /

ENV AUTOMAKE_VERSION=1.16.5

RUN mkdir -p /sources && cd /sources && wget http://ftp.gnu.org/gnu/automake/automake-${AUTOMAKE_VERSION}.tar.xz && \
  tar -xvf automake-${AUTOMAKE_VERSION}.tar.xz && \
  cd automake-${AUTOMAKE_VERSION} && \
  mkdir -p /autoconf && \
  ./configure ${COMMON_ARGS} && \
  make DESTDIR=/autoconf && \
  make DESTDIR=/autoconf install && \
  make install

# XXHash
FROM stage1 AS xxhash

ENV XXHASH_VERSION=0.8.3

COPY --from=sources-downloader /sources/downloads/xxHash-${XXHASH_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf xxHash-${XXHASH_VERSION}.tar.xz && \
  cd xxHash-${XXHASH_VERSION} && \
  mkdir -p /xxhash && \
  CC=gcc make DESTDIR=/xxhash && \
  make DESTDIR=/xxhash install && \
  make install

# Zstd
FROM xxhash AS zstd

ENV ZSTD_VERSION=1.5.7

COPY --from=sources-downloader /sources/downloads/zstd-${ZSTD_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf zstd-${ZSTD_VERSION}.tar.xz && \
  cd zstd-${ZSTD_VERSION} && \
  mkdir -p /zstd && \
  CC=gcc make DESTDIR=/zstd && \
  make DESTDIR=/zstd install && \
  make install

# Zstd
FROM zstd AS lz4

ENV LZ4_VERSION=1.10.0

COPY --from=sources-downloader /sources/downloads/lz4-${LZ4_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf lz4-${LZ4_VERSION}.tar.xz && \
  cd lz4-${LZ4_VERSION} && \
  mkdir -p /lz4 && \
  CC=gcc make PREFIX="/usr" DESTDIR=/lz4 && \
  make DESTDIR=/lz4 install && \
  make install

# Attr
FROM lz4 AS attr

ENV ATTR_VERSION=2.5.2

COPY --from=sources-downloader /sources/downloads/attr-${ATTR_VERSION}.tar.gz /sources/
COPY ./patches/attr/basename.patch /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf attr-${ATTR_VERSION}.tar.xz && \
  cd attr-${ATTR_VERSION} && \
  mkdir -p /attr && \
  patch -p1 < /sources/basename.patch && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --localstatedir=/var \
    --disable-nls && \
  make DESTDIR=/attr && \
  make DESTDIR=/attr install && \
  make install

# ACL
FROM attr AS acl

ENV ACL_VERSION=2.3.2

COPY --from=sources-downloader /sources/downloads/acl-${ACL_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf acl-${ACL_VERSION}.tar.gz && \
  mv acl-${ACL_VERSION} acl && \
  cd acl && mkdir -p /acl && \
  ./configure ${COMMON_ARGS} \
    --prefix=/usr \
    --disable-dependency-tracking \
    --libexecdir=/usr/libexec && \
  make DESTDIR=/acl && \
  make DESTDIR=/acl install && \
  make install

# Popt
FROM acl AS popt

ENV POPT_VERSION=1.19

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.rpm.org/popt/releases/popt-1.x/popt-${POPT_VERSION}.tar.gz && \
  tar -xvf popt-${POPT_VERSION}.tar.gz && \
  mv popt-${POPT_VERSION} popt && \
  cd popt && mkdir -p /popt && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr && \
  make DESTDIR=/popt && \
  make DESTDIR=/popt install && \
  make install

# Zlib
FROM popt AS zlib

ENV ZLIB_VERSION=1.3.1

COPY --from=sources-downloader /sources/downloads/zlib-${ZLIB_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf zlib-${ZLIB_VERSION}.tar.gz && \
  mv zlib-${ZLIB_VERSION} zlib && \
  cd zlib && mkdir -p /zlib && \
  ./configure --prefix=/usr --shared && \
  make DESTDIR=/zlib && \
  make DESTDIR=/zlib install && \
  make install

# Gawk
FROM zlib AS gawk

ENV GAWK_VERSION=5.3.2

COPY --from=sources-downloader /sources/downloads/gawk-${GAWK_VERSION}.tar.xz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf gawk-${GAWK_VERSION}.tar.xz && \
  mv gawk-${GAWK_VERSION} gawk && \
  cd gawk && mkdir -p /gawk && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --disable-nls \
    --disable-pma && \
  make DESTDIR=/gawk && \
  make DESTDIR=/gawk install && \
  make install

# Rsync
FROM gawk AS rsync

ENV RSYNC_VERSION=3.4.1

COPY --from=sources-downloader /sources/downloads/rsync-${RSYNC_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf rsync-${RSYNC_VERSION}.tar.gz && \
  mv rsync-${RSYNC_VERSION} rsync && \
  cd rsync && mkdir -p /rsync && \
  ./configure ${COMMON_ARGS} \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --localstatedir=/var \
    --enable-acl-support \
    --enable-xattr-support \
    --disable-roll-simd \
    --enable-xxhash \
    --with-rrsync \
    --without-included-popt \
    --without-included-zlib \
    --disable-md2man \
    --disable-openssl && \
  make DESTDIR=/rsync && \
  make DESTDIR=/rsync install && \
  make install

# Binutils
FROM stage1 AS binutils

ENV BINUTILS_VERSION=2.44

RUN mkdir /sources && cd /sources && \
  wget https://ftp.gnu.org/gnu/binutils/binutils-${BINUTILS_VERSION}.tar.xz && \
  tar -xvf binutils-${BINUTILS_VERSION}.tar.xz && \
  mv binutils-${BINUTILS_VERSION} binutils && \
  cd binutils && mkdir -p /binutils && \
  ./configure ${COMMON_ARGS} && \
  make DESTDIR=/binutils && \
  make DESTDIR=/binutils install && \
  make install

# Ncurses
FROM stage1 AS ncurses

ENV NCURSES_VERSION=6.5

RUN mkdir /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/ncurses/ncurses-${NCURSES_VERSION}.tar.gz && \
  tar -xvf ncurses-${NCURSES_VERSION}.tar.gz && \
  mv ncurses-${NCURSES_VERSION} ncurses && \
  cd ncurses && mkdir -p /ncurses && \
  sed -i s/mawk// configure && \
  mkdir build && cd build && \
  ../configure ${COMMON_ARGS} && \
  make -C include && make -C progs tic && cd .. && \
  ./configure ${COMMON_ARGS} \
    --mandir=/usr/share/man \
    --with-manpage-format=normal \
    --with-shared \
    --without-debug \
    --without-ada \
    --without-normal \
    --disable-stripping \
    --enable-widec && \
  make -j${JOBS} && \
  make DESTDIR=/ncurses TIC_PATH=/sources/ncurses/build/progs/tic install && \
  make install && \
  echo "INPUT(-lncursesw)" > /ncurses/usr/lib/libncurses.so && \
  cp /ncurses/usr/lib/libncurses.so /usr/lib/libncurses.so

# M4
FROM stage1 AS m4

ENV M4_VERSION=1.4.20

RUN mkdir /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/m4/m4-${M4_VERSION}.tar.xz && \
  tar -xvf m4-${M4_VERSION}.tar.xz && \
  mv m4-${M4_VERSION} m4 && \
  cd m4 && mkdir -p /m4 && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make DESTDIR=/m4 && \
  make DESTDIR=/m4 install && \
  make install

# Readline
FROM stage1 AS readline

ENV READLINE_VERSION=8.3

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/readline/readline-${READLINE_VERSION}.tar.gz && \
  tar -xvf readline-${READLINE_VERSION}.tar.gz && \
  mv readline-${READLINE_VERSION} readline && \
  cd readline && mkdir -p /readline && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make DESTDIR=/readline && \
  make DESTDIR=/readline install && \
  make install

# Bash
FROM readline AS bash

ENV BASH_VERSION=5.3

COPY ./files/bash/bashrc /sources/bashrc
COPY ./files/bash/profile-bashrc.sh /sources/profile-bashrc.sh

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/bash/bash-${BASH_VERSION}.tar.gz && \
  tar -xvf bash-${BASH_VERSION}.tar.gz && \
  mv bash-${BASH_VERSION} bash && \
  cd bash && mkdir -p /bash && \
  ./configure ${COMMON_ARGS} \
    --build=${BUILD} \
    --host=${TARGET} \
    --prefix=/usr \
    --bindir=/bin \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --disable-nls \
    --enable-readline \
    --without-bash-malloc \
    --with-installed-readline && \
  make y.tab.c && make builtins/libbuiltins.a && make && \
  mkdir -p /bash/etc/bash && \
  install -Dm644 /sources/bashrc /bash/etc/bash/bashrc && \
  install -Dm644 /sources/profile-bashrc.sh /bash/etc/profile.d/00-bashrc.sh && \
  make DESTDIR=/bash install && \
  make install

# Perl
FROM m4 AS perl

ENV PERL_VERSION=5.42.0

ENV CFLAGS="-static -Os -ffunction-sections -fdata-sections -Bsymbolic-functions"
ENV LDFLAGS="-Wl,--gc-sections"
ENV PERL_CROSS=1.6.2

RUN cd /sources && \
  wget http://www.cpan.org/src/5.0/perl-${PERL_VERSION}.tar.xz && \
  tar -xvf perl-${PERL_VERSION}.tar.xz && \
  mv perl-${PERL_VERSION} perl && \
  cd perl && \
  ln -s /usr/bin/gcc /usr/bin/cc && \
  ./Configure -des \
    -Dprefix=/usr \
    -Dcccdlflags='-fPIC' \
    -Dccdlflags='-rdynamic' \
    -Dprivlib=/usr/share/perl5/core_perl \
    -Darchlib=/usr/lib/perl5/core_perl \
    -Dvendorprefix=/usr \
    -Dvendorlib=/usr/share/perl5/vendor_perl \
    -Dvendorarch=/usr/lib/perl5/vendor_perl \
    -Dsiteprefix=/usr/local \
    -Dsitelib=/usr/local/share/perl5/site_perl \
    -Dsitearch=/usr/local/lib/perl5/site_perl \
    -Dlocincpth=' ' \
    -Doptimize="-flto=auto -O2" \
    -Duselargefiles \
    -Dusethreads \
    -Duseshrplib \
    -Dd_semctl_semun \
    -Dman1dir=/usr/share/man/man1 \
    -Dman3dir=/usr/share/man/man3 \
    -Dinstallman1dir=/usr/share/man/man1 \
    -Dinstallman3dir=/usr/share/man/man3 \
    -Dman1ext='1' \
    -Dman3ext='3pm' \
    -Dcf_by='mochiilinux' \
    -Dcf_email='mochiilinux@example.com' \
    -Ud_csh \
    -Ud_fpos64_t \
    -Ud_off64_t \
    -Dusenm \
    -Duse64bitint && \
  make libperl.so && \
  make DESTDIR=/perl -j8 && \
  make DESTDIR=/perl install && \
  make install

# OpenSSL
FROM rsync AS openssl

ENV OPENSSL_VERSION=3.5.2

COPY --from=perl /perl /perl
RUN rsync -aHAX --keep-dirlinks /perl/. /

COPY --from=zlib /zlib /zlib
RUN rsync -aHAX --keep-dirlinks /zlib/. /

COPY --from=sources-downloader /sources/downloads/openssl-${OPENSSL_VERSION}.tar.gz /sources/

RUN cd /sources && \
  tar -xvf openssl-${OPENSSL_VERSION}.tar.gz && \
  mv openssl-${OPENSSL_VERSION} openssl && \
  cd openssl && mkdir -p /openssl && \
  ./config --prefix=/usr \
    --openssldir=/etc/ssl \
    --libdir=lib \
    shared \
    zlib-dynamic 2>&1 && \
  make DESTDIR=/openssl 2>&1 && \
  make DESTDIR=/openssl install && \
  make install

# Busybox
FROM openssl AS busybox

COPY --from=busybox-stage0 /sources /sources

ENV BUSYBOX_VERSION=1.37.0

RUN cd /sources && \
  rm -rfv busybox-${BUSYBOX_VERSION} && \
  tar -xvf busybox-${BUSYBOX_VERSION}.tar.bz2 && \
  cd busybox-${BUSYBOX_VERSION} && \
  make -j1 distclean && \
  make defconfig && \
  sed -i 's/\(CONFIG_\)\(.*\)\(INETD\)\(.*\)=y/# \1\2\3\4 is not set/g' .config && \
  sed -i 's/\(CONFIG_IFPLUGD\)=y/# \1 is not set/' .config && \
  sed -i 's/\(CONFIG_FEATURE_WTMP\)=y/# \1 is not set/' .config && \
  sed -i 's/\(CONFIG_FEATURE_UTMP\)=y/# \1 is not set/' .config && \
  sed -i 's/\(CONFIG_UDPSVD\)=y/# \1 is not set/' .config && \
  sed -i 's/\(CONFIG_TCPSVD\)=y/# \1 is not set/' .config && \
  sed -i 's/\(CONFIG_TC\)=y/# \1 is not set/' .config

RUN cd /sources/busybox-${BUSYBOX_VERSION} && \
  make -j1 && \
  make CONFIG_PREFIX="/sysroot" install && \
  make install

# Coreutils
FROM rsync AS coreutils

ENV COREUTILS_VERSION=9.4

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/coreutils/coreutils-${COREUTILS_VERSION}.tar.xz && \
  tar -xvf coreutils-${COREUTILS_VERSION}.tar.xz && \
  mv coreutils-${COREUTILS_VERSION} coreutils && \
  cd coreutils && mkdir -p /coreutils && \
  ./configure ${COMMON_ARGS} \
    --prefix=/usr \
    --bindir=/bin \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --disable-nls \
    --enable-no-install-program=hostname,su,kill,uptime \
    --enable-single-binary=symlinks \
    --enable-single-binary-exceptions=env,fmt,sha512sum \
    --with-openssl \
    --disable-dependency-tracking && \
  make DESTDIR=/coreutils && \
  make DESTDIR=/coreutils install

# Findutils
FROM stage1 AS findutils

ENV FINDUTILS_VERSION=4.10.0

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/findutils/findutils-${FINDUTILS_VERSION}.tar.xz && \
  tar -xvf findutils-${FINDUTILS_VERSION}.tar.xz && \
  mv findutils-${FINDUTILS_VERSION} findutils && \
  cd findutils && mkdir -p /findutils && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make DESTDIR=/findutils && \
  make DESTDIR=/findutils install && \
  make install

# Grep
FROM stage1 AS grep

ENV GREP_VERSION=3.12

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/grep/grep-${GREP_VERSION}.tar.xz && \
  tar -xvf grep-${GREP_VERSION}.tar.xz && \
  mv grep-${GREP_VERSION} grep && \
  cd grep && mkdir -p /grep && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make DESTDIR=/grep && \
  make DESTDIR=/grep install && \
  make install


# CA Certificates
FROM rsync AS ca-certificates

ENV CA_CERTIFICATES_VERSION=20250619

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

COPY --from=perl /perl /perl
RUN rsync -aHAX --keep-dirlinks /perl/. /

COPY --from=bash /bash /bash
RUN rsync -aHAX --keep-dirlinks /bash/. /

COPY --from=readline /readline /readline
RUN rsync -aHAX --keep-dirlinks /readline/. /

COPY --from=acl /acl /acl
RUN rsync -aHAX --keep-dirlinks /acl/. /

COPY --from=attr /attr /attr
RUN rsync -aHAX --keep-dirlinks /attr/. /

COPY --from=findutils /findutils /findutils
RUN rsync -aHAX --keep-dirlinks /findutils/. /

COPY --from=sources-downloader /sources/downloads/ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 && \
  mv ca-certificates-${CA_CERTIFICATES_VERSION} ca-certificates && \
  cd ca-certificates && mkdir -p /ca-certificates && \
  CC=gcc make && \
  make DESTDIR=/ca-certificates install

COPY ./files/ca-certificates/post_install.sh /sources/post_install.sh
RUN bash /sources/post_install.sh


# SQLite3
FROM rsync AS sqlite3

ENV SQLITE3_VERSION=3500400

ENV CFLAGS="${CFLAGS//-Os/-O2} -DSQLITE_ENABLE_FTS3_PARENTHESIS -DSQLITE_ENABLE_COLUMN_METADATA -DSQLITE_SECURE_DELETE -DSQLITE_ENABLE_UNLOCK_NOTIFY -DSQLITE_ENABLE_RTREE -DSQLITE_ENABLE_GEOPOLY -DSQLITE_USE_URI -DSQLITE_ENABLE_DBSTAT_VTAB -DSQLITE_SOUNDEX -DSQLITE_MAX_VARIABLE_NUMBER=250000"

COPY --from=sources-downloader /sources/downloads/sqlite-autoconf-${SQLITE3_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf sqlite-autoconf-${SQLITE3_VERSION}.tar.gz && \
  mv sqlite-autoconf-${SQLITE3_VERSION} sqlite3 && \
  cd sqlite3 && mkdir -p /sqlite3 && \
  ./configure \
    --prefix=/usr \
    --enable-threadsafe \
    --enable-session \
    --enable-static \
    --enable-fts3 \
    --enable-fts4 \
    --enable-fts5 \
    --soname=legacy && \
  make && \
  make DESTDIR=/sqlite3 install && \
  make install


# Util-linux
FROM bash AS util-linux

ENV UTIL_LINUX_VERSION=2.41.1

COPY --from=sources-downloader /sources/downloads/util-linux-${UTIL_LINUX_VERSION}.tar.xz /sources/

RUN rm /bin/sh && ln -s /bin/bash /bin/sh && \
  mkdir -p /sources && cd /sources && \
  tar -xvf util-linux-${UTIL_LINUX_VERSION}.tar.xz && \
  mv util-linux-${UTIL_LINUX_VERSION} util-linux && \
  cd util-linux && mkdir -p /util-linux && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --libdir=/usr/lib \
    --disable-silent-rules \
    --enable-newgrp \
    --disable-uuidd \
    --disable-liblastlog2 \
    --disable-nls \
    --disable-kill \
    --disable-chfn-chsh \
    --with-vendordir=/usr/lib \
    --enable-fs-paths-extra=/usr/sbin && \
  make DESTDIR=/util-linux && \
  make DESTDIR=/util-linux install && \
  make install


# Curl
FROM rsync AS curl

COPY --from=ca-certificates /ca-certificates /ca-certificates
RUN rsync -aHAX --keep-dirlinks /ca-certificates/. /

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

COPY --from=zstd /zstd /zstd
RUN rsync -aHAX --keep-dirlinks /zstd/. /

ENV CURL_VERSION=8.5.0

COPY --from=sources-downloader /sources/downloads/curl-${CURL_VERSION}.tar.gz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf curl-${CURL_VERSION}.tar.gz && \
  mv curl-${CURL_VERSION} curl && \
  cd curl && mkdir -p /curl && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --enable-ipv6 \
    --enable-unix-sockets \
    --enable-static \
    --without-libidn \
    --without-libidn2 \
    --with-ca-bundle=/etc/ssl/certs/ca-certificates.crt \
    --with-ca-path=/etc/ssl/certs \
    --with-zsh-functions-dir \
    --with-fish-functions-dir \
    --disable-ldap \
    --with-pic \
    --enable-websockets \
    --without-libssh2 \
    --with-ssl \
    --with-nghttp2 \
    --without-libssh2 && \
  make DESTDIR=/curl && \
  make DESTDIR=/curl install && \
  make install


# Python
FROM rsync AS python

ENV PYTHON_VERSION=3.12.11

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

COPY --from=bash /bash /bash
RUN rsync -aHAX --keep-dirlinks /bash/. /

COPY --from=zlib /zlib /zlib
RUN rsync -aHAX --keep-dirlinks /zlib/. /

COPY --from=readline /readline /readline
RUN rsync -aHAX --keep-dirlinks /readline/. /

COPY --from=sources-downloader /sources/downloads/Python-${PYTHON_VERSION}.tar.xz /sources/

RUN rm /bin/sh && ln -s /bin/bash /bin/sh && \
  mkdir -p /sources && cd /sources && \
  tar -xvf Python-${PYTHON_VERSION}.tar.xz && \
  mv Python-${PYTHON_VERSION} python && \
  cd python && mkdir -p /python && \
  ./configure --disable-dependency-tracking --prefix=/usr \
    --enable-ipv6 \
    --enable-loadable-sqlite-extensions \
    --enable-optimizations \
    --enable-shared \
    --with-ensurepip=install \
    --with-lto \
    --with-computed-gotos \
    --with-dbmliborder=gdbm:ndbm 2>&1 && \
  make DESTDIR=/python 2>&1 && \
  make DESTDIR=/python install 2>&1 && \
  make install 2>&1


# Libcap
FROM bash AS libcap

ENV LIBCAP_VERSION=2.76

COPY --from=sources-downloader /sources/downloads/libcap-${LIBCAP_VERSION}.tar.xz /sources/

RUN mkdir -p /sources && cd /sources && \
  tar -xvf libcap-${LIBCAP_VERSION}.tar.xz && \
  mv libcap-${LIBCAP_VERSION} libcap && \
  cd libcap && mkdir -p /libcap && \
  make BUILD_CC=gcc CC="${CC:-gcc}" && \
  make DESTDIR=/libcap PAM_LIBDIR=/lib prefix=/usr SBINDIR=/sbin lib=lib RAISE_SETFCAP=no GOLANG=no install && \
  make GOLANG=no PAM_LIBDIR=/lib lib=lib prefix=/usr SBINDIR=/sbin RAISE_SETFCAP=no install


# Gperf
FROM stage1 AS gperf

ENV GPERF_VERSION=3.3

RUN mkdir -p /sources && cd /sources && \
  wget http://ftp.gnu.org/gnu/gperf/gperf-${GPERF_VERSION}.tar.gz && \
  tar -xvf gperf-${GPERF_VERSION}.tar.gz && \
  mv gperf-${GPERF_VERSION} gperf && \
  cd gperf && mkdir -p /gperf && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking --prefix=/usr && \
  make BUILD_CC=gcc CC="${CC:-gcc}" lib=lib prefix=/usr GOLANG=no DESTDIR=/gperf && \
  make DESTDIR=/gperf install && \
  make install


# Systemd
FROM rsync AS systemd

ENV SYSTEMD_VERSION=257.6

COPY --from=gperf /gperf /gperf
RUN rsync -aHAX --keep-dirlinks /gperf/. /

COPY --from=libcap /libcap /libcap
RUN rsync -aHAX --keep-dirlinks /libcap/. /

COPY --from=util-linux /util-linux /util-linux
RUN rsync -aHAX --keep-dirlinks /util-linux/. /

COPY --from=python /python /python
RUN rsync -aHAX --keep-dirlinks /python/. /

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

COPY --from=bash /bash /bash
RUN rsync -aHAX --keep-dirlinks /bash/. /

COPY --from=coreutils /coreutils /coreutils
RUN rsync -aHAX --keep-dirlinks /coreutils/. /

COPY --from=readline /readline /readline
RUN rsync -aHAX --keep-dirlinks /readline/. /

COPY --from=pkgconfig /pkgconfig /pkgconfig
RUN rsync -aHAX --keep-dirlinks /pkgconfig/. /

COPY --from=sources-downloader /sources/downloads/systemd /sources/downloads/systemd

ENV CFLAGS="-D __UAPI_DEF_ETHHDR=0 -D _LARGEFILE64_SOURCE"

RUN rm -fv /bin/sh && ln -s /bin/bash /bin/sh && \
  mkdir -p /sources && cd /sources/downloads/systemd && \
  mkdir -p /systemd && \
  python3 -m pip install meson ninja jinja2 && \
  mkdir -p build && cd build && \
  /usr/bin/meson setup .. \
    --prefix=/usr \
    --buildtype=release \
    -D default-dnssec=no \
    -D firstboot=false \
    -D install-tests=false \
    -D ldconfig=false \
    -D sysusers=false \
    -D rpmmacrosdir=no \
    -D gshadow=false \
    -D idn=false \
    -D localed=false \
    -D nss-myhostname=false \
    -D nss-systemd=false \
    -D userdb=false \
    -D nss-mymachines=disabled \
    -D nss-resolve=disabled \
    -D utmp=false \
    -D homed=disabled \
    -D man=disabled \
    -D mode=release \
    -D pamconfdir=no \
    -D dev-kvm-mode=0660 \
    -D nobody-group=nogroup \
    -D sysupdate=disabled \
    -D ukify=disabled \
    -D docdir=/usr/share/doc/systemd-${SYSTEMD_VERSION} 2>&1 && \
  ninja 2>&1 && \
  DESTDIR=/systemd ninja install && \
  ninja install


# =============================================================================
# Stage 2: Final Image Build
# =============================================================================
# Assembles all built packages into the final root filesystem image.

FROM stage0 AS stage2-merge

RUN apk add rsync

COPY --from=skeleton /sysroot /skeleton

# Musl
COPY --from=musl /sysroot /musl
RUN rsync -aHAX --keep-dirlinks /musl/. /skeleton/

# Coreutils
COPY --from=coreutils /coreutils /coreutils
RUN rsync -aHAX --keep-dirlinks /coreutils/. /skeleton/

# Curl
COPY --from=curl /curl /curl
RUN rsync -aHAX --keep-dirlinks /curl/. /skeleton/

# OpenSSL
COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /skeleton/

# CA Certificates
COPY --from=ca-certificates /ca-certificates /ca-certificates
RUN rsync -aHAX --keep-dirlinks /ca-certificates/. /skeleton/

# Bash
COPY --from=bash /bash /bash
RUN rsync -aHAX --keep-dirlinks /bash/. /skeleton/

# Readline
COPY --from=readline /readline /readline
RUN rsync -aHAX --keep-dirlinks /readline/. /skeleton/

# ACL
COPY --from=acl /acl /acl
RUN rsync -aHAX --keep-dirlinks /acl/. /skeleton/

# Attr
COPY --from=attr /attr /attr
RUN rsync -aHAX --keep-dirlinks /attr/. /skeleton/

# Findutils
COPY --from=findutils /findutils /findutils
RUN rsync -aHAX --keep-dirlinks /findutils/. /skeleton/

# Grep
COPY --from=grep /grep /grep
RUN rsync -aHAX --keep-dirlinks /grep/. /skeleton/

# Zstd
COPY --from=zstd /zstd /zstd
RUN rsync -aHAX --keep-dirlinks /zstd/. /skeleton/

# Zlib
COPY --from=zlib /zlib /zlib
RUN rsync -aHAX --keep-dirlinks /zlib/. /skeleton/

# Libcap
COPY --from=libcap /libcap /libcap
RUN rsync -aHAX --keep-dirlinks /libcap/. /skeleton/

# Util-linux
COPY --from=util-linux /util-linux /util-linux
RUN rsync -aHAX --keep-dirlinks /util-linux/. /skeleton/

# Systemd
COPY --from=systemd /systemd /systemd
RUN rsync -aHAX --keep-dirlinks /systemd/. /skeleton/

# Cleanup - remove headers
RUN rm -rf /skeleton/usr/include


# =============================================================================
# Final Image: The complete mochiilinux distribution
# =============================================================================
FROM scratch AS stage2

COPY --from=stage2-merge /skeleton /


# =============================================================================
# Test Stage 2: Verify the final image works
# =============================================================================
FROM stage2 AS test2

SHELL ["/bin/bash", "-c"]

RUN ls -liah /
RUN curl --version
