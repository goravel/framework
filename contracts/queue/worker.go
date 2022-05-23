package queue

type Worker interface {
	Run() error
}
