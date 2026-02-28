# =============================================================================
# Mochiilinux - A minimal Linux distribution built from scratch
# =============================================================================
# This Dockerfile builds a complete Linux distribution using multi-stage
# Docker builds. It creates a cross-compiler toolchain and builds all
# necessary packages from source.
#
# Build targets:
#   stage0        - Base image with cross-compiler toolchain
#   sources       - Download all source tarballs
#   skeleton      - Base root filesystem
#   stage0-*     - Cross-compiled packages (busybox, musl, gcc, make, binutils)
#   stage1-merge - Merge stage0 packages into single rootfs
#   stage1       - Minimal image with build tools
#   stage1-*     - Native packages for final image
#   stage2-merge - Merge all packages into final rootfs
#   stage2       - Final minimal image
#   test1/test2  - Test stages
# =============================================================================

# =============================================================================
# Stage 0: Base image with cross-compiler toolchain
# =============================================================================
FROM alpine AS stage0

# Build arguments
ARG VENDOR="mochii"
ARG ARCH="x86-64"
ARG BUILD_ARCH="x86_64"
ARG JOBS=16

# Environment variables
ENV VENDOR=${VENDOR}
ENV ARCH=${ARCH}
ENV BUILD_ARCH=${BUILD_ARCH}
ENV JOBS=${JOBS}
ENV TARGET=${BUILD_ARCH}-${VENDOR}-linux-musl
ENV BUILD=${BUILD_ARCH}-pc-linux-musl
ENV COMMON_ARGS="--prefix=/usr --host=${TARGET} --build=${BUILD}"
ENV LC_ALL=POSIX

RUN apk update && apk add \
  git \
  bash \
  wget \
  perl \
  build-base \
  make \
  patch \
  busybox-static \
  curl

RUN git clone https://github.com/firasuke/mussel.git /mussel && \
  cd /mussel && \
  git checkout 95dec40aee2077aa703b7abc7372ba4d34abb889 -b build && \
  ./mussel ${ARCH} -k -l -o -p -s -T ${VENDOR}

ENV PATH=/mussel/toolchain/bin:$PATH


# =============================================================================
# Stage: Source downloads
# =============================================================================
FROM stage0 AS sources

RUN mkdir -p /sources/{downloads,build}

ENV CURL_VERSION=8.5.0
ENV RSYNC_VERSION=3.4.1
ENV XXHASH_VERSION=0.8.3
ENV ZSTD_VERSION=1.5.7
ENV LZ4_VERSION=1.10.0
ENV ZLIB_VERSION=1.3.1
ENV ACL_VERSION=2.3.2
ENV ATTR_VERSION=2.5.2
ENV GAWK_VERSION=5.3.2
ENV CA_CERTIFICATES_VERSION=20250619
ENV SYSTEMD_VERSION=257.6
ENV LIBCAP_VERSION=2.76
ENV UTIL_LINUX_VERSION=2.41.1
ENV UTIL_LINUX_VERSION_MAJOR=2.41
ENV PYTHON_VERSION=3.12.11
ENV SQLITE3_VERSION=3500400
ENV OPENSSL_VERSION=3.5.2
ENV PKGCONFIG_VERSION=1.8.1
ENV OE_CORE_VERSION=bed605e7cd095f210e24f1e07dab6f814d0bbb59

WORKDIR /sources/downloads

RUN apk add git patch && \
  wget -q https://curl.se/download/curl-${CURL_VERSION}.tar.gz & \
  wget -q https://download.samba.org/pub/rsync/rsync-${RSYNC_VERSION}.tar.gz & \
  wget -q https://github.com/Cyan4973/xxHash/archive/refs/tags/v${XXHASH_VERSION}.tar.gz -O xxHash-${XXHASH_VERSION}.tar.gz & \
  wget -q https://github.com/facebook/zstd/archive/v${ZSTD_VERSION}.tar.gz -O zstd-${ZSTD_VERSION}.tar.gz & \
  wget -q https://github.com/lz4/lz4/archive/v${LZ4_VERSION}.tar.gz -O lz4-${LZ4_VERSION}.tar.gz & \
  wget -q https://zlib.net/fossils/zlib-${ZLIB_VERSION}.tar.gz & \
  wget -q https://download.savannah.gnu.org/releases/acl/acl-${ACL_VERSION}.tar.gz & \
  wget -q https://download.savannah.nongnu.org/releases/attr/attr-${ATTR_VERSION}.tar.gz & \
  wget -q https://ftp.gnu.org/gnu/gawk/gawk-${GAWK_VERSION}.tar.xz & \
  wget -q https://gitlab.alpinelinux.org/alpine/ca-certificates/-/archive/${CA_CERTIFICATES_VERSION}/ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 & \
  wget -q https://github.com/systemd/systemd/archive/refs/tags/v${SYSTEMD_VERSION}.tar.gz -O systemd-${SYSTEMD_VERSION}.tar.gz & \
  wait

