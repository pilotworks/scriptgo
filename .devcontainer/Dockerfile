FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04

# Install native compiler toolchain, LLVM, Clang, build tools, and debugging utilities
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
        build-essential \
        clang \
        llvm \
        libclang-rt-dev \
        lld \
        lldb \
        make \
        valgrind \
        pkg-config \
        libbrotli-dev \
        libzstd-dev \
        libssl-dev \
        ca-certificates \
        curl \
        wget \
        xz-utils \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Zig compiler for zero-dependency cross-compilation support
ARG ZIG_VERSION="0.13.0"
RUN ARCH="$(uname -m)" \
    && if [ "$ARCH" = "x86_64" ]; then ZIG_ARCH="x86_64"; \
       elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then ZIG_ARCH="aarch64"; \
       else echo "Unsupported architecture: $ARCH" && exit 1; fi \
    && curl -fsSL "https://ziglang.org/download/${ZIG_VERSION}/zig-linux-${ZIG_ARCH}-${ZIG_VERSION}.tar.xz" -o /tmp/zig.tar.xz \
    && tar -xJf /tmp/zig.tar.xz -C /opt \
    && ln -s "/opt/zig-linux-${ZIG_ARCH}-${ZIG_VERSION}/zig" /usr/local/bin/zig \
    && rm /tmp/zig.tar.xz
