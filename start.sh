#!/bin/bash

# 进入工作目录
cd /apps

# 判断 UPDATE 文件是否存在
if [ -f "/apps/UPDATE" ]; then
    echo "UPDATE 文件存在，开始处理更新..."

    # 备份现有的 ScopeSentry 可执行文件
    if [ -f "/apps/ScopeSentry" ]; then
        echo "备份现有的 ScopeSentry 文件..."
        mv /apps/ScopeSentry /apps/ScopeSentry.bak || { echo "备份失败"; exit 1; }
    fi

    # 从 UPDATE 文件中读取 ZIP 文件的 URL
    zip_url=$(cat /apps/UPDATE)

    if [ -z "$zip_url" ]; then
        echo "ERROR: UPDATE 文件中没有 URL 地址"
        exit 1
    fi

    # 下载 ZIP 文件到 /tmp/main.zip
    echo "从 $zip_url 下载 ZIP 文件..."
    curl -o /tmp/main.zip "$zip_url" || { echo "ERROR: ZIP 文件下载失败"; mv /apps/ScopeSentry.bak /apps/ScopeSentry || exit 1; exit 1; }

    # 解压 ZIP 文件到 /apps 目录
    echo "解压 ZIP 文件到 /apps..."
    unzip -o /tmp/main.zip -d /apps || { echo "ERROR: 解压失败"; mv /apps/ScopeSentry.bak /apps/ScopeSentry || exit 1; exit 1; }

    # 清理临时文件
    rm -f /tmp/main.zip

    # 删除 UPDATE 文件
    rm -f /apps/UPDATE
else
    echo "UPDATE 文件不存在，跳过更新，直接启动应用..."
fi

# 确保新的 ScopeSentry 文件存在并有执行权限
if [ -f "/apps/ScopeSentry" ]; then
    chmod +x /apps/ScopeSentry
    # 运行新的 ScopeSentry 可执行文件
    echo "运行新的 ScopeSentry..."
    /apps/ScopeSentry
else
    echo "ERROR: ScopeSentry 文件不存在，恢复备份并启动..."
    if [ -f "/apps/ScopeSentry.bak" ]; then
        mv /apps/ScopeSentry.bak /apps/ScopeSentry
        chmod +x /apps/ScopeSentry
        /apps/ScopeSentry
    else
        echo "ERROR: 没有备份文件，无法启动应用"
        exit 1
    fi
fi
