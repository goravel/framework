package schema

type Migration interface {
	Signature() string
	Up() error
	Down() error
}
