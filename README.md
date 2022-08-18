# 通信调试助手

## 编译方式
### 1、安装Golang环境
### 2、克隆项目
### 3、执行go get安装依赖
### 4、开始编译

```shell
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o tcpproxy.x64.exe
./upx.exe -9 tcpproxy.x64.exe
```

## 下载方式
通过Release下载 https://gitee.com/tansuyun/tcp-port-forwarding/releases


## TODO功能清单
[√]支持多端口监听  
[√]支持自动切换后端  
[√]支持禁用日志提高性能  
[√]支持hex模式和string模式记录日志  
[√]支持监听失败后间隔5秒再次启动  
[√]支持设定配置文件而不是使用默认配置文件  
[√]使用upx压缩exe文件  
[ ]支持转发内容到AMQP队列  
[ ]支持在配置文件更新后自动重载  
[ ]支持ETCD远程配置  
[ ]支持性能分析  
[ ]支持UI控制  

## 第一步、编写配置文件，支持JSON格式和YAML格式或TOML格式，自选

### proxy.yaml YAML格式配置
```yaml
---
Peers:
# 监听名称，用于输出到日志中
- Name: Debug
# 监听地址，完整地址，不能省略本地地址
  Listen: 0.0.0.0:8986
  # 日志格式 支持 string 或留空表示为 hex
  Log: string
  # 转发的目标地址，支持多个，当一个无法连接是自动连接第二个，如果Targets中一个都没法用，则使用Duplex配置
  Targets:
  - www.oschina.net:80
  - www.oschina.net:443
  # 复制数据流的地址，用于物联网调试或记录相关通信记录
  Duplex: www.oschina.net:5010
- Name: Redis
  Listen: 0.0.0.0:6378
  # 禁用日志，不再输出内容到日志文件中
  Log: "false"
  Targets:
  - www.oschina.net:6379
  - 127.0.0.1:6379
  - 10.0.1.11:6379
```
### proxy.json JSON格式配置
```json
{
    "Peers": [
        {
            "Name": "Debug",
            "Listen": "0.0.0.0:8986",
            "Log": "string",
            "Targets": [
                "www.oschina.net:80",
                "www.oschina.net:443"
            ],
            "Duplex": "www.oschina.net:5010"
        },
        {
            "Name": "Redis",
            "Listen": "0.0.0.0:6378",
            "Targets": [
                "www.oschina.net:6379",
                "127.0.0.1:6379",
                "10.0.1.11:6379"
            ]
        }
    ]
}
```

## 启动程序
```shell
# 或使用默认配置文件 proxy.json/yaml/toml
tcpproxy.x64.exe -c 配置文件路径
```

## 使用效果
### 日志文件内容-文本模式
```
09:48:51 >	[Debug]	127.0.0.1:60230	GET / HTTP/1.1
Host: 127.0.01:8986
User-Agent: curl/7.54.0
Accept: */*


09:48:51 <	[Debug]	127.0.0.1:60230	HTTP/1.1 404 Not Found
Server: nginx
Date: Thu, 18 Aug 2022 01:48:51 GMT
Content-Type: text/html
Content-Length: 146
Connection: keep-alive

<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx</center>
</body>
</html>

```

### 日志文件内容-Hex文本模式
```
09:49:51 >	[Redis]	127.0.0.1:60497	2a310d0a24370d0a434f4d4d414e440d0a
09:49:51 <	[Redis]	127.0.0.1:60497	2a3230300d0a2a360d0a2431320d0a626772657772697465616f660d0a3a310d0a2a320d0a2b61646d696e0d0a2b6e6f7363726970740d0a3a300d0a3a300d0a3a300d0a2a360d0a24340d0a786c656e0d0a3a320d0a2a320d0a2b726561646f6e6c790d0a2b666173740d0a3a310d0a3a310d0a3a310d0a2a360d0a24380d0a627a706f706d61780d0a3a2d330d0a2a330d0a2b77726974650d0a2b6e6f7363726970740d0a2b666173740d0a3a310d0a3a2d320d0a3a310d0a2a360d0a24380d0a6269746669656c640d0a3a2d320d0a2a320d0a2b77726974650d0a2b64656e796f6f6d0d0a3a310d0a3a310d0a3a310d0a2a360d0a2431330d0a7a72616e6765627973636f72650d0a3a2d340d0a2a310d0a2b726561646f6e6c790d0a3a310d0a3a310d0a3a310d0a2a360d0a24380d0a627a706f706d696e0d0a3a2d330d0a2a330d0a2b77726974650d0a2b6e6f7363726970740d0a2b666173740d0a3a310d0a3a2d320d0a3a310d0a2a360d0a2431340d0a7a72656d72616e676562796c65780d0a3a340d0a
```