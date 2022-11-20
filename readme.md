# Centos 监控面板

### 编译运行
```shell
# 启用 go mod
go env -w GO111MODULE=on
#使用七牛云代理
go env -w GOPROXY=https://goproxy.cn,direct

# go mod init dashboard
go mod tidy

# 编译
go build -o dashboard main.go

# 运行
nohup ./dashboard &
```

### 常见问题

- 流量无法统计
```shell
# ifconfig 命令不存在，安装 net-tools
yum -y install net-tools
```

- 内存信息无法识别
```shell
# 安装 dmidecode
yum install dmidecode
```