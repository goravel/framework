package queue

type Task struct {
	Jobs
	Uuid  string `json:"uuid"`
	Chain []Jobs `json:"chain"`
}
