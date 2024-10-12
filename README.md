# OVM IHEXON BRANCH

# 业务使用

## 初始化虚拟机

```
ovm-arm64 \
    --workspace /Users/danhexon/myvm \
    machine init \
    --bootable-image bootable.img.xz \              # support xz / zstd format, bootable.img.xz support @version field
    --external-disk \                               # map a raw disk into virtualMachine's /dev/vdX
    --report-url [remote events listener] \
    --volume /tmp:/mnt/HostDir/tmp                  # Mount macos /tmp/ into virtualMachine's /mnt/HostDir/tmp
```

- workspace 指定数据存储的地方，所有的文件将会被存储在这里，这个参数作为 root 参数对所有的子命令都可见
- machine init 定义了行为，该阶段的行为是初始化虚拟机
- bootable-image 是 machine init 的参数，指定了虚拟机的镜像，该镜像是一个可启动的参数
- report-url 将程序关键的 event 发送给一个 url，这个 url 可以说 unix socks, 也可以是 `tcp://[ip]:[port]`

## 启动虚拟机
```
ovm-arm64 --workspace /Users/danhexon/myvm \
    machine start \
    --ppid [PPID]
```
- ppid 指定一个 PPID，等待这个PPID 消失，虚拟机也会关闭，如果你不指定，**如果不指定 twinpid ，那么 twinpid 是当前进程的 PPID**


## REST API
默认在 `$workspace/tmp/ovm_restapi.socks`

- /apiversion      获取虚拟机VERSION 
- /{name}/info     获取虚拟机配置信息
- /{name}/vmstat   获取虚拟机运行状态
- /{name}/synctime 同步主机时间到虚拟机