# 🌟 生产级：换用自带完整 glibc 环境的 Debian 瘦身版，完美适配 pt-osc
FROM docker.m.daocloud.io/library/debian:11-slim

ENV LANG="en_US.UTF-8"
ENV TZ=Asia/Shanghai

# 1. 顺便安装 wget，并拉取进程优雅退出守护工具
RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/* \
    && wget -q -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.2/dumb-init_1.2.2_amd64 \
    && chmod +x /usr/local/bin/dumb-init

# 2. 直接把你在本地用 M1 交叉编译出的完美 Linux AMD64 二进制文件拷贝进去
COPY bin/goInception /goInception
COPY config/config.toml.default /etc/config.toml

# 3. 🌟 核心修正：使用 apt-get 安装最正宗、带有标准 glibc 动态链接的 percona-toolkit 和 perl 依赖链
RUN apt-get update && apt-get install -y \
    percona-toolkit \
    perl \
    libdbi-perl \
    libdbd-mysql-perl \
    tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /

EXPOSE 4000

# 4. 标准入口
ENTRYPOINT ["/usr/local/bin/dumb-init", "/goInception","--config=/etc/config.toml"]
