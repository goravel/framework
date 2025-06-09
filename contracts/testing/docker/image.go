package docker

type Image struct {
	Env          []string
	ExposedPorts []string
	Repository   string
	Tag          string
	Args         []string
}

type ImageDriver interface {
	// Build the image.
	Build() error
	// Config gets the image configuration.
	Config() ImageConfig
	// Shutdown the image.
	Shutdown() error
}

type ImageConfig struct {
	ContainerID  string
	ExposedPorts map[int]int
}
