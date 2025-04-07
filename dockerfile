FROM python:3.7-slim

WORKDIR /apps
COPY msyh.ttc /usr/share/fonts/

# 更换国内源
RUN rm -rf /etc/apt/sources.list.d/debian.sources && printf "%s\n" \
"deb https://mirrors.aliyun.com/debian/ bookworm main contrib non-free non-free-firmware" \
"deb-src https://mirrors.aliyun.com/debian/ bookworm main contrib non-free non-free-firmware" \
"deb https://mirrors.aliyun.com/debian/ bookworm-updates main contrib non-free non-free-firmware" \
"deb-src https://mirrors.aliyun.com/debian/ bookworm-updates main contrib non-free non-free-firmware" \
"deb https://mirrors.aliyun.com/debian/ bookworm-backports main contrib non-free non-free-firmware" \
"deb-src https://mirrors.aliyun.com/debian/ bookworm-backports main contrib non-free non-free-firmware" \
"deb https://mirrors.aliyun.com/debian-security/ bookworm-security main contrib non-free non-free-firmware" \
"deb-src https://mirrors.aliyun.com/debian-security/ bookworm-security main contrib non-free non-free-firmware" \
> /etc/apt/sources.list

# 更新软件包列表并安装必要的软件包
RUN apt-get update && apt-get install -y \
    libexif-dev \
    udev \
    chromium \
    vim \
    tzdata \
    libpcap-dev \
    && rm -rf /var/lib/apt/lists/*
# 拷贝当前目录下的可执行文件到容器中
COPY dist/ScopeSentry-Scan_linux_amd64_v1/ScopeSentry /apps/ScopeSentry
RUN chmod +x /apps/ScopeSentry
RUN mkdir /apps/ext
RUN mkdir /apps/ext/rad
RUN mkdir /apps/ext/ksubdomain
RUN mkdir /apps/ext/rustscan
RUN mkdir /apps/ext/katana

COPY tools/linux/ksubdomain /apps/ext/ksubdomain/ksubdomain
RUN chmod +x /apps/ext/ksubdomain/ksubdomain
COPY tools/linux/rad /apps/ext/rad/rad
RUN chmod +x /apps/ext/rad/rad
COPY tools/linux/rustscan /apps/ext/rustscan/rustscan
RUN chmod +x /apps/ext/rustscan/rustscan

COPY tools/linux/katana /apps/ext/katana/katana
RUN chmod +x /apps/ext/katana/katana
# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

# 设置编码
ENV LANG C.UTF-8

# 使用tini解决容器中截图及调用任何插件可能导致的僵尸进程问题，原理可参考：https://cn.linux-console.net/?p=20630
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

# 运行golang程序的命令
ENTRYPOINT ["/tini", "--", "/apps/ScopeSentry"]