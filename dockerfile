FROM python:3.7-slim

WORKDIR /apps

# 更新软件包列表并安装必要的软件包
RUN apt-get update && apt-get install -y \
    libexif-dev \
    udev \
    chromium \
    vim \
    tzdata \
    libpcap-dev \
    fonts-noto-cjk \
    fonts-wqy-microhei \
    fonts-arphic-ukai \
    fonts-arphic-uming \
    curl \
    unzip \
    locales \
    && rm -rf /var/lib/apt/lists/*

# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo 'Asia/Shanghai' >/etc/timezone

# 配置语言环境为中文
RUN locale-gen zh_CN.UTF-8
ENV LANG zh_CN.UTF-8
ENV LC_ALL zh_CN.UTF-8

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

# 设置编码
ENV LANG C.UTF-8

# 运行golang程序的命令
# ENTRYPOINT ["/apps/ScopeSentry"]

# 复制 start.sh 脚本到容器
COPY start.sh /usr/local/bin/start.sh

# 给 start.sh 脚本赋予执行权限
RUN chmod +x /usr/local/bin/start.sh

# 设置容器启动命令
ENTRYPOINT ["/usr/local/bin/start.sh"]