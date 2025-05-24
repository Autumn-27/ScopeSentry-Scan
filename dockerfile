FROM python:3.7-slim

WORKDIR /apps
COPY msyh.ttc /usr/share/fonts/

# 更新软件包列表并安装必要的软件包
RUN apt-get update && apt-get install -y \
    libexif-dev \
    udev \
    chromium \
    vim \
    tzdata \
    libpcap-dev \
    default-jdk \
    tini \
    && rm -rf /var/lib/apt/lists/*

# 设置 tini 作为 init 进程（PID 1）
ENTRYPOINT ["/usr/bin/tini", "--"]


RUN pip install uro
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

# 运行golang程序的命令
CMD ["/apps/ScopeSentry"]