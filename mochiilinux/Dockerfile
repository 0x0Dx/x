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
FROM stage1 AS autoconf

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
FROM stage1 AS automake

ENV AUTOMAKE_VERSION=1.16.5

RUN mkdir -p /sources && cd /sources && wget http://ftp.gnu.org/gnu/automake/automake-${AUTOMAKE_VERSION}.tar.xz && \
  tar -xvf automake-${AUTOMAKE_VERSION}.tar.xz && \
  cd automake-${AUTOMAKE_VERSION} && \
  mkdir -p /autoconf && \
  ./configure ${COMMON_ARGS} && \
  make DESTDIR=/autoconf && \
  make DESTDIR=/autoconf install && \
  make install
