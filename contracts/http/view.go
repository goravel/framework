package http

//go:generate mockery --name=View
type View interface {
	Exists(view string) bool
	Share(key string, value any)
	Shared(key string, def ...any) any
	GetShared() map[string]any
}
