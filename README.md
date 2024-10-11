# OVM IHEXON BRANCH

# 业务使用

## 初始化虚拟机
```
$ ovm-arm64 \
    --workdir /Users/danhexon/myvm \
    machine init \
    --report-url xxx.socks \
    --image /Users/danhexon/alpine_virt/alpine_krunkit.raw.xz \
    --image-version "1.0"
```

初始化后，再次初始化的行为根据传入的 --image-version 判断，如果 --image-version 字段和之前不一样，就触发初始化，
如果一样就跳过初始化。

- --workspace 是 ovm-arm64 命令的参数，指定数据存储的地方，所有的文件将会被存储在这里
- machine init 定义了行为，该阶段的行为是初始化虚拟机
- --image 是 machine init 的参数，指定了虚拟机的镜像，该镜像是一个可启动的参数
- --image-version 是 machine init 的参数，指定了虚拟机的镜像版本
- --report-url 将程序关键的 event 发送给一个 url，这个 url 可以说 unix socks, 也可以是 `tcp://[ip]:[port]`

## 启动虚拟机
```
ovm-arm64 --workspace /Users/danhexon/myvm \
    machine start \
    --ppid [PPID]
    --external-disk /Users/danhexon/alpine_virt/mydisk.raw \
    --volume /tmp/:/tmp/macos/tmp \
    --volume /tmp1/:/tmp/macos/tmp1 \
```
- --external-disk 指定一个额外的 raw 虚拟磁盘映射到 /dev/vdX 下
- --volume 指定一个目录映射到虚拟机的目录下
- --twinpid 指定一个 PPID，等待这个PPID 消失，虚拟机也会关闭，如果你不指定，**如果不指定 twinpid ，那么 twinpid 是当前进程的 PPID**


## REST API
```
ovm-arm64 --workdir /Users/danhexon/fucknewhome system service tcp://HOST:PORT
```

默认是 tcp://127.0.0.1:65176
```go
r.Handle(("/apiversion"), s.APIHandler(backend.VersionHandler)).Methods(http.MethodGet)
	r.Handle(("/{name}/diskuage"), s.APIHandler(backend.GetDiskUsage)).Methods(http.MethodGet)
	r.Handle(("/{name}/info"), s.APIHandler(backend.GetInfos)).Methods(http.MethodGet)
	r.Handle(("/{name}/vmstat"), s.APIHandler(backend.GetVMStat)).Methods(http.MethodGet)
	r.Handle(("/{name}/synctime"), s.APIHandler(backend.TimeSync)).Methods(http.MethodGet)
```


- diskuage: 获取磁盘使用量
- info:     获取虚拟机上次运行时的信息
- vmstat:   获取 vm 运行状态[stopped,running]
- synctime: host 时间同步到 guest 的时间
文档以后写....

## 快捷启动脚本

```
./ovm.sh \
    --workdir=/Users/danhexon/myvm/ \
    --image=/Users/danhexon/alpine_virt/bugbox-machine-default-arm64.raw.xz  \
    --image-version="1.0" \
    --volume=/tmp/:/tmp/macos/tmp \
    --external-disk=/Users/danhexon/alpine_virt/mydisk.raw  \
    --twinpid=1000
```