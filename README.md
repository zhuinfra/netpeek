# NetPeek

使用Go语言开发的网络监测命令行工具。

## 功能特性

- **IP查询**: 从多个国内外源获取并显示您的公网IP地址
- **WebSocket测试**: 支持ws://和wss://协议的WebSocket连接测试
- **HTTP网站检测**: 测试网站访问情况，支持设置测试轮数，显示详细性能指标
- **DNS劫持检测**: 通过本地DNS和多个公共DNS服务器对比检测潜在的DNS劫持

## 安装

### 方法一：下载预编译二进制文件（推荐）

使用以下命令自动下载最新版本并安装：

```bash
#!/bin/bash
LATEST_VERSION=$(curl -s https://api.github.com/repos/zhuinfra/netpeek/releases/latest | grep -oP '"tag_name": "\K[^"]+')
wget -O /tmp/netpeek https://github.com/zhuinfra/netpeek/releases/download/${LATEST_VERSION}/netpeek-linux-amd64
chmod +x /tmp/netpeek
sudo mv /tmp/netpeek /usr/local/bin/
netpeek --version
```

将上述内容保存为install.sh并执行，或直接复制粘贴到终端中运行。

### 方法二：从源码构建

```bash
# 克隆仓库
git clone https://github.com/zhuinfra/netpeek.git
cd netpeek

# 构建项目
go build

# 安装到系统PATH目录（可选）
sudo mv netpeek /usr/local/bin/

# 运行应用
netpeek --help
```

## 使用方法

### IP查询
```bash
./netpeek ip
```

该命令会从多个源获取并显示您的公网IP地址：
- 国内源：B站API和腾讯API
- 国外源：Cloudflare API和IPinfo.io API

### WebSocket测试
```bash
./netpeek ws <websocket-url>

# 示例
./netpeek ws wss://echo.websocket.org
```

### HTTP网站检测
```bash
./netpeek http [-n/--num-rounds <轮数>] <网址>

# 示例 (默认3轮)
./netpeek http https://www.baidu.com

# 示例 (自定义5轮)
./netpeek http -n 5 https://httpbin.org/get
```

该命令会测试网站的访问情况，并显示以下性能指标：
- 状态码
- 响应大小
- DNS查询时间
- TCP连接时间
- TLS握手时间
- 首字节时间(TTFB)
- 下载时间
- 总时间

同时会提供多轮测试的统计信息（最快、最慢、平均时间）和评分。

### DNS劫持检测
```bash
./netpeek dns [-t <记录类型>] <域名>

# 示例 (默认IPv4 A记录)
./netpeek dns www.example.com

# 示例 (IPv6 AAAA记录)
./netpeek dns -t AAAA www.example.com
```

该命令会通过以下DNS服务器进行查询对比：
- 本地DNS服务器
- 阿里云DNS (223.5.5.5)
- Google DNS (8.8.8.8)

检测结果包括：
- CNAME解析链
- 各DNS服务器返回的IP地址
- 响应时间
- 劫持风险分析

## 许可证

MIT