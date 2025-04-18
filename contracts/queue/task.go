package queue

type Task struct {
	Jobs
	UUID  string `json:"uuid"`
	Chain []Jobs `json:"chain"`
}