RUN \
  wget -q https://kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-${LIBCAP_VERSION}.tar.xz & \
  wget -q https://kernel.org/pub/linux/utils/util-linux/v${UTIL_LINUX_VERSION_MAJOR}/util-linux-${UTIL_LINUX_VERSION}.tar.xz & \
  wget -q https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tar.xz & \
  wget -q https://www.sqlite.org/2025/sqlite-autoconf-${SQLITE3_VERSION}.tar.gz & \
  wget -q https://www.openssl.org/source/openssl-${OPENSSL_VERSION}.tar.gz & \
  wget -q https://distfiles.dereferenced.org/pkgconf/pkgconf-${PKGCONFIG_VERSION}.tar.xz & \
  wait

RUN tar -xzf systemd-${SYSTEMD_VERSION}.tar.gz && \
  mv systemd-${SYSTEMD_VERSION} systemd && \
  git clone https://github.com/openembedded/openembedded-core /oe-core && \
  cd /oe-core && \
  git checkout ${OE_CORE_VERSION} && \
  cp /mochiilinux/patches/apply_all.sh /apply_all.sh && \
  chmod +x /apply_all.sh && \
  /apply_all.sh /oe-core/meta/recipes-core/systemd/systemd/ /sources/downloads/systemd


# =============================================================================
# Stage: Skeleton root filesystem
# =============================================================================
FROM stage0 AS skeleton

COPY mochiilinux/setup_rootfs.sh /setup_rootfs.sh
RUN chmod +x /setup_rootfs.sh && SYSROOT=/sysroot /setup_rootfs.sh


# =============================================================================
# Stage 0: Cross-compiled packages
# =============================================================================
FROM stage0 AS busybox-stage0

ENV BUSYBOX_VERSION=1.37.0

RUN wget -q https://busybox.net/downloads/busybox-${BUSYBOX_VERSION}.tar.bz2 && \
  tar -xjf busybox-${BUSYBOX_VERSION}.tar.bz2

RUN cd busybox-${BUSYBOX_VERSION} && \
  make distclean && \
  make ARCH="${ARCH}" defconfig && \
  sed -i \
    -e 's/\(CONFIG_\)\(.*\)\(INETD\)\(.*\)=y/# \1\2\3\4 is not set/g' \
    -e 's/\(CONFIG_IFPLUGD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_FEATURE_WTMP\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_FEATURE_UTMP\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_UDPSVD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_TCPSVD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_TC\)=y/# \1 is not set/g' \
    .config && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" && \
  make ARCH="${ARCH}" CROSS_COMPILE="${TARGET}-" CONFIG_PREFIX="/sysroot" install


FROM stage0 AS musl-stage0

ENV MUSL_VERSION=1.2.5

RUN wget -q https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz && \
  tar -xzf musl-${MUSL_VERSION}.tar.gz

RUN cd musl-${MUSL_VERSION} && \
  ./configure CROSS_COMPILE=${TARGET}- --prefix=/usr --disable-static --target=${TARGET} && \
  make -j${JOBS} && \
  DESTDIR=/sysroot make -j${JOBS} install


FROM stage0 AS gcc-stage0

ENV GCC_VERSION=14.3.0
ENV GMP_VERSION=6.3.0
ENV MPC_VERSION=1.3.1
ENV MPFR_VERSION=4.2.2

RUN mkdir -p /sources && cd /sources && \
  wget -q https://ftp.gnu.org/gnu/gcc/gcc-${GCC_VERSION}/gcc-${GCC_VERSION}.tar.xz && \
  wget -q https://ftp.gnu.org/gnu/gmp/gmp-${GMP_VERSION}.tar.bz2 && \
  wget -q https://ftp.gnu.org/gnu/mpc/mpc-${MPC_VERSION}.tar.gz && \
  wget -q https://ftp.gnu.org/gnu/mpfr/mpfr-${MPFR_VERSION}.tar.bz2 && \
  tar -xf gcc-${GCC_VERSION}.tar.xz && \
  tar -xf gmp-${GMP_VERSION}.tar.bz2 && \
  tar -xf mpc-${MPC_VERSION}.tar.gz && \
  tar -xf mpfr-${MPFR_VERSION}.tar.bz2 && \
  mv gmp-${GMP_VERSION} gcc-${GCC_VERSION}/gmp && \
  mv mpc-${MPC_VERSION} gcc-${GCC_VERSION}/mpc && \
  mv mpfr-${MPFR_VERSION} gcc-${GCC_VERSION}/mpfr

