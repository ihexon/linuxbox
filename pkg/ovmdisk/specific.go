package ovmdisk

// Disker WTF is rootfs and bootable image
// Get it ok ?
type Disker interface {
	Get() error
}
