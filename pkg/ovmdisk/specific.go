package ovmdisk

type Disker interface {
	Get() error
}
