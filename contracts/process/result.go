package process

type Result interface {
	Successful() bool
	Failed() bool
	ExitCode() int
	Output() string
	ErrorOutput() string
	Command() string
	SeeInOutput(needle string) bool
	SeeInErrorOutput(needle string) bool
}
