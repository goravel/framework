package recovery

type Handler interface {
	ShouldReport(v interface{}) bool
	Report(v interface{})
}
