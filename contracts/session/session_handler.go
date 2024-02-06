package session

type Handler interface {
	Close() bool
	Destroy(id string) bool
	Gc(maxLifetime int) (int, bool)
	Open(path string, name string) bool
	Read(id string) string
	Write(id string, data string) error
}