RUN mkdir -p /sysroot/usr/include && \
  cd gcc-${GCC_VERSION} && mkdir build && cd build && \
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


FROM stage0 AS make-stage0

ENV MAKE_VERSION=4.4.1

RUN wget -q https://ftp.gnu.org/gnu/make/make-${MAKE_VERSION}.tar.gz && \
  tar -xzf make-${MAKE_VERSION}.tar.gz

RUN cd make-${MAKE_VERSION} && \
  ./configure --prefix=/usr --build=${BUILD_ARCH} --host=${TARGET} && \
  make -j${JOBS} && \
  make -j${JOBS} DESTDIR=/sysroot install


FROM stage0 AS binutils-stage0

ENV BINUTILS_VERSION=2.44

RUN wget -q https://ftp.gnu.org/gnu/binutils/binutils-${BINUTILS_VERSION}.tar.xz && \
  tar -xf binutils-${BINUTILS_VERSION}.tar.xz

RUN cd binutils-${BINUTILS_VERSION} && \
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


FROM alpine AS alpine-headers
RUN apk add linux-headers


# =============================================================================
# Stage 1 Merge: Combine cross-compiled packages
# =============================================================================
FROM stage0 AS stage1-merge

RUN apk add rsync

COPY --from=skeleton /sysroot /skeleton
COPY --from=gcc-stage0 /sysroot /gcc
COPY --from=musl-stage0 /sysroot /musl
COPY --from=busybox-stage0 /sysroot /busybox
COPY --from=make-stage0 /sysroot /make
COPY --from=binutils-stage0 /sysroot /binutils

RUN rsync -aHAX --keep-dirlinks /gcc/. /skeleton && \
  rsync -aHAX --keep-dirlinks /musl/. /skeleton && \
  rsync -aHAX --keep-dirlinks /busybox/. /skeleton && \
  rsync -aHAX --keep-dirlinks /make/. /skeleton && \
  rsync -aHAX --keep-dirlinks /binutils/. /skeleton

COPY --from=alpine-headers /usr/include/linux /linux
COPY --from=alpine-headers /usr/include/asm /asm
COPY --from=alpine-headers /usr/include/asm-generic /asm-generic
COPY --from=alpine-headers /usr/include/mtd /mtd

RUN mkdir -p /skeleton/usr/include/{linux,asm,asm-generic,mtd} && \
  rsync -aHAX --keep-dirlinks /linux/. /skeleton/usr/include/linux && \
  rsync -aHAX --keep-dirlinks /asm/. /skeleton/usr/include/asm && \
  rsync -aHAX --keep-dirlinks /asm-generic/. /skeleton/usr/include/asm-generic && \
  rsync -aHAX --keep-dirlinks /mtd/. /skeleton/usr/include/mtd


# =============================================================================
# Stage 1: Minimal image with build tools
# =============================================================================
FROM scratch AS stage1
COPY --from=stage1-merge /skeleton /


# =============================================================================
# Build dependencies
# =============================================================================
FROM alpine AS build-deps
RUN apk add m4


# =============================================================================
# Test Stage 1
# =============================================================================
FROM stage1 AS test1

COPY mochiilinux/tests/gcc/test.c /test.c

RUN gcc --version && make --version && \
  gcc -Wall /test.c -o /test && /test


# =============================================================================
# Stage 1.x: Native packages
# =============================================================================
FROM stage1 AS musl

ENV MUSL_VERSION=1.2.5

RUN wget -q https://musl.libc.org/releases/musl-${MUSL_VERSION}.tar.gz && \
  tar -xzf musl-${MUSL_VERSION}.tar.gz

RUN cd musl-${MUSL_VERSION} && \
  ./configure --prefix=/usr --disable-static && \
  make -j${JOBS} && \
  DESTDIR=/sysroot make -j${JOBS} install


FROM stage1 AS pkgconfig

ENV PKGCONFIG_VERSION=1.8.1

COPY --from=sources /sources/downloads/pkgconf-${PKGCONFIG_VERSION}.tar.xz /sources/

