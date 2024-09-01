> It is like the game with toy blocks

# VM Provider（但全都没实现）
- ~~support MacOS via vfkit (aarch64)~~(没必要，作用和 krunkit 有重叠)
- ~~Support Linux via Qemu-KVM (aarch64, x86_64)~~(前提是我没被开除)
- support MacOS via krunkit (aarch64)
- Support Windows via WSL (x86_64)(没完成，写了一半，很多地方没跑通)

# 主线任务
- [x] 解压 image.xz 到 `/Users/danhexon/.local/share/oomol/ovm/machine/libkrun`
- [X] 试图正确配置及运行 gvproxy，`.config/oomol/ovm/machine/libkrun/bugbox-machine-default.json`
- [X] 写入 gvproxy fordward 配置到 `~/.config/oomol/ovm/machine/bugbox-connections.json`
- [X] 生成 ssh keys 到 `/Users/danhexon/.local/share/oomol/ovm/machine`

## 淦 ！

 - ~~2024-08-25 VFKIT EFI BOOT 模式方式并不能启动常规的 Linux 发行版，并且连自家的 Fedora Core 也启动不来，原因未知~~
 - Sat Aug 31 15:39:16 HKT 2024 EFI Boot 的坑基本上踩完了，我只能说：不过如此：）