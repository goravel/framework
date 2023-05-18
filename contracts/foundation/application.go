package foundation

//go:generate mockery --name=Application
type Application interface {
	Container
	Boot()
}
