package docker

type Image struct {
	Env          []string
	ExposedPorts []string
	Repository   string
	Tag          string
	Args         []string
}
