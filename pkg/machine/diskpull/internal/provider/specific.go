package provider

// Get image from somewhere
type Disker interface {
	Get() error
}
