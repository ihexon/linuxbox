> It is like the game with toy blocks

# VM Provider（但全都没实现）
- ~~support MacOS via vfkit (aarch64)~~
- ~~Support Linux via Qemu-KVM (aarch64, x86_64)~~
- Support Windows via WSL (x86_64)



# 
- machine reset 会吧 / 一起删了....
- vmcofigs 和 define 有重叠关系，并且关系混乱
- macos 的 provider 还没实现
- windows 的 provider 还没实现

# 主线任务
- [X] 在 macos 上跑通 machine init
  - 实现 macos 的 provider
    - 实现 `applehv.AppleHVStubber`
    - [ ] 实现 libkrun 的 stop 方法 
    - [ ] 实现 libkrun 的 create 方法
    - [X] 实现 libkrun 的 start 方法

## 淦 ！

# 2024-08-25 VFKIT EFI BOOT 模式方式并不能启动常规的 Linux 发行版，并且连自家的 Fedora Core 也启动不来，原因未知
# Sat Aug 31 15:39:16 HKT 2024 EFI Boot 的坑基本上踩完了，我只能说：不过如此：）