RUN tar -xf /sources/pkgconf-${PKGCONFIG_VERSION}.tar.xz && \
  cd pkgconf-${PKGCONFIG_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --localstatedir=/var && \
  make && \
  make DESTDIR=/pkgconfig install && \
  ln -s pkgconf /pkgconfig/usr/bin/pkg-config


FROM build-deps AS autoconf

COPY --from=stage1-merge /skeleton /

ENV AUTOCONF_VERSION=2.71

RUN wget -q http://ftp.gnu.org/gnu/autoconf/autoconf-${AUTOCONF_VERSION}.tar.xz && \
  tar -xf autoconf-${AUTOCONF_VERSION}.tar.xz

RUN cd autoconf-${AUTOCONF_VERSION} && \
  ./configure ${COMMON_ARGS} && \
  make && \
  make DESTDIR=/autoconf install


FROM build-deps AS automake

COPY --from=stage1-merge /skeleton /

ENV AUTOMAKE_VERSION=1.16.5

RUN wget -q http://ftp.gnu.org/gnu/automake/automake-${AUTOMAKE_VERSION}.tar.xz && \
  tar -xf automake-${AUTOMAKE_VERSION}.tar.xz

RUN cd automake-${AUTOMAKE_VERSION} && \
  ./configure ${COMMON_ARGS} && \
  make && \
  make DESTDIR=/automake install


FROM stage1 AS xxhash

ENV XXHASH_VERSION=0.8.3

COPY --from=sources /sources/downloads/xxHash-${XXHASH_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/xxHash-${XXHASH_VERSION}.tar.gz && \
  cd xxHash-${XXHASH_VERSION} && \
  make CC=gcc DESTDIR=/xxhash && \
  make DESTDIR=/xxhash install


FROM xxhash AS zstd

ENV ZSTD_VERSION=1.5.7

COPY --from=sources /sources/downloads/zstd-${ZSTD_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/zstd-${ZSTD_VERSION}.tar.gz && \
  cd zstd-${ZSTD_VERSION} && \
  make CC=gcc DESTDIR=/zstd && \
  make DESTDIR=/zstd install


FROM zstd AS lz4

ENV LZ4_VERSION=1.10.0

COPY --from=sources /sources/downloads/lz4-${LZ4_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/lz4-${LZ4_VERSION}.tar.gz && \
  cd lz4-${LZ4_VERSION} && \
  make CC=gcc PREFIX=/usr DESTDIR=/lz4 && \
  make DESTDIR=/lz4 install


FROM lz4 AS attr

ENV ATTR_VERSION=2.5.2

COPY --from=sources /sources/downloads/attr-${ATTR_VERSION}.tar.gz /sources/
COPY mochiilinux/patches/attr/basename.patch /sources/

RUN tar -xzf /sources/attr-${ATTR_VERSION}.tar.gz && \
  cd attr-${ATTR_VERSION} && \
  patch -p1 < /sources/basename.patch && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --localstatedir=/var \
    --disable-nls && \
  make && \
  make DESTDIR=/attr install


FROM attr AS acl

ENV ACL_VERSION=2.3.2

COPY --from=sources /sources/downloads/acl-${ACL_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/acl-${ACL_VERSION}.tar.gz && \
  cd acl-${ACL_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --prefix=/usr \
    --disable-dependency-tracking \
    --libexecdir=/usr/libexec && \
  make && \
  make DESTDIR=/acl install


FROM acl AS popt

ENV POPT_VERSION=1.19

RUN wget -q http://ftp.rpm.org/popt/releases/popt-1.x/popt-${POPT_VERSION}.tar.gz && \
  tar -xzf popt-${POPT_VERSION}.tar.gz

RUN cd popt-${POPT_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking --prefix=/usr && \
  make && \
  make DESTDIR=/popt install


FROM popt AS zlib

ENV ZLIB_VERSION=1.3.1

COPY --from=sources /sources/downloads/zlib-${ZLIB_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/zlib-${ZLIB_VERSION}.tar.gz && \
  cd zlib-${ZLIB_VERSION} && \
  ./configure --prefix=/usr --shared && \
  make && \
  make DESTDIR=/zlib install


FROM zlib AS gawk

ENV GAWK_VERSION=5.3.2

COPY --from=sources /sources/downloads/gawk-${GAWK_VERSION}.tar.xz /sources/

RUN tar -xf /sources/gawk-${GAWK_VERSION}.tar.xz && \
  cd gawk-${GAWK_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --sysconfdir=/etc \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --disable-nls && \
  make && \
  make DESTDIR=/gawk install


FROM gawk AS rsync

ENV RSYNC_VERSION=3.4.1

COPY --from=sources /sources/downloads/rsync-${RSYNC_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/rsync-${RSYNC_VERSION}.tar.gz && \
  cd rsync-${RSYNC_VERSION} && \
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
  make && \
  make DESTDIR=/rsync install


FROM stage1 AS binutils

ENV BINUTILS_VERSION=2.44

RUN wget -q https://ftp.gnu.org/gnu/binutils/binutils-${BINUTILS_VERSION}.tar.xz && \
  tar -xf binutils-${BINUTILS_VERSION}.tar.xz

RUN cd binutils-${BINUTILS_VERSION} && \
  ./configure ${COMMON_ARGS} && \
  make && \
  make DESTDIR=/binutils install


FROM stage1 AS ncurses

ENV NCURSES_VERSION=6.5

RUN wget -q http://ftp.gnu.org/gnu/ncurses/ncurses-${NCURSES_VERSION}.tar.gz && \
  tar -xzf ncurses-${NCURSES_VERSION}.tar.gz

RUN cd ncurses-${NCURSES_VERSION} && \
  sed -i 's/mawk//' configure && \
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
  echo "INPUT(-lncursesw)" > /ncurses/usr/lib/libncurses.so


FROM stage1 AS m4

ENV M4_VERSION=1.4.20

RUN wget -q http://ftp.gnu.org/gnu/m4/m4-${M4_VERSION}.tar.xz && \
  tar -xf m4-${M4_VERSION}.tar.xz

RUN cd m4-${M4_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make && \
  make DESTDIR=/m4 install


FROM stage1 AS readline

ENV READLINE_VERSION=8.3

RUN wget -q http://ftp.gnu.org/gnu/readline/readline-${READLINE_VERSION}.tar.gz && \
  tar -xzf readline-${READLINE_VERSION}.tar.gz

RUN cd readline-${READLINE_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make && \
  make DESTDIR=/readline install


FROM readline AS bash

ENV BASH_VERSION=5.3

COPY mochiilinux/files/bash/bashrc /bashrc
COPY mochiilinux/files/bash/profile-bashrc.sh /profile-bashrc.sh

RUN wget -q http://ftp.gnu.org/gnu/bash/bash-${BASH_VERSION}.tar.gz && \
  tar -xzf bash-${BASH_VERSION}.tar.gz

RUN cd bash-${BASH_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --prefix=/usr \
    --bindir=/bin \
    --mandir=/usr/share/man \
    --infodir=/usr/share/info \
    --disable-nls \
    --enable-readline \
    --without-bash-malloc \
    --with-installed-readline && \
  make && \
  mkdir -p /bash/etc/bash && \
  install -Dm644 /bashrc /bash/etc/bash/bashrc && \
  install -Dm644 /profile-bashrc.sh /bash/etc/profile.d/00-bashrc.sh && \
  make DESTDIR=/bash install


FROM m4 AS perl

ARG BUILD_EMAIL="builder@localhost"

ENV PERL_VERSION=5.42.0
ENV CFLAGS="-static -Os -ffunction-sections -fdata-sections -Bsymbolic-functions"
ENV LDFLAGS="-Wl,--gc-sections"

RUN wget -q http://www.cpan.org/src/5.0/perl-${PERL_VERSION}.tar.xz && \
  tar -xf perl-${PERL_VERSION}.tar.xz

RUN cd perl-${PERL_VERSION} && \
  ln -sf /usr/bin/gcc /usr/bin/cc && \
  ./Configure -des \
    -Dprefix=/usr \
    -Dprivlib=/usr/share/perl5/core_perl \
    -Darchlib=/usr/lib/perl5/core_perl \
    -Dvendorprefix=/usr \
    -Dvendorlib=/usr/share/perl5/vendor_perl \
    -Dvendorarch=/usr/lib/perl5/vendor_perl \
    -Dsiteprefix=/usr/local \
    -Dsitelib=/usr/local/share/perl5/site_perl \
    -Dsitearch=/usr/local/lib/perl5/site_perl \
    -Doptimize="-flto=auto -O2" \
    -Dusethreads \
    -Duseshrplib \
    -Dman1dir=/usr/share/man/man1 \
    -Dman3dir=/usr/share/man/man3 \
    -Dcf_by=mochiilinux \
    -Dcf_email=${BUILD_EMAIL} \
    -Ud_csh \
    -Ud_fpos64_t \
    -Ud_off64_t \
    -Dusenm \
    -Duse64bitint && \
  make libperl.so && \
  make DESTDIR=/perl -j8 && \
  make DESTDIR=/perl install


FROM rsync AS openssl

ENV OPENSSL_VERSION=3.5.2

COPY --from=perl /perl /perl
COPY --from=zlib /zlib /zlib

RUN rsync -aHAX --keep-dirlinks /perl/. / && \
  rsync -aHAX --keep-dirlinks /zlib/. /

COPY --from=sources /sources/downloads/openssl-${OPENSSL_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/openssl-${OPENSSL_VERSION}.tar.gz && \
  cd openssl-${OPENSSL_VERSION} && \
  ./config --prefix=/usr --openssldir=/etc/ssl --libdir=lib shared zlib-dynamic && \
  make && \
  make DESTDIR=/openssl install


FROM openssl AS busybox

COPY --from=busybox-stage0 /sources /sources

ENV BUSYBOX_VERSION=1.37.0

RUN tar -xjf /sources/busybox-${BUSYBOX_VERSION}.tar.bz2 && \
  cd busybox-${BUSYBOX_VERSION} && \
  make distclean && \
  make defconfig && \
  sed -i \
    -e 's/\(CONFIG_\)\(.*\)\(INETD\)\(.*\)=y/# \1\2\3\4 is not set/g' \
    -e 's/\(CONFIG_IFPLUGD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_FEATURE_WTMP\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_FEATURE_UTMP\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_UDPSVD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_TCPSVD\)=y/# \1 is not set/g' \
    -e 's/\(CONFIG_TC\)=y/# \1 is not set/g' \
    .config && \
  make && \
  make CONFIG_PREFIX=/sysroot install


FROM rsync AS coreutils

ENV COREUTILS_VERSION=9.4

COPY --from=openssl /openssl /openssl
RUN rsync -aHAX --keep-dirlinks /openssl/. /

RUN wget -q http://ftp.gnu.org/gnu/coreutils/coreutils-${COREUTILS_VERSION}.tar.xz && \
  tar -xf coreutils-${COREUTILS_VERSION}.tar.xz

RUN cd coreutils-${COREUTILS_VERSION} && \
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
    --with-openssl && \
  make && \
  make DESTDIR=/coreutils install


FROM stage1 AS findutils

ENV FINDUTILS_VERSION=4.10.0

RUN wget -q http://ftp.gnu.org/gnu/findutils/findutils-${FINDUTILS_VERSION}.tar.xz && \
  tar -xf findutils-${FINDUTILS_VERSION}.tar.xz

RUN cd findutils-${FINDUTILS_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make && \
  make DESTDIR=/findutils install


FROM stage1 AS grep

ENV GREP_VERSION=3.12

RUN wget -q http://ftp.gnu.org/gnu/grep/grep-${GREP_VERSION}.tar.xz && \
  tar -xf grep-${GREP_VERSION}.tar.xz

RUN cd grep-${GREP_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking && \
  make && \
  make DESTDIR=/grep install


FROM rsync AS ca-certificates

ENV CA_CERTIFICATES_VERSION=20250619

COPY --from=openssl /openssl /openssl
COPY --from=perl /perl /perl
COPY --from=bash /bash /bash
COPY --from=readline /readline /readline
COPY --from=acl /acl /acl
COPY --from=attr /attr /attr
COPY --from=findutils /findutils /findutils

RUN rsync -aHAX --keep-dirlinks /openssl/. / && \
  rsync -aHAX --keep-dirlinks /perl/. / && \
  rsync -aHAX --keep-dirlinks /bash/. / && \
  rsync -aHAX --keep-dirlinks /readline/. / && \
  rsync -aHAX --keep-dirlinks /acl/. / && \
  rsync -aHAX --keep-dirlinks /attr/. / && \
  rsync -aHAX --keep-dirlinks /findutils/. /

COPY --from=sources /sources/downloads/ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 /sources/

RUN tar -xjf /sources/ca-certificates-${CA_CERTIFICATES_VERSION}.tar.bz2 && \
  cd ca-certificates-${CA_CERTIFICATES_VERSION} && \
  make && \
  make DESTDIR=/ca-certificates install

COPY mochiilinux/files/ca-certificates/post_install.sh /post_install.sh
RUN bash /post_install.sh


FROM rsync AS sqlite3

ENV SQLITE3_VERSION=3500400

ENV CFLAGS="-O2 -DSQLITE_ENABLE_FTS3_PARENTHESIS -DSQLITE_ENABLE_COLUMN_METADATA -DSQLITE_SECURE_DELETE -DSQLITE_ENABLE_UNLOCK_NOTIFY -DSQLITE_ENABLE_RTREE -DSQLITE_ENABLE_GEOPOLY -DSQLITE_USE_URI -DSQLITE_ENABLE_DBSTAT_VTAB -DSQLITE_SOUNDEX -DSQLITE_MAX_VARIABLE_NUMBER=250000"

COPY --from=sources /sources/downloads/sqlite-autoconf-${SQLITE3_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/sqlite-autoconf-${SQLITE3_VERSION}.tar.gz && \
  cd sqlite-autoconf-${SQLITE3_VERSION} && \
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
  make DESTDIR=/sqlite3 install


FROM bash AS util-linux

ENV UTIL_LINUX_VERSION=2.41.1

COPY --from=sources /sources/downloads/util-linux-${UTIL_LINUX_VERSION}.tar.xz /sources/

RUN rm -f /bin/sh && ln -sf /bin/bash /bin/sh && \
  tar -xf /sources/util-linux-${UTIL_LINUX_VERSION}.tar.xz

RUN cd util-linux-${UTIL_LINUX_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --prefix=/usr \
    --libdir=/usr/lib \
    --disable-silent-rules \
    --enable-newgrp \
    --disable-uuidd \
    --disable-nls \
    --disable-kill \
    --with-vendordir=/usr/lib && \
  make && \
  make DESTDIR=/util-linux install


FROM rsync AS curl

ENV CURL_VERSION=8.5.0

COPY --from=ca-certificates /ca-certificates /ca-certificates
COPY --from=openssl /openssl /openssl
COPY --from=zstd /zstd /zstd

RUN rsync -aHAX --keep-dirlinks /ca-certificates/. / && \
  rsync -aHAX --keep-dirlinks /openssl/. / && \
  rsync -aHAX --keep-dirlinks /zstd/. /

COPY --from=sources /sources/downloads/curl-${CURL_VERSION}.tar.gz /sources/

RUN tar -xzf /sources/curl-${CURL_VERSION}.tar.gz && \
  cd curl-${CURL_VERSION} && \
  ./configure ${COMMON_ARGS} \
    --disable-dependency-tracking \
    --enable-ipv6 \
    --enable-unix-sockets \
    --enable-static \
    --without-libidn \
    --without-libidn2 \
    --with-ca-bundle=/etc/ssl/certs/ca-certificates.crt \
    --with-ca-path=/etc/ssl/certs \
    --disable-ldap \
    --with-pic \
    --enable-websockets \
    --without-libssh2 \
    --with-ssl \
    --with-nghttp2 && \
  make && \
  make DESTDIR=/curl install


FROM rsync AS python

ENV PYTHON_VERSION=3.12.11

COPY --from=openssl /openssl /openssl
COPY --from=bash /bash /bash
COPY --from=zlib /zlib /zlib
COPY --from=readline /readline /readline

RUN rsync -aHAX --keep-dirlinks /openssl/. / && \
  rsync -aHAX --keep-dirlinks /bash/. / && \
  rsync -aHAX --keep-dirlinks /zlib/. / && \
  rsync -aHAX --keep-dirlinks /readline/. /

COPY --from=sources /sources/downloads/Python-${PYTHON_VERSION}.tar.xz /sources/

RUN rm -f /bin/sh && ln -sf /bin/bash /bin/sh && \
  tar -xf /sources/Python-${PYTHON_VERSION}.tar.xz

RUN cd Python-${PYTHON_VERSION} && \
  ./configure --disable-dependency-tracking --prefix=/usr \
    --enable-ipv6 \
    --enable-loadable-sqlite-extensions \
    --enable-optimizations \
    --enable-shared \
    --with-ensurepip=install \
    --with-lto \
    --with-computed-gotos \
    --with-dbmliborder=gdbm:ndbm && \
  make && \
  make DESTDIR=/python install


FROM bash AS libcap

ENV LIBCAP_VERSION=2.76

COPY --from=sources /sources/downloads/libcap-${LIBCAP_VERSION}.tar.xz /sources/

RUN tar -xf /sources/libcap-${LIBCAP_VERSION}.tar.xz && \
  cd libcap-${LIBCAP_VERSION} && \
  make BUILD_CC=gcc CC=gcc && \
  make DESTDIR=/libcap PAM_LIBDIR=/lib prefix=/usr SBINDIR=/sbin lib=lib GOLANG=no install


FROM stage1 AS gperf

ENV GPERF_VERSION=3.3

RUN wget -q http://ftp.gnu.org/gnu/gperf/gperf-${GPERF_VERSION}.tar.gz && \
  tar -xzf gperf-${GPERF_VERSION}.tar.gz

RUN cd gperf-${GPERF_VERSION} && \
  ./configure ${COMMON_ARGS} --disable-dependency-tracking --prefix=/usr && \
  make && \
  make DESTDIR=/gperf install


FROM rsync AS systemd

ENV SYSTEMD_VERSION=257.6

COPY --from=gperf /gperf /gperf
COPY --from=libcap /libcap /libcap
COPY --from=util-linux /util-linux /util-linux
COPY --from=python /python /python
COPY --from=openssl /openssl /openssl
COPY --from=bash /bash /bash
COPY --from=coreutils /coreutils /coreutils
COPY --from=readline /readline /readline
COPY --from=pkgconfig /pkgconfig /pkgconfig

RUN rsync -aHAX --keep-dirlinks /gperf/. / && \
  rsync -aHAX --keep-dirlinks /libcap/. / && \
  rsync -aHAX --keep-dirlinks /util-linux/. / && \
  rsync -aHAX --keep-dirlinks /python/. / && \
  rsync -aHAX --keep-dirlinks /openssl/. / && \
  rsync -aHAX --keep-dirlinks /bash/. / && \
  rsync -aHAX --keep-dirlinks /coreutils/. / && \
  rsync -aHAX --keep-dirlinks /readline/. / && \
  rsync -aHAX --keep-dirlinks /pkgconfig/. /

COPY --from=sources /sources/downloads/systemd /sources/systemd

ENV CFLAGS="-D __UAPI_DEF_ETHHDR=0 -D _LARGEFILE64_SOURCE"

RUN rm -f /bin/sh && ln -sf /bin/bash /bin/sh && \
  python3 -m pip install meson ninja jinja2

RUN cd /sources/systemd && \
  mkdir build && cd build && \
  meson setup .. \
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
    -D ukify=disabled && \
  ninja && \
  DESTDIR=/systemd ninja install


# =============================================================================
# Stage 2 Merge: Final rootfs
# =============================================================================
FROM stage0 AS stage2-merge

RUN apk add rsync

COPY --from=skeleton /sysroot /skeleton

COPY --from=musl /sysroot /musl
COPY --from=coreutils /coreutils /coreutils
COPY --from=curl /curl /curl
COPY --from=openssl /openssl /openssl
COPY --from=ca-certificates /ca-certificates /ca-certificates
COPY --from=bash /bash /bash
COPY --from=readline /readline /readline
COPY --from=acl /acl /acl
COPY --from=attr /attr /attr
COPY --from=findutils /findutils /findutils
COPY --from=grep /grep /grep
COPY --from=zstd /zstd /zstd
COPY --from=zlib /zlib /zlib
COPY --from=libcap /libcap /libcap
COPY --from=util-linux /util-linux /util-linux
COPY --from=systemd /systemd /systemd

RUN rsync -aHAX --keep-dirlinks /musl/. /skeleton && \
  rsync -aHAX --keep-dirlinks /coreutils/. /skeleton && \
  rsync -aHAX --keep-dirlinks /curl/. /skeleton && \
  rsync -aHAX --keep-dirlinks /openssl/. /skeleton && \
  rsync -aHAX --keep-dirlinks /ca-certificates/. /skeleton && \
  rsync -aHAX --keep-dirlinks /bash/. /skeleton && \
  rsync -aHAX --keep-dirlinks /readline/. /skeleton && \
  rsync -aHAX --keep-dirlinks /acl/. /skeleton && \
  rsync -aHAX --keep-dirlinks /attr/. /skeleton && \
  rsync -aHAX --keep-dirlinks /findutils/. /skeleton && \
  rsync -aHAX --keep-dirlinks /grep/. /skeleton && \
  rsync -aHAX --keep-dirlinks /zstd/. /skeleton && \
  rsync -aHAX --keep-dirlinks /zlib/. /skeleton && \
  rsync -aHAX --keep-dirlinks /libcap/. /skeleton && \
  rsync -aHAX --keep-dirlinks /util-linux/. /skeleton && \
  rsync -aHAX --keep-dirlinks /systemd/. /skeleton

RUN rm -rf /skeleton/usr/include


# =============================================================================
# Stage 2: Final image
# =============================================================================
FROM scratch AS stage2
COPY --from=stage2-merge /skeleton /


# =============================================================================
# Test Stage 2
# =============================================================================
FROM stage2 AS test2

SHELL ["/bin/bash", "-c"]

RUN ls -la / && \
  curl --version && \
  echo "Build successful!"
