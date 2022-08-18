# 通信调试助手
编写配置文件


## proxy.json JSON格式配置
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

## proxy.yaml YAML格式配置
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
  Targets:
  - www.oschina.net:6379
  - 127.0.0.1:6379
  - 10.0.1.11:6379
```