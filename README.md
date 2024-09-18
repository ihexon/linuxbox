> It is like the game with toy blocks

# VM Provider（但全都没实现）
- ~~support MacOS via vfkit (aarch64)~~(没必要，作用和 krunkit 有重叠)
- ~~Support Linux via Qemu-KVM (aarch64, x86_64)~~(前提是我没被开除)
- Support Windows via WSL (x86_64) (没完成，写了一半，很多地方没跑通)
- support MacOS via krunkit (aarch64)

# 已经实现的子命令
- [X] `machine init [vm_name]` 初始化虚拟机
- [X] `machine start [vm_name]` 启动虚拟机
- [X] `machine stop  [vm_name]` 停止虚拟机
- [ ] `machine rm    [vm_name]` 删除虚拟机
- [ ] `machine reset`  重置所有系统
- [ ] `machine set`    修改虚拟机配置文件
- [ ] `支持点火器 :)`

#  oomol studio 相关的参数
- twinpid [PID]
监视 PID ，如果 PID 找不到，则退出 ovm

- evtsock [SOCK_PATH]
发送虚拟机事件给 SOCK_PATH，这个 SOCK_PATH 应该是一个 UDF

```bash
machine init --twinpid [PID] --evtsock [SOCK_PATH]
```

## TODO LIST

 - ~~2024-08-25 VFKIT EFI BOOT 模式方式并不能启动常规的 Linux 发行版，并且连自家的 Fedora Core 也启动不来，原因未知~~
 - ~~Sat Aug 31 15:39:16 HKT 2024 EFI Boot 的坑基本上踩完了，我只能说：不过如此：）~~
 - ~~Mon Sep  9 16:50:17 HKT 2024 现在 machine start 可以启动一个基于 alpine rootfs 的虚拟机~~
 - ~~machine start 似乎不会退出主进程，我的期望是退出主进程~~
 - machine stop 需要更多的测试
 - machine init 需要更多的测试
 - ~~gvproxy 生成的 socks file 会与 podman 本身的 socks files 冲突。~~
 - machine stop 无法正确停止 gvproxy 和 krunkit
 - wsl2v1 完整抄袭 podman，先跑起来再说...
 - wsl2v2 重新实现