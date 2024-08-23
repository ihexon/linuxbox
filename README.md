> It is like the game with toy blocks


# VM Provider（但全都没实现）
- support MacOS via vfkit (aarch64)
- Support Windows via WSL (x86_64)
- Support Linux via Qemu-KVM (aarch64, x86_64)


# 全是问题
- machine init 全是 BUG
- machine reset 会吧 / 一起删了....
- vmcofigs 和 define 有重叠关系，并且关系混乱
- macos 的 provider 还没实现
- windows 的 provider 还没实现


# 主线任务
- 在 macos 上跑通 machine init
  - 实现 macos 的 provider
    - 实现 `applehv.AppleHVStubber`
    - [X] 实现 vfkit 的 stop 方法
    - [ ] 实现 appleHV 的 getDisk 方法

## 淦 ！